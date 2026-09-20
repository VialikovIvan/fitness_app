package domain

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrInvalidPrice    = errors.New("invalid price")
	ErrInvalidCurrency = errors.New("invalid currency")
)

type Plan struct {
	ID           int64
	Name         string
	DurationDays int
	Price        decimal.Decimal
	Currency     string
}

func IsValidPrice(price decimal.Decimal) bool {
	return price.GreaterThan(decimal.Zero) &&
		price.Exponent() == -2 &&
		price.LessThan(decimal.NewFromInt(100000000))
}

func IsValidCurrency(currency string) bool {
	return len(currency) == 3
}
