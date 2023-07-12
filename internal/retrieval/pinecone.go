package retrieval

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
	"golang.org/x/sync/errgroup"
)

// NewPinecone initializes a new Pinecone vector store.
func NewPinecone(ctx context.Context, embedder embeddings.Embedder, companyID uint) (*pinecone.Store, error) {
	store, err := pinecone.New(
		ctx,
		pinecone.WithProjectName(os.Getenv("PINECONE_PROJECT")),
		pinecone.WithIndexName(os.Getenv("PINECONE_INDEX")),
		pinecone.WithEnvironment(os.Getenv("PINECONE_ENVIRONMENT")),
		pinecone.WithEmbedder(embedder),
		pinecone.WithAPIKey(os.Getenv("PINECONE_API_KEY")),
		pinecone.WithNameSpace(fmt.Sprint(companyID)),
	)

	if err != nil {
		return nil, err
	}

	return &store, nil
}

// StoreChunks stores document chunks in Pinecone.
func StoreChunks(store vectorstores.VectorStore, documentID uint, chunks []schema.Document) error {
	const BATCH_SIZE = 50

	// Set document metadata. We only set the ID that matches the internal ID.
	//
	// Langchain sets document text for us.
	for i := range chunks {
		// Modify chunks in-place. They are not pointers.
		chunks[i].Metadata = map[string]interface{}{
			"document_id": documentID,
		}
	}

	errs, ctx := errgroup.WithContext(context.Background())
	for i := 0; i <= len(chunks); i += BATCH_SIZE {
		end := i + BATCH_SIZE
		if end > len(chunks) {
			end = len(chunks)
		}

		func(i, end int) {
			errs.Go(func() error {
				err := store.AddDocuments(ctx, chunks[i:end])
				if err != nil {
					return err
				}

				return nil
			})
		}(i, end)

	}

	return errs.Wait()
}
