-- Add multi-factor authentication (RFC 6238 TOTP) to users table.
-- mfa_secret stores the 32-character Base32 shared secret; mfa_enabled is the
-- enrollment flag flipped to true only after the user confirms the first
-- code in admin settings.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS mfa_secret VARCHAR(64),
    ADD COLUMN IF NOT EXISTS mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- Partial index — most users won't have MFA, so a sparse index is sufficient
-- for analytics ("how many admins enrolled?") without bloating storage.
CREATE INDEX IF NOT EXISTS idx_users_mfa_enabled
    ON users(mfa_enabled)
    WHERE mfa_enabled = TRUE;

COMMENT ON COLUMN users.mfa_secret IS 'Base32-encoded 160-bit TOTP secret (RFC 6238)';
COMMENT ON COLUMN users.mfa_enabled IS 'TRUE once user has confirmed their first TOTP code';
