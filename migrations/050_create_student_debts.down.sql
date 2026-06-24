-- Rollback for migration 050 — reverse order: drop the trigger and its
-- function, then the child table, then the root. ON DELETE CASCADE on
-- debt_resit_attempts.debt_id makes the child drop redundant when the
-- root is dropped, but we drop it explicitly for symmetry and clarity.

DROP TRIGGER IF EXISTS trigger_student_debts_updated_at ON student_debts;
DROP FUNCTION IF EXISTS update_student_debts_updated_at();

DROP TABLE IF EXISTS debt_resit_attempts CASCADE;
DROP TABLE IF EXISTS student_debts        CASCADE;
