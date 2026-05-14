-- Rollback for migration 037 — drop the brand_settings singleton.
-- The data loss is acceptable on rollback because the seed row is
-- recreated by re-applying 037.up.sql.
DROP TABLE IF EXISTS brand_settings;
