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

## Stress-testing

### 1. Modify Register() and Login() in `internal/user/transport/handler.go`

Set the Secure flag in cookie settings to false

```Go
cookie := http.Cookie{
		Name:     "authToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(cookieValidityTime),
	}
```

> ⚠️ **Remember to revert this change.** Leaving `Secure: false` in production 
> allows cookies to be sent over plain HTTP, exposing auth tokens.

### 2. Run the app and wait until `docker compose ps` shows all services as `healthy`.

```bash
docker compose up -d
```

### 3. Run the stress test in a separate terminal

```bash
docker run --rm -i --network=fitness_app_default grafana/k6 run - < ./loadtest/stress-test.js
```

### 4. View results

- **K6 Summary** — printed in the terminal after the test
- **Grafana** — `http://localhost:3000` → dashboards

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
