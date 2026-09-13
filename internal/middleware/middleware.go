package middleware

import (
	"context"
	"log/slog"
	"sync"
)

type Limiter interface {
	AllowRequestByIP(ctx context.Context, ip string) (bool, error)
}

type Manager struct {
	logger    *slog.Logger
	limiter   Limiter
	muMap     map[string]*sync.Mutex
	muMapLock sync.Mutex
}

func New(l *slog.Logger, lim Limiter) *Manager {
	return &Manager{
		logger:  l,
		limiter: lim,
		muMap: make(map[string]*sync.Mutex),
	}
}
