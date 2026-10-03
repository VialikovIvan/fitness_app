package mystripe

import (
	"errors"
	"github.com/stripe/stripe-go/v86"
)

const (
	relativeCancelPath  string = "/plans"
	relativeSuccessPath string = "/checkout/success"
)

var (
	ErrNilDefaultPrice        = errors.New("stripe returned nil default price")
	ErrNilInvoiceSubscription = errors.New("stripe returned nil parent subscription")
)

type StripeManager struct {
	sc            *stripe.Client
	baseURL       string
	webhookSecret string
}

func New(sc *stripe.Client, baseURL string, secret string) *StripeManager {
	return &StripeManager{
		sc:            sc,
		baseURL:       baseURL,
		webhookSecret: secret,
	}
}
