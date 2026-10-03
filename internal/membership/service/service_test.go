package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/VialikovIvan/fitness_app/internal/membership/service"
	"github.com/shopspring/decimal"
)

type MockPlanRepo struct {
	CreatePlanFunc func(ctx context.Context, plan domain.Plan) error
}

func (m *MockPlanRepo) CreatePlan(ctx context.Context, plan domain.Plan) error {
	return m.CreatePlanFunc(ctx, plan)
}

func (m *MockPlanRepo) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	return nil, nil
}

type MockStripe struct {
	CreatePlanFunc func(ctx context.Context, p domain.Plan) (domain.Plan, error)
}

func (m *MockStripe) CreatePlan(ctx context.Context, p domain.Plan) (domain.Plan, error) {
	return m.CreatePlanFunc(ctx, p)
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
