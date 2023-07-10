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
	router.POST("/conversations", r.ConversationsController.Respond)

	router.GET("/companies", r.CompaniesController.GetCompanies)
	router.POST("/auth", r.AuthController.SignIn)

	//
	// Authorized Requests
	//
	authorized := router.Group("/", RequireAuth)
	authorized.GET("/users/me", r.UsersController.GetCurrentUser)

	authorized.GET("/payments/prices", r.PaymentsController.GetPrices)
	authorized.POST("/payments/checkout", r.PaymentsController.Checkout)
}
