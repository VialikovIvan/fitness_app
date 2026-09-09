package middleware_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/VialikovIvan/fitness_app/internal/infrastructure/middleware"
)

func TestManager_TrackRequest(t *testing.T) {
	var logBuffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuffer, nil))
	m := middleware.New(logger)

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
