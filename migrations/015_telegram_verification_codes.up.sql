-- ============================================================================
-- TELEGRAM VERIFICATION CODES - For secure linking of Telegram accounts
-- ============================================================================

-- Verification codes for linking Telegram accounts
CREATE TABLE IF NOT EXISTS telegram_verification_codes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(8) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for looking up codes
CREATE INDEX idx_telegram_verification_codes_code ON telegram_verification_codes(code) WHERE used_at IS NULL;
CREATE INDEX idx_telegram_verification_codes_user_id ON telegram_verification_codes(user_id);
CREATE INDEX idx_telegram_verification_codes_expires ON telegram_verification_codes(expires_at) WHERE used_at IS NULL;

-- Clean up expired codes (can be run periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_telegram_codes()
RETURNS void AS $$
BEGIN
    DELETE FROM telegram_verification_codes
    WHERE expires_at < NOW() OR used_at IS NOT NULL;
END;
$$ LANGUAGE plpgsql;
