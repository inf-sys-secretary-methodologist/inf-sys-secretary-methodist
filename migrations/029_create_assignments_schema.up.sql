-- ============================================================================
-- ASSIGNMENTS - Academic homework context (separate from project-mgmt tasks)
-- ============================================================================
-- Establishes the Assignments bounded context: a teacher publishes an
-- Assignment for a student group, and per-student Submissions hold the grade.
-- This is intentionally NOT colocated with the existing tasks module, which
-- models project-management work items (issue tracker semantics).
-- ============================================================================

-- Assignment aggregate root
CREATE TABLE IF NOT EXISTS assignments (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    teacher_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    group_name VARCHAR(100) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    max_score INT NOT NULL,
    due_date TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_assignments_max_score_positive CHECK (max_score > 0),
    CONSTRAINT chk_assignments_group_name_nonempty CHECK (length(trim(group_name)) > 0),
    CONSTRAINT chk_assignments_title_nonempty CHECK (length(trim(title)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_assignments_teacher_id ON assignments(teacher_id);
CREATE INDEX IF NOT EXISTS idx_assignments_group_name ON assignments(group_name);
CREATE INDEX IF NOT EXISTS idx_assignments_due_date ON assignments(due_date);

-- Submission child entity (one per student per assignment)
CREATE TABLE IF NOT EXISTS submissions (
    id BIGSERIAL PRIMARY KEY,
    assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    grade_value INT,
    feedback TEXT,
    graded_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    graded_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_submissions_status_enum CHECK (status IN ('pending', 'graded', 'returned')),
    CONSTRAINT chk_submissions_grade_nonnegative CHECK (grade_value IS NULL OR grade_value >= 0),
    CONSTRAINT chk_submissions_graded_consistency CHECK (
        (status = 'graded' AND grade_value IS NOT NULL AND graded_by IS NOT NULL AND graded_at IS NOT NULL)
        OR (status <> 'graded')
    ),
    CONSTRAINT chk_submissions_feedback_length CHECK (feedback IS NULL OR length(feedback) <= 4096),
    CONSTRAINT uq_submissions_assignment_student UNIQUE (assignment_id, student_id)
);

CREATE INDEX IF NOT EXISTS idx_submissions_assignment_id ON submissions(assignment_id);
CREATE INDEX IF NOT EXISTS idx_submissions_student_id ON submissions(student_id);
CREATE INDEX IF NOT EXISTS idx_submissions_status ON submissions(status);

COMMENT ON TABLE assignments IS 'Academic assignments published by teachers for student groups';
COMMENT ON TABLE submissions IS 'Per-student submission record, holds grade once teacher grades it';

-- Reuse the shared updated_at trigger function defined in earlier migrations
-- (see migrations/021 update_attendance_updated_at). We expose new triggers
-- bound to the same function to keep updated_at fresh.
CREATE TRIGGER tr_assignments_updated_at
    BEFORE UPDATE ON assignments
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();

CREATE TRIGGER tr_submissions_updated_at
    BEFORE UPDATE ON submissions
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();
