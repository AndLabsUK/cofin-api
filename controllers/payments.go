package controllers

import (
	"cofin/api"
	"cofin/api/middleware"
	"cofin/integrations"
	"github.com/gin-gonic/gin"
)

type PaymentsController struct{}

func (p PaymentsController) GetPrices(c *gin.Context) {
	stripe := integrations.Stripe{}
	prices := stripe.GetPrices()
	api.ResultData(c, prices)
}

func (p PaymentsController) Checkout(c *gin.Context) {
	type checkoutParams struct {
		StripePriceId string `json:"stripe_price_id"`
	}

	var payload checkoutParams

	if err := c.BindJSON(&payload); err != nil {
		api.ResultError(c, []string{"invalid_request"})
		return
	}

	user := middleware.CurrentUser(c)

	stripe := integrations.Stripe{}
	checkoutUrl, err := stripe.CreateCheckout(user, &payload.StripePriceId)
	if err != nil {
		api.ResultError(c, []string{"invalid_request"})
		return
	}

	api.ResultData(c, checkoutUrl)
}
