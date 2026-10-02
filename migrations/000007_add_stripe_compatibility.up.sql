ALTER TABLE plans
ADD COLUMN stripe_product_id VARCHAR(255) NOT NULL,
ADD COLUMN stripe_price_id VARCHAR(255) NOT NULL;

ALTER TABLE subscriptions
ADD COLUMN stripe_subscription_id VARCHAR(255) NOT NULL;

ALTER TABLE users
ADD COLUMN stripe_customer_id VARCHAR(255) NULL UNIQUE;