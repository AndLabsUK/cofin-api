package controllers

import (
	"cofin/internal/retrieval"
	"cofin/models"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Conversation struct {
	CompanyID uint      `json:"company_id" binding:"required"`
	Messages  []Message `json:"messages" binding:"required"`
}

// Message describes input or output of a conversation.
type Message struct {
	Author models.MessageAuthor `json:"author" binding:"required"`
	// Text input from the user.
	Text      string          `json:"text" binding:"required"`
	Sources   []models.Source `json:"sources,omitempty"`
	CreatedAt time.Time       `json:"created_at,omitempty"`
}

type ConversationsController struct {
	DB        *gorm.DB
	Logger    *zap.SugaredLogger
	Generator *retrieval.Generator
}

func (cc ConversationsController) PostConversation(c *gin.Context) {
	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	message := Message{}
	err = c.BindJSON(&message)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	company, err := models.GetCompanyByID(cc.DB, uint(companyID))
	if err != nil {
		cc.Logger.Errorf("Error getting company: %w", err)
		RespondInternalErr(c)
		return
	}

	if company == nil {
		RespondBadRequestErr(c, []error{ErrUnknownCompany})
		return
	}

	// TODO: use company ID as the namespace in Pinecone.
	ticker := company.Ticker
	retriever, err := retrieval.NewRetriever(cc.DB, ticker)
	if err != nil {
		cc.Logger.Errorf("Error creating retriever: %w", err)
		RespondInternalErr(c)
		return
	}

	documents, err := retriever.GetDocuments(company.ID)
	if err != nil {
		cc.Logger.Errorf("Error getting documents: %w", err)
		RespondInternalErr(c)
		return
	}

	if documents == nil {
		cc.Logger.Errorf("No documents found for ticker %v", ticker)
		RespondInternalErr(c)
		return
	}

	var allChunks = make([][]string, 0, len(documents))
	var sources = make([]models.Source, 0, len(documents))
	for _, document := range documents {
		chunks, err := retriever.GetSemanticChunks(c.Request.Context(), ticker, document.UUID, message.Text)
		if err != nil {
			cc.Logger.Errorf("Error getting semantic chunks for ticker %v document %v: %w", ticker, document.ID, err)
			RespondInternalErr(c)
			return
		}

		allChunks = append(allChunks, chunks)
		sources = append(sources, models.Source{
			ID:        document.ID,
			Kind:      document.Kind,
			FiledAt:   document.FiledAt,
			OriginURL: document.OriginURL,
		})
	}

	response, err := cc.Generator.Continue(c.Request.Context(), *company, documents, allChunks, message.Text)
	if err != nil {
		cc.Logger.Errorf("Error generating AI response: %w", err)
		RespondInternalErr(c)
		return
	}

	user := CurrentUser(c)
	if err := cc.DB.Transaction(func(tx *gorm.DB) error {
		_, err := models.CreateUserMessage(tx, user.ID, company.ID, message.Text)
		if err != nil {
			return err
		}

		_, err = models.CreateAIMessage(tx, user.ID, company.ID, response, sources)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		cc.Logger.Errorf("Error creating messages: %w", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, Message{Author: models.AIAuthor, Text: response, Sources: sources})
}

func (cc ConversationsController) GetConversation(c *gin.Context) {
	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	messages, err := models.GetMessagesForCompany(cc.DB, uint(companyID), offset, limit)
	if err != nil {
		cc.Logger.Errorf("Error getting messages: %w", err)
		RespondInternalErr(c)
		return
	}

	retrievedMessages := make([]Message, 0, len(messages))
	for _, message := range messages {
		annotation := models.Annotation{}
		err := json.Unmarshal(message.Annotation, &annotation)
		if err != nil {
			cc.Logger.Errorf("Error unmarshalling annotation: %w", err)
			RespondInternalErr(c)
			return
		}

		retrievedMessages = append(retrievedMessages, Message{
			Author:    message.Author,
			Text:      message.Text,
			Sources:   annotation.Sources,
			CreatedAt: message.CreatedAt,
		})
	}

	RespondOK(c, Conversation{
		CompanyID: uint(companyID),
		Messages:  retrievedMessages,
	})
}
