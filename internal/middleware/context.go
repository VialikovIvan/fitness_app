package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
)

const (
	httpRequestTimeout = time.Second*15
)

func (m *Manager) EnrichContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "traceID", generateTraceID())
		ctx, cancel := context.WithTimeout(ctx, httpRequestTimeout)
		defer cancel()

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func generateTraceID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func GetTraceID(r *http.Request) (string, bool) {
	rawTraceID := r.Context().Value("traceID")
	if rawTraceID == nil {
		return "", false
	}
	traceID, ok := rawTraceID.(string)
	return traceID, ok
}
