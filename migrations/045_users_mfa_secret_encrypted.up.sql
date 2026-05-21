-- ============================================================================
-- USERS — at-rest encryption marker for users.mfa_secret (v0.159.0 #279 ADR-4)
-- ============================================================================
-- Adds users.mfa_secret_encrypted boolean so the repository can tell whether
-- an existing mfa_secret row is the historical plaintext Base32 form or the
-- new AES-256-GCM ciphertext (base64-encoded nonce || sealed). Lazy migration
-- pattern: existing rows default to FALSE, the repository decrypts only when
-- both KEK is wired AND the column is TRUE. The first Save() under the KEK
-- rewraps the secret and flips the column to TRUE.
--
-- Side-stepping a forced backfill avoids a deploy step that would otherwise
-- require KEK access during a one-off rewrap script — diploma deployment has
-- zero MFA-enabled users at the time of cutover, so lazy migration is safe.
-- ============================================================================

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS mfa_secret_encrypted BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN users.mfa_secret_encrypted IS
    'v0.159.0 ADR-4 — TRUE when mfa_secret is AES-256-GCM ciphertext (base64), FALSE for legacy plaintext rows awaiting lazy rewrap on next Save under KEK';
