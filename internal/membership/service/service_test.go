package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/VialikovIvan/fitness_app/internal/membership/service"
	"github.com/shopspring/decimal"
)

type MockRepo struct {
	CreatePlanFunc func(ctx context.Context, plan domain.Plan) error
}

func (m *MockRepo) CreatePlan(ctx context.Context, plan domain.Plan) error {
	return m.CreatePlanFunc(ctx, plan)
}

func (m *MockRepo) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	return nil, nil
}

func TestService_CreatePlan(t *testing.T) {
	tests := []struct {
		name    string
		p       domain.Plan
		r       service.MembershipRepository
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
			r: &MockRepo{
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
			r: &MockRepo{
				CreatePlanFunc: func(ctx context.Context, plan domain.Plan) error {
					return nil
				},
			},
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
			r: &MockRepo{
				CreatePlanFunc: func(ctx context.Context, plan domain.Plan) error {
					return errors.New("error")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(tt.r)
			err := s.CreatePlan(t.Context(), tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
