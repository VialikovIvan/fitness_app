package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
)

type MembershipRepository interface {
	GetPlans(ctx context.Context) ([]domain.Plan, error)
	CreatePlan(ctx context.Context, plan domain.Plan) error
}

type Service struct {
	membershipRepo MembershipRepository
}

func New(r MembershipRepository) *Service {
	return &Service{
		membershipRepo: r,
	}
}

func (s *Service) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	plans, err := s.membershipRepo.GetPlans(ctx)
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

	if !domain.IsValidCurrency(p.Currency) {
		return domain.ErrInvalidCurrency
	}

	p.Currency = strings.ToUpper(p.Currency)

	err := s.membershipRepo.CreatePlan(ctx, p)
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}

	return nil
}
