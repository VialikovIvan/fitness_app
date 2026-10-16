package qrcode

import (
	"log/slog"

	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	membershipRepo "github.com/VialikovIvan/fitness_app/internal/membership/repository"
	myjwt "github.com/VialikovIvan/fitness_app/internal/pkg/jwt"
	"github.com/VialikovIvan/fitness_app/internal/qrcode/service"
	"github.com/VialikovIvan/fitness_app/internal/qrcode/transport"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(db *pgxpool.Pool, tokens *myjwt.Manager, c *myredis.RedisCacher, l *slog.Logger, baseURL string) *transport.Handler {
	l = l.With("feature", "membership")
	mr := membershipRepo.NewPostgres(db)
	svc := service.New(mr, tokens, c, l, baseURL)
	handler := transport.New(svc, l)

	return handler
}
