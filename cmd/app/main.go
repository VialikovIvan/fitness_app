package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VialikovIvan/fitness_app/internal/health"
	"github.com/VialikovIvan/fitness_app/internal/infrastructure/config"
	myhttp "github.com/VialikovIvan/fitness_app/internal/infrastructure/http"
	"github.com/VialikovIvan/fitness_app/internal/infrastructure/postgres"
	myrabbitmq "github.com/VialikovIvan/fitness_app/internal/infrastructure/rabbitmq"
	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	mystripe "github.com/VialikovIvan/fitness_app/internal/infrastructure/stripe"
	"github.com/VialikovIvan/fitness_app/internal/membership"
	"github.com/VialikovIvan/fitness_app/internal/middleware"
	"github.com/VialikovIvan/fitness_app/internal/pkg/jwt"
	"github.com/VialikovIvan/fitness_app/internal/user"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v86"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if _, err := os.Stat("../../.env"); err == nil {
		err := godotenv.Load("../../.env")
		if err != nil {
			logger.Error(".env file loading error", "err", err)
			os.Exit(1)
		}
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("config loading error", "err", err)
		os.Exit(1)
	}
	logger.Info("config loaded")

	db, err := postgres.InitDB(cfg.DBURL)
	if err != nil {
		logger.Error("database connection error", "err", err)
		os.Exit(1)
	}
	logger.Info("database connection established")

	defer postgres.CloseDB(db)

	rdb, err := myredis.InitRedis(cfg.RedisURL)
	if err != nil {
		logger.Error("redis connection error", "err", err)
		os.Exit(1)
	}
	logger.Info("redis connection established")

	defer myredis.CloseRedis(rdb)

	rbtmq, err := myrabbitmq.InitRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		logger.Error("rabbitMQ connection error", "err", err)
		os.Exit(1)
	}

	defer myrabbitmq.CloseRabbitMQ(rbtmq)

	if err := myrabbitmq.SetupTopology(rbtmq); err != nil {
		logger.Error("rabbitMQ topology setup error", "err", err)
		os.Exit(1)
	}

	sc := stripe.NewClient(cfg.StripeSecretKey)

	cacher := myredis.NewCacher(rdb)
	tokenValidator := jwt.New(cfg.JWTSecret)
	stripeManager := mystripe.New(sc, cfg.BaseURL, cfg.StripeWebhookSecret)
	rbtmqPublisher := myrabbitmq.NewPublisher(rbtmq)

	mwManager := middleware.New(logger, cacher, tokenValidator)

	healthHandler := health.New()
	membershipHandler, eventConsumer := membership.New(db, rdb, rbtmq, stripeManager, cacher, rbtmqPublisher, logger)
	userHandler := user.New(db, cfg.JWTSecret, logger)

	router := myhttp.NewRouter(mwManager, healthHandler, membershipHandler, userHandler)
	server := myhttp.NewServer(cfg.Port, router.Setup())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("event consumer started")
		if err := eventConsumer.RunStripeWebhookConsumer(ctx); err != nil {
			logger.Error("event consumer error", "err", err)
			os.Exit(1)
		}
	}()

	go func() {
		logger.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown forced", "err", err)

	}

	logger.Info("server successfully stopped")
}
