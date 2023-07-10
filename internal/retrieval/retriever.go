package retrieval

import (
	"cofin/models"
	"context"

	"github.com/google/uuid"
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

func NewRetriever(db *gorm.DB, ticker string) (*Retriever, error) {
	embedder, err := NewEmbedder()
	if err != nil {
		return nil, err
	}

	store, err := NewPinecone(context.Background(), embedder, ticker)
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

// For now, we always use two most recent documents available for any given
// ticker. The ability to customise what documents to use will come in the paid
// plan as we build up functionality. Ideally, we should be recognising what
// period to retrieve documents for based on free-form user input.
func (r *Retriever) GetDocuments(companyID uint) ([]models.Document, error) {
	documents, err := models.GetRecentCompanyDocuments(r.db, companyID, 2)
	if err != nil {
		return nil, err
	}

	return documents, nil
}

func (r *Retriever) GetSemanticChunks(ctx context.Context, ticker string, documentUUID uuid.UUID, text string) ([]string, error) {
	// TODO: should I set the namespace here or in the constructor?
	docs, err := r.store.SimilaritySearch(context.Background(), text, r.topK, vectorstores.WithNameSpace(ticker), vectorstores.WithFilters(map[string]string{
		"document_uuid": documentUUID.String(),
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
