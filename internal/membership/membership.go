package membership

import (
	"log/slog"

	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	mystripe "github.com/VialikovIvan/fitness_app/internal/infrastructure/stripe"
	membershipRepo "github.com/VialikovIvan/fitness_app/internal/membership/repository"
	"github.com/VialikovIvan/fitness_app/internal/membership/service"
	"github.com/VialikovIvan/fitness_app/internal/membership/transport"
	paymentRepo "github.com/VialikovIvan/fitness_app/internal/payment/repository"
	userRepo "github.com/VialikovIvan/fitness_app/internal/user/repository"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func New(db *pgxpool.Pool, rdb *redis.Client, strp *mystripe.StripeManager, c *myredis.RedisCacher, l *slog.Logger) *transport.Handler {
	l = l.With("feature", "membership")
	mr := membershipRepo.NewPostgres(db)
	ur := userRepo.NewPostgres(db)
	pmr := paymentRepo.NewPostgres(db)
	svc := service.New(mr, mr, mr, ur, pmr, c, strp, l)
	v := validator.New()
	handler := transport.New(svc, v, l)

	return handler
}
