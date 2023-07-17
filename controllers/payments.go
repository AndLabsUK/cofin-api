package controllers

import (
	"cofin/internal/stripe_api"
	"cofin/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v74"
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

func (pc PaymentsController) PostCheckout(c *gin.Context) {
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

func (pc PaymentsController) PostEvent(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		pc.Logger.Errorf("Could not read Stripe webhook request body: %v", err)
		RespondBadRequestErr(c, []error{ErrBadInput})
		return
	}

	event, err := pc.StripeAPI.ConstructEvent(body, c.Request.Header.Get("Stripe-Signature"))
	if err != nil {
		pc.Logger.Errorf("Could not construct Stripe webhook event: %v", err)
		RespondBadRequestErr(c, []error{ErrBadInput})
		return
	}

	switch event.Type {
	case "customer.subscription.created":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			pc.Logger.Errorf("Could not unmarshal Stripe subscription: %v", err)
			RespondBadRequestErr(c, []error{ErrBadInput})
			return
		}

		if err := models.SetUserSubscriptionByStripeCustomerID(pc.DB, subscription.Customer.ID, true); err != nil {
			pc.Logger.Errorf("Could not set user subscription: %v", err)
			RespondInternalErr(c)
			return
		}
	case "customer.subscription.deleted":
		var subscription stripe.Subscription
		err := json.Unmarshal(event.Data.Raw, &subscription)
		if err != nil {
			pc.Logger.Errorf("Could not unmarshal Stripe subscription: %v", err)
			RespondBadRequestErr(c, []error{ErrBadInput})
			return
		}

		if err := models.SetUserSubscriptionByStripeCustomerID(pc.DB, subscription.Customer.ID, false); err != nil {
			pc.Logger.Errorf("Could not set user subscription: %v", err)
			RespondInternalErr(c)
			return
		}
	default:
		fmt.Fprintf(os.Stderr, "Unhandled event type: %s\n", event.Type)
	}

	RespondOK(c, nil)
}

func (pc PaymentsController) PostBillingPortal(c *gin.Context) {
	user := CurrentUser(c)
	portalURL, err := pc.StripeAPI.CreatePortal(user)
	if err != nil {
		pc.Logger.Errorf("Could not create Stripe billing portal: %v", err)
		RespondInternalErr(c)
		return
	}

	RespondOK(c, *portalURL)
}
