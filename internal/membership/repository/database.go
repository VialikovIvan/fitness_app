package repository

import (
	"context"
	"time"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
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
