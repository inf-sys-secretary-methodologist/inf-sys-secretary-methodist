-- Rollback: Drop Integration Module Schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_sync_conflicts_updated_at ON sync_conflicts;
DROP TRIGGER IF EXISTS update_external_students_updated_at ON external_students;
DROP TRIGGER IF EXISTS update_external_employees_updated_at ON external_employees;
DROP TRIGGER IF EXISTS update_sync_logs_updated_at ON sync_logs;

-- Drop function
DROP FUNCTION IF EXISTS update_integration_updated_at_column();

-- Drop tables (in correct order due to foreign keys)
DROP TABLE IF EXISTS sync_conflicts;
DROP TABLE IF EXISTS external_students;
DROP TABLE IF EXISTS external_employees;
DROP TABLE IF EXISTS sync_logs;

-- Drop enums
DROP TYPE IF EXISTS conflict_resolution;
DROP TYPE IF EXISTS sync_entity_type;
DROP TYPE IF EXISTS sync_direction;
DROP TYPE IF EXISTS sync_status;
