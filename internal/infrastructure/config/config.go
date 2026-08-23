package config

import (
	"errors"
	"os"
)

var ErrDBURLVariableNotFound = errors.New("expected .env variable not found: DATABASE_URL")
var ErrJWTSecretVariableNotFound = errors.New("expected .env variable not found: JWT_SECRET")

type Config struct {
	DBURL     string
	Port      string
	JWTSecret string
}

func NewConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	if port == "" {
		port = "8080"
	}

	if dbURL == "" {
		return nil, ErrDBURLVariableNotFound
	} else if jwtSecret == "" {
		return nil, ErrJWTSecretVariableNotFound
	}

	return &Config{
		DBURL:     dbURL,
		Port:      port,
		JWTSecret: jwtSecret,
	}, nil
}
