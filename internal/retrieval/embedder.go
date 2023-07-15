package retrieval

import (
	"github.com/tmc/langchaingo/embeddings"
)

// This embedder initializes the embedding model using the OPENAI_MODEL
// environment variable.
func NewEmbedder() (embeddings.Embedder, error) {
	// TODO: set model explicitly (not available in the library currently).
	embedder, err := embeddings.NewOpenAI()
	if err != nil {
		return nil, err
	}
	embedder.BatchSize = 512
	embedder.StripNewLines = false

	return &embedder, nil
}
