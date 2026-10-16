package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/VialikovIvan/fitness_app/internal/user/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u domain.User, hash string) (domain.User, error)
	GetCredentialsByEmail(ctx context.Context, email string) (domain.User, string, error)
}

type PasswordManager interface {
	Hash(ctx context.Context, plainPassword string) (string, error)
	Verify(encoded, providedPassword string) (bool, error)
}

type TokenGenerator interface {
	GenerateAuth(ctx context.Context, userID int64, role string) (string, error)
}

type Service struct {
	userRepo        UserRepository
	passwordManager PasswordManager
	tokenGenerator  TokenGenerator
}

func New(r UserRepository, p PasswordManager, t TokenGenerator) *Service {
	return &Service{
		userRepo:        r,
		passwordManager: p,
		tokenGenerator:  t,
	}
}

func (s *Service) Register(ctx context.Context, u domain.User, plainPassword string) (string, error) {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	if !domain.IsValidEmailFormat(u.Email) {
		return "", domain.ErrInvalidEmailFormat
	}

	if !domain.IsValidPasswordLength(plainPassword) {
		return "", domain.ErrInvalidPasswordLength
	}

	hash, err := s.passwordManager.Hash(ctx, plainPassword)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	user, err := s.userRepo.Create(ctx, u, hash)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			return "", err
		}
		return "", fmt.Errorf("create user: %w", err)
	}

	token, err := s.tokenGenerator.GenerateAuth(ctx, user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("generate auth token: %w", err)
	}

	return token, nil
}

func (s *Service) Login(ctx context.Context, u domain.User, plainPassword string) (string, error) {
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	if !domain.IsValidEmailFormat(u.Email) {
		return "", domain.ErrInvalidEmailFormat
	}

	if !domain.IsValidPasswordLength(plainPassword) {
		return "", domain.ErrInvalidPasswordLength
	}

	user, hash, err := s.userRepo.GetCredentialsByEmail(ctx, u.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserDoesNotExist) {
			return "", err
		}
		return "", fmt.Errorf("get credentials by email: %w", err)
	}

	verified, err := s.passwordManager.Verify(hash, plainPassword)
	if err != nil {
		return "", fmt.Errorf("verify password: %w", err)
	}
	if !verified {
		return "", domain.ErrInvalidPassword
	}

	token, err := s.tokenGenerator.GenerateAuth(ctx, user.ID, user.Role)
	if err != nil {
		return "", fmt.Errorf("generate auth token: %w", err)
	}

	return token, nil
}
