-- ============================================================================
-- Submissions: return-for-revision audit triple
-- ============================================================================
-- v0.111.0 introduces a return-for-revision flow on submissions: a teacher,
-- methodist, academic_secretary or system_admin can transition a submission
-- back into 'returned' status, clearing any prior grade. The audit trail of
-- who did it, when, and why must persist on the submission row itself so
-- the grading UI can render it without a join, and so the consistency rule
-- "status=returned implies the grade fields are NULL" can be enforced at
-- the schema layer (defense in depth — the entity already enforces it on
-- writes, but Reconstitute bypasses that and so does any direct SQL).
-- ============================================================================

ALTER TABLE submissions
    ADD COLUMN return_reason TEXT,
    ADD COLUMN returned_by   BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN returned_at   TIMESTAMPTZ;

-- Consistency: when status=returned, the audit triple must be filled and
-- any prior grade must be cleared. The reverse direction is guaranteed
-- by the entity (Return() always nils the grade fields), but the CHECK
-- catches direct SQL inserts/updates that bypass the domain model.
ALTER TABLE submissions
    ADD CONSTRAINT chk_submissions_returned_consistency CHECK (
        (status = 'returned'
         AND return_reason IS NOT NULL
         AND returned_by IS NOT NULL
         AND returned_at IS NOT NULL
         AND grade_value IS NULL
         AND graded_by IS NULL
         AND graded_at IS NULL)
        OR (status <> 'returned')
    );

-- Length cap mirrors the entity's 4096-char invariant on Return.reason.
-- Allowing NULL keeps non-returned rows free of any stored reason.
ALTER TABLE submissions
    ADD CONSTRAINT chk_submissions_return_reason_length CHECK (
        return_reason IS NULL OR length(return_reason) <= 4096
    );

CREATE INDEX IF NOT EXISTS idx_submissions_returned_by ON submissions(returned_by);

COMMENT ON COLUMN submissions.return_reason IS 'Teacher / methodist explanation when status=returned';
COMMENT ON COLUMN submissions.returned_by   IS 'Actor who returned the submission for revision';
COMMENT ON COLUMN submissions.returned_at   IS 'When the submission was returned for revision';
