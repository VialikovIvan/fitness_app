package mystripe

import (
	"context"
	"strings"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/stripe/stripe-go/v86"
)

func (m *StripeManager) CreatePlan(ctx context.Context, p domain.Plan) (domain.Plan, error) {
	params := &stripe.ProductCreateParams{
		DefaultPriceData: &stripe.ProductCreateDefaultPriceDataParams{
			Currency: stripe.String(strings.ToLower(p.Currency)),
			Recurring: &stripe.ProductCreateDefaultPriceDataRecurringParams{
				Interval:      stripe.String("day"),
				IntervalCount: stripe.Int64(int64(p.DurationDays)),
			},
			UnitAmount: stripe.Int64(p.Price.Shift(2).IntPart()),
		},
		Name: stripe.String(p.Name),
	}
	params.AddExpand("default_price")

	product, err := m.sc.V1Products.Create(ctx, params)
	if err != nil {
		return domain.Plan{}, err
	}
	// Checking for nil to not get nil pointer dereference if stripe behavior changes
	if product.DefaultPrice == nil {
		return domain.Plan{}, ErrNilDefaultPrice
	}

	p.StripeProductID = product.ID
	p.StripePriceID = product.DefaultPrice.ID

	return p, nil
}
