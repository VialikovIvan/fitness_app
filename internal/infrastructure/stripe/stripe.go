package mystripe

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
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

func (m *StripeManager) CreateCheckoutSession(ctx context.Context, p domain.Checkout) (string, error) {
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

func (m *StripeManager) VerifyWebhook(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
	event, err := m.sc.ConstructEvent(body, sig, m.webhookSecret)
	if err != nil {
		return domain.WebhookEvent{}, err
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession

		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return domain.WebhookEvent{}, err
		}

		userID, err := strconv.ParseInt(session.ClientReferenceID, 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(session.Metadata["plan_id"], 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		sub, err := m.GetSubscriptionByID(ctx, session.Subscription.ID)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		return domain.WebhookEvent{
			ID:           event.ID,
			Type:         domain.EventCheckoutCompleted,
			UserID:       userID,
			PlanID:       planID,
			CustomerID:   session.Customer.ID,
			Subscription: sub,
		}, nil

	case "customer.subscription.deleted":
		var subscription stripe.Subscription

		if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
			return domain.WebhookEvent{}, err
		}

		userID, err := strconv.ParseInt(subscription.Metadata["user_id"], 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(subscription.Metadata["plan_id"], 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		return domain.WebhookEvent{
			ID:     event.ID,
			Type:   domain.EventCustomerSubscriptionDeleted,
			UserID: userID,
			PlanID: planID,
			Subscription: domain.Subscription{
				StripeID:          subscription.ID,
				Status:            string(subscription.Status),
				CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
				CanceledAt:        time.Unix(subscription.CanceledAt, 0),
			},
		}, nil

	case "invoice.paid":
		var invoice stripe.Invoice

		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return domain.WebhookEvent{}, err
		}

		if invoice.BillingReason == "subscription_create" {
			return domain.WebhookEvent{
				ID:   event.ID,
				Type: domain.EventToSkip,
			}, nil
		}

		userID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["user_id"], 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["plan_id"], 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		sub, err := m.GetSubscriptionByID(ctx, invoice.Parent.SubscriptionDetails.Subscription.ID)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		if invoice.Parent.SubscriptionDetails.Subscription == nil || invoice.Parent.SubscriptionDetails.Subscription.ID == "" {
			return domain.WebhookEvent{}, ErrNilInvoiceSubscription
		}

		return domain.WebhookEvent{
			ID:           event.ID,
			Type:         domain.EventInvoicePaid,
			UserID:       userID,
			PlanID:       planID,
			Subscription: sub,
		}, nil

	case "invoice.payment_failed":
		var invoice stripe.Invoice

		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			return domain.WebhookEvent{}, err
		}

		if invoice.BillingReason == "subscription_create" {
			return domain.WebhookEvent{
				ID:   event.ID,
				Type: domain.EventToSkip,
			}, nil
		}

		userID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["user_id"], 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		planID, err := strconv.ParseInt(invoice.Parent.SubscriptionDetails.Metadata["plan_id"], 10, 64)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		sub, err := m.GetSubscriptionByID(ctx, invoice.Parent.SubscriptionDetails.Subscription.ID)
		if err != nil {
			return domain.WebhookEvent{}, err
		}

		if invoice.Parent.SubscriptionDetails.Subscription == nil || invoice.Parent.SubscriptionDetails.Subscription.ID == "" {
			return domain.WebhookEvent{}, ErrNilInvoiceSubscription
		}

		return domain.WebhookEvent{
			ID:           event.ID,
			Type:         domain.EventInvoicePaymentFailed,
			UserID:       userID,
			PlanID:       planID,
			Subscription: sub,
		}, nil

	default:
		return domain.WebhookEvent{
			ID:   event.ID,
			Type: domain.EventType(event.Type),
		}, nil
	}
}
