package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/user/domain"
	"github.com/VialikovIvan/fitness_app/internal/user/service"
)

type MockRepo struct {
	CreateFunc func(ctx context.Context, u domain.User, hash string) (int64, error)
}

func (m *MockRepo) Create(ctx context.Context, u domain.User, hash string) (int64, error) {
	return m.CreateFunc(ctx, u, hash)
}

type MockPWHasher struct {
	HashFunc func(ctx context.Context, plainPassword string) (string, error)
}

func (m *MockPWHasher) Hash(ctx context.Context, plainPassword string) (string, error) {
	return m.HashFunc(ctx, plainPassword)
}

type MockTKGenerator struct {
	GenerateFunc func(ctx context.Context, userID int64) (string, error)
}

func (m *MockTKGenerator) Generate(ctx context.Context, userID int64) (string, error) {
	return m.GenerateFunc(ctx, userID)
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name          string
		u             domain.User
		plainPassword string
		r             service.UserRepository
		h             service.PasswordHasher
		t             service.TokenGenerator
		want          string
		wantErr       bool
	}{
		{
			name: "successful registration",
			u: domain.User{
				FirstName:  "John",
				MiddleName: "Middle",
				LastName:   "Great",
				Email:      "valid.email@gmail.com",
			},
			plainPassword: "12345678",
			r: &MockRepo{
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (int64, error) {
					return 1, nil
				},
			},
			h: &MockPWHasher{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "aFakePasswordHash", nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64) (string, error) {
					return "aTokenToBeReturned", nil
				},
			},
			want:    "aTokenToBeReturned",
			wantErr: false,
		},
		{
			name: "hashing error",
			u: domain.User{
				FirstName:  "John",
				MiddleName: "Middle",
				LastName:   "Great",
				Email:      "valid.email@gmail.com",
			},
			plainPassword: "12345678",
			r: &MockRepo{
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (int64, error) {
					return 1, nil
				},
			},
			h: &MockPWHasher{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "", errors.New("error")
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64) (string, error) {
					return "aTokenToBeReturned", nil
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "user already exists (database error)",
			u: domain.User{
				FirstName:  "John",
				MiddleName: "Middle",
				LastName:   "Great",
				Email:      "valid.email@gmail.com",
			},
			plainPassword: "12345678",
			r: &MockRepo{
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (int64, error) {
					return 0, domain.ErrUserAlreadyExists
				},
			},
			h: &MockPWHasher{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "aFakePasswordHash", nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64) (string, error) {
					return "aTokenToBeReturned", nil
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "token generation error",
			u: domain.User{
				FirstName:  "John",
				MiddleName: "Middle",
				LastName:   "Great",
				Email:      "valid.email@gmail.com",
			},
			plainPassword: "12345678",
			r: &MockRepo{
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (int64, error) {
					return 1, nil
				},
			},
			h: &MockPWHasher{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "aFakePasswordHash", nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64) (string, error) {
					return "", errors.New("error")
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(tt.r, tt.h, tt.t)
			got, err := s.Register(t.Context(), tt.u, tt.plainPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
