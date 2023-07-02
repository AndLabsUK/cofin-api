package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/jaytaylor/html2text"
	"github.com/joho/godotenv"
	loader "github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

// TODO: make this a loop with jobs
// TODO: pull companies that need to be fetched from a database
// TODO: create proper interfaces for embeddings & interaction with Pinecone
// TODO: annotate documents in Pinecone with metadata

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}

	client := &http.Client{}
	for {
		cik := "CIK0001477333"

		// Get SEC submissions for a company.
		req, err := http.NewRequest("GET", fmt.Sprintf("https://data.sec.gov/submissions/%v.json", cik), nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("User-Agent", "andlabs.co.uk")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		var submissions map[string]interface{}
		err = json.Unmarshal(b, &submissions)
		if err != nil {
			log.Fatal(err)
		}

		recentFilings := submissions["filings"].(map[string]interface{})["recent"].(map[string]interface{})
		reportDates := recentFilings["reportDate"].([]interface{})
		accessionNumbers := recentFilings["accessionNumber"].([]interface{})
		forms := recentFilings["form"].([]interface{})
		primaryDocuments := recentFilings["primaryDocument"].([]interface{})

		// Get indexes of 10-Qs.
		indexes := make([]int, 0)
		for i, form := range forms {
			if strings.ToUpper(form.(string)) == "10-Q" {
				indexes = append(indexes, i)
			}
		}

		// Log 10-Qs.
		for _, i := range indexes {
			log.Printf("Filed %v on %v, accession number %v, primary document %v", forms[i], reportDates[i], accessionNumbers[i], primaryDocuments[i])
		}

		// Get the raw file for the last 10-Q.
		lastIndex := indexes[len(indexes)-1]
		// For whatever reason, in this URL SEC expects us to remove 0s from the
		// CIK prefix and remove dashes from the accession number.
		cik = strings.ReplaceAll(strings.ReplaceAll(cik, "CIK", ""), "0", "")
		accessionNumber := strings.ReplaceAll(accessionNumbers[lastIndex].(string), "-", "")
		req, err = http.NewRequest("GET", fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%v/%v/%v", cik, accessionNumber, primaryDocuments[lastIndex]), nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("User-Agent", "andlabs.co.uk")
		resp, err = client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		b, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		html, err := html2text.FromString(string(b))
		if err != nil {
			panic(err)
		}

		// Split the document into overlapping chunks and upload them to
		// Pinecone.
		text := loader.NewText(strings.NewReader(html))
		splitter, err := NewSplitter(1000, 100)
		if err != nil {
			panic(err)
		}

		docs, err := text.LoadAndSplit(context.Background(), splitter)
		if err != nil {
			panic(err)
		}

		embedder, err := embeddings.NewOpenAI()
		embedder.BatchSize = 30
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
			log.Fatal(err)
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

		return
	}
}
