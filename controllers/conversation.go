package controllers

import (
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

// UserMessage describes user input to converse with the AI.
type UserMessage struct {
	// Text input from the user.
	Text string `json:"text" binding:"required"`
	// Ticker to which the user's message pertains.
	Ticker string `json:"ticker" binding:"required"`
}

type Source struct {
	ID        uint              `json:"id" binding:"required"`
	Kind      models.SourceKind `json:"kind" binding:"required"`
	FiledAt   time.Time         `json:"filed_at" binding:"required"`
	OriginURL string            `json:"origin_url" binding:"required"`
}

// AIMessage describes AI response to the user.
type AIMessage struct {
	// Text response from the AI.
	Text    string   `json:"text" binding:"required"`
	Sources []Source `json:"sources" binding:"required"`
}

// Exchange describes a series of messages exchanged between the user and the
// AI.
type Exchange struct {
	UserMessage UserMessage `json:"user_message" binding:"required"`
	AIMessage   AIMessage   `json:"ai_message" binding:"required"`
}

// ConversationInput describes an accumulated user-AI conversation.
type ConversationInput struct {
	// Latest user message that needs to be responded to.
	UserMessage UserMessage `json:"user_message" binding:"required"`
}

type ConversationController struct {
	DB        *gorm.DB
	Generator *internal.Generator
}

// TODO: clean up response codes, user vs internal errors, logging, naming.
func (convo ConversationController) Respond(c *gin.Context) {
	input := ConversationInput{}
	err := c.BindJSON(&input)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	informationRetriever, err := internal.NewInformationRetriever(convo.DB, strings.ToUpper(input.UserMessage.Ticker))
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	company, documents, err := informationRetriever.GetDocuments(c.Request.Context(), input.UserMessage.Ticker)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if company == nil {
		c.JSON(400, gin.H{"error": errors.New("Unknown ticker").Error()})
		return
	}

	if documents == nil {
		c.JSON(400, gin.H{"error": errors.New("No documents found for the ticker").Error()})
		return
	}

	var allChunks = make([][]string, 0, len(documents))
	var sources = make([]Source, 0, len(documents))
	for _, document := range documents {
		chunks, err := informationRetriever.GetSemanticChunks(c.Request.Context(), input.UserMessage.Ticker, document.UUID, input.UserMessage.Text)
		if err != nil {
			log.Println(err)
			c.JSON(500, gin.H{"error": err.Error()})
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

	response, err := convo.Generator.Continue(c.Request.Context(), *company, documents, allChunks, input.UserMessage.Text)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, AIMessage{Text: response, Sources: sources})
}
