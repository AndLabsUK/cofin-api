package controllers

import (
	"github.com/gin-gonic/gin"
)

type Router struct {
	HealthController        *HealthController
	AuthController          *AuthController
	UsersController         *UsersController
	CompaniesController     *CompaniesController
	PaymentsController      *PaymentsController
	ConversationsController *ConversationsController
}

func (r Router) RegisterRoutes(router gin.IRouter) {
	//
	// Anonymous requests
	//
	router.GET("/health", r.HealthController.Status)
	router.GET("/companies", r.CompaniesController.GetCompanies)
	router.GET("/companies/:company_id", r.CompaniesController.GetCompany)
	router.GET("/companies/:company_id/documents", r.CompaniesController.GetCompanyDocuments)
	router.POST("/auth", r.AuthController.SignIn)
	router.POST("/payments/webhook", r.PaymentsController.PostEvent)

	//
	// Authorized Requests
	//
	authorized := router.Group("/", RequireAuth)
	authorized.GET("/users/me", r.UsersController.GetCurrentUser)

	conversations := authorized.Group("/conversations")
	conversations.GET("/:company_id", r.ConversationsController.GetConversation)
	conversations.POST("/:company_id", r.ConversationsController.PostConversation)

	authorized.GET("/payments/prices", r.PaymentsController.GetPrices)
	authorized.POST("/payments/checkout", r.PaymentsController.PostCheckout)
	authorized.POST("/payments/portal", r.PaymentsController.PostBillingPortal)
}
