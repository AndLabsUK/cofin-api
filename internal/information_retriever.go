package internal

import (
	"cofin/models"
	"context"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"gorm.io/gorm"
)

// InformationRetriever can retrieve information based on keywords and
// semantics.
type InformationRetriever struct {
	db       *gorm.DB
	embedder embeddings.Embedder
	store    vectorstores.VectorStore
	topK     int
}

func NewInformationRetriever(db *gorm.DB, ticker string) (*InformationRetriever, error) {
	embedder, err := NewEmbedder()
	if err != nil {
		return nil, err
	}

	store, err := NewPinecone(context.Background(), embedder, ticker)
	if err != nil {
		return nil, err
	}

	return &InformationRetriever{
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
func (ir *InformationRetriever) GetDocuments(ctx context.Context, ticker string) (*models.Company, []models.Document, error) {
	company, err := models.GetCompany(ir.db, ticker)
	if err != nil {
		return nil, nil, err
	}

	documents, err := models.GetRecentCompanyDocuments(ir.db, company.ID, 2)
	if err != nil {
		return nil, nil, err
	}

	return company, documents, nil
}

func (ir *InformationRetriever) GetSemanticChunks(ctx context.Context, ticker string, documentUUID uuid.UUID, text string) ([]string, error) {
	// TODO: should I set the namespace here or in the constructor?
	docs, err := ir.store.SimilaritySearch(context.Background(), text, ir.topK, vectorstores.WithNameSpace(ticker), vectorstores.WithFilters(map[string]string{
		"documentUUID": documentUUID.String(),
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
