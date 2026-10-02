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
	ID               int64
	FirstName        string
	MiddleName       string
	LastName         string
	PasswordHash     string
	Email            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Role             string
	StripeCustomerID *string
}

func (p *PostgresRepository) Create(ctx context.Context, u domain.User, hash string) (domain.User, error) {
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
	RETURNING id, role`,
		user.FirstName, user.MiddleName, user.LastName, user.PasswordHash, user.Email,
	).Scan(&user.ID, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrUserAlreadyExists
		}
		return domain.User{}, err
	}

	return domain.User{
		ID:   user.ID,
		Role: user.Role,
	}, nil
}

func (p *PostgresRepository) GetCredentialsByEmail(ctx context.Context, email string) (domain.User, string, error) {
	user := postgresUser{
		Email: email,
	}

	err := p.db.QueryRow(ctx, `
	SELECT id, password_hash, role FROM users
	WHERE email = $1
	`, user.Email,
	).Scan(&user.ID, &user.PasswordHash, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, "", domain.ErrUserDoesNotExist
		}
		return domain.User{}, "", err
	}

	return domain.User{
		ID:   user.ID,
		Role: user.Role,
	}, user.PasswordHash, nil
}

func (p *PostgresRepository) GetStripeCustomerIDByID(ctx context.Context, userID int64) (string, error) {
	user := postgresUser{
		ID: userID,
	}

	err := p.db.QueryRow(ctx, `
	SELECT stripe_customer_id FROM users
	WHERE id = $1
	`, user.ID).Scan(&user.StripeCustomerID)
	if err != nil {
		return "", err
	}

	if user.StripeCustomerID == nil {
		return "", nil
	}
	
	return *user.StripeCustomerID, nil
} 
