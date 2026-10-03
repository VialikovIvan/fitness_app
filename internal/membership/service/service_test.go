package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/VialikovIvan/fitness_app/internal/membership/service"
	"github.com/shopspring/decimal"
)

type MockPlanRepo struct {
	CreatePlanFunc               func(ctx context.Context, plan domain.Plan) error
	GetStripePriceIDByPlanIDFunc func(ctx context.Context, planID int64) (string, error)
}

func (m *MockPlanRepo) CreatePlan(ctx context.Context, plan domain.Plan) error {
	return m.CreatePlanFunc(ctx, plan)
}

func (m *MockPlanRepo) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	return nil, nil
}

func (m *MockPlanRepo) GetStripePriceIDByPlanID(ctx context.Context, planID int64) (string, error) {
	return m.GetStripePriceIDByPlanIDFunc(ctx, planID)
}

type MockSubscriptionRepo struct {
	CreateSubscriptionFunc            func(ctx context.Context, event domain.WebhookEvent) error
	ProlongSubscriptionFunc           func(ctx context.Context, event domain.WebhookEvent) error
	MarkSubscriptionPaymentFailedFunc func(ctx context.Context, event domain.WebhookEvent) error
	CancelSubscriptionFunc            func(ctx context.Context, event domain.WebhookEvent) error
}

func (m *MockSubscriptionRepo) CreateSubscription(ctx context.Context, event domain.WebhookEvent) error {
	return m.CreateSubscriptionFunc(ctx, event)
}

func (m *MockSubscriptionRepo) ProlongSubscription(ctx context.Context, event domain.WebhookEvent) error {
	return m.ProlongSubscriptionFunc(ctx, event)
}

func (m *MockSubscriptionRepo) MarkSubscriptionPaymentFailed(ctx context.Context, event domain.WebhookEvent) error {
	return m.MarkSubscriptionPaymentFailedFunc(ctx, event)
}

func (m *MockSubscriptionRepo) CancelSubscription(ctx context.Context, event domain.WebhookEvent) error {
	return m.CancelSubscriptionFunc(ctx, event)
}

type MockMembershipRepo struct {
	HasActiveMembershipByIDFunc func(ctx context.Context, userID int64) (bool, error)
}

func (m *MockMembershipRepo) HasActiveMembershipByID(ctx context.Context, userID int64) (bool, error) {
	return m.HasActiveMembershipByIDFunc(ctx, userID)
}

type MockUserRepo struct {
	GetStripeCustomerIDByIDFunc func(ctx context.Context, userID int64) (string, error)
}

func (m *MockUserRepo) GetStripeCustomerIDByID(ctx context.Context, userID int64) (string, error) {
	return m.GetStripeCustomerIDByIDFunc(ctx, userID)
}

type MockCache struct {
	GetCheckoutURLFunc     func(ctx context.Context, idx string) (string, error)
	SetCheckoutURLFunc     func(ctx context.Context, idx string, url string) error
	DeleteCheckoutURLFunc  func(ctx context.Context, idx string) error
	EventIsProcessedFunc   func(ctx context.Context, eventID string) (bool, error)
	MarkEventProcessedFunc func(ctx context.Context, eventID string) error
}

func (m *MockCache) GetCheckoutURL(ctx context.Context, idx string) (string, error) {
	return m.GetCheckoutURLFunc(ctx, idx)
}

func (m *MockCache) SetCheckoutURL(ctx context.Context, idx string, url string) error {
	return m.SetCheckoutURLFunc(ctx, idx, url)
}

func (m *MockCache) DeleteCheckoutURL(ctx context.Context, idx string) error {
	return m.DeleteCheckoutURLFunc(ctx, idx)
}

func (m *MockCache) EventIsProcessed(ctx context.Context, eventID string) (bool, error) {
	return m.EventIsProcessedFunc(ctx, eventID)
}

func (m *MockCache) MarkEventProcessed(ctx context.Context, eventID string) error {
	return m.MarkEventProcessedFunc(ctx, eventID)
}

type MockStripe struct {
	CreatePlanFunc            func(ctx context.Context, p domain.Plan) (domain.Plan, error)
	CreateCheckoutSessionFunc func(ctx context.Context, p domain.Checkout) (string, error)
	VerifyWebhookFunc         func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error)
	GetSubscriptionByIDFunc   func(ctx context.Context, subscriptionID string) (domain.Subscription, error)
}

func (m *MockStripe) CreatePlan(ctx context.Context, p domain.Plan) (domain.Plan, error) {
	return m.CreatePlanFunc(ctx, p)
}

func (m *MockStripe) CreateCheckoutSession(ctx context.Context, p domain.Checkout) (string, error) {
	return m.CreateCheckoutSessionFunc(ctx, p)
}

func (m *MockStripe) VerifyWebhook(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
	return m.VerifyWebhookFunc(ctx, body, sig)
}

func (m *MockStripe) GetSubscriptionByID(ctx context.Context, subscriptionID string) (domain.Subscription, error) {
	return m.GetSubscriptionByIDFunc(ctx, subscriptionID)
}

func TestService_CreatePlan(t *testing.T) {
	tests := []struct {
		name    string
		p       domain.Plan
		strp    service.Stripe
		r       service.PlanRepository
		wantErr bool
	}{
		{
			name: "creates a plan",
			p: domain.Plan{
				Name:         "month plan",
				DurationDays: 30,
				Price:        decimal.RequireFromString("49.99"),
				Currency:     "eur",
			},
			strp: &MockStripe{
				CreatePlanFunc: func(ctx context.Context, p domain.Plan) (domain.Plan, error) {
					return domain.Plan{
						StripeProductID: "fakeID",
						Name:            "month plan",
						DurationDays:    30,
						Price:           decimal.RequireFromString("49.99"),
						Currency:        "eur",
						StripePriceID:   "fakeID",
					}, nil
				},
			},
			r: &MockPlanRepo{
				CreatePlanFunc: func(ctx context.Context, plan domain.Plan) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "invalid price",
			p: domain.Plan{
				Name:         "month plan",
				DurationDays: 30,
				Price:        decimal.RequireFromString("-49.99"),
				Currency:     "eur",
			},
			strp:    &MockStripe{},
			r:       &MockPlanRepo{},
			wantErr: true,
		},
		{
			name: "stripe error",
			p: domain.Plan{
				Name:         "month plan",
				DurationDays: 30,
				Price:        decimal.RequireFromString("49.99"),
				Currency:     "eur",
			},
			strp: &MockStripe{
				CreatePlanFunc: func(ctx context.Context, p domain.Plan) (domain.Plan, error) {
					return domain.Plan{}, errors.New("error")
				},
			},
			r:       &MockPlanRepo{},
			wantErr: true,
		},
		{
			name: "database error",
			p: domain.Plan{
				Name:         "month plan",
				DurationDays: 30,
				Price:        decimal.RequireFromString("49.99"),
				Currency:     "eur",
			},
			strp: &MockStripe{
				CreatePlanFunc: func(ctx context.Context, p domain.Plan) (domain.Plan, error) {
					return domain.Plan{
						StripeProductID: "fakeID",
						Name:            "month plan",
						DurationDays:    30,
						Price:           decimal.RequireFromString("49.99"),
						Currency:        "eur",
						StripePriceID:   "fakeID",
					}, nil
				},
			},
			r: &MockPlanRepo{
				CreatePlanFunc: func(ctx context.Context, plan domain.Plan) error {
					return errors.New("error")
				},
			},
			wantErr: true,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(tt.r, &MockSubscriptionRepo{},
				&MockMembershipRepo{}, &MockUserRepo{},
				&MockCache{}, tt.strp, logger)

			err := s.CreatePlan(t.Context(), tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateCheckoutSession(t *testing.T) {
	tests := []struct {
		name    string
		params  domain.Checkout
		c       service.Cache
		r       service.MembershipRepository
		pr      service.PlanRepository
		ur      service.UserRepository
		strp    service.Stripe
		want    string
		wantErr bool
	}{
		{
			name: "success — cache miss, creates session",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "", myredis.ErrNoSuchKey
				},
				SetCheckoutURLFunc: func(ctx context.Context, idx string, url string) error {
					return nil
				},
			},
			r: &MockMembershipRepo{
				HasActiveMembershipByIDFunc: func(ctx context.Context, userID int64) (bool, error) {
					return false, nil
				},
			},
			pr: &MockPlanRepo{
				GetStripePriceIDByPlanIDFunc: func(ctx context.Context, planID int64) (string, error) {
					return "price_ABC", nil
				},
			},
			ur: &MockUserRepo{
				GetStripeCustomerIDByIDFunc: func(ctx context.Context, userID int64) (string, error) {
					return "cus_ABC", nil
				},
			},
			strp: &MockStripe{
				CreateCheckoutSessionFunc: func(ctx context.Context, p domain.Checkout) (string, error) {
					return "https://checkout.stripe.com/abc", nil
				},
			},
			want:    "https://checkout.stripe.com/abc",
			wantErr: false,
		},
		{
			name: "cache hit — returns existing URL",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "https://cached.url/abc", nil
				},
			},
			r:       &MockMembershipRepo{},
			pr:      &MockPlanRepo{},
			ur:      &MockUserRepo{},
			strp:    &MockStripe{},
			want:    "https://cached.url/abc",
			wantErr: false,
		},
		{
			name: "cache error — non-miss",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "", errors.New("redis down")
				},
			},
			r:       &MockMembershipRepo{},
			pr:      &MockPlanRepo{},
			ur:      &MockUserRepo{},
			strp:    &MockStripe{},
			want:    "",
			wantErr: true,
		},
		{
			name: "user already has active membership",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "", myredis.ErrNoSuchKey
				},
			},
			r: &MockMembershipRepo{
				HasActiveMembershipByIDFunc: func(ctx context.Context, userID int64) (bool, error) {
					return true, nil
				},
			},
			pr:      &MockPlanRepo{},
			ur:      &MockUserRepo{},
			strp:    &MockStripe{},
			want:    "",
			wantErr: true,
		},
		{
			name: "database error",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "", myredis.ErrNoSuchKey
				},
			},
			r: &MockMembershipRepo{
				HasActiveMembershipByIDFunc: func(ctx context.Context, userID int64) (bool, error) {
					return false, errors.New("db down")
				},
			},
			pr:      &MockPlanRepo{},
			ur:      &MockUserRepo{},
			strp:    &MockStripe{},
			want:    "",
			wantErr: true,
		},
		{
			name: "plan does not exist",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "", myredis.ErrNoSuchKey
				},
			},
			r: &MockMembershipRepo{
				HasActiveMembershipByIDFunc: func(ctx context.Context, userID int64) (bool, error) {
					return false, nil
				},
			},
			pr: &MockPlanRepo{
				GetStripePriceIDByPlanIDFunc: func(ctx context.Context, planID int64) (string, error) {
					return "", domain.ErrPlanDoesNotExist
				},
			},
			ur:      &MockUserRepo{},
			strp:    &MockStripe{},
			want:    "",
			wantErr: true,
		},
		{
			name: "stripe error",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "", myredis.ErrNoSuchKey
				},
			},
			r: &MockMembershipRepo{
				HasActiveMembershipByIDFunc: func(ctx context.Context, userID int64) (bool, error) {
					return false, nil
				},
			},
			pr: &MockPlanRepo{
				GetStripePriceIDByPlanIDFunc: func(ctx context.Context, planID int64) (string, error) {
					return "price_ABC", nil
				},
			},
			ur: &MockUserRepo{
				GetStripeCustomerIDByIDFunc: func(ctx context.Context, userID int64) (string, error) {
					return "cus_ABC", nil
				},
			},
			strp: &MockStripe{
				CreateCheckoutSessionFunc: func(ctx context.Context, p domain.Checkout) (string, error) {
					return "", errors.New("stripe error")
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "cache set error — still returns URL",
			params: domain.Checkout{
				UserID: 1,
				PlanID: 5,
			},
			c: &MockCache{
				GetCheckoutURLFunc: func(ctx context.Context, idx string) (string, error) {
					return "", myredis.ErrNoSuchKey
				},
				SetCheckoutURLFunc: func(ctx context.Context, idx string, url string) error {
					return errors.New("redis down")
				},
			},
			r: &MockMembershipRepo{
				HasActiveMembershipByIDFunc: func(ctx context.Context, userID int64) (bool, error) {
					return false, nil
				},
			},
			pr: &MockPlanRepo{
				GetStripePriceIDByPlanIDFunc: func(ctx context.Context, planID int64) (string, error) {
					return "price_ABC", nil
				},
			},
			ur: &MockUserRepo{
				GetStripeCustomerIDByIDFunc: func(ctx context.Context, userID int64) (string, error) {
					return "cus_ABC", nil
				},
			},
			strp: &MockStripe{
				CreateCheckoutSessionFunc: func(ctx context.Context, p domain.Checkout) (string, error) {
					return "https://checkout.stripe.com/abc", nil
				},
			},
			want:    "https://checkout.stripe.com/abc",
			wantErr: false,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(tt.pr, &MockSubscriptionRepo{},
				tt.r, tt.ur,
				tt.c, tt.strp, logger)

			got, err := s.CreateCheckoutSession(t.Context(), tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_HandleStripeWebhook(t *testing.T) {
	tests := []struct {
		name    string
		c       service.Cache
		strp    service.Stripe
		r       service.SubscriptionRepository
		wantErr bool
	}{
		{
			name: "checkout completed — success",
			c: &MockCache{
				DeleteCheckoutURLFunc: func(ctx context.Context, idx string) error {
					return nil
				},
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return false, nil
				},
				MarkEventProcessedFunc: func(ctx context.Context, eventID string) error {
					return nil
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_1",
						Type: domain.EventCheckoutCompleted,
					}, nil
				},
			},
			r: &MockSubscriptionRepo{
				CreateSubscriptionFunc: func(ctx context.Context, event domain.WebhookEvent) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "subscription deleted — success",
			c: &MockCache{
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return false, nil
				},
				MarkEventProcessedFunc: func(ctx context.Context, eventID string) error {
					return nil
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_2",
						Type: domain.EventCustomerSubscriptionDeleted,
					}, nil
				},
			},
			r: &MockSubscriptionRepo{
				CancelSubscriptionFunc: func(ctx context.Context, event domain.WebhookEvent) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "invoice paid — success",
			c: &MockCache{
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return false, nil
				},
				MarkEventProcessedFunc: func(ctx context.Context, eventID string) error {
					return nil
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_3",
						Type: domain.EventInvoicePaid,
					}, nil
				},
			},
			r: &MockSubscriptionRepo{
				ProlongSubscriptionFunc: func(ctx context.Context, event domain.WebhookEvent) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "invoice payment failed — success",
			c: &MockCache{
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return false, nil
				},
				MarkEventProcessedFunc: func(ctx context.Context, eventID string) error {
					return nil
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_4",
						Type: domain.EventInvoicePaymentFailed,
					}, nil
				},
			},
			r: &MockSubscriptionRepo{
				MarkSubscriptionPaymentFailedFunc: func(ctx context.Context, event domain.WebhookEvent) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "unhandled event — returns nil",
			c: &MockCache{
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return false, nil
				},
				MarkEventProcessedFunc: func(ctx context.Context, eventID string) error {
					return nil
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_5",
						Type: domain.EventType("event.not.handled"),
					}, nil
				},
			},
			r:       &MockSubscriptionRepo{},
			wantErr: false,
		},
		{
			name: "signature verification fails",
			c:    &MockCache{},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{}, errors.New("invalid")
				},
			},
			r:       &MockSubscriptionRepo{},
			wantErr: true,
		},
		{
			name: "event already processed — returns nil",
			c: &MockCache{
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return true, nil
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_6",
						Type: domain.EventInvoicePaid,
					}, nil
				},
			},
			r:       &MockSubscriptionRepo{},
			wantErr: false,
		},
		{
			name: "cache check error",
			c: &MockCache{
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return false, errors.New("redis down")
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_7",
						Type: domain.EventInvoicePaid,
					}, nil
				},
			},
			r:       &MockSubscriptionRepo{},
			wantErr: true,
		},
		{
			name: "database error",
			c: &MockCache{
				EventIsProcessedFunc: func(ctx context.Context, eventID string) (bool, error) {
					return false, nil
				},
			},
			strp: &MockStripe{
				VerifyWebhookFunc: func(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error) {
					return domain.WebhookEvent{
						ID:   "evt_8",
						Type: domain.EventInvoicePaid,
					}, nil
				},
			},
			r: &MockSubscriptionRepo{
				ProlongSubscriptionFunc: func(ctx context.Context, event domain.WebhookEvent) error {
					return errors.New("db error")
				},
			},
			wantErr: true,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	body := []byte("...")
	sig := "sig"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(&MockPlanRepo{}, tt.r,
				&MockMembershipRepo{}, &MockUserRepo{},
				tt.c, tt.strp, logger)

			err := s.HandleStripeWebhook(t.Context(), body, sig)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
