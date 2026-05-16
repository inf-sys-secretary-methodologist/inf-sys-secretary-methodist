-- v0.149.0 — documents workflow Phase 2 (issue #230)
-- Register transition extends migration 039 с registered_by audit trail
-- + partial UNIQUE index on registration_number to enforce uniqueness
-- only for documents that actually traversed the register gate.

ALTER TABLE documents
    ADD COLUMN IF NOT EXISTS registered_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL;

-- Partial UNIQUE — pre-v0.149.0 rows may have registration_number set but
-- without a registered_at audit (legacy import scenario). Index only the
-- post-v0.149.0 registered subset so existing data does не trigger conflict.
CREATE UNIQUE INDEX IF NOT EXISTS documents_registration_number_unique
    ON documents (registration_number)
    WHERE registered_by IS NOT NULL;
