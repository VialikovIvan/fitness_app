package health

import (
	"github.com/VialikovIvan/fitness_app/internal/health/service"
	"github.com/VialikovIvan/fitness_app/internal/health/transport"
)

func New() *transport.Handler {
	svc := service.New()
	handler := transport.New(svc)

	return handler
}