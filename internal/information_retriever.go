package internal

import (
	"context"
	"errors"
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

// InformationRetriever can retrieve information based on keywords and
// semantics.
type InformationRetriever struct {
	Embedder embeddings.Embedder
	Store    vectorstores.VectorStore
	TopK     int
}

func NewInformationRetriever() (*InformationRetriever, error) {
	embedder, err := embeddings.NewOpenAI()
	if err != nil {
		return nil, err
	}
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
		return nil, err
	}

	return &InformationRetriever{
		Embedder: embedder,
		Store:    store,
		TopK:     10,
	}, nil
}

func (ir *InformationRetriever) Get(ctx context.Context, ticker string, year int, quarter Quarter, sourceKind SourceKind, text string) ([]string, error) {
	if ticker != "$NET" {
		return nil, errors.New("TODO: remove this check. Please use $NET as the ticker.")
	}

	if year != 2023 {
		return nil, errors.New("TODO: remove this check. Please use 2023 as the year.")
	}

	if quarter != 1 {
		return nil, errors.New("TODO: remove this check. Please use 1 as the quarter.")
	}

	if sourceKind != "10-Q" {
		return nil, errors.New("TODO: remove this check. Please use 10-Q as the source type.")
	}

	docs, err := ir.Store.SimilaritySearch(context.Background(), text, ir.TopK)
	if err != nil {
		return nil, err
	}

	docStrings := make([]string, len(docs))
	for i, doc := range docs {
		docStrings[i] = doc.PageContent
	}

	return docStrings, nil
}
