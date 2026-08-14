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

type HealthResponse struct {
	Status string `json:"status"`
}

func NewHandler(service HealthService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	_ = h.service.CheckHealth()

	resp := HealthResponse{
		Status: "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("Encoding error:", slog.Any("err", err))
		return
	}
}