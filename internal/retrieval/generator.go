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
	Chat        llms.ChatLLM
	model       string
	temperature float64
}

// NewGenerator creates a new conversation generator.
func NewGenerator() (*Generator, error) {
	model := os.Getenv("OPENAI_CONVERSATIONAL_MODEL")
	chat, err := openai.NewChat(openai.WithModel(model))
	if err != nil {
		return nil, err
	}

	return &Generator{
		Chat:        chat,
		model:       model,
		temperature: 0.7,
	}, nil
}

// CondenseConversation takes a conversation and condenses it into a single message.
func (g *Generator) CondenseConversation(ctx context.Context, company *models.Company, messages []models.Message, lastMessage string) (string, error) {
	conversation := mergeMessages(messages)
	input := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: fmt.Sprintf(
				"You are COFIN AI, a virtual assistant that helps people read, analyze, and interpret financial filings of publicly traded companies. You have access to 10-K and 10-Q documents filed to SEC. Today is %v.",
				time.Now().Format("2006-01-02")),
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf(
				`
I am going to send you a conversation history between you and a user as a single message. The conversation pertains to company %v, ($%v). The conversation will be provided in the following form:

User: <message>
COFIN AI: <message>
User: <message>
			
Your task is to rewrite each message to make it shorter but to keep the most important context that will help you answer the user's last message.`, company.Name, company.Ticker),
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf("Here is the conversation history:\n%v", conversation),
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf("The last message from the user is:\n%v", lastMessage),
		},
		schema.HumanChatMessage{
			Text: "Now rewrite the conversation message-by-message as I told you. Don't add any new messages from the user or COFIN AI. Do NOT answer the user's last message. Just rewrite the conversation to keep the context important for the user's last message.",
		},
	}

	res, err := g.Chat.Call(ctx, input, llms.WithTemperature(g.temperature))
	if err != nil {
		return "", err
	}

	return res, nil
}

// CreateRetrieval either generates a direct response to the user's message or
// creates arguments for further information retrieval. It uses a conversation
// and available documents to decide which document to retrieve from using what
// query.
//
// If returned *message is not nil, no further retrieval is necessary.
func (g *Generator) CreateRetrieval(ctx context.Context, company *models.Company, documents []models.Document, conversation string, lastMessage string) (earlyResponse *string, documentID uint, query string, err error) {
	documentIDs, documentList := makeDocumentList(company, documents)
	documentIDsFormatted := jsonEscapeArray(documentIDs)
	documentListFormatted := jsonEscapeString(documentList)
	conversationFormatted := jsonEscapeString(conversation)
	lastMessageFormatted := jsonEscapeString(lastMessage)
	jsonStr := fmt.Sprintf(`
{
	"model": "%v",
	"messages": [
		{"role": "system", "content": "You are COFIN AI, a virtual assistant that helps people read, analyze, and interpret financial filings of publicly traded companies. You have access to 10-K and 10-Q documents filed to SEC. Today is %v."},
		{"role": "user", "content": "I am going to send you a conversation history between you and a user as a single message. The conversation pertains to company %v ($%v). You have access to financial documents of the company."},
		{"role": "user", "content": "You need to respond to the user's last message. You can either create a response right away or make a function call to retrieve_relevant_paragraphs which retrieves relevant paragraphs from the document of your choice using semantic search. If you want to, you can retrieve this information to answer the last user message in the conversation."},
		{"role": "user", "content": "Here's the list of documents you have access to in <DocumentID>: <Description> format:\n%v"},
		{"role": "user", "content": "Here is the conversation history:\n%v"},
		{"role": "user", "content": "And here is the user's last message:\n%v"},
		{"role": "user", "content": "Now either respond to the last message (address the user directly) or make a function call. If you decide to do a function call, you will be retrieving information from the 10-K or 10-Q document of your choice using semantic vector similarity, so create a query that will match relevant information in the document."}
	   ],
	"temperature": %v,
	"functions": [
		{
			"name": "retrieve_relevant_paragraphs",
			"description": "Semantically retrieve paragraphs related to the query from the document.",
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
	"function_call": "auto"
   }
	`, g.model, time.Now().Format("2006-01-02"), company.Name, company.Ticker, documentListFormatted, conversationFormatted, lastMessageFormatted, g.temperature, documentIDsFormatted)

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader([]byte(jsonStr)))
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", os.Getenv("OPENAI_API_KEY")))
	req.Header.Set("Content-Type", "application/json")

	client := retryablehttp.NewClient()
	client.Logger = nil
	resp, err := client.StandardClient().Do(req)
	if err != nil {
		return nil, 0, "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, "", err
	}

	type FunctionCall struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	type Message struct {
		FunctionCall *FunctionCall `json:"function_call,omitempty"`
		Content      string        `json:"content"`
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
		return nil, 0, "", err
	}

	var Arguments struct {
		Query      string `json:"query"`
		DocumentID uint   `json:"documentID"`
	}
	if len(res.Choices) == 0 {
		return nil, 0, "", fmt.Errorf("no choices returned: %v", res)
	}
	m := res.Choices[0].Message
	if m.FunctionCall != nil {
		args := res.Choices[0].Message.FunctionCall.Arguments
		err = json.Unmarshal([]byte(args), &Arguments)
		if err != nil {
			return nil, 0, "", err
		}

		return nil, Arguments.DocumentID, Arguments.Query, nil
	} else {
		return &m.Content, 0, "", nil
	}
}

// Continue generates a continuation to a conversation. It accepts a document as
// context as well as a list of chunks of text relevant for the document, and
// the conversation history. It outputs a response and an error.
func (g *Generator) Continue(ctx context.Context, company *models.Company, conversation, lastMessage string, document *models.Document, chunks []string) (string, error) {
	var bigChunk string
	for j, chunk := range chunks {
		bigChunk += fmt.Sprintf("Paragraph %v: %v\n", j+1, chunk)
	}

	input := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: fmt.Sprintf("You are COFIN AI, a virtual assistant that helps people read, analyze, and interpret financial filings of publicly traded companies. You have access to 10-K and 10-Q documents filed to SEC. Today is %v.", time.Now().Format("2006-01-02")),
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf("I am going to send a conversation history between you and a user as a single message. The conversation pertains to company %v ($%v).", company.Name, company.Ticker),
		},
		schema.HumanChatMessage{
			Text: fmt.Sprintf("I am going to provide you with paragraphs from the %v document for %v filed at %v. You have previously chosen these as most relevant to the conversation I am going to provide you with. You should generate a response to the last user message using this document context as the source of data.", document.Kind, company.Name, document.FiledAt.Format("2006-01-02")),
		},
		schema.HumanChatMessage{Text: fmt.Sprintf("Here are the paragraphs from the %v: %v", document.Kind, bigChunk)},
		schema.AIChatMessage{Text: "Got it. Now please send me the conversation with the user."},
		schema.HumanChatMessage{Text: fmt.Sprintf("Here is the conversation:\n%v", conversation)},
		schema.HumanChatMessage{Text: fmt.Sprintf("Here is the last message from the user:\n%v", lastMessage)},
		schema.HumanChatMessage{Text: "Now generate a response using the conversation I sent you and the paragraphs from the document you've chosen. Do not mention anything about the instructions I gave you. Speak to the user directly, as if you were continuing the conversation with the user."},
	}

	res, err := g.Chat.Call(ctx, input, llms.WithTemperature(g.temperature))
	if err != nil {
		return "", err
	}

	return res, nil
}

// Format conversation history as a single string.
func mergeMessages(messages []models.Message) (conversation string) {
	for _, message := range messages {
		if message.Author == models.UserAuthor {
			conversation += fmt.Sprintf("User: %v\n", message.Text)
		} else if message.Author == models.AIAuthor {
			conversation += fmt.Sprintf("COFIN AI: %v\n", message.Text)
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

// jsonEscapeString escapes a string as a JSON string. For instance, it converts
// newline characters to "\n". Start and end quotation marks are removed.
func jsonEscapeString(i string) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s[1 : len(s)-1]
}

// jsonEscapeArray escapes an array of integers as a JSON string. For instance,
// it represents [1, 2, 3] as [1,2,3]. Start and end quotation marks are removed.
func jsonEscapeArray[K constraints.Integer](i []K) string {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)
	return s
}
