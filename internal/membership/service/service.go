package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
)

type PlanRepository interface {
	GetPlans(ctx context.Context) ([]domain.Plan, error)
	CreatePlan(ctx context.Context, plan domain.Plan) error
}

type Stripe interface {
	CreatePlan(ctx context.Context, p domain.Plan) (domain.Plan, error)
}

type Service struct {
	planRepo PlanRepository
	stripe   Stripe
}

func New(pr PlanRepository, s Stripe) *Service {
	return &Service{
		planRepo: pr,
		stripe:   s,
	}
}

func (s *Service) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	plans, err := s.planRepo.GetPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("get plans: %w", err)
	}

	return plans, nil
}

func (s *Service) CreatePlan(ctx context.Context, p domain.Plan) error {
	p.Price = p.Price.Round(2)

	if !domain.IsValidPrice(p.Price) {
		return domain.ErrInvalidPrice
	}

	p.Currency = strings.ToUpper(p.Currency)

	if !domain.IsValidCurrency(p.Currency) {
		return domain.ErrInvalidCurrency
	}

	var err error
	p, err = s.stripe.CreatePlan(ctx, p)
	if err != nil {
		return fmt.Errorf("stripe create plan: %w", err)
	}

	err = s.planRepo.CreatePlan(ctx, p)
	if err != nil {
		return fmt.Errorf("repo create plan: %w", err)
	}

	return nil
}
