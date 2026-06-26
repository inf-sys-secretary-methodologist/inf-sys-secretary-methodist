-- Rollback migration 053: drop the signing_keys table (#140 PR3).
DROP TABLE IF EXISTS signing_keys;
