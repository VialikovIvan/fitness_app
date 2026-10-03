package repository

import (
	"context"
	"errors"
	"time"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgres(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

type postgresPlan struct {
	ID              int64
	Name            string
	DurationDays    int
	Price           string
	Currency        string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	StripeProductID string
	StripePriceID   string
}

func (p *PostgresRepository) GetPlans(ctx context.Context) ([]domain.Plan, error) {
	rows, err := p.db.Query(ctx, `
	SELECT id, name, duration_days, price, currency
	FROM plans
	`)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var plans []postgresPlan

	for rows.Next() {
		var p postgresPlan

		err = rows.Scan(&p.ID, &p.Name, &p.DurationDays, &p.Price, &p.Currency)
		if err != nil {
			return nil, err
		}

		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	res := make([]domain.Plan, 0, len(plans))

	for _, p := range plans {
		prc, err := decimal.NewFromString(p.Price)
		if err != nil {
			return nil, err
		}

		res = append(res, domain.Plan{
			ID:           p.ID,
			Name:         p.Name,
			DurationDays: p.DurationDays,
			Price:        prc,
			Currency:     p.Currency,
		})
	}

	return res, nil
}

func (p *PostgresRepository) CreatePlan(ctx context.Context, pl domain.Plan) error {
	plan := postgresPlan{
		Name:            pl.Name,
		DurationDays:    pl.DurationDays,
		Price:           pl.Price.String(),
		Currency:        pl.Currency,
		StripeProductID: pl.StripeProductID,
		StripePriceID:   pl.StripePriceID,
	}

	_, err := p.db.Exec(ctx, `
	INSERT INTO plans (name, duration_days, price, currency, stripe_product_id, stripe_price_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	`, plan.Name, plan.DurationDays, plan.Price, plan.Currency, plan.StripeProductID, plan.StripePriceID)

	return err
}

func (p *PostgresRepository) HasActiveMembershipByID(ctx context.Context, userID int64) (bool, error) {
	var exists bool

	err := p.db.QueryRow(ctx, `
	SELECT EXISTS(SELECT 1 FROM memberships WHERE user_id = $1 AND end_date > NOW() AND status = 'active')
	`, userID).Scan(&exists)

	return exists, err
}

func (p *PostgresRepository) GetStripePriceIDByPlanID(ctx context.Context, planID int64) (string, error) {
	plan := postgresPlan{
		ID: planID,
	}

	err := p.db.QueryRow(ctx, `
	SELECT stripe_price_id FROM plans
	WHERE id = $1 
	`, plan.ID).Scan(&plan.StripePriceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrPlanDoesNotExist
		}
		return "", err
	}

	return plan.StripePriceID, nil
}

func (p *PostgresRepository) CreateSubscription(ctx context.Context, event domain.WebhookEvent) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
	UPDATE users
	SET stripe_customer_id = $1, updated_at = NOW()
	WHERE id = $2 AND stripe_customer_id IS NULL
	`, event.CustomerID, event.UserID)
	if err != nil {
		return err
	}

	err = tx.QueryRow(ctx, `
	INSERT INTO subscriptions(user_id, plan_id, status, current_period_start,
	current_period_end, cancel_at_period_end, stripe_subscription_id)
	VALUES($1, $2, $3, $4, $5, $6, $7)
	RETURNING id
	`, event.UserID, event.PlanID, event.Status, event.Subscription.CurrentPeriodStart,
		event.Subscription.CurrentPeriodEnd, event.Subscription.CancelAtPeriodEnd,
		event.Subscription.StripeID).Scan(&event.Subscription.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO memberships(user_id, plan_id,
	subscription_id, start_date, end_date, status)
	VALUES($1, $2, $3, $4, $5, 'active')
	`, event.UserID, event.PlanID, event.Subscription.ID,
		event.Subscription.CurrentPeriodStart,
		event.Subscription.CurrentPeriodEnd)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *PostgresRepository) CancelSubscription(ctx context.Context, event domain.WebhookEvent) error {
	_, err := p.db.Exec(ctx, `
	UPDATE subscriptions
	SET status = 'canceled', canceled_at = $1, cancel_at_period_end = true, updated_at = NOW()
	WHERE stripe_subscription_id = $2
	`, event.Subscription.CanceledAt, event.Subscription.StripeID)

	return err
}

func (p *PostgresRepository) ProlongSubscription(ctx context.Context, event domain.WebhookEvent) error {
	tx, err := p.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
	UPDATE subscriptions
	SET status = $1, current_period_start = $2,
	current_period_end = $3, updated_at = NOW()
	WHERE stripe_subscription_id = $4
	RETURNING id
	`, event.Subscription.Status, event.Subscription.CurrentPeriodStart,
		event.Subscription.CurrentPeriodEnd, event.Subscription.StripeID).Scan(&event.Subscription.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
	INSERT INTO memberships(user_id, plan_id,
	subscription_id, start_date, end_date, status)
	VALUES($1, $2, $3, $4, $5, $6)
	`, event.UserID, event.PlanID, event.Subscription.ID,
		event.Subscription.CurrentPeriodStart,
		event.Subscription.CurrentPeriodEnd,
		event.Subscription.Status)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (p *PostgresRepository) MarkSubscriptionPaymentFailed(ctx context.Context, event domain.WebhookEvent) error {
	_, err := p.db.Exec(ctx, `
	UPDATE subscriptions
	SET status = $1, current_period_start = $2,
	current_period_end = $3, updated_at = NOW()
	WHERE stripe_subscription_id = $4
	`, event.Subscription.Status,
		event.Subscription.CurrentPeriodStart,
		event.Subscription.CurrentPeriodEnd,
		event.Subscription.StripeID)

	return err
}
