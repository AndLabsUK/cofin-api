package internal

import (
	"context"
	"os"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

// NewPinecone initializes a new Pinecone vector store.
func NewPinecone(ctx context.Context, embedder embeddings.Embedder) (*pinecone.Store, error) {
	store, err := pinecone.New(
		ctx,
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

	return &store, nil
}

// StoreChunks stores document chunks in Pinecone.
func StoreChunks(store vectorstores.VectorStore, documentUUID uuid.UUID, chunks []schema.Document) error {
	// Set document metadata. We only set the ID that matches the internal ID.
	//
	// Langchain sets document text for us.
	for i := range chunks {
		// Modify chunks in-place. They are not pointers.
		chunks[i].Metadata = map[string]interface{}{
			"documentUUID": documentUUID,
		}
	}

	// Go in batches of 50 when pushing to Pinecone.
	for i := 0; i <= len(chunks); i += 50 {
		end := i + 50
		if end > len(chunks) {
			end = len(chunks)
		}

		err := store.AddDocuments(context.Background(), chunks[i:end])
		if err != nil {
			return err
		}
	}

	return nil
}
