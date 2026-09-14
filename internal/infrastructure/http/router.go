package http

import (
	"net/http"
	"slices"

	health "github.com/VialikovIvan/fitness_app/internal/health/transport"
	"github.com/VialikovIvan/fitness_app/internal/middleware"
	user "github.com/VialikovIvan/fitness_app/internal/user/transport"
)

type Router struct {
	mw     *middleware.Manager
	mux    *http.ServeMux
	health *health.Handler
	user   *user.Handler
}

func NewRouter(mw *middleware.Manager, h *health.Handler, u *user.Handler) *Router {
	return &Router{
		mw:     mw,
		mux:    http.NewServeMux(),
		health: h,
		user:   u,
	}
}

func (r *Router) Setup() http.Handler {
	r.public("GET /health", r.health.HandleHealth)

	r.guestOnly("POST /register", r.user.Register)
	r.guestOnly("POST /login", r.user.Login)

	r.authOnly("POST /logout", r.user.Logout)

	// Return globally wrapped
	return chainMiddlewares(r.mux, r.mw.Recovery, r.mw.EnrichContext, r.mw.TrackRequest, r.mw.RateLimit)
}

func (r *Router) authOnly(pattern string, h func(http.ResponseWriter, *http.Request)) {
	r.mux.Handle(pattern, r.mw.AuthOnly(http.HandlerFunc(h)))
}

func (r *Router) guestOnly(pattern string, h func(http.ResponseWriter, *http.Request)) {
	r.mux.Handle(pattern, r.mw.GuestOnly(http.HandlerFunc(h)))
}

func (r *Router) public(pattern string, h func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(pattern, h)
}

func chainMiddlewares(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range slices.Backward(middlewares) {
		handler = m(handler)
	}
	return handler
}
