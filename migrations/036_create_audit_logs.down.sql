-- Reverse of 036_create_audit_logs.up.sql.
-- Rollback drops the table; CASCADE handles index cleanup.

DROP TABLE IF EXISTS audit_logs CASCADE;
