package demo

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/dslipak/pdf"
	"github.com/joho/godotenv"
	loader "github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/pinecone"
)

func run() {
	godotenv.Load()

	content, err := readPdf("cloudflare-10-q-march-2023.pdf")
	if err != nil {
		panic(err)
	}

	text := loader.NewText(strings.NewReader(content))
	splitter, err := NewSplitter(1000, 100)
	if err != nil {
		panic(err)
	}

	docs, err := text.LoadAndSplit(context.Background(), splitter)
	if err != nil {
		panic(err)
	}

	embedder, err := embeddings.NewOpenAI()
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
		log.Fatal(err)
	}

	// Go in batches of 15 when pushing to Pinecone.
	for i := 0; i <= len(docs); i += 30 {
		end := i + 15
		if end > len(docs) {
			end = len(docs)
		}

		// err = store.AddDocuments(context.Background(), docs[i:end])
		// if err != nil {
		// log.Fatal(err)
		// }
	}

	userMessage := "Hey, who is the CFO of Cloudflare?"

	docs, err = store.SimilaritySearch(context.Background(), userMessage, 5)
	if err != nil {
		log.Fatal(err)
	}

	docStrings := make([]string, len(docs))
	for i, doc := range docs {
		docStrings[i] = doc.PageContent
	}
	// log.Println(strings.Join(docStrings, "\n"))

	// TODO: calculate token limit
	// tkm, err := tiktoken.EncodingForModel("gpt-3.5-turbo")
	chat, err := openai.NewChat(openai.WithModel("gpt-3.5-turbo"))
	res, err := chat.Call(context.Background(), []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: `
			You are Artificial Intelligence built purposely to analyse financial filings of publicly traded companies and answer questions based on them.
			Below you will find a list of paragraphs from the 10-Q filing of Cloudflare, Inc. for the quarter ended March 31, 2023.
			Please read them and answer the user message below. You can optionally use the provided paragraph as context to answer the user.
			You may choose not to use the paragraphs in your answer.
			Think step by step and explain your reasoning.
			`,
		},
		schema.AIChatMessage{
			Text: `Understood.`,
		},
		schema.SystemChatMessage{
			Text: fmt.Sprintf("Here are the paragraphs from Cloudflare's ($NET) 10-Q filing from March 31, 2023 most relevant to user's input:\n%v", strings.Join(docStrings, "\n")),
		},
		schema.AIChatMessage{
			Text: `Sounds good.`,
		},
		schema.HumanChatMessage{
			Text: userMessage,
		},
	})

	log.Printf("Question: %v\n\nAnswer: %v\n", userMessage, res)
}

func readPdf(path string) (string, error) {
	r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}

	buf.ReadFrom(b)
	return buf.String(), nil
}
