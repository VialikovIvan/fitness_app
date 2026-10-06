package repository

import (
	"context"

	"github.com/VialikovIvan/fitness_app/internal/payment/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgres(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

/*
type postgresPayment struct {
	ID                int64
	UserID            int64
	SubscriptionID    int64
	Provider          string
	ProviderPaymentID string
	Amount            decimal.Decimal
	Currency          string
	Status            string
	ProviderCreatedAt time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	PaidAt            *time.Time
}
*/

func (p *PostgresRepository) Log(ctx context.Context, pmt domain.Payment) error {
	_, err := p.db.Exec(ctx, `
	INSERT INTO payments(user_id, subscription_id, provider,
	provider_payment_id, amount, currency,
	status, provider_created_at, paid_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, pmt.UserID, pmt.SubscriptionID, pmt.Provider,
		pmt.ProviderPaymentID, pmt.Amount, pmt.Currency,
		pmt.Status, pmt.ProviderCreatedAt, pmt.PaidAt)

	return err
}
