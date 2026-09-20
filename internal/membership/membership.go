package membership

import (
	"log/slog"

	"github.com/VialikovIvan/fitness_app/internal/membership/repository"
	"github.com/VialikovIvan/fitness_app/internal/membership/service"
	"github.com/VialikovIvan/fitness_app/internal/membership/transport"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(db *pgxpool.Pool, l *slog.Logger) *transport.Handler {
	l = l.With("feature", "membership")
	repo := repository.NewPostgres(db)
	svc := service.New(repo)
	v := validator.New()
	handler := transport.New(svc, v, l)

	return handler
}