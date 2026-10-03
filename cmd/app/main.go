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

	"github.com/VialikovIvan/fitness_app/internal/infrastructure/config"
	myhttp "github.com/VialikovIvan/fitness_app/internal/infrastructure/http"
	"github.com/VialikovIvan/fitness_app/internal/infrastructure/postgres"
	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	mystripe "github.com/VialikovIvan/fitness_app/internal/infrastructure/stripe"
	"github.com/VialikovIvan/fitness_app/internal/membership"
	"github.com/VialikovIvan/fitness_app/internal/middleware"
	"github.com/VialikovIvan/fitness_app/internal/pkg/jwt"
	"github.com/joho/godotenv"
	"github.com/stripe/stripe-go/v86"

	"github.com/VialikovIvan/fitness_app/internal/health"
	"github.com/VialikovIvan/fitness_app/internal/user"
)

const shutdownTimeout = 10 * time.Second

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if _, err := os.Stat("../../.env"); err == nil {
		err := godotenv.Load("../../.env")
		if err != nil {
			logger.Error(".env file loading error", slog.Any("err", err))
			os.Exit(1)
		}
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Error("config loading error", slog.Any("err", err))
		os.Exit(1)
	}
	logger.Info("config loaded")

	db, err := postgres.InitDB(cfg.DBURL)
	if err != nil {
		logger.Error("database connection error", slog.Any("err", err))
		os.Exit(1)
	}
	logger.Info("database connection established")

	defer postgres.CloseDB(db)

	rdb, err := myredis.InitRedis(cfg.RedisURL)
	if err != nil {
		logger.Error("redis connection error", slog.Any("err", err))
		os.Exit(1)
	}
	logger.Info("redis connection established")

	defer myredis.CloseRedis(rdb)

	sc := stripe.NewClient(cfg.StripeSecretKey)

	cacher := myredis.NewCacher(rdb)
	tokenValidator := jwt.New(cfg.JWTSecret)
	stripeManager := mystripe.New(sc, cfg.BaseURL, cfg.StripeWebhookSecret)

	mwManager := middleware.New(logger, cacher, tokenValidator)

	healthHandler := health.New()
	membershipHandler := membership.New(db, rdb, stripeManager, cacher, logger)
	userHandler := user.New(db, cfg.JWTSecret, logger)

	router := myhttp.NewRouter(mwManager, healthHandler, membershipHandler, userHandler)
	server := myhttp.NewServer(cfg.Port, router.Setup())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown forced", slog.Any("err", err))

	}

	logger.Info("server successfully stopped")
}
