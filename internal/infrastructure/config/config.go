package config

import (
	"errors"
	"os"
)

var (
	ErrDBURLVariableNotFound               = errors.New("expected .env variable not found: DATABASE_URL")
	ErrJWTAuthSecretVariableNotFound       = errors.New("expected .env variable not found: JWT_AUTH_SECRET")
	ErrJWTAccessSecretVariableNotFound     = errors.New("expected .env variable not found: JWT_ACCESS_SECRET")
	ErrStripeSecretKeyVariableNotFound     = errors.New("expected .env variable not found: STRIPE_SECRET_KEY")
	ErrStripeWebhookSecretVariableNotFound = errors.New("expected .env variable not found: STRIPE_WEBHOOK_SECRET")
	ErrRedisURLVariableNotFound            = errors.New("expected .env variable not found: REDIS_URL")
	ErrBaseURLVariableNotFound             = errors.New("expected .env variable not found: BASE_URL")
	ErrRabbitMQURLVariableNotFound         = errors.New("expected .env variable not found: RABBITMQ_URL")
	ErrTurnstileAPIKeyVariableNotFound     = errors.New("expected .env variable not found: TURNSTILE_API_KEY")
	ErrMetricsAPIKeyVariableNotFound       = errors.New("expected .env variable not found: METRICS_API_KEY")
	ErrGrafanaPasswordVariableNotFound     = errors.New("expected .env variable not found: GF_SECURITY_ADMIN_PASSWORD")
)

type Config struct {
	DBURL               string
	Port                string
	JWTAuthSecret       string
	JWTAccessSecret     string
	RedisURL            string
	StripeSecretKey     string
	StripeWebhookSecret string
	BaseURL             string
	RabbitMQURL         string
	TurnstileAPIKey     string
	MetricsAPIKey       string
	GrafanaPassword     string
}

func NewConfig() (*Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	jwtAuthSecret := os.Getenv("JWT_AUTH_SECRET")
	jwtAccessSecret := os.Getenv("JWT_ACCESS_SECRET")
	redisURL := os.Getenv("REDIS_URL")
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	baseURL := os.Getenv("BASE_URL")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	turnstileAPIKey := os.Getenv("TURNSTILE_API_KEY")
	metricsAPIKey := os.Getenv("METRICS_API_KEY")
	grafanaPassword := os.Getenv("GF_SECURITY_ADMIN_PASSWORD")

	if port == "" {
		port = "8080"
	}

	switch {
	case dbURL == "":
		return nil, ErrDBURLVariableNotFound
	case jwtAuthSecret == "":
		return nil, ErrJWTAuthSecretVariableNotFound
	case jwtAccessSecret == "":
		return nil, ErrJWTAccessSecretVariableNotFound
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
	case turnstileAPIKey == "":
		return nil, ErrTurnstileAPIKeyVariableNotFound
	case metricsAPIKey == "":
		return nil, ErrMetricsAPIKeyVariableNotFound
	case grafanaPassword == "":
		return nil, ErrGrafanaPasswordVariableNotFound
	}

	return &Config{
		DBURL:               dbURL,
		Port:                port,
		JWTAuthSecret:       jwtAuthSecret,
		JWTAccessSecret:     jwtAccessSecret,
		RedisURL:            redisURL,
		StripeSecretKey:     stripeSecretKey,
		StripeWebhookSecret: stripeWebhookSecret,
		BaseURL:             baseURL,
		RabbitMQURL:         rabbitMQURL,
		TurnstileAPIKey:     turnstileAPIKey,
		MetricsAPIKey:       metricsAPIKey,
		GrafanaPassword:     grafanaPassword,
	}, nil
}
