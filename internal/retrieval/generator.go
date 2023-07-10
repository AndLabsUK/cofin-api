package retrieval

import (
	"cofin/models"
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

// Generator is a type that completes conversations with AI.
type Generator struct {
	// Chat is the underlying chatbot.
	Chat llms.ChatLLM
}

// NewGenerator creates a new conversation generator.
func NewGenerator() (*Generator, error) {
	// TODO: decide what model to use
	chat, err := openai.NewChat(openai.WithModel("gpt-3.5-turbo-16k"))
	if err != nil {
		return nil, err
	}

	return &Generator{
		Chat: chat,
	}, nil
}

// Continue generates a continuation to a conversation. It accepts a list of
// documents as context as well as a list of chunks of text relevant from each
// document, and the user message. It outputs a response and an error.
//
// TODO: use existing messages
//
// TODO: use prompt templates
//
// TODO: use langchain memory
func (g *Generator) Continue(ctx context.Context, company models.Company, documents []models.Document, chunks [][]string, message string) (string, error) {
	messages := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: fmt.Sprintf(`
			You are Artificial Intelligence built purposely to analyse financial filings of publicly traded companies and answer questions based on them.
			Below you will find a list of paragraphs from the documents of %v ($%v). 
			Please read them and answer the user message below. You can optionally use the provided paragraph as context to answer the user.
			You may choose not to use the paragraphs in your answer.
			Think step by step and explain your reasoning. 
			Also keep it mind that your name is COFIN AI and you have nothing to do with OpenAI or ChatGPT.
			`, company.Name, company.Ticker),
		},
		schema.AIChatMessage{
			Text: `Understood.`,
		},
	}

	for i, document := range documents {
		var bigChunk string
		for j, chunk := range chunks[i] {
			bigChunk += fmt.Sprintf("Paragraph %v: %v\n", j+1, chunk)
		}

		messages = append(messages, schema.SystemChatMessage{
			Text: fmt.Sprintf(`
			Below are a few select paragraphs from the %v document filed at %v.
			These paragraphs are not necessarily consecutive and may have been extracted from different parts of the file.
			%v
			`, document.Kind, document.FiledAt, bigChunk),
		})
	}

	messages = append(messages, schema.AIChatMessage{
		Text: "Excellent. I am happy to help analyse this information. Now please provide me with the user question.",
	})

	messages = append(messages, schema.HumanChatMessage{
		Text: message,
	})

	// TODO: why do I have to set the model here and not in the constructor?
	res, err := g.Chat.Call(ctx, messages, func(o *llms.CallOptions) { o.Model = "gpt-3.5-turbo-16k" })
	if err != nil {
		return "", err
	}

	return res, nil
}
