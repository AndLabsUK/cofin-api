package main

import (
	"cofin/core"
	"cofin/internal"
	"cofin/models"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jaytaylor/html2text"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
	"gorm.io/gorm"
)

var SECFilingKinds = []internal.SourceKind{internal.K10, internal.Q10}

func GetOrCreateCompanyFromSECListing(db *gorm.DB, listing SECListing) (*models.Company, error) {
	ticker := strings.ToUpper(listing.Ticker)

	var company models.Company
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := db.Where("ticker = ?", ticker).First(&company).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				company = models.Company{
					Name:   listing.Name,
					Ticker: ticker,
					CIK:    listing.CIK,
				}

				if err := db.Create(&company).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &company, nil
}

func GetMostRecentDocumentOfType(db *gorm.DB, company *models.Company, kind internal.SourceKind) (*models.Document, error) {
	var document models.Document
	err := db.Where("company_id = ? AND kind = ?", company.ID, kind).Order("filed_at DESC").First(&document).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &document, nil
}

func CreateDocument(db *gorm.DB, company *models.Company, filedAt time.Time, kind internal.SourceKind, originURL, rawContent string) (*models.Document, error) {
	document := models.Document{
		CompanyID:  company.ID,
		FiledAt:    filedAt,
		Kind:       kind,
		OriginURL:  originURL,
		RawContent: rawContent,
	}

	if err := db.Create(&document).Error; err != nil {
		return nil, err
	}

	return &document, nil
}

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
	embedder, err := embeddings.NewOpenAI()
	embedder.BatchSize = 30

	// Initialize vector document store.
	store, err := pinecone.New(
		context.Background(),
		pinecone.WithProjectName(os.Getenv("PINECONE_PROJECT")),
		pinecone.WithIndexName(os.Getenv("PINECONE_INDEX")),
		pinecone.WithEnvironment(os.Getenv("PINECONE_ENVIRONMENT")),
		pinecone.WithEmbedder(embedder),
		pinecone.WithAPIKey(os.Getenv("PINECONE_API_KEY")),
		pinecone.WithNameSpace("$NET"),
	)
	if err != nil {
		panic(err)
	}

	// Initialize text splitter.
	splitter, err := NewSplitter(1000, 100)
	if err != nil {
		panic(err)
	}

	// We could run this once a month and that'd be enough.
	t := time.NewTicker(1 * time.Hour)
	for ; true; <-t.C {
		for _, exchange := range stockExchanges {
			listings, err := getTradedCompanies(exchange)
			if err != nil {
				log.Printf("Failed to fetch listings for %v: %v\n", exchange, err)
				continue
			}

			for _, listing := range listings {
				if listing.IsDelisted {
					continue
				}

				company, err := GetOrCreateCompanyFromSECListing(db, listing)
				if err != nil {
					log.Printf("Could not create company for %v (%v): %v\n", listing.Name, listing.Ticker, err)
					continue
				}

				for _, filingKind := range SECFilingKinds {
					document, err := GetMostRecentDocumentOfType(db, company, filingKind)
					if err != nil {
						log.Printf("Failed to fetch most recent document for %v (%v): %v\n", listing.Name, listing.Ticker, err)
						continue
					}

					// Query for the last year of documents if we have no
					// documents for the company. Otherwise query for documents
					// since the last document.
					var lastFiledAt = time.Now().Add(-365 * 24 * time.Hour)
					if document != nil {
						lastFiledAt = document.FiledAt
					}

					// TODO: debug -- looks like some companies return no
					// documents. Why?
					filings, err := getFilingsSince(listing.CIK, filingKind, lastFiledAt)
					if err != nil {
						log.Printf("Failed to fetch filings for %v (%v): %v\n", listing.Name, listing.Ticker, err)
						continue
					}

					for _, filing := range filings {
						originURL, f, err := getFilingFile(filing)
						if err != nil {
							log.Printf("Failed to fetch filing file for %v (%v): %v\n", listing.Name, listing.Ticker, err)
							continue
						}

						// TODO: do some basic cleanup on the HTML. (Too much
						// whitespace, etc.)
						html, err := html2text.FromString(string(f))
						if err != nil {
							log.Printf("Failed to parse filing file for %v (%v): %v\n", listing.Name, listing.Ticker, err)
							continue
						}

						// TODO: fix URL.
						filingAt, err := time.Parse(time.RFC3339, filing.FiledAt)
						if err != nil {
							log.Printf("Failed to parse filing time for %v (%v): %v\n", listing.Name, listing.Ticker, err)
							continue
						}

						_, err = CreateDocument(db, company, filingAt, filingKind, originURL, html)
						if err != nil {
							log.Printf("Failed to create document for %v (%v): %v\n", listing.Name, listing.Ticker, err)
							continue
						}

						text := documentloaders.NewText(strings.NewReader(html))
						docs, err := text.LoadAndSplit(context.Background(), splitter)
						if err != nil {
							log.Printf("Failed to split document for %v (%v): %v\n", listing.Name, listing.Ticker, err)
							continue
						}

						// Set document metadata. We only set the ID that matches
						// the internal ID.
						//
						// Langchain sets document text for us.
						for _, doc := range docs {
							doc.Metadata = map[string]interface{}{
								"ID": "123",
							}
						}

						// Go in batches of 50 when pushing to Pinecone.
						for i := 0; i <= len(docs); i += 50 {
							end := i + 50
							if end > len(docs) {
								end = len(docs)
							}

							err = store.AddDocuments(context.Background(), docs[i:end])
							if err != nil {
								log.Fatal(err)
							}
						}
					}
				}
			}
		}
	}
}
