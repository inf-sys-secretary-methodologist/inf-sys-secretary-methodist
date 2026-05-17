-- v0.151.0 — rollback of execution audit trail (issue #232)
ALTER TABLE documents
    DROP COLUMN IF EXISTS executed_at,
    DROP COLUMN IF EXISTS executed_by,
    DROP COLUMN IF EXISTS executor_due_date,
    DROP COLUMN IF EXISTS executor_assigned_at,
    DROP COLUMN IF EXISTS executor_assigned_to;
