package main

import (
	"cofin/core"
	"cofin/internal/real_stonks"
	"cofin/internal/retrieval"
	"cofin/internal/sec_api"
	"cofin/models"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const CHUNK_SIZE = 3000
const CHUNK_OVERLAP = 100
const MAX_FILINGS_PER_COMPANY_PER_BATCH = 20

var SEC_API_KEY = ""

func main() {
	godotenv.Load()

	SEC_API_KEY = os.Getenv("SEC_API_KEY")

	// connect to the database
	db, err := core.InitDB()
	if err != nil {
		panic(err)
	}

	// auto migrate the database
	err = db.Debug().AutoMigrate(
		&models.User{},
		&models.Company{},
		&models.Document{},
		&models.AccessToken{},
		&models.Message{},
	)
	if err != nil {
		panic(err)
	}

	fetcher, err := newDocumentFetcher(db)
	if err != nil {
		panic(err)
	}

	fetcher.run()
}

type documentFetcher struct {
	db       *gorm.DB
	embedder embeddings.Embedder
	splitter *retrieval.Splitter
	logger   *zap.SugaredLogger
}

func newDocumentFetcher(db *gorm.DB) (*documentFetcher, error) {
	embedder, err := retrieval.NewEmbedder()
	if err != nil {
		return nil, err
	}

	splitter, err := retrieval.NewSplitter(CHUNK_SIZE, CHUNK_OVERLAP)
	if err != nil {
		panic(err)
	}

	logger, err := core.NewLogger()
	if err != nil {
		panic(err)
	}

	return &documentFetcher{
		db:       db,
		embedder: embedder,
		splitter: splitter,
		logger:   logger,
	}, nil
}

func (f *documentFetcher) run() {
	logger := f.logger
	embedder := f.embedder
	splitter := f.splitter
	db := f.db

	fetchDocuments(db, logger, embedder, splitter)
}

func fetchDocuments(db *gorm.DB, logger *zap.SugaredLogger, embedder embeddings.Embedder, splitter *retrieval.Splitter) {
	logger.Info("Running fetching job...")

	var allListings []sec_api.Listing
	// Go over all stock exchanges.
	for _, exchange := range sec_api.StockExchanges {
		// Get listings for the exchange.
		listings, err := sec_api.GetTradedCompanies(SEC_API_KEY, exchange)
		if err != nil {
			logger.Errorw(fmt.Errorf("failed to get companies traded on an exchange: %v", err).Error(), "exchange", exchange)
			continue
		}

		for _, listing := range listings {
			// Skip delisted companies.
			if listing.IsDelisted {
				logger.Infof("Skipping delisted company: %v", listing.Ticker)
				continue
			}

			if strings.ToUpper(listing.Exchange) != strings.ToUpper(string(exchange)) {
				logger.Infof("Skipping company on the wrong exchange: $%v (%v)", listing.Ticker, listing.Exchange)
				continue
			}

			allListings = append(allListings, listing)
		}
	}

	// Allow limiting the number of companies to process. Useful in staging.
	if maxCompanies := os.Getenv("MAX_COMPANIES"); maxCompanies != "" {
		limit, err := strconv.Atoi(maxCompanies)
		if err != nil {
			logger.Errorf("failed to parse MAX_COMPANIES: %w", err)
			return
		}

		if limit < len(allListings) {
			allListings = allListings[:limit]
		}
	}

	for _, listing := range allListings {
		logger.Infof("Processing company: %v", listing.Ticker)

		// Create the company if it doesn't exist, fetchDocuments documents, and
		// store them.
		err := processListing(db, logger, listing, embedder, splitter)
		if err != nil {
			logger.Errorw(fmt.Errorf("failed to process a listing: %v", err).Error(), "ticker", listing.Ticker)
			continue
		}
	}
}

// processListing creates a company if it doesn't exist, fetches documents, and
// stores them.
func processListing(db *gorm.DB, logger *zap.SugaredLogger, listing sec_api.Listing, embedder embeddings.Embedder, splitter *retrieval.Splitter) error {
	// Create or get a company in a transaction.
	var company *models.Company
	err := db.Transaction(func(tx *gorm.DB) (err error) {
		company, err = models.GetCompanyByTicker(tx, listing.Ticker)
		if err != nil {
			company = nil
			return err
		}

		if company != nil {
			return nil
		}

		logger.Infof("Creating company: %v", listing.Ticker)
		company, err = models.CreateCompany(tx, listing.Name, listing.Ticker, listing.CIK, time.Time{})
		return err
	})
	if err != nil {
		return fmt.Errorf("could not create company for %v (%v): %w\n", listing.Name, listing.Ticker, err)
	}

	// Initialize the vector store.
	store, err := retrieval.NewPinecone(context.Background(), embedder, company.ID)
	if err != nil {
		panic(err)
	}

	// If the company documents were fetched in the past 72 hours, don't fetch
	// documents for the company again.
	if !company.LastFetchedAt.IsZero() && company.LastFetchedAt.Add(72*time.Hour).After(time.Now()) {
		logger.Infow(fmt.Sprintf("Skipping company %v because it has been fetched in the past 72 hours", listing.Ticker), "companyID", company.ID)
		return nil
	}

	for _, filingKind := range []models.SourceKind{models.K10, models.Q10} {
		logger.Infof("Processing filing kind: %v", filingKind)
		if err := processFilingKind(db, logger, company, splitter, store, filingKind); err != nil {
			logger.Errorw(fmt.Errorf("failed to process a filing kind for a company: %v", err).Error(), "companyID", company.ID, "filingKind", filingKind)
			continue
		}
	}

	realStonks := real_stonks.RealStonks{}
	marketInformation, err := realStonks.GetMarketData(company.Ticker)
	if err == nil {
		company.Currency = marketInformation.Currency
		company.Price = marketInformation.Price
		company.Change = marketInformation.Change
		company.TotalVolume = marketInformation.TotalVolume

		tx := db.Save(company)
		if tx.Error != nil {
			logger.Infof("Unable to update market data for %v: %v", listing.Ticker, tx.Error)
		}

	} else {
		logger.Infof("Unable to fetchDocuments market data for %v", listing.Ticker)
	}

	return nil
}

// processFilingKind fetches filings of a particular kind for a company,
// processes and stores them.
func processFilingKind(db *gorm.DB, logger *zap.SugaredLogger, company *models.Company, splitter *retrieval.Splitter, store vectorstores.VectorStore, filingKind models.SourceKind) error {
	// Get the most recent document of the kind for the company.
	document, err := models.GetCompanyDocumentsOfKindInverseChronological(db, company.ID, filingKind)
	if err != nil {
		return fmt.Errorf("failed to fetchDocuments most recent document for %v (%v): %w\n", company.Name, company.Ticker, err)
	}

	// Query for the last one year of documents if we have no documents for the
	// company. Otherwise query for documents since the last document.
	var lastFiledAt = time.Now().Add(-365 * 1 * 24 * time.Hour)
	if document != nil {
		lastFiledAt = document.FiledAt.Add(1 * time.Second)
	} else {
		logger.Infow(fmt.Sprintf("No documents found for %v (%v) of kind %v, fetching all documents since %v", company.Name, company.Ticker, filingKind, lastFiledAt), "companyID", company.ID, "filingKind", filingKind)
	}

	// Get filings since the last filed time.
	filings, err := sec_api.GetFilingsSince(SEC_API_KEY, company.CIK, filingKind, lastFiledAt, MAX_FILINGS_PER_COMPANY_PER_BATCH)
	if err != nil {
		return fmt.Errorf("failed to fetchDocuments filings for %v (%v): %w\n", company.Name, company.Ticker, err)
	}

	// Process filings. We only process up to MAX_FILINGS_PER_COMPANY_PER_BATCH
	// at a time. This guarantees that no company hogs the fetching pipeline for
	// too long. If not all documents are fetched, next time the company is due
	// for re-fetching we will continue where we left off by checking most
	// recent document's filing time.
	if len(filings) > MAX_FILINGS_PER_COMPANY_PER_BATCH {
		logger.Infof("Company %v has %v filings, processing only %v", company.Ticker, len(filings), MAX_FILINGS_PER_COMPANY_PER_BATCH)
		filings = filings[:MAX_FILINGS_PER_COMPANY_PER_BATCH]
	}

	for _, filing := range filings {
		// Process the filing in a transaction. Processing a filing is atomic
		// and involves three things: storing the file in the DB, storing the
		// chunks in vector store, and updating the company. If any of these
		// suboperations fail, we revert and abort.
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := processFiling(tx, logger, company, splitter, store, filingKind, filing); err != nil {
				return fmt.Errorf("failed to process a filing with accession number %v: %v", filing.AccessionNo, err.Error())
			}

			// Update the company's last fetched time after successfully
			// processing a filing for it.
			company.LastFetchedAt = time.Now()
			logger.Infof("Updating company %v (%v) last fetched time to %v", company.Name, company.Ticker, company.LastFetchedAt)
			err = tx.Save(&company).Error
			if err != nil {
				return fmt.Errorf("failed to update company for %v (%v): %w\n", company.Name, company.Ticker, err)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("failed to process a filing for a company: %v", err.Error())
		}
	}

	return nil
}

// processFiling processes a filing and stores it.
func processFiling(db *gorm.DB, logger *zap.SugaredLogger, company *models.Company, splitter *retrieval.Splitter, store vectorstores.VectorStore, filingKind models.SourceKind, filing sec_api.Filing) error {
	var sections []models.Section
	if filingKind == models.Q10 {
		sections = models.Q10Sections
	} else if filingKind == models.K10 {
		sections = models.K10Sections
	}

	var rawContent string
	originURL := sec_api.GetFilingOriginURL(filing)
	for _, section := range sections {
		// Get the filing file from the SEC.
		sectionContent, err := sec_api.ExtractSectionContent(SEC_API_KEY, originURL, section)
		if err != nil {
			return fmt.Errorf("failed to fetchDocuments filing file (accession number %v) for %v (%v): %w\n", filing.AccessionNo, company.Name, company.Ticker, err)
		}

		if sectionContent != "" {
			rawContent += "\n\n" + sectionContent
		}
	}

	if rawContent == "" {
		logger.Infow(fmt.Sprintf("failed to fetchDocuments filing file (accession number %v) for %v (%v): no content (%v)\n", filing.AccessionNo, company.Name, company.Ticker, originURL), "companyID", company.ID, "filingKind", filingKind)
		return nil
	}

	// Create the document.
	filedAt, err := time.Parse(time.RFC3339, filing.FiledAt)
	if err != nil {
		return fmt.Errorf("failed to parse filing time (accession number %v) for %v (%v): %w\n", filing.AccessionNo, company.Name, company.Ticker, err)
	}

	// Wrap document creation and semantic indexing into a single transaction.
	if err = db.Transaction(func(tx *gorm.DB) error {
		logger.Infof("Creating document (accession number %v) for %v (%v) filed at %v", filing.AccessionNo, company.Name, company.Ticker, filedAt)
		document, err := models.CreateDocument(tx, company, filedAt, filingKind, originURL, rawContent)
		if err != nil {
			return fmt.Errorf("failed to create document (accession number %v) for %v (%v): %w\n", filing.AccessionNo, company.Name, company.Ticker, err)
		}

		text := documentloaders.NewText(strings.NewReader(rawContent))
		chunks, err := text.LoadAndSplit(context.Background(), splitter)
		if err != nil {
			return fmt.Errorf("failed to split document (accession number %v) for %v (%v): %w\n", filing.AccessionNo, company.Name, company.Ticker, err)
		}

		// Store chunks in the vector store.
		err = retrieval.StoreChunks(store, document.ID, chunks)
		if err != nil {
			return fmt.Errorf("failed to store chunks (accession number %v) for %v (%v): %w\n", filing.AccessionNo, company.Name, company.Ticker, err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
