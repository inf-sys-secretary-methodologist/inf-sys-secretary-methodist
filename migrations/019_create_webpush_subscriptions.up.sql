-- Web Push Subscriptions Table
-- Stores browser push notification subscriptions for users

CREATE TABLE IF NOT EXISTS webpush_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL UNIQUE,
    p256dh_key TEXT NOT NULL,
    auth_key TEXT NOT NULL,
    user_agent TEXT,
    device_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for efficient user subscription lookups
CREATE INDEX idx_webpush_subscriptions_user_id ON webpush_subscriptions(user_id);

-- Index for active subscriptions
CREATE INDEX idx_webpush_subscriptions_active ON webpush_subscriptions(user_id, is_active) WHERE is_active = true;

-- Trigger for auto-updating updated_at
CREATE OR REPLACE FUNCTION update_webpush_subscriptions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_webpush_subscriptions_updated_at
    BEFORE UPDATE ON webpush_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_webpush_subscriptions_updated_at();

COMMENT ON TABLE webpush_subscriptions IS 'Browser push notification subscriptions';
COMMENT ON COLUMN webpush_subscriptions.endpoint IS 'Unique push service endpoint URL';
COMMENT ON COLUMN webpush_subscriptions.p256dh_key IS 'P-256 Diffie-Hellman public key for encryption';
COMMENT ON COLUMN webpush_subscriptions.auth_key IS 'Authentication secret for message encryption';
COMMENT ON COLUMN webpush_subscriptions.user_agent IS 'Browser user agent string';
COMMENT ON COLUMN webpush_subscriptions.device_name IS 'User-friendly device/browser name';
COMMENT ON COLUMN webpush_subscriptions.is_active IS 'Whether the subscription is active';
COMMENT ON COLUMN webpush_subscriptions.last_used_at IS 'Last time a push was successfully sent';
