package middleware_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/middleware"
)

type MockLimiter struct {
	AllowRequestByIPFunc func(ctx context.Context, ip string) (bool, error)
}

func (m *MockLimiter) AllowRequestByIP(ctx context.Context, ip string) (bool, error) {
	return m.AllowRequestByIPFunc(ctx, ip)
}

type MockTokenValidator struct {
	ValidateAuthFunc func(ctx context.Context, tokenStr string) (int64, string, error)
}

func (m *MockTokenValidator) ValidateAuth(ctx context.Context, tokenStr string) (int64, string, error) {
	return m.ValidateAuthFunc(ctx, tokenStr)
}

func TestManager_Recovery(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.Handler
		wantStatus int
	}{
		{
			name: "no panic, returns 200",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			wantStatus: 200,
		},
		{
			name: "recovers from panic, returns 500",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("fake panic")
			}),
			wantStatus: 500,
		},
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := middleware.New(logger, &MockLimiter{}, &MockTokenValidator{})

			handler := m.Recovery(tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestManager_EnrichContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	m := middleware.New(logger, &MockLimiter{}, &MockTokenValidator{})

	handler := m.EnrichContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawTraceID := r.Context().Value("traceID")

		if rawTraceID == nil {
			t.Error("expected trace id to exist")
		}

		traceID, ok := rawTraceID.(string)
		if !ok {
			t.Error("expected trace id to be a string")
		}

		if traceID == "" {
			t.Error("expected trace id to exist")
		}

		if _, ok := r.Context().Deadline(); !ok {
			t.Error("expected context to have deadline")
		}
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
}

func TestManager_TrackRequest(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	m := middleware.New(logger, &MockLimiter{}, &MockTokenValidator{})

	handler := m.TrackRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequestWithContext(context.WithValue(context.Background(), "traceID", "test-trace-id"),
		"GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	logOutput := logBuffer.String()

	if !strings.Contains(logOutput, "request received") || !strings.Contains(logOutput, "request processed") {
		t.Error("expected both 'request received' and 'request processed' in logs")
	}

	if !strings.Contains(logOutput, "test-trace-id") {
		t.Error("expected trace id in logs")
	}
	if !strings.Contains(logOutput, "status=500") {
		t.Error("expected status 500 in logs")
	}
	if !strings.Contains(logOutput, "duration") {
		t.Error("expected duration in logs")
	}
}

func TestManager_RateLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		newManager func() *middleware.Manager
		wantStatus int
	}{
		{
			name: "not rate limited",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{
					AllowRequestByIPFunc: func(ctx context.Context, ip string) (bool, error) {
						return true, nil
					},
				}, &MockTokenValidator{})
			},
			wantStatus: 200,
		},
		{
			name: "rate limited",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{
					AllowRequestByIPFunc: func(ctx context.Context, ip string) (bool, error) {
						return false, nil
					},
				}, &MockTokenValidator{})
			},
			wantStatus: 429,
		},
		{
			name: "redis error",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{
					AllowRequestByIPFunc: func(ctx context.Context, ip string) (bool, error) {
						return true, errors.New("error")
					},
				}, &MockTokenValidator{})
			},
			wantStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.newManager()

			handler := m.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestManager_AuthOnly(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name            string
		newManager      func() *middleware.Manager
		cookieSetup     *http.Cookie
		wantStatus      int
		wantIDInContext int64
	}{
		{
			name: "authorized",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{}, &MockTokenValidator{
					ValidateAuthFunc: func(ctx context.Context, tokenStr string) (int64, string, error) {
						return 1, "user", nil
					},
				})
			},
			cookieSetup: &http.Cookie{
				Name:  "authToken",
				Value: "aTestToken",
			},
			wantStatus:      200,
			wantIDInContext: 1,
		},
		{
			name: "no cookie",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{}, &MockTokenValidator{
					ValidateAuthFunc: func(ctx context.Context, tokenStr string) (int64, string, error) {
						return 0, "", nil
					},
				})
			},
			wantStatus: 401,
		},
		{
			name: "validation error",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{}, &MockTokenValidator{
					ValidateAuthFunc: func(ctx context.Context, tokenStr string) (int64, string, error) {
						return 0, "", errors.New("error")
					},
				})
			},
			cookieSetup: &http.Cookie{
				Name:  "authToken",
				Value: "aTestToken",
			},
			wantStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.newManager()

			var gotID int64

			handler := m.AuthOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if raw := r.Context().Value("userID"); raw != nil {
					gotID = raw.(int64)
				}

				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.cookieSetup != nil {
				req.AddCookie(tt.cookieSetup)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if tt.wantIDInContext != 0 && gotID != tt.wantIDInContext {
				t.Errorf("ID in context: got = %v, want %v", gotID, tt.wantIDInContext)
			}

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestManager_GuestOnly(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name        string
		newManager  func() *middleware.Manager
		cookieSetup *http.Cookie
		wantStatus  int
	}{
		{
			name: "guest allowed (no cookie)",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{}, &MockTokenValidator{
					ValidateAuthFunc: func(ctx context.Context, tokenStr string) (int64, string, error) {
						return 1, "user", nil
					},
				})
			},
			wantStatus: 200,
		},
		{
			name: "guest allowed (invalid token)",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{}, &MockTokenValidator{
					ValidateAuthFunc: func(ctx context.Context, tokenStr string) (int64, string, error) {
						return 0, "", errors.New("error")
					},
				})
			},
			cookieSetup: &http.Cookie{
				Name:  "authToken",
				Value: "aTestToken",
			},
			wantStatus: 200,
		},
		{
			name: "blocked (valid token)",
			newManager: func() *middleware.Manager {
				return middleware.New(logger, &MockLimiter{}, &MockTokenValidator{
					ValidateAuthFunc: func(ctx context.Context, tokenStr string) (int64, string, error) {
						return 1, "user", nil
					},
				})
			},
			cookieSetup: &http.Cookie{
				Name:  "authToken",
				Value: "aTestToken",
			},
			wantStatus: 403,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.newManager()

			handler := m.GuestOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.cookieSetup != nil {
				req.AddCookie(tt.cookieSetup)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestManager_RequireAdmin(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		role       string // Fill in unless want status 500
		wantStatus int
	}{
		{
			name:       "access allowed",
			role:       "admin",
			wantStatus: 200,
		},
		{
			name:       "access denied",
			role:       "user",
			wantStatus: 403,
		},
		{
			name:       "no role in context",
			wantStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := middleware.New(logger, &MockLimiter{}, &MockTokenValidator{})

			handler := m.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			var req *http.Request
			if tt.role != "" {
				req = httptest.NewRequestWithContext(context.WithValue(context.Background(), "role", tt.role),
					"GET", "/test", nil)
			} else {
				req = httptest.NewRequest("GET", "/test", nil)
			}
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status code: got = %v, want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
