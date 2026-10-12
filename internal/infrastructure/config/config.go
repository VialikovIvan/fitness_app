package config

import (
	"errors"
	"os"
)

var (
	ErrDBURLVariableNotFound               = errors.New("expected .env variable not found: DATABASE_URL")
	ErrJWTSecretVariableNotFound           = errors.New("expected .env variable not found: JWT_SECRET")
	ErrStripeSecretKeyVariableNotFound     = errors.New("expected .env variable not found: STRIPE_SECRET_KEY")
	ErrStripeWebhookSecretVariableNotFound = errors.New("expected .env variable not found: STRIPE_WEBHOOK_SECRET")
	ErrRedisURLVariableNotFound            = errors.New("expected .env variable not found: REDIS_URL")
	ErrBaseURLVariableNotFound             = errors.New("expected .env variable not found: BASE_URL")
	ErrRabbitMQURLVariableNotFound         = errors.New("expected .env variable not found: RABBITMQ_URL")
)

type Config struct {
	DBURL               string
	Port                string
	JWTSecret           string
	RedisURL            string
	StripeSecretKey     string
	StripeWebhookSecret string
	BaseURL             string
	RabbitMQURL         string
}

func NewConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")
	redisURL := os.Getenv("REDIS_URL")
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	baseURL := os.Getenv("BASE_URL")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")

	if port == "" {
		port = "8080"
	}

	switch {
	case dbURL == "":
		return nil, ErrDBURLVariableNotFound
	case jwtSecret == "":
		return nil, ErrJWTSecretVariableNotFound
	case redisURL == "":
		return nil, ErrRedisURLVariableNotFound
	case stripeSecretKey == "":
		return nil, ErrStripeSecretKeyVariableNotFound
	case stripeWebhookSecret == "":
		return nil, ErrStripeWebhookSecretVariableNotFound
	case baseURL == "":
		return nil, ErrBaseURLVariableNotFound
	case rabbitMQURL == "":
		return nil, ErrRabbitMQURLVariableNotFound
	}

	return &Config{
		DBURL:               dbURL,
		Port:                port,
		JWTSecret:           jwtSecret,
		RedisURL:            redisURL,
		StripeSecretKey:     stripeSecretKey,
		StripeWebhookSecret: stripeWebhookSecret,
		BaseURL:             baseURL,
		RabbitMQURL:         rabbitMQURL,
	}, nil
}
