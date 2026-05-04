DROP TRIGGER IF EXISTS tr_submissions_updated_at ON submissions;
DROP TRIGGER IF EXISTS tr_assignments_updated_at ON assignments;

DROP INDEX IF EXISTS idx_submissions_status;
DROP INDEX IF EXISTS idx_submissions_student_id;
DROP INDEX IF EXISTS idx_submissions_assignment_id;
DROP INDEX IF EXISTS idx_assignments_due_date;
DROP INDEX IF EXISTS idx_assignments_group_name;
DROP INDEX IF EXISTS idx_assignments_teacher_id;

DROP TABLE IF EXISTS submissions;
DROP TABLE IF EXISTS assignments;
