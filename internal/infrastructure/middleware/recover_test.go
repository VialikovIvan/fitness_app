package middleware_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/infrastructure/middleware"
)

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
			m := middleware.New(logger)

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
