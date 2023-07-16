package retrieval

import (
	"context"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"gorm.io/gorm"
)

// Retriever can retrieve information based on keywords and semantics.
type Retriever struct {
	db       *gorm.DB
	embedder embeddings.Embedder
	store    vectorstores.VectorStore
	topK     int
}

// NewRetriever creates a new Retriever namespaces to the given company.
func NewRetriever(db *gorm.DB, companyID uint) (*Retriever, error) {
	embedder, err := NewEmbedder()
	if err != nil {
		return nil, err
	}

	store, err := NewPinecone(context.Background(), embedder, companyID)
	if err != nil {
		return nil, err
	}

	return &Retriever{
		db:       db,
		embedder: embedder,
		store:    store,
		topK:     10,
	}, nil
}

func (r *Retriever) GetSemanticChunks(ctx context.Context, companyID, documentID uint, text string) ([]string, error) {
	docs, err := r.store.SimilaritySearch(context.Background(), text, r.topK, vectorstores.WithFilters(map[string]any{
		// This is type-sensitive. Setting this to a string, for example, will
		// return no results.
		"document_id": documentID,
	}))
	if err != nil {
		return nil, err
	}

	docStrings := make([]string, len(docs))
	for i, doc := range docs {
		docStrings[i] = doc.PageContent
	}

	return docStrings, nil
}
