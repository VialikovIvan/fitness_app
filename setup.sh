#!/bin/bash
set -e

if [ -f .env ]; then
    cp .env .env.bak
    echo "Existing .env backed up to .env.bak"
fi

if [ -f ./rabbitmq/definitions.json ]; then
    cp ./rabbitmq/definitions.json ./rabbitmq/definitions.json.bak
    echo "Existing definitions.json backed up to definitions.json.bak"
fi

if [ -f ./metrics/metrics-token ]; then
    cp ./metrics/metrics-token ./metrics/metrics-token.bak
    echo "Existing metrics-token backed up to metrics-token.bak"
fi

cp .env.example .env

read -p "Enter your Stripe secret key (If you don't have one yet, visit https://docs.stripe.com/keys): " STRIPE_SECRET_KEY
sed -i "s|your_stripe_secret_key_here|$STRIPE_SECRET_KEY|g" .env

sed -i "s|your_stripe_webhook_secret_here|$(stripe listen --print-secret)|g" .env
echo "Stripe webhook secret generated and set! Proceeding to the next step.."

sed -i "s|your_jwt_auth_secret_here|$(openssl rand -base64 32)|g" .env
echo "Random JWT auth secret generated and set! Proceeding to the next step.."

sed -i "s|your_jwt_access_secret_here|$(openssl rand -base64 32)|g" .env
echo "Random JWT access secret generated and set! Proceeding to the next step.."

sed -i "s|your_turnstile_api_key_here|$(openssl rand -base64 32)|g" .env
echo "Random turnstile API key generated and set! Proceeding to the next step.."

METRICS_API_KEY=$(openssl rand -base64 32)
sed -i "s|your_metrics_api_key_here|$METRICS_API_KEY|g" .env
echo "Random metrics API key generated and set! Proceeding to the next step.."

read -p "Choose a database password: " DB_PASSWORD
sed -i "s|your_db_password_here|$DB_PASSWORD|g" .env

read -p "Choose a RabbitMQ password: " RABBIT_PASSWORD
sed -i "s|your_rabbitmq_password_here|$RABBIT_PASSWORD|g" .env

read -p "Choose a Grafana password: " GRAFANA_PASSWORD
sed -i "s|your_grafana_password_here|$GRAFANA_PASSWORD|g" .env

echo ".env configured! Proceeding to the next step.."

cp ./rabbitmq/definitions.json.example ./rabbitmq/definitions.json
sed -i "s|your_rabbitmq_password_here|$RABBIT_PASSWORD|g" ./rabbitmq/definitions.json
echo "definitions.json configured! Proceeding to the next step.."

cp ./metrics/metrics-token.example ./metrics/metrics-token
sed -i "s|your_metrics_api_key_here|$METRICS_API_KEY|g" .metrics/metrics-token
echo "metrics-token configured! Proceeding to the next step.."

echo "Open a new terminal tab and run the following command: 'stripe listen --forward-to localhost:8080/webhooks/stripe --all-snapshot'"
read -p "Press Enter once done..."

docker compose up --build -d
echo "Docker containers successfully loaded! Proceeding to the next step.."

docker compose exec app sh -c 'migrate -path ./migrations -database "$DATABASE_URL" up'
echo "Migrations successfully applied! Application is running."
