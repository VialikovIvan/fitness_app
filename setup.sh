#!/bin/bash

if [ -f .env ]; then
    cp .env .env.bak
    echo "Existing .env backed up to .env.bak"
fi

cp .env.example .env

read -p "Enter your preferred database password: " DB_PASSWORD
sed -i "s/your_password_here/$DB_PASSWORD/g" .env

echo ".env configured! You may proceed to the next step now."