package internal

import "github.com/tmc/langchaingo/embeddings"

// This embedder initializes the embedding model using the OPENAI_MODEL
// environment variable.
func NewEmbedder() (*embeddings.OpenAI, error) {
	embedder, err := embeddings.NewOpenAI()
	if err != nil {
		return nil, err
	}
	embedder.BatchSize = 30

	return &embedder, nil
}
