package controllers

import (
	"cofin/internal/retrieval"
	"cofin/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TODO: log internal errors but don't expose them to the user

// Message describes input or output of a conversation.
type Message struct {
	Author models.MessageAuthor `json:"author" binding:"required"`
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

type ConversationsController struct {
	DB        *gorm.DB
	Logger    *zap.SugaredLogger
	Generator *retrieval.Generator
}

// TODO: clean up response codes, user vs internal errors, logging, naming.
func (cc ConversationsController) Respond(c *gin.Context) {
	message := Message{}
	err := c.BindJSON(&message)
	if err != nil {
		cc.Logger.Errorf("Error querying companies: %w", err)
		WriteBadRequestError(c, []error{err})
		return
	}

	retriever, err := retrieval.NewRetriever(cc.DB, strings.ToUpper(message.Ticker))
	if err != nil {
		cc.Logger.Errorf("Error creating retriever: %w", err)
		WriteInternalError(c)
		return
	}

	company, documents, err := retriever.GetDocuments(message.Ticker)
	if err != nil {
		cc.Logger.Errorf("Error getting documents: %w", err)
		WriteInternalError(c)
		return
	}

	if company == nil {
		WriteBadRequestError(c, []error{ErrUnknownTicker})
		return
	}

	if documents == nil {
		cc.Logger.Error("No documents found")
		WriteInternalError(c)
		return
	}

	var allChunks = make([][]string, 0, len(documents))
	var sources = make([]Source, 0, len(documents))
	for _, document := range documents {
		chunks, err := retriever.GetSemanticChunks(c.Request.Context(), message.Ticker, document.UUID, message.Text)
		if err != nil {
			cc.Logger.Errorf("Error getting semantic chunks: %w", err)
			WriteInternalError(c)
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

	response, err := cc.Generator.Continue(c.Request.Context(), *company, documents, allChunks, message.Text)
	if err != nil {
		cc.Logger.Errorf("Error generating AI response: %w", err)
		WriteInternalError(c)
		return
	}

	WriteSuccess(c, Message{Ticker: message.Ticker, Author: models.AIAuthor, Text: response, Sources: sources})
}
