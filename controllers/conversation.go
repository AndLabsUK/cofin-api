package controllers

import (
	"cofin/internal"
	"cofin/models"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// TODO: log internal errors but don't expose them to the user

// UserMessage describes user input to converse with the AI.
type UserMessage struct {
	// Text input from the user.
	Text string `json:"text" binding:"required"`
	// Ticker to which the user's message pertains.
	Ticker string `json:"ticker" binding:"required"`
	// Financial report year the user is interested in. By default, the most
	// recent year is chosen.
	Year int `json:"year" binding:"required"`
	// Financial report quarter the user is interested in. By default, the most
	// recent quarter is chosen.
	Quarter models.Quarter `json:"quarter"`
	// Document type the user is interested in. By default, 10-K is chosen, if
	// available. Otherwise, 10-Q is chosen.
	SourceKind models.SourceKind `json:"source_type"`
}

// AIMessage describes AI response to the user.
type AIMessage struct {
	// Text response from the AI.
	Text string `json:"text" binding:"required"`
	// TODO: add sources of information used in the response.
}

// Exchange describes a series of messages exchanged between the user and the
// AI.
type Exchange struct {
	UserMessage UserMessage `json:"user_message" binding:"required"`
	AIMessage   AIMessage   `json:"ai_message" binding:"required"`
}

// ConversationInput describes an accumulated user-AI conversation.
type ConversationInput struct {
	// Previous exchanges.
	Exchanges []Exchange `json:"exchanges" binding:"required"`
	// Latest user message that needs to be responded to.
	UserMessage UserMessage `json:"user_message" binding:"required"`
}

type ConversationController struct {
	Generator *internal.Generator
}

func (convo ConversationController) Respond(c *gin.Context) {
	input := ConversationInput{}
	err := c.BindJSON(&input)
	if err != nil {
		log.Println(err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	informationRetriever, err := internal.NewInformationRetriever(strings.ToUpper(input.UserMessage.Ticker))
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	docs, err := informationRetriever.Get(c.Request.Context(), input.UserMessage.Ticker, input.UserMessage.Year, input.UserMessage.Quarter, input.UserMessage.SourceKind, input.UserMessage.Text)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response, err := convo.Generator.Continue(docs, input.UserMessage.Text)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, AIMessage{Text: response})
}
