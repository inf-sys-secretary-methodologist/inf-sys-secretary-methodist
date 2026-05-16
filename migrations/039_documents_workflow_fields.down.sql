-- Reverse of 039_documents_workflow_fields.up.sql.
-- Drops the workflow audit columns. Existing rejected/approved/submitted
-- rows lose их forensic trail но не break runtime (status column itself
-- preserved).

ALTER TABLE documents
    DROP COLUMN IF EXISTS rejected_reason,
    DROP COLUMN IF EXISTS rejected_by,
    DROP COLUMN IF EXISTS rejected_at,
    DROP COLUMN IF EXISTS approved_by,
    DROP COLUMN IF EXISTS approved_at,
    DROP COLUMN IF EXISTS submitted_by,
    DROP COLUMN IF EXISTS submitted_at;
