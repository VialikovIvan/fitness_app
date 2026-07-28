package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/VialikovIvan/fitness_app/internal/config"
	"github.com/VialikovIvan/fitness_app/internal/transport"
	"github.com/VialikovIvan/fitness_app/internal/repository"
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

	db, err := repository.InitDB(cfg.DBURL)
	if err != nil {
		log.Fatal("Database connection error: ", err)
	}

	defer repository.CloseDB(db)

	server := transport.NewServer(cfg.Port)

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server error:", slog.Any("err", err))
	}
}
