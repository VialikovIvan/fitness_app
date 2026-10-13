package myrabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const StripeWebhooksQueue string = "stripe_webhooks"

func InitRabbitMQ(amqpAddr string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(amqpAddr)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	return conn, nil
}

func CloseRabbitMQ(conn *amqp.Connection) {
	if conn != nil {
		conn.Close()
	}
}

func SetupTopology(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare("stripe.dlx", "direct", true, false, false, false, nil)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare("stripe.dlq", true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind("stripe.dlq", "stripe.failed", "stripe.dlx", false, nil)
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(StripeWebhooksQueue, true, false, false, false, nil)
	return err
}
