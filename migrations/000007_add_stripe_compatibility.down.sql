ALTER TABLE plans
DROP COLUMN stripe_price_id,
DROP COLUMN stripe_product_id;

ALTER TABLE subscriptions
DROP COLUMN stripe_subscription_id;

ALTER TABLE users
DROP COLUMN stripe_customer_id;