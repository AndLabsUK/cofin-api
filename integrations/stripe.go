package integrations

import (
	"cofin/models"
	"errors"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
	"gorm.io/gorm"
	"os"
)

type Stripe struct{}

func (s Stripe) CreateCustomer(user *models.User, db *gorm.DB) {
	stripe.Key = os.Getenv("STRIPE_API_KEY")

	if user.StripeCustomerId != "" {
		return
	}

	params := &stripe.CustomerParams{
		Name:  &user.FullName,
		Email: &user.Email,
	}

	c, err := customer.New(params)
	if err != nil {
		//log err
		return
	}

	user.StripeCustomerId = c.ID
	result := db.Save(&user)
	if result.Error != nil {
		//TODO: Log error
		return
	}
}

func (s Stripe) GetPrices() []*stripe.Price {
	stripe.Key = os.Getenv("STRIPE_API_KEY")

	productId := os.Getenv("STRIPE_PRODUCT_ID")

	params := &stripe.PriceListParams{
		Product: &productId,
	}

	var stripePrices []*stripe.Price
	i := price.List(params)
	for i.Next() {
		stripePrice := i.Price()
		stripePrices = append(stripePrices, stripePrice)
	}

	return stripePrices
}

func (s Stripe) CreateCheckout(user *models.User, priceId *string) (*string, error) {
	stripe.Key = os.Getenv("STRIPE_API_KEY")

	if len(user.StripeCustomerId) == 0 {
		return nil, errors.New("user is not registered on Stripe")
	}

	stripePrice, err := price.Get(*priceId, nil)
	if err != nil {
		return nil, err
	}

	if stripePrice.Product.ID != os.Getenv("STRIPE_PRODUCT_ID") {
		return nil, errors.New("specified price does not belong to specified plan")
	}

	params := &stripe.CheckoutSessionParams{
		Customer: &user.StripeCustomerId,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(*priceId),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"),
		CancelURL:  stripe.String("https://" + os.Getenv("UI_DOMAIN") + "/?payment=cancelled"),
		SuccessURL: stripe.String("https://" + os.Getenv("UI_DOMAIN") + "/?payment=success"),
	}

	stripeCheckoutSession, err := session.New(params)
	if err != nil {
		return nil, err
	}

	return &stripeCheckoutSession.URL, nil
}
