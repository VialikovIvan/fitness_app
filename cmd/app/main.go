package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/VialikovIvan/fitness_app/internal/infrastructure/config"
	"github.com/VialikovIvan/fitness_app/internal/infrastructure/postgres"
	myhttp "github.com/VialikovIvan/fitness_app/internal/infrastructure/http"

	healthservice "github.com/VialikovIvan/fitness_app/internal/health/service"
	healthtransport "github.com/VialikovIvan/fitness_app/internal/health/transport"
)

func main() {
	if _, err := os.Stat("../../.env"); err == nil {
		err := godotenv.Load("../../.env")
		if err != nil {
			log.Fatal(".env file loading error: ", err)
		}
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := postgres.InitDB(cfg.DBURL)
	if err != nil {
		log.Fatal("Database connection error: ", err)
	}

	defer postgres.CloseDB(db)

	service := healthservice.NewService()
	handler := healthtransport.NewHandler(service)

	server := myhttp.NewServer(cfg.Port, handler)

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server error:", slog.Any("err", err))
	}
}
