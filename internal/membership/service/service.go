package service

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	mbspDomain "github.com/VialikovIvan/fitness_app/internal/membership/domain"
	pmtDomain "github.com/VialikovIvan/fitness_app/internal/payment/domain"
)

const numLocks uint64 = 64

type PlanRepository interface {
	GetPlans(ctx context.Context) ([]mbspDomain.Plan, error)
	CreatePlan(ctx context.Context, plan mbspDomain.Plan) error
	GetStripePriceIDByPlanID(ctx context.Context, planID int64) (string, error)
}

type SubscriptionRepository interface {
	CreateSubscription(ctx context.Context, event mbspDomain.WebhookEvent) (int64, error)
	ProlongSubscription(ctx context.Context, event mbspDomain.WebhookEvent) (int64, error)
	MarkSubscriptionPaymentFailed(ctx context.Context, event mbspDomain.WebhookEvent) (int64, error)
	CancelSubscription(ctx context.Context, event mbspDomain.WebhookEvent) error
}

type MembershipRepository interface {
	HasActiveMembershipByID(ctx context.Context, userID int64) (bool, error)
}

type UserRepository interface {
	GetStripeCustomerIDByID(ctx context.Context, userID int64) (string, error)
}

type PaymentRepository interface {
	Log(ctx context.Context, pmt pmtDomain.Payment) error
}

type Cache interface {
	GetCheckoutURL(ctx context.Context, idx string) (string, error)
	SetCheckoutURL(ctx context.Context, idx string, url string) error
	DeleteCheckoutURL(ctx context.Context, idx string) error
	EventIsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkEventProcessed(ctx context.Context, eventID string) error
}

type Stripe interface {
	CreatePlan(ctx context.Context, p mbspDomain.Plan) (mbspDomain.Plan, error)
	CreateCheckoutSession(ctx context.Context, p mbspDomain.Checkout) (string, error)
	VerifyWebhook(ctx context.Context, body []byte, sig string) (mbspDomain.WebhookEvent, error)
	GetSubscriptionByID(ctx context.Context, subscriptionID string) (mbspDomain.Subscription, error)
}

type Publisher interface {
	PublishStripeWebhook(ctx context.Context, event mbspDomain.WebhookEvent) error
}

type CheckoutsStartedCounter interface {
	Inc(planID string)
}

type CheckoutsCompletedCounter interface {
	Inc(planID string)
}

type Metrics struct {
	checkoutsStartedTotal   CheckoutsStartedCounter
	checkoutsCompletedTotal CheckoutsCompletedCounter
}

func NewMetricsManager(csc CheckoutsStartedCounter, ccc CheckoutsCompletedCounter) *Metrics {
	return &Metrics{
		checkoutsStartedTotal:   csc,
		checkoutsCompletedTotal: ccc,
	}
}

type Service struct {
	planRepo         PlanRepository
	subscriptionRepo SubscriptionRepository
	membershipRepo   MembershipRepository
	userRepo         UserRepository
	paymentRepo      PaymentRepository
	cache            Cache
	stripe           Stripe
	publisher        Publisher
	logger           *slog.Logger
	metrics          *Metrics
	locks            [numLocks]sync.Mutex
}

func New(pr PlanRepository, sr SubscriptionRepository, mr MembershipRepository,
	ur UserRepository, pmr PaymentRepository, c Cache, s Stripe, p Publisher,
	l *slog.Logger, m *Metrics) *Service {
	return &Service{
		planRepo:         pr,
		subscriptionRepo: sr,
		membershipRepo:   mr,
		userRepo:         ur,
		paymentRepo:      pmr,
		cache:            c,
		stripe:           s,
		publisher:        p,
		logger:           l.With("layer", "service"),
		metrics:          m,
	}
}

func (s *Service) GetPlans(ctx context.Context) ([]mbspDomain.Plan, error) {
	plans, err := s.planRepo.GetPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("get plans: %w", err)
	}

	return plans, nil
}

func (s *Service) CreatePlan(ctx context.Context, p mbspDomain.Plan) error {
	p.Price = p.Price.Round(2)

	if !mbspDomain.IsValidPrice(p.Price) {
		return mbspDomain.ErrInvalidPrice
	}

	p.Currency = strings.ToUpper(p.Currency)

	if !mbspDomain.IsValidCurrency(p.Currency) {
		return mbspDomain.ErrInvalidCurrency
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

func (s *Service) CreateCheckoutSession(ctx context.Context, params mbspDomain.Checkout) (string, error) {
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
		return "", mbspDomain.ErrUserAlreadyHasActiveMembership
	}

	params.StripePriceID, err = s.planRepo.GetStripePriceIDByPlanID(ctx, params.PlanID)
	if err != nil {
		if errors.Is(err, mbspDomain.ErrPlanDoesNotExist) {
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

	s.metrics.checkoutsStartedTotal.Inc(strconv.FormatInt(params.PlanID, 10))

	if err := s.cache.SetCheckoutURL(ctx, uniqueIdx, params.URL); err != nil {
		s.logger.Error("failed to cache checkout url", "err", err)
	}

	return params.URL, nil
}

func (s *Service) HandleStripeWebhook(ctx context.Context, body []byte, sig string) error {
	event, err := s.stripe.VerifyWebhook(ctx, body, sig)
	if err != nil {
		return mbspDomain.ErrInvalidWebhook
	}

	isProcessed, err := s.cache.EventIsProcessed(ctx, event.ID)
	if err != nil {
		return fmt.Errorf("event is processd: %w", err)
	}
	if isProcessed {
		// Returning nil to return 200 in handler (request processed), so stripe stops sending webhooks
		return nil
	}

	if err := s.publisher.PublishStripeWebhook(ctx, event); err != nil {
		return fmt.Errorf("publish stripe webhook: %w", err)
	}

	return nil
}

func (s *Service) HandleEventCheckoutCompleted(ctx context.Context, event mbspDomain.WebhookEvent) error {
	s.metrics.checkoutsCompletedTotal.Inc(strconv.FormatInt(event.PlanID, 10))

	var err error
	event.Payment.SubscriptionID, err = s.subscriptionRepo.CreateSubscription(ctx, event)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	if err := s.paymentRepo.Log(ctx, event.Payment); err != nil {
		s.logger.Error("failed to log the payment", "err", err)
	}

	if err := s.cache.DeleteCheckoutURL(ctx, checkoutIdx(event.UserID, event.PlanID)); err != nil {
		s.logger.Error("failed to delete checkout url cache", "err", err)
	}

	err = s.cache.MarkEventProcessed(ctx, event.ID)
	if err != nil {
		s.logger.Error("unexpected cache error", "err", err)
	}

	return nil
}

func (s *Service) HandleEventCustomerSubscriptionDeleted(ctx context.Context, event mbspDomain.WebhookEvent) error {
	if err := s.subscriptionRepo.CancelSubscription(ctx, event); err != nil {
		return fmt.Errorf("cancel subscription: %w", err)
	}

	err := s.cache.MarkEventProcessed(ctx, event.ID)
	if err != nil {
		s.logger.Error("unexpected cache error", "err", err)
	}

	return nil
}

func (s *Service) HandleEventInvoicePaid(ctx context.Context, event mbspDomain.WebhookEvent) error {
	var err error
	event.Payment.SubscriptionID, err = s.subscriptionRepo.ProlongSubscription(ctx, event)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	if err := s.paymentRepo.Log(ctx, event.Payment); err != nil {
		s.logger.Error("failed to log the payment", "err", err)
	}

	err = s.cache.MarkEventProcessed(ctx, event.ID)
	if err != nil {
		s.logger.Error("unexpected cache error", "err", err)
	}

	return nil
}

func (s *Service) HandleEventInvoicePaymentFailed(ctx context.Context, event mbspDomain.WebhookEvent) error {
	var err error
	event.Payment.SubscriptionID, err = s.subscriptionRepo.MarkSubscriptionPaymentFailed(ctx, event)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}

	if err := s.paymentRepo.Log(ctx, event.Payment); err != nil {
		s.logger.Error("failed to log the payment", "err", err)
	}

	err = s.cache.MarkEventProcessed(ctx, event.ID)
	if err != nil {
		s.logger.Error("unexpected cache error", "err", err)
	}

	return nil
}

func (s *Service) HandleUnknownEvent(ctx context.Context, event mbspDomain.WebhookEvent) {
	if err := s.cache.MarkEventProcessed(ctx, event.ID); err != nil {
		s.logger.Error("unexpected cache error", "err", err)
	}
}

func (s *Service) getMutexByIdx(idx string) *sync.Mutex {
	h := fnv.New64a()
	h.Write([]byte(idx))
	return &s.locks[h.Sum64()%numLocks]
}

func checkoutIdx(userID, planID int64) string {
	return strconv.FormatInt(userID, 10) + ":" + strconv.FormatInt(planID, 10)
}
