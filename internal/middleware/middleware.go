package middleware

import (
	"context"
	"log/slog"
	"sync"
)

const numLocks uint64 = 64

type Limiter interface {
	AllowRequestByIP(ctx context.Context, ip string) (bool, error)
}

type TokenValidator interface {
	ValidateAuth(ctx context.Context, tokenStr string) (int64, string, error)
}

type HTTPRequestsCounter interface {
	Inc(pattern, status string)
}

type HTTPRequestDurationHistogram interface {
	Observe(pattern string, durationSeconds float64)
}

type HTTPRequestsInFlightGauge interface {
	Inc()
	Dec()
}

type Metrics struct {
	httpRequestsTotal      HTTPRequestsCounter
	httpRequestDuration    HTTPRequestDurationHistogram
	httpRequestsProcessing HTTPRequestsInFlightGauge
}

func NewMetricsManager(hrc HTTPRequestsCounter, hrdh HTTPRequestDurationHistogram, hrp HTTPRequestsInFlightGauge) *Metrics {
	return &Metrics{
		httpRequestsTotal:      hrc,
		httpRequestDuration:    hrdh,
		httpRequestsProcessing: hrp,
	}
}

type Manager struct {
	logger         *slog.Logger
	limiter        Limiter
	tokenValidator TokenValidator
	metrics        *Metrics
	locks          [numLocks]sync.Mutex
}

func New(l *slog.Logger, lim Limiter, t TokenValidator, m *Metrics) *Manager {
	return &Manager{
		logger:         l,
		limiter:        lim,
		tokenValidator: t,
		metrics:        m,
	}
}
