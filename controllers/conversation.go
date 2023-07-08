package controllers

import (
	"cofin/cmd/api"
	"cofin/internal"
	"cofin/models"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TODO: log internal errors but don't expose them to the user

type MessageKind string

const (
	// MessageKindInput is a message from the user.
	MessageKindInput MessageKind = "user"
	// MessageKindOutput is a message from the Artificial Intelligence.
	MessageKindOutput MessageKind = "ai"
)

// Message describes input or output of a conversation.
type Message struct {
	MessageKind MessageKind `json:"message_kind" binding:"required"`
	// Text input from the user.
	Text string `json:"text" binding:"required"`
	// Ticker is the copmany ticker. It is the "namespace" of the conversation.
	Ticker  string   `json:"ticker" binding:"required"`
	Sources []Source `json:"sources,omitempty"`
}

type Source struct {
	ID        uint              `json:"id" binding:"required"`
	Kind      models.SourceKind `json:"kind" binding:"required"`
	FiledAt   time.Time         `json:"filed_at" binding:"required"`
	OriginURL string            `json:"origin_url" binding:"required"`
}

type ConversationController struct {
	DB        *gorm.DB
	Generator *internal.Generator
}

// TODO: clean up response codes, user vs internal errors, logging, naming.
func (convo ConversationController) Respond(c *gin.Context) {
	message := Message{}
	err := c.BindJSON(&message)
	if err != nil {
		log.Println(err)
		api.ResultError(c, []string{err.Error()})
		return
	}

	retriever, err := internal.NewRetriever(convo.DB, strings.ToUpper(message.Ticker))
	if err != nil {
		log.Println(err)
		api.ResultError(c, nil)
		return
	}

	company, documents, err := retriever.GetDocuments(c.Request.Context(), message.Ticker)
	if err != nil {
		log.Println(err)
		api.ResultError(c, nil)
		return
	}

	if company == nil {
		api.ResultError(c, []string{errors.New("Unknown ticker").Error()})
		return
	}

	if documents == nil {
		api.ResultError(c, []string{errors.New("No documents found for the ticker").Error()})
		return
	}

	var allChunks = make([][]string, 0, len(documents))
	var sources = make([]Source, 0, len(documents))
	for _, document := range documents {
		chunks, err := retriever.GetSemanticChunks(c.Request.Context(), message.Ticker, document.UUID, message.Text)
		if err != nil {
			log.Println(err)
			api.ResultError(c, nil)
			return
		}

		allChunks = append(allChunks, chunks)
		sources = append(sources, Source{
			ID:        document.ID,
			Kind:      document.Kind,
			FiledAt:   document.FiledAt,
			OriginURL: document.OriginURL,
		})
	}

	response, err := convo.Generator.Continue(c.Request.Context(), *company, documents, allChunks, message.Text)
	if err != nil {
		log.Println(err)
		api.ResultError(c, nil)
		return
	}

	api.ResultData(c, Message{Ticker: message.Ticker, MessageKind: MessageKindOutput, Text: response, Sources: sources})
}
