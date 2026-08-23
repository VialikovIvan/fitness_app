package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type HealthService interface{
	CheckHealth() bool
}

type Handler struct {
	service HealthService
}

type healthResponse struct {
	Status string `json:"status"`
}

func New(svc HealthService) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	_ = h.service.CheckHealth()

	resp := healthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("Encoding error:", slog.Any("err", err))
		return
	}
}