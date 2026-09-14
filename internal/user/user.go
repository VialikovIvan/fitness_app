package user

import (
	"log/slog"

	"github.com/VialikovIvan/fitness_app/internal/pkg/crypto"
	myjwt "github.com/VialikovIvan/fitness_app/internal/pkg/jwt"
	"github.com/VialikovIvan/fitness_app/internal/user/repository"
	"github.com/VialikovIvan/fitness_app/internal/user/service"
	"github.com/VialikovIvan/fitness_app/internal/user/transport"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(db *pgxpool.Pool, jwtSecret string, l *slog.Logger) *transport.Handler {
	l = l.With("feature", "user")
	repo := repository.NewPostgres(db)
	hasher := crypto.NewArgon2()
	tokens := myjwt.New(jwtSecret)
	svc := service.New(repo, hasher, tokens)
	v := validator.New()
	handler := transport.New(svc, v, l)

	return handler
}
