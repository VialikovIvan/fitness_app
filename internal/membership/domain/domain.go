package domain

import (
	"errors"
	"time"

	pmtDomain "github.com/VialikovIvan/fitness_app/internal/payment/domain"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidPrice                   = errors.New("invalid price")
	ErrInvalidCurrency                = errors.New("invalid currency")
	ErrUserAlreadyHasActiveMembership = errors.New("user already has an active membership")
	ErrPlanDoesNotExist               = errors.New("plan does not exist")
	ErrInvalidWebhook                 = errors.New("invalid webhook")
)

type Plan struct {
	ID              int64
	StripeProductID string
	Name            string
	DurationDays    int
	Price           decimal.Decimal
	StripePriceID   string
	Currency        string
}

type Checkout struct {
	UserID           int64
	StripeCustomerID string
	PlanID           int64
	StripePriceID    string
	URL              string
}

type WebhookEvent struct {
	ID           string
	Type         EventType
	UserID       int64
	PlanID       int64
	CustomerID   string
	Subscription Subscription
	Payment      pmtDomain.Payment
}

type Subscription struct {
	ID                 int64
	StripeID           string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         time.Time
}

type EventType string

const (
	EventCheckoutCompleted           EventType = "checkout.session.completed"
	EventCustomerSubscriptionDeleted EventType = "customer.subscription.deleted"
	EventInvoicePaid                 EventType = "invoice.paid"
	EventInvoicePaymentFailed        EventType = "invoice.payment_failed"
	EventToSkip                      EventType = "event chosen to skip"
)

func IsValidPrice(price decimal.Decimal) bool {
	return price.GreaterThan(decimal.Zero) &&
		price.Exponent() == -2 &&
		price.LessThan(decimal.NewFromInt(100000000))
}

func IsValidCurrency(currency string) bool {
	return currency == "USD" || currency == "EUR"
}
