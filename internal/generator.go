package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

// Generator is a type that completes conversations with AI.
type Generator struct {
	// Chat is the underlying chatbot.
	Chat llms.ChatLLM
}

func NewGenerator() (*Generator, error) {
	chat, err := openai.NewChat(openai.WithModel("gpt-3.5-turbo"))
	if err != nil {
		return nil, err
	}

	return &Generator{
		Chat: chat,
	}, nil
}

func (g *Generator) Continue(sources []string, text string) (string, error) {
	// TODO: use prompt templates
	// TODO: use langchain memory
	res, err := g.Chat.Call(context.Background(), []schema.ChatMessage{
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
			Text: fmt.Sprintf("Here are the paragraphs from Cloudflare's ($NET) 10-Q filing from March 31, 2023 most relevant to user's input:\n%v", strings.Join(sources, "\n")),
		},
		schema.AIChatMessage{
			Text: `Sounds good.`,
		},
		schema.HumanChatMessage{
			Text: text,
		},
	})

	if err != nil {
		return "", err
	}

	return res, nil
}
