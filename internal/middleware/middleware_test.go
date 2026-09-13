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
			m := middleware.New(logger, &MockLimiter{})

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
	m := middleware.New(logger, &MockLimiter{})

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
	m := middleware.New(logger, &MockLimiter{})

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
				})
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
				})
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
				})
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
