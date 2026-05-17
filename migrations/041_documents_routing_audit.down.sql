-- v0.150.0 — rollback of routing + visa audit trail (issue #231)
ALTER TABLE documents
    DROP COLUMN IF EXISTS visa_signed_at,
    DROP COLUMN IF EXISTS visa_signed_by,
    DROP COLUMN IF EXISTS routed_at,
    DROP COLUMN IF EXISTS routed_by;
