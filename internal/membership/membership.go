package membership

import (
	"log/slog"

	myrabbitmq "github.com/VialikovIvan/fitness_app/internal/infrastructure/rabbitmq"
	myredis "github.com/VialikovIvan/fitness_app/internal/infrastructure/redis"
	mystripe "github.com/VialikovIvan/fitness_app/internal/infrastructure/stripe"
	membershipRepo "github.com/VialikovIvan/fitness_app/internal/membership/repository"
	"github.com/VialikovIvan/fitness_app/internal/membership/service"
	"github.com/VialikovIvan/fitness_app/internal/membership/transport"
	paymentRepo "github.com/VialikovIvan/fitness_app/internal/payment/repository"
	userRepo "github.com/VialikovIvan/fitness_app/internal/user/repository"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	myprometheus "github.com/VialikovIvan/fitness_app/internal/infrastructure/prometheus"
	amqp "github.com/rabbitmq/amqp091-go"
)

func New(db *pgxpool.Pool, conn *amqp.Connection, strp *mystripe.StripeManager,
	c *myredis.RedisCacher, p *myrabbitmq.RabbitMQPublisher, l *slog.Logger, m *myprometheus.Metrics) (*transport.Handler, *transport.RabbitMQConsumer) {
	l = l.With("feature", "membership")
	mr := membershipRepo.NewPostgres(db)
	ur := userRepo.NewPostgres(db)
	pmr := paymentRepo.NewPostgres(db)
	svcMetrics := service.NewMetricsManager(m.CheckoutsStartedCounter, m.CheckoutsCompletedCounter)
	svc := service.New(mr, mr, mr, ur, pmr, c, strp, p, l, svcMetrics)
	v := validator.New()
	handler := transport.NewHandler(svc, v, l)
	consumer := transport.NewRabbitMQConsumer(conn, svc, l)

	return handler, consumer
}
