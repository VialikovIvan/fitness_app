#!/bin/bash
set -e

if [ -f .env ]; then
    cp .env .env.bak
    echo "Existing .env backed up to .env.bak"
fi

cp .env.example .env
sed -i "s|your_jwt_secret_here|$(openssl rand -base64 32)|g" .env
echo "jwt secret set! Proceeding to the next step.."
read -p "Choose a database password: " DB_PASSWORD
sed -i "s|your_db_password_here|$DB_PASSWORD|g" .env
echo ".env configured! Proceeding to the next step.."

docker compose up --build -d
echo "Docker containers successfully loaded! Proceeding to the next step.."

docker compose exec app sh -c 'migrate -path ./migrations -database "$DATABASE_URL" up'
echo "Migrations successfully applied! Application is running."