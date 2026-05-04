DROP INDEX IF EXISTS idx_submissions_returned_by;

ALTER TABLE submissions DROP CONSTRAINT IF EXISTS chk_submissions_return_reason_length;
ALTER TABLE submissions DROP CONSTRAINT IF EXISTS chk_submissions_returned_consistency;

ALTER TABLE submissions
    DROP COLUMN IF EXISTS returned_at,
    DROP COLUMN IF EXISTS returned_by,
    DROP COLUMN IF EXISTS return_reason;
