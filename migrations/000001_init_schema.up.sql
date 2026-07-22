CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE plans (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    duration_days INTEGER NOT NULL CHECK (duration_days > 0),
    price DECIMAL(10, 2) NOT NULL CHECK (price > 0),
    currency CHAR(3) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE TABLE subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE NO ACTION,
    plan_id BIGINT NOT NULL REFERENCES plans(id) ON DELETE NO ACTION,
    status VARCHAR(30) NOT NULL,
    current_period_start TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMP WITH TIME ZONE NOT NULL,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    canceled_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT check_subscription_period
        CHECK (current_period_end > current_period_start)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE NO ACTION,
    provider VARCHAR(100) NOT NULL,
    provider_payment_id TEXT NOT NULL UNIQUE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount >= 0),
    currency CHAR(3) NOT NULL,
    status VARCHAR(30) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_payments_subscription_id ON payments(subscription_id);

CREATE TABLE qr_codes (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    token UUID NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_qr_codes_subscription_id ON qr_codes(subscription_id);