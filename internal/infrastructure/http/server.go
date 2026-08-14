package http

import (
	"net/http"

	healthtransport "github.com/VialikovIvan/fitness_app/internal/health/transport"
)

func NewServer(port string, h *healthtransport.Handler) *http.Server {
	router := setupRouter(h)

	server := &http.Server{
		Addr: ":"+port,
		Handler: router,
	}

	return server
}