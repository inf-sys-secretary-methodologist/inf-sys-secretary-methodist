-- ============================================================================
-- NOTIFICATIONS MODULE - Rollback
-- ============================================================================

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_slack_connections_updated_at ON user_slack_connections;
DROP TRIGGER IF EXISTS trigger_telegram_connections_updated_at ON user_telegram_connections;
DROP TRIGGER IF EXISTS trigger_notification_preferences_updated_at ON notification_preferences;
DROP TRIGGER IF EXISTS trigger_notifications_updated_at ON notifications;

-- Drop function
DROP FUNCTION IF EXISTS update_notification_updated_at();

-- Restore original constraint on event_reminders
ALTER TABLE event_reminders
    DROP CONSTRAINT IF EXISTS event_reminders_reminder_type_check;

ALTER TABLE event_reminders
    ADD CONSTRAINT event_reminders_reminder_type_check
    CHECK (reminder_type IN ('email', 'push', 'in_app'));

-- Drop tables
DROP TABLE IF EXISTS notification_delivery_log;
DROP TABLE IF EXISTS user_slack_connections;
DROP TABLE IF EXISTS user_telegram_connections;
DROP TABLE IF EXISTS notification_preferences;
DROP TABLE IF EXISTS notifications;

-- Drop types
DROP TYPE IF EXISTS notification_priority;
DROP TYPE IF EXISTS notification_type;
