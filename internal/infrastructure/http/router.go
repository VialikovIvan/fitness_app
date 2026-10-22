package http

import (
	"net/http"
	"slices"

	health "github.com/VialikovIvan/fitness_app/internal/health/transport"
	membership "github.com/VialikovIvan/fitness_app/internal/membership/transport"
	"github.com/VialikovIvan/fitness_app/internal/middleware"
	qrcode "github.com/VialikovIvan/fitness_app/internal/qrcode/transport"
	user "github.com/VialikovIvan/fitness_app/internal/user/transport"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	mw              *middleware.Manager
	mux             *http.ServeMux
	health          *health.Handler
	user            *user.Handler
	membership      *membership.Handler
	qrcode          *qrcode.Handler
	turnstileAPIKey string
	metricsAPIKey   string
	reg             *prometheus.Registry
}

func NewRouter(mw *middleware.Manager, h *health.Handler, m *membership.Handler,
	u *user.Handler, qr *qrcode.Handler, tKey string, mKey string, reg *prometheus.Registry) *Router {
	return &Router{
		mw:              mw,
		mux:             http.NewServeMux(),
		health:          h,
		membership:      m,
		user:            u,
		qrcode:          qr,
		turnstileAPIKey: tKey,
		metricsAPIKey:   mKey,
		reg:             reg,
	}
}

func (r *Router) Setup() http.Handler {
	r.public("GET /health", r.health.HandleHealth)
	r.public("POST /webhooks/stripe", r.membership.HandleStripeWebhook)

	r.guestOnly("POST /register", r.user.Register)
	r.guestOnly("POST /login", r.user.Login)

	r.authOnly("POST /logout", r.user.Logout)
	r.authOnly("GET /plans", r.membership.GetPlans)
	r.authOnly("POST /checkout", r.membership.Checkout)
	r.authOnly("GET /qr", r.qrcode.GetQR)

	r.requireAdmin("POST /plans", r.membership.CreatePlan)

	r.turnstileOnly("GET /access", r.qrcode.VerifyAccess)

	r.metricsOnly("GET /metrics", promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{}).ServeHTTP)

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

func (r *Router) requireAdmin(pattern string, h func(http.ResponseWriter, *http.Request)) {
	r.mux.Handle(pattern, r.mw.AuthOnly(r.mw.RequireAdmin(http.HandlerFunc(h))))
}

func (r *Router) turnstileOnly(pattern string, h func(http.ResponseWriter, *http.Request)) {
	r.mux.Handle(pattern, r.mw.RequireBearer(http.HandlerFunc(h), r.turnstileAPIKey))
}

func (r *Router) metricsOnly(pattern string, h func(http.ResponseWriter, *http.Request)) {
	r.mux.Handle(pattern, r.mw.RequireBearer(http.HandlerFunc(h), r.metricsAPIKey))
}

func chainMiddlewares(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range slices.Backward(middlewares) {
		handler = m(handler)
	}
	return handler
}
