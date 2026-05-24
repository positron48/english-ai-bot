ALTER TABLE users ADD COLUMN IF NOT EXISTS subscription_tier TEXT NOT NULL DEFAULT 'free';

CREATE INDEX IF NOT EXISTS idx_users_subscription_tier ON users(subscription_tier);
