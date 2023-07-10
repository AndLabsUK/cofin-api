package main

import (
	"cofin/api/middleware"
	"cofin/controllers"

	"github.com/gin-gonic/gin"
)

type Router struct {
	healthController       *controllers.HealthController
	authController         *controllers.AuthController
	usersController        *controllers.UsersController
	companiesController    *controllers.CompaniesController
	paymentsController     *controllers.PaymentsController
	conversationController *controllers.ConversationController
}

func (r Router) RegisterRoutes(router gin.IRouter) {

	//
	// Anonymous requests
	//
	router.GET("/health", r.healthController.Status)
	router.POST("/conversation", r.conversationController.Respond)

	router.GET("/companies", r.companiesController.GetCompanies)
	router.POST("/auth", r.authController.SignIn)

	//
	// Authorized Requests
	//
	authorized := router.Group("/", middleware.RequireAuth)

	authorized.GET("/users/me", r.usersController.GetCurrentUser)

	authorized.GET("/payments/prices", r.paymentsController.GetPrices)
	authorized.POST("/payments/checkout", r.paymentsController.Checkout)
}
