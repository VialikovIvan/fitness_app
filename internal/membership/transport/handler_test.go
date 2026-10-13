package transport_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/VialikovIvan/fitness_app/internal/membership/transport"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type MockService struct {
	CreatePlanFunc            func(ctx context.Context, p domain.Plan) error
	CreateCheckoutSessionFunc func(ctx context.Context, params domain.Checkout) (string, error)
	HandleStripeWebhookFunc   func(ctx context.Context, body []byte, sig string) error
}

func (m *MockService) CreatePlan(ctx context.Context, p domain.Plan) error {
	return m.CreatePlanFunc(ctx, p)
}

func (m *MockService) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	return nil, nil
}

func (m *MockService) CreateCheckoutSession(ctx context.Context, params domain.Checkout) (string, error) {
	return m.CreateCheckoutSessionFunc(ctx, params)
}

func (m *MockService) HandleStripeWebhook(ctx context.Context, body []byte, sig string) error {
	return m.HandleStripeWebhookFunc(ctx, body, sig)
}

func TestHandler_CreatePlan(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		s          transport.MembershipService
		wantJSON   string // Keep empty unless expect JSON response
		wantText   string // Keep empty unless expect text response
		wantStatus int
	}{
		{
			name: "creates a plan",
			body: `{"name":"month plan","duration_days":30,"price":"44.99","currency":"eur"}`,
			s: &MockService{
				CreatePlanFunc: func(ctx context.Context, p domain.Plan) error {
					return nil
				},
			},
			wantJSON:   `{"message":"created the plan successfully"}`,
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid price",
			body: `{"name":"month plan","duration_days":30,"price":"NotAPrice","currency":"eur"}`,
			s: &MockService{
				CreatePlanFunc: func(ctx context.Context, p domain.Plan) error {
					return nil
				},
			},
			wantText:   "Invalid JSON payload",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "validation error",
			body: `{"name":"month plan","price":"44.99","currency":"eur"}`,
			s: &MockService{
				CreatePlanFunc: func(ctx context.Context, p domain.Plan) error {
					return nil
				},
			},
			wantText:   "Invalid data",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{"name":"month plan","duration_days":30,"price":"44.99","currency":"eur"}`,
			s: &MockService{
				CreatePlanFunc: func(ctx context.Context, p domain.Plan) error {
					return errors.New("error")
				},
			},
			wantText:   "Internal Server Error",
			wantStatus: http.StatusInternalServerError,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := transport.NewHandler(tt.s, validator.New(), logger)

			req := httptest.NewRequest("POST", "/plans", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.CreatePlan(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantJSON != "" {
				assert.JSONEq(t, tt.wantJSON, rr.Body.String())
			} else if tt.wantText != "" {
				assert.Equal(t, tt.wantText+"\n", rr.Body.String())
			}
		})
	}
}

func TestHandler_Checkout(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		userID     int64
		hasUserID  bool
		s          transport.MembershipService
		wantJSON   string
		wantText   string
		wantStatus int
	}{
		{
			name:      "creates a checkout session",
			body:      `{"plan_id":5}`,
			userID:    1,
			hasUserID: true,
			s: &MockService{
				CreateCheckoutSessionFunc: func(ctx context.Context, params domain.Checkout) (string, error) {
					return "https://checkout.stripe.com/abc", nil
				},
			},
			wantJSON:   `{"redirect_url":"https://checkout.stripe.com/abc"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing user id in context",
			body:       `{"plan_id":5}`,
			hasUserID:  false,
			s:          &MockService{},
			wantText:   "Internal Server Error",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid json payload",
			body:       `{"plan_id":}`,
			userID:     1,
			hasUserID:  true,
			s:          &MockService{},
			wantText:   "Invalid JSON payload",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "validation error",
			body:       `{}`,
			userID:     1,
			hasUserID:  true,
			s:          &MockService{},
			wantText:   "Invalid data",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "service error",
			body:      `{"plan_id":5}`,
			userID:    1,
			hasUserID: true,
			s: &MockService{
				CreateCheckoutSessionFunc: func(ctx context.Context, params domain.Checkout) (string, error) {
					return "", errors.New("error")
				},
			},
			wantText:   "Internal Server Error",
			wantStatus: http.StatusInternalServerError,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := transport.NewHandler(tt.s, validator.New(), logger)

			req := httptest.NewRequest("POST", "/checkout", strings.NewReader(tt.body))
			if tt.hasUserID {
				ctx := context.WithValue(req.Context(), "userID", tt.userID)
				req = req.WithContext(ctx)
			}
			rr := httptest.NewRecorder()

			h.Checkout(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantJSON != "" {
				assert.JSONEq(t, tt.wantJSON, rr.Body.String())
			} else if tt.wantText != "" {
				assert.Equal(t, tt.wantText+"\n", rr.Body.String())
			}
		})
	}
}

func TestHandler_HandleStripeWebhook(t *testing.T) {
	tests := []struct {
		name       string
		s          transport.MembershipService
		wantText   string
		wantStatus int
	}{
		{
			name: "success",
			s: &MockService{
				HandleStripeWebhookFunc: func(ctx context.Context, body []byte, sig string) error {
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid webhook signature",
			s: &MockService{
				HandleStripeWebhookFunc: func(ctx context.Context, body []byte, sig string) error {
					return domain.ErrInvalidWebhook
				},
			},
			wantText:   domain.ErrInvalidWebhook.Error(),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			s: &MockService{
				HandleStripeWebhookFunc: func(ctx context.Context, body []byte, sig string) error {
					return errors.New("error")
				},
			},
			wantText:   "Internal Server Error",
			wantStatus: http.StatusInternalServerError,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	body := `{"id":"evt_1"}`
	sig := "sig"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := transport.NewHandler(tt.s, validator.New(), logger)

			req := httptest.NewRequest("POST", "/webhooks/stripe", strings.NewReader(body))
			req.Header.Set("Stripe-Signature", sig)
			rr := httptest.NewRecorder()

			h.HandleStripeWebhook(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}

			if tt.wantText != "" {
				assert.Equal(t, tt.wantText+"\n", rr.Body.String())
			}
		})
	}
}
