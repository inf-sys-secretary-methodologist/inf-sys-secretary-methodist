-- Rollback for migration 049 — reverse order: drop the revision audit FK
-- column, then the junction, then the orders table. work_program_revisions
-- (migration 048) and work_programs (047) are otherwise untouched.

DROP INDEX IF EXISTS idx_wp_revisions_triggered_by_order_id;
ALTER TABLE work_program_revisions DROP COLUMN IF EXISTS triggered_by_order_id;

DROP TABLE IF EXISTS minobrnauki_order_affected CASCADE;
DROP TABLE IF EXISTS minobrnauki_orders         CASCADE;
