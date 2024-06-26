package controllers

import (
	"cofin/internal/amplitude"
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
	Amplitude amplitude.Amplitude
}

func (cc ConversationsController) PostConversation(c *gin.Context) {
	user := CurrentUser(c)

	companyID, err := strconv.ParseUint(c.Param("company_id"), 10, 32)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	if !user.IsSubscribed && user.RemainingMessageAllowance <= 0 {
		RespondCustomStatusErr(c, http.StatusPaymentRequired, []error{ErrUnpaidUser})
		return
	}

	userMessage := models.Message{}
	err = c.BindJSON(&userMessage)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	company, err := models.GetCompanyByID(cc.DB, uint(companyID))
	if err != nil {
		cc.Logger.Errorf("Error getting company: %w", err)
		RespondInternalErr(c)
		return
	} else if company == nil {
		RespondCustomStatusErr(c, http.StatusNotFound, []error{ErrUnknownCompany})
		return
	}

	messageHistory, err := models.GetMessagesForCompanyInverseChronological(cc.DB, user.ID, company.ID, 0, 6)
	if err != nil {
		cc.Logger.Errorf("Error getting messages: %w", err)
		RespondInternalErr(c)
	}
	messageHistory = reverseMessageArray(messageHistory)

	if _, err := models.CreateUserMessage(cc.DB, user.ID, company.ID, userMessage.Text); err != nil {
		cc.Logger.Errorf("Error saving user message: %w", err)
		RespondInternalErr(c)
		return
	}

	cc.Amplitude.TrackEvent(user.FirebaseSubjectID, "user_sent_message", map[string]interface{}{
		"company_ticker": company.Ticker,
	})

	cc.Logger.Infow(fmt.Sprintf("Answering user message: %v", userMessage.Text), "userID", user.ID, "companyID", company.ID)

	documents, err := models.GetCompanyDocumentsInverseChronological(cc.DB, company.ID, 0, 10)
	if err != nil {
		cc.Logger.Errorf("Error getting documents: %w", err)
		RespondInternalErr(c)
		return
	}

	if len(documents) == 0 {
		var earlyResponse = "Sorry, I'm afraid no recent documents are available for this company."
		cc.Logger.Infow(fmt.Sprintf("Early response \"%v\" for user messsage", earlyResponse), "userID", user.ID, "companyID", company.ID)
		aiMessage, err := models.CreateAIMessage(cc.DB, user.ID, company.ID, earlyResponse, []models.Source{})
		if err != nil {
			cc.Logger.Errorf("Error saving messages: %w", err)
			RespondInternalErr(c)
			return
		}

		RespondOK(c, aiMessage)
		return
	}

	var conversation string = mergeMessages(user, messageHistory)
	if len(messageHistory) != 0 {
		conversation, err = cc.Generator.CondenseConversation(c.Request.Context(), user, company, conversation, userMessage.Text)
		if err != nil {
			cc.Logger.Errorf("Error condensing conversation: %w", err)
			RespondInternalErr(c)
		}
		cc.Logger.Infow(fmt.Sprintf("Condensed the conversation to:\n%v", conversation), "userID", user.ID, "companyID", company.ID)
	}

	documentIDs, documentList := makeDocumentList(company, documents)
	earlyResponse, documentID, query, err := cc.Generator.CreateRetrieval(c.Request.Context(), user, company, documentIDs, documentList, conversation, userMessage.Text)
	if err != nil {
		cc.Logger.Errorf("Error creating retrieval: %w", err)
		RespondInternalErr(c)
		return
	}
	if earlyResponse != nil {
		cc.Logger.Infow(fmt.Sprintf("Early response \"%v\" for user messsage", *earlyResponse), "userID", user.ID, "companyID", company.ID)
		aiMessage, err := models.CreateAIMessage(cc.DB, user.ID, company.ID, *earlyResponse, []models.Source{})
		if err != nil {
			cc.Logger.Errorf("Error saving messages: %w", err)
			RespondInternalErr(c)
			return
		}

		RespondOK(c, aiMessage)
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
	retriever, err := retrieval.NewRetriever(cc.DB, company.ID)
	if err != nil {
		cc.Logger.Errorf("Error creating retriever: %w", err)
		RespondInternalErr(c)
		return
	}
	chunks, err := retriever.GetSemanticChunks(c.Request.Context(), company.ID, documentID, query)
	if err != nil {
		cc.Logger.Errorf("Error getting semantic chunks for namespace %v document %v: %w", company.ID, documentID, err)
		RespondInternalErr(c)
		return
	}

	sources = append(sources, models.Source{
		ID:        document.ID,
		Kind:      document.Kind,
		FiledAt:   document.FiledAt,
		OriginURL: document.OriginURL,
	})

	response, err := cc.Generator.Continue(c.Request.Context(), user, company, documentList, conversation, userMessage.Text, document, chunks)
	if err != nil {
		cc.Logger.Errorf("Error generating AI response: %w", err)
		RespondInternalErr(c)
		return
	}
	cc.Logger.Infow(fmt.Sprintf("Generated response: %v", response), "userID", user.ID, "companyID", company.ID)

	aiMessage, err := models.CreateAIMessage(cc.DB, user.ID, company.ID, response, sources)
	if err != nil {
		cc.Logger.Errorf("Error saving messages: %w", err)
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

	messages, err := models.GetMessagesForCompanyInverseChronological(cc.DB, CurrentUserID(c), uint(companyID), offset, limit)
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

func reverseMessageArray(a []models.Message) (b []models.Message) {
	for j := len(a) - 1; j >= 0; j-- {
		b = append(b, a[j])
	}

	return
}

// Format conversation history as a single string.
func mergeMessages(user *models.User, messages []models.Message) (conversation string) {
	for _, message := range messages {
		if message.Author == models.UserAuthor {
			conversation += fmt.Sprintf("%v: %v\n", user.FullName, message.Text)
		} else if message.Author == models.AIAuthor {
			conversation += fmt.Sprintf("COFIN: %v\n", message.Text)
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
