CREATE UNIQUE INDEX idx_one_active_subscription
ON subscriptions (user_id)
WHERE status = 'active';