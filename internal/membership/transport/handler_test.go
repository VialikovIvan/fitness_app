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
	CreatePlanFunc func(ctx context.Context, p domain.Plan) error
}

func (m *MockService) CreatePlan(ctx context.Context, p domain.Plan) error {
	return m.CreatePlanFunc(ctx, p)
}

func (m *MockService) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	return nil, nil
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
			wantText: "Invalid JSON payload",
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
			wantText: "Invalid data",
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
			wantText: "Internal Server Error",
			wantStatus: http.StatusInternalServerError,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := transport.New(tt.s, validator.New(), logger)

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
