package service

import (
	"github.com/VialikovIvan/fitness_app/internal/domain"
)

func CheckHealth() domain.HealthResponse {
	return domain.HealthResponse{Status: "ok"}
}