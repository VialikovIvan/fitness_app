package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Payment struct {
	ID                int64
	UserID            int64
	SubscriptionID    int64
	Provider          string
	ProviderPaymentID string
	Amount            decimal.Decimal
	Currency          string
	Status            string
	ProviderCreatedAt time.Time
	PaidAt            *time.Time
}
