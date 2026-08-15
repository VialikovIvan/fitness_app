package http

import (
	"net/http"

	healthtransport "github.com/VialikovIvan/fitness_app/internal/health/transport"
)

func setupRouter(h *healthtransport.Handler) http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("GET /health", h.HealthHandler)

	return mux
}