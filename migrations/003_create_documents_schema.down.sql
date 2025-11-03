-- Drop tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS document_history;
DROP TABLE IF EXISTS document_tag_relations;
DROP TABLE IF EXISTS document_tags;
DROP TABLE IF EXISTS document_relations;
DROP TABLE IF EXISTS document_permissions;
DROP TABLE IF EXISTS document_routes;
DROP TABLE IF EXISTS document_versions;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS document_categories;
DROP TABLE IF EXISTS document_types;
