-- v0.152.0 — rollback of archive audit trail (issue #233)
ALTER TABLE documents
    DROP COLUMN IF EXISTS archived_at,
    DROP COLUMN IF EXISTS archived_by;
