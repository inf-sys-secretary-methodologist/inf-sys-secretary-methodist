-- ============================================================================
-- TASK_REMINDERS — per-user reminders for project-mgmt tasks (v0.138.0)
-- ============================================================================
-- Initiative: docs/plans/2026-05-14-v0138-0-task-reminders.md (local-only).
--
-- ADR-1: separate table from `event_reminders` (migration 007 + 014).
-- DDD bounded contexts: tasks (project management) and events (calendar)
-- are different aggregates. Shared table would force cross-module
-- concept leakage; FK targets differ (tasks(id) vs events(id)).
--
-- ADR-2: minutes_before is relative to tasks.due_date. Survives task
-- rescheduling: if methodist moves due_date, the reminder auto-shifts.
-- An absolute timestamp would orphan reminders on deadline changes.
-- Reminders for a task with NULL due_date stay dormant until a
-- due_date is set (the scheduler query JOINs tasks and filters NULL).
--
-- reminder_type CHECK mirrors migration 014's event_reminders CHECK
-- so that the dispatch surface (email/push/in_app/telegram) is
-- identical across both reminder tables.
-- ============================================================================

CREATE TABLE IF NOT EXISTS task_reminders (
    id              BIGSERIAL    PRIMARY KEY,
    task_id         BIGINT       NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id         BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reminder_type   VARCHAR(50)  NOT NULL DEFAULT 'in_app'
                    CHECK (reminder_type IN ('email', 'push', 'in_app', 'telegram')),
    minutes_before  INT          NOT NULL DEFAULT 15
                    CHECK (minutes_before > 0),
    is_sent         BOOLEAN      NOT NULL DEFAULT FALSE,
    sent_at         TIMESTAMP,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- Lookup indices mirror к event_reminders index pattern.
CREATE INDEX IF NOT EXISTS idx_task_reminders_task_id   ON task_reminders(task_id);
CREATE INDEX IF NOT EXISTS idx_task_reminders_user_id   ON task_reminders(user_id);
CREATE INDEX IF NOT EXISTS idx_task_reminders_is_sent   ON task_reminders(is_sent);
