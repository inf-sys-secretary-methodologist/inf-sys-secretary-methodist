-- ============================================================================
-- USERS MODULE - Откат миграции
-- ============================================================================

-- Удаление триггеров
DROP TRIGGER IF EXISTS trigger_user_profiles_updated_at ON user_profiles;
DROP TRIGGER IF EXISTS trigger_positions_updated_at ON positions;
DROP TRIGGER IF EXISTS trigger_org_departments_updated_at ON org_departments;

-- Удаление функции
DROP FUNCTION IF EXISTS update_users_module_updated_at();

-- Удаление таблиц в правильном порядке (из-за foreign keys)
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS org_departments;
