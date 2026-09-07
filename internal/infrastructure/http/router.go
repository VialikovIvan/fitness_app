package http

import (
	"net/http"

	health "github.com/VialikovIvan/fitness_app/internal/health/transport"
	user "github.com/VialikovIvan/fitness_app/internal/user/transport"
)

func SetupRouter(h *health.Handler, u *user.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.HandleHealth)
	mux.HandleFunc("POST /register", u.Register)
	mux.HandleFunc("POST /login", u.Login)
	mux.HandleFunc("POST /logout", u.Logout)

	return mux
}
