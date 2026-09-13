package middleware

import (
	"context"
	"log/slog"
	"sync"
)

type Limiter interface {
	AllowRequestByIP(ctx context.Context, ip string) (bool, error)
}

type TokenValidator interface {
	Validate(ctx context.Context, tokenStr string) (int64, error)
}

type Manager struct {
	logger         *slog.Logger
	limiter        Limiter
	tokenValidator TokenValidator
	muMap          map[string]*sync.Mutex
	muMapLock      sync.Mutex
}

func New(l *slog.Logger, lim Limiter, t TokenValidator) *Manager {
	return &Manager{
		logger:         l,
		limiter:        lim,
		tokenValidator: t,
		muMap:          make(map[string]*sync.Mutex),
	}
}
