package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/user/domain"
	"github.com/VialikovIvan/fitness_app/internal/user/service"
)

type MockRepo struct {
	CreateFunc              func(ctx context.Context, u domain.User, hash string) (domain.User, error)
	GetCredentialsByEmailFunc func(ctx context.Context, email string) (domain.User, string, error)
}

func (m *MockRepo) Create(ctx context.Context, u domain.User, hash string) (domain.User, error) {
	return m.CreateFunc(ctx, u, hash)
}

func (m *MockRepo) GetCredentialsByEmail(ctx context.Context, email string) (domain.User, string, error) {
	return m.GetCredentialsByEmailFunc(ctx, email)
}

type MockPWManager struct {
	HashFunc   func(ctx context.Context, plainPassword string) (string, error)
	VerifyFunc func(encoded, providedPassword string) (bool, error)
}

func (m *MockPWManager) Hash(ctx context.Context, plainPassword string) (string, error) {
	return m.HashFunc(ctx, plainPassword)
}

func (m *MockPWManager) Verify(encoded, providedPassword string) (bool, error) {
	return m.VerifyFunc(encoded, providedPassword)
}

type MockTKGenerator struct {
	GenerateFunc func(ctx context.Context, userID int64, role string) (string, error)
}

func (m *MockTKGenerator) Generate(ctx context.Context, userID int64, role string) (string, error) {
	return m.GenerateFunc(ctx, userID, role)
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name          string
		u             domain.User
		plainPassword string
		r             service.UserRepository
		p             service.PasswordManager
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
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (domain.User, error) {
					return domain.User{
						ID: 1,
						Role: "user",
					}, nil
				},
			},
			p: &MockPWManager{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "aFakePasswordHash", nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
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
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (domain.User, error) {
					return domain.User{
						ID: 1,
						Role: "user",
					}, nil
				},
			},
			p: &MockPWManager{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "", errors.New("error")
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
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
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (domain.User, error) {
					return domain.User{}, domain.ErrUserAlreadyExists
				},
			},
			p: &MockPWManager{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "aFakePasswordHash", nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
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
				CreateFunc: func(ctx context.Context, u domain.User, hash string) (domain.User, error) {
					return domain.User{
						ID: 1,
						Role: "user",
					}, nil
				},
			},
			p: &MockPWManager{
				HashFunc: func(ctx context.Context, plainPassword string) (string, error) {
					return "aFakePasswordHash", nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
					return "", errors.New("error")
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(tt.r, tt.p, tt.t)
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

func TestService_Login(t *testing.T) {
	tests := []struct {
		name          string
		u             domain.User
		plainPassword string
		r             service.UserRepository
		p             service.PasswordManager
		t             service.TokenGenerator
		want          string
		wantErr       bool
	}{
		{
			name: "successful login",
			u: domain.User{
				Email: "valid.email@gmail.com",
			},
			plainPassword: "12345678",
			r: &MockRepo{
				GetCredentialsByEmailFunc: func(ctx context.Context, email string) (domain.User, string, error) {
					return domain.User{
						ID: 1,
						Role: "user",
					}, "aValidHash", nil
				},
			},
			p: &MockPWManager{
				VerifyFunc: func(encoded, providedPassword string) (bool, error) {
					return true, nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
					return "aTokenToBeReturned", nil
				},
			},
			want:    "aTokenToBeReturned",
			wantErr: false,
		},
		{
			name: "invalid password",
			u: domain.User{
				Email: "valid.email@gmail.com",
			},
			plainPassword: "invalidPassword",
			r: &MockRepo{
				GetCredentialsByEmailFunc: func(ctx context.Context, email string) (domain.User, string, error) {
					return domain.User{
						ID: 1,
						Role: "user",
					}, "aValidHash", nil
				},
			},
			p: &MockPWManager{
				VerifyFunc: func(encoded, providedPassword string) (bool, error) {
					return false, nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
					return "aValidToken", nil
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "user does not exist",
			u: domain.User{
				Email: "valid.email@gmail.com",
			},
			plainPassword: "invalidPassword",
			r: &MockRepo{
				GetCredentialsByEmailFunc: func(ctx context.Context, email string) (domain.User, string, error) {
					return domain.User{}, "", domain.ErrUserDoesNotExist
				},
			},
			p: &MockPWManager{
				VerifyFunc: func(encoded, providedPassword string) (bool, error) {
					return true, nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
					return "aValidToken", nil
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "token generation error",
			u: domain.User{
				Email: "valid.email@gmail.com",
			},
			plainPassword: "12345678",
			r: &MockRepo{
				GetCredentialsByEmailFunc: func(ctx context.Context, email string) (domain.User, string, error) {
					return domain.User{
						ID: 1,
						Role: "user",
					}, "aValidHash", nil
				},
			},
			p: &MockPWManager{
				VerifyFunc: func(encoded, providedPassword string) (bool, error) {
					return true, nil
				},
			},
			t: &MockTKGenerator{
				GenerateFunc: func(ctx context.Context, userID int64, role string) (string, error) {
					return "", errors.New("error")
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service.New(tt.r, tt.p, tt.t)
			got, err := s.Login(t.Context(), tt.u, tt.plainPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
