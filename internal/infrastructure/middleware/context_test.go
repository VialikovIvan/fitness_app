package middleware_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/infrastructure/middleware"
)

func TestManager_EnrichContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	m := middleware.New(logger)

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
