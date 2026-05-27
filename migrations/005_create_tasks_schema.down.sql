-- Drop tables in reverse order
-- task_reminders ownership moved to migration 038 (v0.167.0 cleanup) —
-- rolling back 005 no longer touches it; 038's own down migration
-- handles task_reminders teardown.
DROP TABLE IF EXISTS task_templates;
DROP TABLE IF EXISTS task_history;
DROP TABLE IF EXISTS task_dependencies;
DROP TABLE IF EXISTS task_checklist_items;
DROP TABLE IF EXISTS task_checklists;
DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS task_attachments;
DROP TABLE IF EXISTS task_watchers;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
