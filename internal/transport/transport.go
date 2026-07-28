package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/VialikovIvan/fitness_app/internal/service"
)

func NewServer(port string) *http.Server {
	router := setupRouter()

	server := &http.Server{
		Addr: ":"+port,
		Handler: router,
	}

	return server
}

func setupRouter() http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", healthHandler)

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	resp := service.CheckHealth()

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		slog.Error("Encoding error:", slog.Any("err", err))
		return
	}
}