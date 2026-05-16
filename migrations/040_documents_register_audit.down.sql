DROP INDEX IF EXISTS documents_registration_number_unique;

ALTER TABLE documents
    DROP COLUMN IF EXISTS registered_by;
