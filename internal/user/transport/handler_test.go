package transport_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/user/domain"
	"github.com/VialikovIvan/fitness_app/internal/user/transport"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type MockService struct {
	RegisterFunc func(ctx context.Context, u domain.User, plainPassword string) (string, error)
	LoginFunc    func(ctx context.Context, u domain.User, plainPassword string) (string, error)
}

func (m *MockService) Register(ctx context.Context, u domain.User, plainPassword string) (string, error) {
	return m.RegisterFunc(ctx, u, plainPassword)
}

func (m *MockService) Login(ctx context.Context, u domain.User, plainPassword string) (string, error) {
	return m.LoginFunc(ctx, u, plainPassword)
}

func TestHandler_Register(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		s           transport.AuthService
		wantJSON    string // Keep empty unless expect JSON response
		wantText    string // Keep empty unless expect text response
		wantCookies int
		wantStatus  int
	}{
		{
			name: "successfully returns token",
			body: `{"first_name":"John","middle_name":"Middle","last_name":"Great","password":"12345678","email":"valid.email@gmail.com"}`,
			s: &MockService{
				RegisterFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "aTokenToBeReturned", nil
				},
			},
			wantJSON:    `{"token":"aTokenToBeReturned"}`,
			wantCookies: 1,
			wantStatus:  http.StatusCreated,
		},
		{
			name: "malformed json error",
			body: `{"first_name":"John",middle_name":"Middle","last_name":"Great","password":"12345678","email":"valid.email@gmail.com"}`,
			s: &MockService{
				RegisterFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "", errors.New("error")
				},
			},
			wantText:    "Invalid JSON payload",
			wantCookies: 0,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "validation error",
			body: `{"first_name":"John","middle_name":"Middle","last_name":"Great","password":"12345678","email":"invalid.email.gmail.com"}`,
			s: &MockService{
				RegisterFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "", errors.New("error")
				},
			},
			wantText:    "Invalid data",
			wantCookies: 0,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "user already exists",
			body: `{"first_name":"John","middle_name":"Middle","last_name":"Great","password":"12345678","email":"valid.email@gmail.com"}`,
			s: &MockService{
				RegisterFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "", domain.ErrUserAlreadyExists
				},
			},
			wantText:    domain.ErrUserAlreadyExists.Error(),
			wantCookies: 0,
			wantStatus:  http.StatusConflict,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := transport.New(tt.s, validator.New(), logger)

			req := httptest.NewRequest("POST", "/register", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.Register(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}

			if len(rr.Result().Cookies()) != tt.wantCookies {
				t.Errorf("cookies: got = %v, want %v", len(rr.Result().Cookies()), tt.wantCookies)
			}

			if tt.wantJSON != "" {
				assert.JSONEq(t, tt.wantJSON, rr.Body.String())
			} else if tt.wantText != "" {
				assert.Equal(t, tt.wantText+"\n", rr.Body.String())
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		s           transport.AuthService
		wantJSON    string // Keep empty unless expect JSON response
		wantText    string // Keep empty unless expect text response
		wantCookies int
		wantStatus  int
	}{
		{
			name: "successfully returns token",
			body: `{"password":"12345678","email":"valid.email@gmail.com"}`,
			s: &MockService{
				LoginFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "aTokenToBeReturned", nil
				},
			},
			wantJSON:    `{"token":"aTokenToBeReturned"}`,
			wantCookies: 1,
			wantStatus:  http.StatusOK,
		},
		{
			name: "invalid password",
			body: `{"password":"12345678","email":"valid.email@gmail.com"}`,
			s: &MockService{
				LoginFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "", domain.ErrInvalidPassword
				},
			},
			wantText:    domain.ErrInvalidPassword.Error(),
			wantCookies: 0,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name: "malformed json error",
			body: `{"password":"12345678",email":"valid.email@gmail.com"}`,
			s: &MockService{
				LoginFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "", errors.New("error")
				},
			},
			wantText:    "Invalid JSON payload",
			wantCookies: 0,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "validation error",
			body: `{"password":"12345678","email":"invalid.email.gmail.com"}`,
			s: &MockService{
				LoginFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "", errors.New("error")
				},
			},
			wantText:    "Invalid data",
			wantCookies: 0,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name: "user does not exist",
			body: `{"password":"12345678","email":"valid.email@gmail.com"}`,
			s: &MockService{
				LoginFunc: func(ctx context.Context, u domain.User, plainPassword string) (string, error) {
					return "", domain.ErrUserDoesNotExist
				},
			},
			wantText:    domain.ErrUserDoesNotExist.Error(),
			wantCookies: 0,
			wantStatus:  http.StatusUnauthorized,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := transport.New(tt.s, validator.New(), logger)

			req := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			h.Login(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}

			if len(rr.Result().Cookies()) != tt.wantCookies {
				t.Errorf("cookies: got = %v, want %v", len(rr.Result().Cookies()), tt.wantCookies)
			}

			if tt.wantJSON != "" {
				assert.JSONEq(t, tt.wantJSON, rr.Body.String())
			} else if tt.wantText != "" {
				assert.Equal(t, tt.wantText+"\n", rr.Body.String())
			}
		})
	}
}
