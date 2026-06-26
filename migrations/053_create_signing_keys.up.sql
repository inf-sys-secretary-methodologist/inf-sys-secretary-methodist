-- ============================================================================
-- SIGNING_KEYS — per-user ключи электронной подписи документов
-- ============================================================================
-- Initiative: #140 (cryptographic e-signature, PR3 of 6).
--
-- У каждого подписанта одна ECDSA P-256 пара ключей, обёрнутая в самоподписанный
-- X.509-сертификат. Приватный ключ НИКОГДА не хранится в открытом виде:
-- encrypted_private_key = AES-256-GCM(PEM приватного ключа) под KEK из env
-- DOC_SIGNING_ENC_KEY (тот же примитив, что MFA-секреты, Issue #279 ADR-4).
--
-- Один ключ на пользователя (UNIQUE user_id). ON DELETE CASCADE: удаление
-- пользователя уносит его ключ (исторические подписи самодостаточны — они
-- несут собственную копию сертификата в document_signatures).
-- ============================================================================

CREATE TABLE IF NOT EXISTS signing_keys (
    id                    BIGSERIAL    PRIMARY KEY,
    user_id               BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    algorithm             VARCHAR(32)  NOT NULL,
    certificate_pem       TEXT         NOT NULL,           -- самоподписанный X.509 (PEM)
    encrypted_private_key TEXT         NOT NULL,           -- AES-256-GCM(PEM приватного ключа), base64
    created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT signing_keys_algorithm_valid
        CHECK (algorithm IN ('ECDSA_P256_SHA256')),
    CONSTRAINT signing_keys_certificate_nonempty
        CHECK (length(btrim(certificate_pem)) > 0),
    CONSTRAINT signing_keys_private_key_nonempty
        CHECK (length(btrim(encrypted_private_key)) > 0)
);
