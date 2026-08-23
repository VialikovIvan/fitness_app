#!/bin/bash
set -e

if [ -f .env ]; then
    cp .env .env.bak
    echo "Existing .env backed up to .env.bak"
fi

cp .env.example .env
read -p "Enter your preferred database password: " DB_PASSWORD
sed -i "s|your_password_here|$DB_PASSWORD|g" .env
echo ".env configured! Proceeding to the next step.."

docker compose up --build -d
echo "Docker containers successfully loaded! Proceeding to the next step.."

docker compose exec app sh -c 'migrate -path ./migrations -database "$DATABASE_URL" up'
echo "Migrations successfully applied! Applications is running and ready to use."