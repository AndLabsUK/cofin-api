package stripe_api

import (
	"cofin/models"
	"errors"
	"os"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/checkout/session"
	"github.com/stripe/stripe-go/v74/customer"
	"github.com/stripe/stripe-go/v74/price"
)

type StripeAPI struct {
	apiKey    string
	productID string
	domain    string
}

func NewStripeAPI() StripeAPI {
	return StripeAPI{
		apiKey:    os.Getenv("STRIPE_API_KEY"),
		domain:    os.Getenv("UI_DOMAIN"),
		productID: os.Getenv("STRIPE_PRODUCT_ID"),
	}
}

func (s StripeAPI) CreateCustomer(email, fullName string) (stripeCustomerID string, err error) {
	stripe.Key = s.apiKey
	params := &stripe.CustomerParams{
		Name:  &fullName,
		Email: &email,
	}

	stripeCustomer, err := customer.New(params)
	if err != nil {
		return "", err
	}

	return stripeCustomer.ID, nil
}

func (s StripeAPI) GetPrices() []*stripe.Price {
	stripe.Key = s.apiKey
	productID := s.productID
	params := &stripe.PriceListParams{
		Product: &productID,
	}

	var stripePrices []*stripe.Price
	i := price.List(params)
	for i.Next() {
		stripePrice := i.Price()
		stripePrices = append(stripePrices, stripePrice)
	}

	return stripePrices
}

func (s StripeAPI) CreateCheckout(user *models.User, productID string) (*string, error) {
	stripe.Key = s.apiKey
	stripePrice, err := price.Get(productID, nil)
	if err != nil {
		return nil, err
	}

	if stripePrice.Product.ID != s.productID {
		return nil, errors.New("specified price does not belong to specified plan")
	}

	params := &stripe.CheckoutSessionParams{
		Customer: &user.StripeCustomerID,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(productID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("subscription"),
		CancelURL:  stripe.String("https://" + s.domain + "/?payment=cancelled"),
		SuccessURL: stripe.String("https://" + s.domain + "/?payment=success"),
		ConsentCollection: &stripe.CheckoutSessionConsentCollectionParams{
			TermsOfService: stripe.String("required"),
		},
	}

	stripeCheckoutSession, err := session.New(params)
	if err != nil {
		return nil, err
	}

	return &stripeCheckoutSession.URL, nil
}
