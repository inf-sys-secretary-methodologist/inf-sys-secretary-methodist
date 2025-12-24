-- Drop custom_reports table and related objects

DROP TRIGGER IF EXISTS trigger_custom_reports_updated_at ON custom_reports;
DROP FUNCTION IF EXISTS update_custom_reports_updated_at();
DROP TABLE IF EXISTS custom_reports;
