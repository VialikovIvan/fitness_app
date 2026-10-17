# Fitness Club App

An application for managing fitness club access with integrated payment system.

## Tech Stack

- Go 1.26
- PostgreSQL 16
- Redis 7
- Docker
- Stripe API
- RabbitMQ 3.13
- Prometheus 3.15
- Grafana 13.2

## Requirements

- Docker
- [Stripe CLI](https://docs.stripe.com/cli)
- Be logged into Stripe:
```bash
stripe login
```

## First start (localhost)

### 1. Clone the repository.

```bash
git clone https://github.com/VialikovIvan/fitness_app.git
cd fitness_app
```
### 2. Run the setup script

Linux/Mac/Windows(Git Bash):

```bash
chmod +x setup.sh
./setup.sh
```

## Following starts

```bash
docker compose up -d
```

In a separate terminal:

```bash
stripe listen --forward-to localhost:8080/webhooks/stripe --all-snapshot
```

## Ports

| Service | Port |
|---------|------|
| App | 8080 |
| Postgres | 5432 |
| Redis | 6379 |
| RabbitMQ AMQP | 5672 |
| RabbitMQ UI | 15672 |
| Prometheus | 9090 |
| Grafana | 3000 |
