package retrieval

import (
	"bytes"
	"cofin/models"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
	"golang.org/x/exp/constraints"
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

// CondenseConversation takes a conversation and condenses it into a single message.
func (g *Generator) CondenseConversation(ctx context.Context, company *models.Company, messages []models.Message) (string, error) {
	conversation := mergeMessages(messages)
	input := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: fmt.Sprintf(
				`You are COFIN AI, a virtual assistant that helps people read, analyze, and interpret financial filings of publicly traded companies. You have access to 10-K and 10-Q documents. Today is %v.`,
				time.Now().Format("2006-01-02")),
		},
		schema.AIChatMessage{
			Text: `Understood.`,
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf(
				`
I am going to send you a conversation history between a user and a virtual assistant as a single message. The conversation pertains to company %v, ($%v). The conversation will be provided in the following form:

User: <Message>
Assistant: <Message>
User: <Message>
			
Make the conversation shorter by rewording each message. Keep the order of messages. DO NOT add any new messages to the conversation or any text not already in the conversation. Keep the last message intact.`, company.Name, company.Ticker),
		},
		schema.AIChatMessage{
			Text: `Sounds good. I will take the conversation history, make it shorter but will preserve the meaning, order, and key points. I will preserve the last message as is. Now please send the conversation.`,
		},
		schema.HumanChatMessage{
			Text: conversation,
		},
		schema.HumanChatMessage{
			Text: `Now make the summary of the above as I asked you.`,
		},
	}

	res, err := g.Chat.Call(ctx, input, func(o *llms.CallOptions) { o.Model = "gpt-3.5-turbo-16k" })
	if err != nil {
		return "", err
	}

	return res, nil
}

func (g *Generator) CreateRetrieval(ctx context.Context, company *models.Company, documents []models.Document, conversation string) (documentID uint, query string, err error) {
	documentIDs, documentList := makeDocumentList(company, documents)
	jsonStr := fmt.Sprintf(`
{
	"model": "gpt-3.5-turbo-0613",
	"messages": [
		{"role": "system", "content": "You are COFIN AI, a virtual assistant that helps people read, analyze, and interpret financial filings of publicly traded companies. You have access to 10-K and 10-Q documents. Today is %v."},
		{"role": "assistant", "content": "Sounds good."},
		{"role": "user", "content": "I am going to send you a conversation history between a user and a virtual assistant as a single message. The conversation pertains to company %v ($%v). The conversation will be provided in the following format:\nUser: <Message>\nAssistant: <Message>\nUser: <Message>\n"},
		{"role": "assistant", "content": "Sounds good. What should I do with this conversation?"},
		{"role": "user", "content": "You have access to multiple financial documents about the company. Your task is to make a function call to retrieve_relevant_paragraphs which retrieves relevant paragraphs from the document of your choice using semantic search. You should use this function to answer the last user message in the conversation."},
		{"role": "assistant", "content": "Sounds good. What documents do I have access to?"},
		{"role": "user", "content": "Here's the list of documents you have access to in <DocumentID>: <Description> format:\n%v"},
		{"role": "assistant", "content": "Excellent. Now what is the conversation history?"},
		{"role": "user", "content": "%v"},
		{"role": "user", "content": "Now make the function call with the query and document ID"}
	   ],
	"temperature": 0.7,
	"functions": [
		{
			"name": "retrieve_relevant_paragraphs",
			"description": "Retrieve paragraphs related to the query from a document",
			"parameters": {
			   "type": "object",
			   "properties": {
				   "query": {
					   "type": "string",
					   "description": "Query to retrieve relevant paragraphs for."
				   },
				   "documentID": {"type": "number", "enum": %v}
			   },
			   "required": ["query", "documentID"]
			}
		}
	],
	"function_call": {"name": "retrieve_relevant_paragraphs"}
   }
	`, time.Now().Format("2006-01-02"), company.Name, company.Ticker, jsonEscapeString(documentList), jsonEscapeString(conversation), jsonEscapeArray(documentIDs))

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader([]byte(jsonStr)))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Logger = nil
	resp, err := client.StandardClient().Do(req)
	if err != nil {
		return 0, "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, "", err
	}

	type FunctionCall struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	type Message struct {
		FunctionCall FunctionCall `json:"function_call"`
	}
	type Choice struct {
		Message Message `json:"message"`
	}
	type Response struct {
		Choices []Choice `json:"choices"`
	}

	var res Response
	err = json.Unmarshal(b, &res)
	if err != nil {
		return 0, "", err
	}

	var Arguments struct {
		Query      string `json:"query"`
		DocumentID uint   `json:"documentID"`
	}
	args := res.Choices[0].Message.FunctionCall.Arguments
	err = json.Unmarshal([]byte(args), &Arguments)
	if err != nil {
		return 0, "", err
	}

	return Arguments.DocumentID, Arguments.Query, nil
}

// Continue generates a continuation to a conversation. It accepts a list of
// documents as context as well as a list of chunks of text relevant from each
// document, and the user message. It outputs a response and an error.
func (g *Generator) Continue(ctx context.Context, company *models.Company, conversation string, document *models.Document, chunks []string) (string, error) {
	var bigChunk string
	for j, chunk := range chunks {
		bigChunk += fmt.Sprintf("Paragraph %v: %v\n", j+1, chunk)
	}

	input := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: fmt.Sprintf(`You are COFIN AI, an Artificial Intelligence built purposely to analyze financial filings of publicly traded companies and answer questions based on them. You have access to 10-Q and 10-K filings. Today is %v.`, time.Now().Format("2006-01-02")),
		},
		schema.AIChatMessage{
			Text: `Understood.`,
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf("I am going to send a previous conversation history between you and a user as a single message. The conversation pertains to company %v ($%v)", company.Name, company.Ticker),
		},
		schema.AIChatMessage{
			Text: fmt.Sprintf("Sounds good. What should I do with this conversation?"),
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf("I am going to provide you with paragraphs from the %v document for %v filed at %v. These paragraphs are the most relevant to our conversation and you have chosen this document to as useful context for the conversation. You should generate a response to the last user message using this document context as the source of data.",
				document.Kind, company.Name, document.FiledAt),
		},
		schema.AIChatMessage{
			Text: "Perfect! I understand.",
		},
		schema.HumanChatMessage{Text: fmt.Sprintf(`Here are the paragraphs from the %v: %v`, document.Kind, bigChunk)},
		schema.AIChatMessage{Text: "Got it. Now please send me the conversation with the user."},
		schema.HumanChatMessage{Text: conversation},
		schema.HumanChatMessage{Text: "Now generate a response using the conversation I sent you and the paragraphs from the document you've chosen."},
	}

	// TODO: why do I have to set the model here and not in the constructor?
	res, err := g.Chat.Call(ctx, input, func(o *llms.CallOptions) { o.Model = "gpt-3.5-turbo-16k" })
	if err != nil {
		return "", err
	}

	return res, nil
}

// Format conversation history as a single string.
// TODO: count message length and cut off at threshold.
func mergeMessages(messages []models.Message) (conversation string) {
	if len(messages) > 5 {
		messages = messages[len(messages)-5:]
	}
	for _, message := range messages {
		if message.Author == models.UserAuthor {
			conversation += fmt.Sprintf("User: %v\n", message.Text)
		} else if message.Author == models.AIAuthor {
			conversation += fmt.Sprintf("Assistant: %v\n", message.Text)
		}
	}

	return conversation
}

func makeDocumentList(company *models.Company, documents []models.Document) (documentIDs []uint, documentList string) {
	for _, document := range documents {
		documentIDs = append(documentIDs, document.ID)
		documentList += fmt.Sprintf("%v: $%v %v %v\n", document.ID, company.Ticker, document.FiledAt.Format("2006-01-02"), document.Kind)
	}

	return documentIDs, documentList
}

func jsonEscapeString(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

func jsonEscapeArray[K constraints.Integer](i []K) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s
}
