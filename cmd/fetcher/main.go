package main

import (
	"cofin/core"
	"cofin/internal"
	"cofin/models"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jaytaylor/html2text"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/vectorstores"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const MAX_FILINGS_PER_COMPANY_PER_BATCH = 20

func main() {
	// Load environment variables.
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// Connect to the database.
	db, err := core.InitDB()
	if err != nil {
		panic(err)
	}

	// Initialize the embedder.
	embedder, err := internal.NewEmbedder()
	if err != nil {
		panic(err)
	}

	// Initialize the vector store.
	store, err := internal.NewPinecone(context.Background(), embedder)
	if err != nil {
		panic(err)
	}

	// Initialize text splitter.
	splitter, err := NewSplitter(1000, 100)
	if err != nil {
		panic(err)
	}

	logger, err := internal.NewLogger()
	if err != nil {
		panic(err)
	}

	// We could run this once a month and that'd be enough.
	t := time.NewTicker(1 * time.Hour)
	for ; true; <-t.C {
		// Go over all stock exchanges.
		for _, exchange := range stockExchanges {
			// Get listings for the exchange.
			listings, err := getTradedCompanies(exchange)
			if err != nil {
				logger.Errorw(fmt.Errorf("failed to get companies traded on an exchange: %v", err.Error()).Error(), "exchange", exchange)
				continue
			}

			for _, listing := range listings {
				// Skip delisted companies.
				if listing.IsDelisted {
					continue
				}

				// Create the company if it doesn't exist, fetch documents, and
				// store them.
				err := processListing(db, logger, listing, splitter, store)
				if err != nil {
					logger.Errorw(fmt.Errorf("failed to process a listing: %v", err.Error()).Error(), "ticker", listing.Ticker)
					continue
				}
			}
		}
	}
}

// processListing creates a company if it doesn't exist, fetches documents, and
// stores them.
func processListing(db *gorm.DB, logger *zap.SugaredLogger, listing SECListing, splitter *Splitter, store vectorstores.VectorStore) error {
	// Create or get a company in a transaction.
	var company *models.Company
	err := db.Transaction(func(tx *gorm.DB) (err error) {
		company, err = models.GetCompany(db, listing.Ticker)
		if err != nil {
			company = nil
			return err
		}

		if company != nil {
			return nil
		}

		company, err = models.CreateCompany(db, listing.Name, listing.Ticker, listing.CIK, time.Time{})
		return err
	})
	if err != nil {
		return fmt.Errorf("could not create company for %v (%v): %w\n", listing.Name, listing.Ticker, err)
	}

	// If the company documents were fetched in the past 24 hours, don't fetch
	// the company again.
	if !company.LastFetchedAt.IsZero() && company.LastFetchedAt.Add(24*time.Hour).After(time.Now()) {
		return nil
	}

	for _, filingKind := range []models.SourceKind{models.K10, models.Q10} {
		if err := processFilingKind(db, logger, company, splitter, store, filingKind); err != nil {
			logger.Errorw(fmt.Errorf("failed to process a filing kind for a company: %v", err.Error()).Error(), "ticker", company.Ticker, "filingKind", filingKind)
			continue
		}
	}

	return nil
}

// processFilingKind fetches filings of a particular kind for a company,
// processes and stores them.
func processFilingKind(db *gorm.DB, logger *zap.SugaredLogger, company *models.Company, splitter *Splitter, store vectorstores.VectorStore, filingKind models.SourceKind) error {
	// Get the most recent document of the kind for the company.
	document, err := models.GetMostRecentCompanyDocumentOfKind(db, company.ID, filingKind)
	if err != nil {
		return fmt.Errorf("failed to fetch most recent document for %v (%v): %w\n", company.Name, company.Ticker, err)
	}

	// Query for the last two years of documents if we have no documents for the
	// company. Otherwise query for documents since the last document.
	var lastFiledAt = time.Now().Add(-365 * 2 * 24 * time.Hour)
	if document != nil {
		lastFiledAt = document.FiledAt
	}

	// Get filings since the last filed time.
	filings, err := getFilingsSince(company.CIK, filingKind, lastFiledAt)
	if err != nil {
		return fmt.Errorf("failed to fetch filings for %v (%v): %w\n", company.Name, company.Ticker, err)
	}

	// Process filings. We only process up to MAX_FILINGS_PER_COMPANY_PER_BATCH
	// at a time. This guarantees that no company hogs the fetching pipeline for
	// too long. If not all documents are fetched, next time the company is due
	// for re-fetching we will continue where we left off by checking most
	// recent document's filing time.
	for _, filing := range filings[:MAX_FILINGS_PER_COMPANY_PER_BATCH] {
		// Process the filing in a transaction. Processing a filing is atomic
		// and involves three things: storing the file in the DB, storing the
		// chunks in vector store, and updating the company. If any of these
		// suboperations fail, we revert and abort.
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := processFiling(db, company, splitter, store, filingKind, filing); err != nil {
				return fmt.Errorf("failed to process a filing with accession number %v for a company: %v", filing.AccessionNo, err.Error())
			}

			// Update the company's last fetched time after successfully
			// processing a filing for it.
			company.LastFetchedAt = time.Now()
			err = db.Save(&company).Error
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
func processFiling(db *gorm.DB, company *models.Company, splitter *Splitter, store vectorstores.VectorStore, filingKind models.SourceKind, filing Filing) error {
	// Get the filing file from the SEC.
	originURL, file, err := getFilingFile(filing)
	if err != nil {
		return fmt.Errorf("failed to fetch filing file for %v (%v): %w\n", company.Name, company.Ticker, err)
	}

	// Convert the file to text.
	html, err := html2text.FromString(string(file))
	if err != nil {
		return fmt.Errorf("failed to parse filing file for %v (%v): %w\n", company.Name, company.Ticker, err)
	}

	// Create the document.
	filedAt, err := time.Parse(time.RFC3339, filing.FiledAt)
	if err != nil {
		return fmt.Errorf("failed to parse filing time for %v (%v): %w\n", company.Name, company.Ticker, err)
	}

	// By wrapping this piece of code in a transaction we ensure that the vector
	// DB and documents in SQL are in sync. If a document fails to create, we
	// obviously won't upload chunks to Pinecone. But if chunks fail to upload,
	// we will revert document creation.
	//
	// Note that chunks are uploaded in batches, so a batch might succeed
	// uploading and then a subsequent batch will fail, in which case there will
	// be lingering chunks in vector store. This is not an acute issue -- we
	// always query the vector store by filtering for document IDs, and if we
	// don't have the document saved with this ID, the chunks will simply be
	// "dead".
	if err = db.Transaction(func(tx *gorm.DB) error {
		document, err := models.CreateDocument(db, company, filedAt, filingKind, originURL, html)
		if err != nil {
			return fmt.Errorf("failed to create document for %v (%v): %w\n", company.Name, company.Ticker, err)
		}

		// Split the document into chunks.
		text := documentloaders.NewText(strings.NewReader(html))
		chunks, err := text.LoadAndSplit(context.Background(), splitter)
		if err != nil {
			return fmt.Errorf("failed to split document for %v (%v): %w\n", company.Name, company.Ticker, err)
		}

		// Store chunks in the vector store. In the future, we might want to store
		// chunks in the SQL DB. This will largely depend on our supportability and
		// debugging needs.
		err = internal.StoreChunks(store, document.UUID, chunks)
		if err != nil {
			return fmt.Errorf("failed to store chunks for %v (%v): %w\n", company.Name, company.Ticker, err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
