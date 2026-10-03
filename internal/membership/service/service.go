package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
)

type PlanRepository interface {
	GetPlans(ctx context.Context) ([]domain.Plan, error)
	CreatePlan(ctx context.Context, plan domain.Plan) error
	GetStripePriceIDByPlanID(ctx context.Context, planID int64) (string, error)
}

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, event domain.WebhookEvent) error
	ProlongSubscription(ctx context.Context, event domain.WebhookEvent) error
	MarkSubscriptionPaymentFailed(ctx context.Context, event domain.WebhookEvent) error
	CancelSubscription(ctx context.Context, event domain.WebhookEvent) error
}

type MembershipRepository interface {
	HasActiveMembershipByID(ctx context.Context, userID int64) (bool, error)
}

type UserRepository interface {
	GetStripeCustomerIDByID(ctx context.Context, userID int64) (string, error)
}

type Cache interface {
	GetCheckoutURL(ctx context.Context, idx string) (string, error)
	SetCheckoutURL(ctx context.Context, idx string, url string) error
	DeleteCheckoutURL(ctx context.Context, idx string) error
	EventIsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventID string) (error)
}

type Stripe interface {
	CreatePlan(ctx context.Context, p domain.Plan) (domain.Plan, error)
	CreateCheckoutSession(ctx context.Context, p domain.Checkout) (string, error)
	VerifyWebhook(ctx context.Context, body []byte, sig string) (domain.WebhookEvent, error)
	GetSubscriptionByID(ctx context.Context, subscriptionID string) (domain.Subscription, error)
}

type Service struct {
	planRepo         PlanRepository
	subscriptionRepo SubscriptionRepository
	membershipRepo   MembershipRepository
	userRepo         UserRepository
	cache            Cache
	stripe           Stripe
	logger           *slog.Logger
	muMap            map[string]*sync.Mutex
	muMapLock        sync.Mutex
}

func New(pr PlanRepository, sr SubscriptionRepository, mr MembershipRepository, ur UserRepository, c Cache, s Stripe, l *slog.Logger) *Service {
	return &Service{
		planRepo:         pr,
		subscriptionRepo: sr,
		membershipRepo:   mr,
		userRepo:         ur,
		cache:            c,
		stripe:           s,
		logger:           l.With("layer", "service"),
		muMap:            make(map[string]*sync.Mutex),
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

func (s *Service) CreateCheckoutSession(ctx context.Context, params domain.Checkout) (string, error) {
	uniqueIdx := checkoutIdx(params.UserID, params.PlanID)
	mu := s.getMutexByIdx(uniqueIdx)

	mu.Lock()
	defer mu.Unlock()

	var err error
	params.URL, err = s.cache.GetCheckoutURL(ctx, uniqueIdx)
	if err == nil {
		return params.URL, nil
	}
	if !errors.Is(err, myredis.ErrNoSuchKey) {
		return "", fmt.Errorf("get checkout url: %w", err)
	}

	hasActive, err := s.membershipRepo.HasActiveMembershipByID(ctx, params.UserID)
	if err != nil {
		return "", fmt.Errorf("has active membership: %w", err)
	}
	if hasActive {
		return "", domain.ErrUserAlreadyHasActiveMembership
	}

	params.StripePriceID, err = s.planRepo.GetStripePriceIDByPlanID(ctx, params.PlanID)
	if err != nil {
		if errors.Is(err, domain.ErrPlanDoesNotExist) {
			return "", err
		}
		return "", fmt.Errorf("get price id: %w", err)
	}

	params.StripeCustomerID, err = s.userRepo.GetStripeCustomerIDByID(ctx, params.UserID)
	if err != nil {
		return "", fmt.Errorf("get stripe customer id: %w", err)
	}

	params.URL, err = s.stripe.CreateCheckoutSession(ctx, params)
	if err != nil {
		return "", fmt.Errorf("create checkout session: %w", err)
	}

	if err := s.cache.SetCheckoutURL(ctx, uniqueIdx, params.URL); err != nil {
		s.logger.Error("failed to cache checkout url", "err", err)
	}

	return params.URL, nil
}

func (s *Service) HandleStripeWebhook(ctx context.Context, body []byte, sig string) error {
	event, err := s.stripe.VerifyWebhook(ctx, body, sig)
	if err != nil {
		return domain.ErrInvalidWebhook
	}

	isProcessed, err := s.cache.EventIsProcessed(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("event is processd: %w", err)
	}
	if isProcessed {
		// Returning nil to return 200 in handler (request processed), so stripe stops sending webhooks
		return nil
	}

	// TODO: Swap the service call for the publish to RabbitMQ
	// TODO: Make use of payments table
	switch event.Type {
	case domain.EventCheckoutCompleted:
		err = s.HandleEventCheckoutCompleted(ctx, event)
	case domain.EventCustomerSubscriptionDeleted:
		err = s.HandleEventCustomerSubscriptionDeleted(ctx, event)
	case domain.EventInvoicePaid:
		err = s.HandleEventInvoicePaid(ctx, event)
	case domain.EventInvoicePaymentFailed:
		err = s.HandleEventInvoicePaymentFailed(ctx, event)
	default:
		s.logger.Info("unhandled webhook event", "type", event.Type, "id", event.ID)
	}
	if err != nil {
		return err
	}

	err = s.cache.MarkEventProcessed(ctx, event.ID)
	if err != nil {
		s.logger.Error("unexpected cache error", "err", err)
	}

	return nil
}

func (s *Service) HandleEventCheckoutCompleted(ctx context.Context, event domain.WebhookEvent) error {
	if err := s.subscriptionRepo.CreateSubscription(ctx, event); err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	if err := s.cache.DeleteCheckoutURL(ctx, checkoutIdx(event.UserID, event.PlanID)); err != nil {
		s.logger.Error("failed to delete checkout url cache", "err", err)
	}

	return nil
}

func (s *Service) HandleEventCustomerSubscriptionDeleted(ctx context.Context, event domain.WebhookEvent) error {
	if err := s.subscriptionRepo.CancelSubscription(ctx, event); err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}

	return nil
}

func (s *Service) HandleEventInvoicePaid(ctx context.Context, event domain.WebhookEvent) error {
	if err := s.subscriptionRepo.ProlongSubscription(ctx, event); err != nil {
		return fmt.Errorf("prolong subscription: %w", err)
	}

	return nil
}

func (s *Service) HandleEventInvoicePaymentFailed(ctx context.Context, event domain.WebhookEvent) error {
	if err := s.subscriptionRepo.MarkSubscriptionPaymentFailed(ctx, event); err != nil {
		return fmt.Errorf("mark subscription payment failed: %w", err)
	}

	return nil
}

func (s *Service) getMutexByIdx(idx string) *sync.Mutex {
	s.muMapLock.Lock()
	defer s.muMapLock.Unlock()
	if s.muMap[idx] == nil {
		s.muMap[idx] = &sync.Mutex{}
	}
	return s.muMap[idx]
}

func checkoutIdx(userID, planID int64) string {
	return strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(planID, 10)
}
