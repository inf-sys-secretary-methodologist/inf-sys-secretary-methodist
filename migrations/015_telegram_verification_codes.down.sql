-- ============================================================================
-- DROP TELEGRAM VERIFICATION CODES
-- ============================================================================

DROP FUNCTION IF EXISTS cleanup_expired_telegram_codes();
DROP TABLE IF EXISTS telegram_verification_codes;
