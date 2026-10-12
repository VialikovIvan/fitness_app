package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	myrabbitmq "github.com/VialikovIvan/fitness_app/internal/infrastructure/rabbitmq"
	mbspDomain "github.com/VialikovIvan/fitness_app/internal/membership/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrChannelClosed = errors.New("channel closed")

type StripeEventService interface {
	HandleEventCheckoutCompleted(ctx context.Context, event mbspDomain.WebhookEvent) error
	HandleEventCustomerSubscriptionDeleted(ctx context.Context, event mbspDomain.WebhookEvent) error
	HandleEventInvoicePaid(ctx context.Context, event mbspDomain.WebhookEvent) error
	HandleEventInvoicePaymentFailed(ctx context.Context, event mbspDomain.WebhookEvent) error
	HandleUnknownEvent(ctx context.Context, event mbspDomain.WebhookEvent)
}

type RabbitMQConsumer struct {
	conn    *amqp.Connection
	service StripeEventService
	logger  *slog.Logger
}

func NewRabbitMQConsumer(conn *amqp.Connection, svc StripeEventService, l *slog.Logger) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		conn:    conn,
		service: svc,
		logger:  l.With("layer", "transport"),
	}
}

func (c *RabbitMQConsumer) RunStripeWebhookConsumer(ctx context.Context) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set Qos: %w", err)
	}

	deliveries, err := ch.ConsumeWithContext(ctx, myrabbitmq.StripeWebhooksQueue, "stripe-webhook-consumer", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-deliveries:
			if !ok {
				return ErrChannelClosed
			}

			if err := c.processStripeWebhook(ctx, msg.Body); err != nil {
				c.logger.Error("failed to process a stripe webhook", "err", err)
				msg.Nack(false, false)
				continue
			}

			msg.Ack(false)
		}
	}
}

func (c *RabbitMQConsumer) processStripeWebhook(ctx context.Context, body []byte) error {
	var event mbspDomain.WebhookEvent

	if err := json.Unmarshal(body, &event); err != nil {
		return err
	}

	switch event.Type {
	case mbspDomain.EventCheckoutCompleted:
		return c.service.HandleEventCheckoutCompleted(ctx, event)
	case mbspDomain.EventCustomerSubscriptionDeleted:
		return c.service.HandleEventCustomerSubscriptionDeleted(ctx, event)
	case mbspDomain.EventInvoicePaid:
		return c.service.HandleEventInvoicePaid(ctx, event)
	case mbspDomain.EventInvoicePaymentFailed:
		return c.service.HandleEventInvoicePaymentFailed(ctx, event)
	default:
		c.service.HandleUnknownEvent(ctx, event)
		return nil
	}
}
