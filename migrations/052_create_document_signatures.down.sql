-- Rollback migration 052: drop the document_signatures table (#140 PR2).
DROP TABLE IF EXISTS document_signatures;
