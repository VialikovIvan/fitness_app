package config

import (
	"errors"
	"os"
)

var ErrDBURLVariableNotFound = errors.New("expected .env variable not found: DATABASE_URL")
var ErrJWTSecretVariableNotFound = errors.New("expected .env variable not found: JWT_SECRET")
var ErrRedisURLVariableNotFound = errors.New("expected .env variable not found: REDIS_URL")

type Config struct {
	DBURL     string
	Port      string
	JWTSecret string
	RedisURL  string
}

func NewConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	redisURL := os.Getenv("REDIS_URL")

	if port == "" {
		port = "8080"
	}

	if dbURL == "" {
		return nil, ErrDBURLVariableNotFound
	}

	if jwtSecret == "" {
		return nil, ErrJWTSecretVariableNotFound
	}

	if redisURL == "" {
		return nil, ErrRedisURLVariableNotFound
	}

	return &Config{
		DBURL:     dbURL,
		Port:      port,
		JWTSecret: jwtSecret,
		RedisURL:  redisURL,
	}, nil
}
