package retrieval

import (
	"cofin/models"
	"context"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/vectorstores"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Retriever can retrieve information based on keywords and semantics.
type Retriever struct {
	db       *gorm.DB
	logger   *zap.SugaredLogger
	embedder embeddings.Embedder
	store    vectorstores.VectorStore
	topK     int
}

// NewRetriever creates a new Retriever namespaces to the given company.
func NewRetriever(db *gorm.DB, logger *zap.SugaredLogger, companyID uint) (*Retriever, error) {
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
		logger:   logger,
		embedder: embedder,
		store:    store,
		topK:     3,
	}, nil
}

func (r *Retriever) GetSemanticChunks(ctx context.Context, documentID uint, text string) ([]string, error) {
	chunks, err := r.store.SimilaritySearch(context.Background(), text, r.topK, vectorstores.WithFilters(map[string]any{
		// This is type-sensitive. Setting this to a string, for example, will
		// return no results.
		"document_id": documentID,
	}))
	if err != nil {
		return nil, err
	}

	chunkStrings := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		if id, ok := chunk.Metadata["id"]; ok {
			internalChunk, err := models.GetChunkByID(r.db, id.(uint))
			if err != nil {
				return nil, err
			}

			if internalChunk == nil {
				r.logger.Errorf("chunk with id %d not found (document ID %d)", id, documentID)
				continue
			}
		} else {
			chunkStrings = append(chunkStrings, chunk.PageContent)
		}
	}

	return chunkStrings, nil
}
