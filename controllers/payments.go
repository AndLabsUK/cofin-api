package controllers

import (
	"cofin/internal/stripe_api"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentsController struct {
	DB        *gorm.DB
	Logger    *zap.SugaredLogger
	StripeAPI stripe_api.StripeAPI
}

func (pc PaymentsController) GetPrices(c *gin.Context) {
	prices := pc.StripeAPI.GetPrices()
	RespondOK(c, prices)
}

func (pc PaymentsController) Checkout(c *gin.Context) {
	type checkoutParams struct {
		StripePriceID string `json:"stripe_price_id"`
	}

	var payload checkoutParams

	if err := c.BindJSON(&payload); err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	user := CurrentUser(c)
	checkoutUrl, err := pc.StripeAPI.CreateCheckout(user, payload.StripePriceID)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	RespondOK(c, checkoutUrl)
}
