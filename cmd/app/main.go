package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/VialikovIvan/fitness_app/internal/infrastructure/config"
	myhttp "github.com/VialikovIvan/fitness_app/internal/infrastructure/http"
	"github.com/VialikovIvan/fitness_app/internal/infrastructure/postgres"
	"github.com/joho/godotenv"

	"github.com/VialikovIvan/fitness_app/internal/health"
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

	healthHandler := health.New()

	router := myhttp.SetupRouter(healthHandler)
	server := myhttp.NewServer(cfg.Port, router)

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server error:", slog.Any("err", err))
	}
}
