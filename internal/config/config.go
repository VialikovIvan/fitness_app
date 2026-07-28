package config

import (
	"errors"
	"os"
)

var ErrEnvVariableNotFound = errors.New("expected .env variable not found")

type Config struct {
	DBURL string
	Port  string
}

func NewConfig() (*Config, error) {
	db_url := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	if db_url == "" {
		return nil, ErrEnvVariableNotFound
	}

	return &Config{
		DBURL: db_url,
		Port:  port,
	}, nil
}
