package myprometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestsCounter          *HTTPRequestsCounter
	HTTPRequestDurationHistogram *HTTPRequestDurationHistogram
	HTTPRequestsInFlightGauge    *HTTPRequestsInFlightGauge
	CheckoutsStartedCounter      *CheckoutsStartedCounter
	CheckoutsCompletedCounter    *CheckoutsCompletedCounter
}

func InitMetrics(reg *prometheus.Registry) *Metrics {
	return &Metrics{
		HTTPRequestsCounter:          newHttpRequestsCounter(reg),
		HTTPRequestDurationHistogram: newHttpRequestDurationHistogram(reg),
		HTTPRequestsInFlightGauge:    newHttpRequestsInFlightGauge(reg),
		CheckoutsStartedCounter:      newCheckoutsStartedCounter(reg),
		CheckoutsCompletedCounter:    newCheckoutsCompletedCounter(reg),
	}
}

type HTTPRequestsCounter struct {
	vec *prometheus.CounterVec
}

func newHttpRequestsCounter(reg *prometheus.Registry) *HTTPRequestsCounter {
	return &HTTPRequestsCounter{
		vec: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests processed.",
			},
			[]string{"pattern", "status"},
		),
	}
}

func (c *HTTPRequestsCounter) Inc(pattern, status string) {
	c.vec.WithLabelValues(pattern, status).Inc()
}

type HTTPRequestDurationHistogram struct {
	vec *prometheus.HistogramVec
}

func newHttpRequestDurationHistogram(reg *prometheus.Registry) *HTTPRequestDurationHistogram {
	return &HTTPRequestDurationHistogram{
		vec: promauto.With(reg).NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds.",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"pattern"},
		),
	}
}

func (h *HTTPRequestDurationHistogram) Observe(pattern string, durationSeconds float64) {
	h.vec.WithLabelValues(pattern).Observe(durationSeconds)
}

type HTTPRequestsInFlightGauge struct {
	gauge prometheus.Gauge
}

func newHttpRequestsInFlightGauge(reg *prometheus.Registry) *HTTPRequestsInFlightGauge {
	return &HTTPRequestsInFlightGauge{
		gauge: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "A number of currently active HTTP requests.",
			},
		),
	}
}

func (g *HTTPRequestsInFlightGauge) Inc() {
	g.gauge.Inc()
}

func (g *HTTPRequestsInFlightGauge) Dec() {
	g.gauge.Dec()
}

type CheckoutsStartedCounter struct {
	vec *prometheus.CounterVec
}

func newCheckoutsStartedCounter(reg *prometheus.Registry) *CheckoutsStartedCounter {
	return &CheckoutsStartedCounter{
		vec: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "checkouts_started_total",
				Help: "Total number of started checkouts.",
			},
			[]string{"plan_id"},
		),
	}
}

func (c *CheckoutsStartedCounter) Inc(planID string) {
	c.vec.WithLabelValues(planID).Inc()
}

type CheckoutsCompletedCounter struct {
	vec *prometheus.CounterVec
}

func newCheckoutsCompletedCounter(reg *prometheus.Registry) *CheckoutsCompletedCounter {
	return &CheckoutsCompletedCounter{
		vec: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Name: "checkouts_completed_total",
				Help: "Total number of completed checkouts.",
			},
			[]string{"plan_id"},
		),
	}
}

func (c *CheckoutsCompletedCounter) Inc(planID string) {
	c.vec.WithLabelValues(planID).Inc()
}
