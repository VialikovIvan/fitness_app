package middleware

import "log/slog"

type Manager struct {
	logger *slog.Logger
}

func New(l *slog.Logger) *Manager {
	return &Manager{
		logger: l,
	}
}