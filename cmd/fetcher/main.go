package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jaytaylor/html2text"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

func main() {
	// Load environment variables.
	err := godotenv.Load()
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
				// TODO: Retry here and everywhere else.
				log.Printf("Failed to fetch listings for %v: %v\n", exchange, err)
				continue
			}

			for _, listing := range listings {
				if listing.IsDelisted {
					continue
				}

				fmt.Printf("Fetching %v (%v)...\n", listing.Name, listing.Ticker)
				// TODO: Support more document types.
				filings, err := getFilings(listing.CIK, "10-Q")
				if err != nil {
					log.Printf("Failed to fetch filings for %v (%v): %v\n", listing.Name, listing.Ticker, err)
					continue
				}

				for _, filing := range filings {
					f, err := getFilingFile(filing)
					if err != nil {
						log.Printf("Failed to fetch filing file for %v (%v): %v\n", listing.Name, listing.Ticker, err)
						continue
					}

					html, err := html2text.FromString(string(f))
					if err != nil {
						log.Printf("Failed to parse filing file for %v (%v): %v\n", listing.Name, listing.Ticker, err)
						continue
					}

					text := documentloaders.NewText(strings.NewReader(html))
					docs, err := text.LoadAndSplit(context.Background(), splitter)
					if err != nil {
						log.Printf("Failed to split document for %v (%v): %v\n", listing.Name, listing.Ticker, err)
						continue
					}

					// Set document metadata. This will match what is set in the
					// DB and will be used for filtering.
					for _, doc := range docs {
						doc.Metadata = map[string]interface{}{
							"ID":               "123",
							"accession_number": filing.AccessionNo,
							"ticker":           listing.Ticker,
							"name":             listing.Name,
							"date":             filing.FiledAt,
							"type":             "10-Q",
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
