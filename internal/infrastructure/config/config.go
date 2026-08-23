package config

import (
	"errors"
	"os"
)

var ErrDBURLVariableNotFound = errors.New("expected .env variable not found: DATABASE_URL")

type Config struct {
	DBURL     string
	Port      string
}

func NewConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	if dbURL == "" {
		return nil, ErrDBURLVariableNotFound
	}

	return &Config{
		DBURL:     dbURL,
		Port:      port,
	}, nil
}
