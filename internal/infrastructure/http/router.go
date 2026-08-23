package http

import (
	"net/http"

	health "github.com/VialikovIvan/fitness_app/internal/health/transport"
)

func SetupRouter(h *health.Handler) http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /health", h.HandleHealth)

	return mux
}