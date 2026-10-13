package mystripe

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	mbspDomain "github.com/VialikovIvan/fitness_app/internal/membership/domain"
	pmtDomain "github.com/VialikovIvan/fitness_app/internal/payment/domain"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v86"
)

const (
	PaymentProvider     string = "stripe"
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

func (m *StripeManager) CreateCheckoutSession(ctx context.Context, p mbspDomain.Checkout) (string, error) {
	params := &stripe.CheckoutSessionCreateParams{
		CancelURL:             stripe.String(m.baseURL + relativeCancelPath),
		ClientReferenceID:     stripe.String(strconv.FormatInt(p.UserID, 10)),
		IntegrationIdentifier: stripe.String("fitness_app_membership"),
		LineItems: []*stripe.CheckoutSessionCreateLineItemParams{
			{
				Price:    stripe.String(p.StripePriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Metadata: map[string]string{
			"user_id": strconv.FormatInt(p.UserID, 10),
			"plan_id": strconv.FormatInt(p.PlanID, 10),
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SubscriptionData: &stripe.CheckoutSessionCreateSubscriptionDataParams{
			Metadata: map[string]string{
				"user_id": strconv.FormatInt(p.UserID, 10),
				"plan_id": strconv.FormatInt(p.PlanID, 10),
			},
		},
		SuccessURL: stripe.String(m.baseURL + relativeSuccessPath),
	}

	if p.StripeCustomerID != "" {
		params.Customer = stripe.String(p.StripeCustomerID)
	}

	session, err := m.sc.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return "", err
	}

	return session.URL, nil
}

func (m *StripeManager) VerifyWebhook(ctx context.Context, body []byte, sig string) (mbspDomain.WebhookEvent, error) {
	event, err := m.sc.ConstructEvent(body, sig, m.webhookSecret)
	if err != nil {
		return mbspDomain.WebhookEvent{}, err
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession

		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		userID, err := strconv.ParseInt(session.ClientReferenceID, 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(session.Metadata["plan_id"], 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		sub, err := m.GetSubscriptionByID(ctx, session.Subscription.ID)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		return mbspDomain.WebhookEvent{
			ID:           event.ID,
			Type:         mbspDomain.EventCheckoutCompleted,
			UserID:       userID,
			PlanID:       planID,
			CustomerID:   session.Customer.ID,
			Subscription: sub,
			Payment: pmtDomain.Payment{
				UserID:            userID,
				Provider:          PaymentProvider,
				ProviderPaymentID: session.ID,
				Amount:            decimal.NewFromInt(session.AmountTotal).Shift(-2),
				Currency:          strings.ToUpper(string(session.Currency)),
				Status:            string(session.PaymentStatus),
				ProviderCreatedAt: time.Unix(session.Created, 0),
				PaidAt:            unixToTimePtr(event.Created),
			},
		}, nil

	case "customer.subscription.deleted":
		var subscription stripe.Subscription

		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		userID, err := strconv.ParseInt(subscription.Metadata["user_id"], 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(subscription.Metadata["plan_id"], 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		return mbspDomain.WebhookEvent{
			ID:     event.ID,
			Type:   mbspDomain.EventCustomerSubscriptionDeleted,
			UserID: userID,
			PlanID: planID,
			Subscription: mbspDomain.Subscription{
				StripeID:          subscription.ID,
				Status:            string(subscription.Status),
				CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
				CanceledAt:        time.Unix(subscription.CanceledAt, 0),
			},
		}, nil

	case "invoice.paid":
		var invoice stripe.Invoice

		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		if invoice.BillingReason == "subscription_create" {
			return mbspDomain.WebhookEvent{
				ID:   event.ID,
				Type: mbspDomain.EventToSkip,
			}, nil
		}

		userID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["user_id"], 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["plan_id"], 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		sub, err := m.GetSubscriptionByID(ctx, invoice.Parent.SubscriptionDetails.Subscription.ID)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		if invoice.Parent.SubscriptionDetails.Subscription == nil || invoice.Parent.SubscriptionDetails.Subscription.ID == "" {
			return mbspDomain.WebhookEvent{}, ErrNilInvoiceSubscription
		}

		return mbspDomain.WebhookEvent{
			ID:           event.ID,
			Type:         mbspDomain.EventInvoicePaid,
			UserID:       userID,
			PlanID:       planID,
			Subscription: sub,
			Payment: pmtDomain.Payment{
				UserID:            userID,
				Provider:          PaymentProvider,
				ProviderPaymentID: invoice.ID,
				Amount:            decimal.NewFromInt(invoice.AmountPaid).Shift(-2),
				Currency:          strings.ToUpper(string(invoice.Currency)),
				Status:            string(invoice.Status),
				ProviderCreatedAt: time.Unix(invoice.Created, 0),
				PaidAt:            unixToTimePtr(invoice.StatusTransitions.PaidAt),
			},
		}, nil

	case "invoice.payment_failed":
		var invoice stripe.Invoice

		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		if invoice.BillingReason == "subscription_create" {
			return mbspDomain.WebhookEvent{
				ID:   event.ID,
				Type: mbspDomain.EventToSkip,
			}, nil
		}

		userID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["user_id"], 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["plan_id"], 10, 64)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		sub, err := m.GetSubscriptionByID(ctx, invoice.Parent.SubscriptionDetails.Subscription.ID)
		if err != nil {
			return mbspDomain.WebhookEvent{}, err
		}

		if invoice.Parent.SubscriptionDetails.Subscription == nil || invoice.Parent.SubscriptionDetails.Subscription.ID == "" {
			return mbspDomain.WebhookEvent{}, ErrNilInvoiceSubscription
		}

		return mbspDomain.WebhookEvent{
			ID:           event.ID,
			Type:         mbspDomain.EventInvoicePaymentFailed,
			UserID:       userID,
			PlanID:       planID,
			Subscription: sub,
			Payment: pmtDomain.Payment{
				UserID:            userID,
				Provider:          PaymentProvider,
				ProviderPaymentID: invoice.ID,
				Amount:            decimal.NewFromInt(invoice.AmountPaid).Shift(-2),
				Currency:          strings.ToUpper(string(invoice.Currency)),
				Status:            string(invoice.Status),
				ProviderCreatedAt: time.Unix(invoice.Created, 0),
				PaidAt:            unixToTimePtr(invoice.StatusTransitions.PaidAt),
			},
		}, nil

	default:
		return mbspDomain.WebhookEvent{
			ID:   event.ID,
			Type: mbspDomain.EventType(event.Type),
		}, nil
	}
}

func unixToTimePtr(sec int64) *time.Time {
	if sec <= 0 {
		return nil
	}
	t := time.Unix(sec, 0)
	return &t
}
