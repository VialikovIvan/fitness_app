package myrabbitmq

import (
	"context"
	"encoding/json"
	"time"

	mbspDomain "github.com/VialikovIvan/fitness_app/internal/membership/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQPublisher struct {
	conn *amqp.Connection
}

func NewPublisher(conn *amqp.Connection) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		conn: conn,
	}
}

func (p *RabbitMQPublisher) PublishStripeWebhook(ctx context.Context, event mbspDomain.WebhookEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.PublishWithContext(ctx, "", StripeWebhooksQueue, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    event.ID,
		Timestamp:    time.Now(),
		Body:         body,
	})
}
