package controllers

import (
	"cofin/internal/stripe_api"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PaymentsController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (p PaymentsController) GetPrices(c *gin.Context) {
	stripe := stripe_api.StripeAPI{}
	prices := stripe.GetPrices()
	RespondOK(c, prices)
}

func (p PaymentsController) Checkout(c *gin.Context) {
	type checkoutParams struct {
		StripePriceId string `json:"stripe_price_id"`
	}

	var payload checkoutParams

	if err := c.BindJSON(&payload); err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	user := CurrentUser(c)
	stripe := stripe_api.StripeAPI{}
	checkoutUrl, err := stripe.CreateCheckout(user, &payload.StripePriceId)
	if err != nil {
		RespondBadRequestErr(c, []error{err})
		return
	}

	RespondOK(c, checkoutUrl)
}
