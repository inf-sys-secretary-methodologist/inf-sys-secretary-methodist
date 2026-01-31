-- Rollback Web Push Subscriptions Table

DROP TRIGGER IF EXISTS trigger_webpush_subscriptions_updated_at ON webpush_subscriptions;
DROP FUNCTION IF EXISTS update_webpush_subscriptions_updated_at();
DROP INDEX IF EXISTS idx_webpush_subscriptions_active;
DROP INDEX IF EXISTS idx_webpush_subscriptions_user_id;
DROP TABLE IF EXISTS webpush_subscriptions;
