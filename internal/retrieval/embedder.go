package retrieval

import (
	"os"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/embeddings/openai"
	llm "github.com/tmc/langchaingo/llms/openai"
)

// This embedder initializes the embedding model using the OPENAI_MODEL
// environment variable.
func NewEmbedder() (embeddings.Embedder, error) {
	client, err := llm.New(llm.WithModel(os.Getenv("OPENAI_EMBEDDING_MODEL")), llm.WithBaseURL(os.Getenv("OPENAI_BASE_URL")), llm.WithToken(os.Getenv("OPENAI_API_KEY")))
	if err != nil {
		return nil, err
	}

	embedder, err := openai.NewOpenAI(openai.WithClient(*client), openai.WithBatchSize(512), openai.WithStripNewLines(false), openai.WithStripNewLines(false))
	if err != nil {
		return nil, err
	}

	return &embedder, nil
}
