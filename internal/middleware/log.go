package middleware

import (
	"net/http"
	"strconv"
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

		m.metrics.httpRequestsProcessing.Inc()

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

		defer func() {
			duration := time.Since(start)

			m.logger.Info("request processed",
				"traceID", traceID,
				"duration", duration.Milliseconds(),
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
			)

			m.metrics.httpRequestDuration.Observe(r.Pattern, duration.Seconds())
			m.metrics.httpRequestsTotal.Inc(r.Pattern, strconv.Itoa(sw.status))
			m.metrics.httpRequestsProcessing.Dec()
		}()

		next.ServeHTTP(sw, r)
	})
}
