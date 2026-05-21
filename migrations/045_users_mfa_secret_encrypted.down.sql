-- Rollback v0.159.0 #279 ADR-4 — drop users.mfa_secret_encrypted column.
-- Note: this does NOT decrypt rows whose mfa_secret column already holds
-- ciphertext. The rollback assumes either (a) the encrypted rows were
-- rewrapped to plaintext by an operator-run script before reverting OR
-- (b) the deployment had no encrypted rows yet (true on first apply, the
-- common case for the diploma cutover).
ALTER TABLE users DROP COLUMN IF EXISTS mfa_secret_encrypted;
