package repository

import (
	"context"
	"errors"
	"time"

	"github.com/VialikovIvan/fitness_app/internal/user/domain"
	"github.com/jackc/pgx/v5"
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

type postgresUser struct {
	ID           int64
	FirstName    string
	MiddleName   string
	LastName     string
	PasswordHash string
	Email        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    time.Time
}

func (p *PostgresRepository) Create(ctx context.Context, u domain.User, hash string) (int64, error) {
	user := postgresUser{
		FirstName:    u.FirstName,
		MiddleName:   u.MiddleName,
		LastName:     u.LastName,
		PasswordHash: hash,
		Email:        u.Email,
	}

	err := p.db.QueryRow(ctx, `
	INSERT INTO users (first_name, middle_name, last_name, password_hash, email)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (email) DO NOTHING
	RETURNING id`,
		user.FirstName, user.MiddleName, user.LastName, user.PasswordHash, user.Email,
	).Scan(&user.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrUserAlreadyExists
		}
		return 0, err
	}

	return user.ID, nil
}

func (p *PostgresRepository) GetIDAndHashByEmail(ctx context.Context, email string) (int64, string, error) {
	user := postgresUser{
		Email: email,
	}

	err := p.db.QueryRow(ctx, `
	SELECT id, password_hash FROM users
	WHERE email = $1
	`, user.Email,
	).Scan(&user.ID, &user.PasswordHash)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", domain.ErrUserDoesNotExist
		}
		return 0, "", err
	}

	return user.ID, user.PasswordHash, nil
}
