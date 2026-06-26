-- ============================================================================
-- DOCUMENT_SIGNATURES — криптографические электронные подписи документов
-- ============================================================================
-- Initiative: #140 (cryptographic e-signature, PR2 of 6).
--
-- Каждая строка — одна ЭЦП конкретной ВЕРСИИ документа конкретным подписантом.
-- Подпись = ECDSA P-256 над SHA-256-дайджестом канонического payload
-- (см. internal/modules/documents/domain/entities/document_signature.go,
-- ComputeSigningDigest). Храним сырую подпись (DER), X.509-сертификат
-- подписанта (PEM) и сам подписанный дайджест, чтобы verify был возможен
-- независимо от любого серверного ключа.
--
-- document_version денормализован: verify сверяет его с текущей версией
-- документа, чтобы поймать изменение тела после подписи. Документ может нести
-- несколько подписей (несколько подписантов / переподпись новой версии),
-- поэтому уникальность по подписанту НЕ навязывается на уровне БД.
--
-- Defense-in-depth: каждый доменный инвариант из NewDocumentSignature
-- продублирован CHECK-ом. signature_algorithm CHECK совпадает byte-for-byte
-- с доменной константой SignatureAlgorithmECDSAP256SHA256.
-- ============================================================================

CREATE TABLE IF NOT EXISTS document_signatures (
    id                  BIGSERIAL    PRIMARY KEY,

    -- Что подписано. ON DELETE CASCADE: удаление документа уносит его подписи.
    document_id         BIGINT       NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    document_version    INT          NOT NULL,

    -- Кто подписал. ON DELETE RESTRICT: нельзя удалить пользователя, у которого
    -- есть подписи — подпись это юридический след. signer_name денормализован
    -- (натуральная подпись «кто подписал» переживает переименование/удаление).
    signer_id           BIGINT       NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    signer_name         VARCHAR(255) NOT NULL,

    -- Криптоматериал.
    signature_algorithm VARCHAR(32)  NOT NULL,
    digest_hex          CHAR(64)     NOT NULL,           -- подписанный SHA-256 дайджест (hex)
    signature_der       BYTEA        NOT NULL,           -- сырая ECDSA-подпись (ASN.1 DER)
    certificate_pem     TEXT         NOT NULL,           -- X.509-сертификат подписанта (PEM)

    signed_at           TIMESTAMPTZ  NOT NULL,           -- момент подписи (binds в дайджест по секундам)
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT document_signatures_version_positive
        CHECK (document_version >= 1),
    CONSTRAINT document_signatures_signer_name_nonempty
        CHECK (length(btrim(signer_name)) > 0),
    CONSTRAINT document_signatures_algorithm_valid
        CHECK (signature_algorithm IN ('ECDSA_P256_SHA256')),
    CONSTRAINT document_signatures_digest_hex_format
        CHECK (digest_hex ~ '^[0-9a-f]{64}$'),
    CONSTRAINT document_signatures_signature_nonempty
        CHECK (length(signature_der) > 0),
    CONSTRAINT document_signatures_certificate_nonempty
        CHECK (length(btrim(certificate_pem)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_document_signatures_document_id
    ON document_signatures(document_id);
CREATE INDEX IF NOT EXISTS idx_document_signatures_signer_id
    ON document_signatures(signer_id);
