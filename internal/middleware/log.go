package middleware

import (
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func newStatusWriter(w http.ResponseWriter) *statusWriter {
	return &statusWriter{
		ResponseWriter: w,
		status:         0,
	}
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (sw *statusWriter) Write(b []byte) (int, error) {
	if sw.status == 0 {
		sw.status = http.StatusOK
	}
	return sw.ResponseWriter.Write(b)
}

func (m *Manager) TrackRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		traceID, ok := GetTraceID(r)
		if !ok {
			traceID = ""
		}

		m.logger.Info("request received",
			"traceID", traceID,
			"method", r.Method,
			"path", r.URL.Path,
		)

		sw := newStatusWriter(w)

		next.ServeHTTP(sw, r)

		durationMS := time.Since(start).Milliseconds()
		m.logger.Info("request processed",
			"traceID", traceID,
			"duration", durationMS,
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
		)
	})
}

func GetTraceID(r *http.Request) (string, bool) {
	rawTraceID := r.Context().Value("traceID")
	if rawTraceID == nil {
		return "", false
	}
	traceID, ok := rawTraceID.(string)
	return traceID, ok
}
