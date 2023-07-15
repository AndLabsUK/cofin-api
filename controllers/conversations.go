package controllers

import (
	"cofin/internal/retrieval"
	"cofin/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Conversation struct {
	CompanyID uint             `json:"company_id" binding:"required"`
	Messages  []models.Message `json:"messages" binding:"required"`
}

type ConversationsController struct {
	DB        *gorm.DB
	Logger    *zap.SugaredLogger
	Generator *retrieval.Generator
}

const MAX_MESSAGES_UNPAID = 3

func (cc ConversationsController) PostConversation(c *gin.Context) {
	user := CurrentUser(c)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	if !user.IsSubscribed {
		messageCount, err := models.CountUserMessages(cc.DB, user.ID)
		if err != nil {
			cc.Logger.Errorf("Error getting messages: %w", err)
			RespondInternalErr(c)
			return
		}

		if messageCount >= MAX_MESSAGES_UNPAID {
			RespondCustomStatusErr(c, http.StatusPaymentRequired, []error{ErrUnpaidUser})
			return
		}
	}

	message := models.Message{}
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
		RespondCustomStatusErr(c, http.StatusNotFound, []error{ErrUnknownCompany})
		return
	}

	messageHistory, err := models.GetMessagesForCompanyChronological(cc.DB, user.ID, company.ID, 0, 15)
	if err != nil {
		cc.Logger.Errorf("Error getting messages: %w", err)
		RespondInternalErr(c)
	}

	ticker := company.Ticker
	retriever, err := retrieval.NewRetriever(cc.DB, company.ID)
	if err != nil {
		cc.Logger.Errorf("Error creating retriever: %w", err)
		RespondInternalErr(c)
		return
	}

	documents, err := models.GetRecentCompanyDocuments(cc.DB, company.ID, 10)
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

	cc.Logger.Infow(fmt.Sprintf("Answering user message: %v", message.Text), "userID", user.ID, "companyID", company.ID)

	conversation, err := cc.Generator.CondenseConversation(c.Request.Context(), company, append(messageHistory, message))
	if err != nil {
		cc.Logger.Errorf("Error condensing conversation: %w", err)
		RespondInternalErr(c)
	}

	cc.Logger.Infow(fmt.Sprintf("Condensed the conversation to: %v", conversation), "userID", user.ID, "companyID", company.ID)

	documentID, query, err := cc.Generator.CreateRetrieval(c.Request.Context(), company, documents, conversation)
	if err != nil {
		cc.Logger.Errorf("Error creating retrieval: %w", err)
		RespondInternalErr(c)
		return
	}

	cc.Logger.Infow(fmt.Sprintf("Created retrieval for document %v with query %v", documentID, query), "userID", user.ID, "companyID", company.ID)

	document, err := models.GetDocumentByID(cc.DB, documentID)
	if err != nil {
		cc.Logger.Errorf("Error getting document: %w", err)
		RespondInternalErr(c)
		return
	}

	var sources = make([]models.Source, 0, len(documents))
	chunks, err := retriever.GetSemanticChunks(c.Request.Context(), company.ID, documentID, query)
	if err != nil {
		cc.Logger.Errorf("Error getting semantic chunks for namespace %v document %v: %w", company.ID, documentID, err)
		RespondInternalErr(c)
		return
	}

	cc.Logger.Infow(fmt.Sprintf("Got (%v) semantic chunks for document %v", chunks, documentID), "userID", user.ID, "companyID", company.ID)

	sources = append(sources, models.Source{
		ID:        document.ID,
		Kind:      document.Kind,
		FiledAt:   document.FiledAt,
		OriginURL: document.OriginURL,
	})

	response, err := cc.Generator.Continue(c.Request.Context(), company, conversation, document, chunks)
	if err != nil {
		cc.Logger.Errorf("Error generating AI response: %w", err)
		RespondInternalErr(c)
		return
	}

	cc.Logger.Infow(fmt.Sprintf("Generated response: %v", response), "userID", user.ID, "companyID", company.ID)

	var aiMessage *models.Message
	if err := cc.DB.Transaction(func(tx *gorm.DB) error {
		_, err := models.CreateUserMessage(tx, user.ID, company.ID, message.Text)
		if err != nil {
			return err
		}

		aiMessage, err = models.CreateAIMessage(tx, user.ID, company.ID, response, sources)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		cc.Logger.Errorf("Error creating messages: %w", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, aiMessage)
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

	messages, err := models.GetMessagesForCompanyInverseChronological(cc.DB, CurrentUserId(c), uint(companyID), offset, limit)
	if err != nil {
		cc.Logger.Errorf("Error getting messages: %w", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, Conversation{
		CompanyID: uint(companyID),
		Messages:  messages,
	})
}
