-- ============================================================================
-- NOTIFICATIONS MODULE - In-app notifications and user preferences
-- ============================================================================

-- Notification types enum
DO $$ BEGIN
    CREATE TYPE notification_type AS ENUM (
        'info', 'success', 'warning', 'error',
        'reminder', 'task', 'document', 'event', 'system'
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Notification priority enum
DO $$ BEGIN
    CREATE TYPE notification_priority AS ENUM ('low', 'normal', 'high', 'urgent');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- In-app notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type notification_type NOT NULL DEFAULT 'info',
    priority notification_priority NOT NULL DEFAULT 'normal',
    title VARCHAR(500) NOT NULL,
    message TEXT NOT NULL,
    link VARCHAR(1000),
    image_url VARCHAR(1000),
    is_read BOOLEAN NOT NULL DEFAULT false,
    read_at TIMESTAMP,
    expires_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for notifications
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_user_is_read ON notifications(user_id, is_read);
CREATE INDEX idx_notifications_user_created ON notifications(user_id, created_at DESC);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_priority ON notifications(priority);
CREATE INDEX idx_notifications_expires_at ON notifications(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);

-- User notification preferences table
CREATE TABLE IF NOT EXISTS notification_preferences (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Channel preferences
    email_enabled BOOLEAN NOT NULL DEFAULT true,
    push_enabled BOOLEAN NOT NULL DEFAULT true,
    in_app_enabled BOOLEAN NOT NULL DEFAULT true,
    telegram_enabled BOOLEAN NOT NULL DEFAULT false,
    slack_enabled BOOLEAN NOT NULL DEFAULT false,

    -- Quiet hours
    quiet_hours_enabled BOOLEAN NOT NULL DEFAULT false,
    quiet_hours_start VARCHAR(5) NOT NULL DEFAULT '22:00',
    quiet_hours_end VARCHAR(5) NOT NULL DEFAULT '07:00',
    timezone VARCHAR(50) NOT NULL DEFAULT 'Europe/Moscow',

    -- Digest settings
    digest_enabled BOOLEAN NOT NULL DEFAULT false,
    digest_frequency VARCHAR(20) NOT NULL DEFAULT 'daily',
    digest_time VARCHAR(5) NOT NULL DEFAULT '09:00',

    -- Type-specific preferences (stored as JSONB)
    type_preferences JSONB,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for preferences
CREATE INDEX idx_notification_preferences_user_id ON notification_preferences(user_id);

-- Telegram connections for notifications
CREATE TABLE IF NOT EXISTS user_telegram_connections (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    telegram_chat_id BIGINT NOT NULL,
    telegram_username VARCHAR(255),
    telegram_first_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    connected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_telegram_connections_chat_id ON user_telegram_connections(telegram_chat_id);

-- Slack connections for notifications
CREATE TABLE IF NOT EXISTS user_slack_connections (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    slack_user_id VARCHAR(255) NOT NULL,
    slack_channel_id VARCHAR(255),
    slack_workspace_id VARCHAR(255),
    slack_username VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    connected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_slack_connections_user_id ON user_slack_connections(slack_user_id);

-- Notification delivery log (for tracking sent notifications across channels)
CREATE TABLE IF NOT EXISTS notification_delivery_log (
    id BIGSERIAL PRIMARY KEY,
    notification_id BIGINT REFERENCES notifications(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel VARCHAR(50) NOT NULL, -- 'email', 'push', 'telegram', 'slack', 'in_app'
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'sent', 'delivered', 'failed', 'bounced'
    external_id VARCHAR(255), -- external message ID from the channel
    error_message TEXT,
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_log_notification_id ON notification_delivery_log(notification_id);
CREATE INDEX idx_delivery_log_user_channel ON notification_delivery_log(user_id, channel);
CREATE INDEX idx_delivery_log_status ON notification_delivery_log(status);
CREATE INDEX idx_delivery_log_created_at ON notification_delivery_log(created_at DESC);

-- Add telegram reminder type to event_reminders if not exists
ALTER TABLE event_reminders
    DROP CONSTRAINT IF EXISTS event_reminders_reminder_type_check;

ALTER TABLE event_reminders
    ADD CONSTRAINT event_reminders_reminder_type_check
    CHECK (reminder_type IN ('email', 'push', 'in_app', 'telegram', 'slack'));

-- Update function for updated_at
CREATE OR REPLACE FUNCTION update_notification_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
DROP TRIGGER IF EXISTS trigger_notifications_updated_at ON notifications;
CREATE TRIGGER trigger_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_notification_updated_at();

DROP TRIGGER IF EXISTS trigger_notification_preferences_updated_at ON notification_preferences;
CREATE TRIGGER trigger_notification_preferences_updated_at
    BEFORE UPDATE ON notification_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_notification_updated_at();

DROP TRIGGER IF EXISTS trigger_telegram_connections_updated_at ON user_telegram_connections;
CREATE TRIGGER trigger_telegram_connections_updated_at
    BEFORE UPDATE ON user_telegram_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_notification_updated_at();

DROP TRIGGER IF EXISTS trigger_slack_connections_updated_at ON user_slack_connections;
CREATE TRIGGER trigger_slack_connections_updated_at
    BEFORE UPDATE ON user_slack_connections
    FOR EACH ROW
    EXECUTE FUNCTION update_notification_updated_at();
