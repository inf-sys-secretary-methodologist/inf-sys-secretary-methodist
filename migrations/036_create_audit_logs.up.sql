-- ============================================================================
-- AUDIT_LOGS — forensic trail persistence (v0.130.0)
-- ============================================================================
-- Initiative: docs/plans/2026-05-11-audit-logs.md (local-only).
--
-- Prior to migration 036 the AuditLogger (internal/shared/infrastructure/
-- logging) emitted events only as structured stdout logs; this migration
-- adds the persistent table that the same logger now writes to in
-- addition (ADR-2 sync write — independent of any business transaction
-- so failed/denied business ops still get audited; the writer calls
-- db.ExecContext directly, never via *sql.Tx).
--
-- ADR-3: a failed INSERT INTO audit_logs is logged and NOT propagated —
-- audit emit must never break a business operation.
--
-- ADR-6: actor_user_id is BIGINT WITHOUT REFERENCES users(id) — audit
-- log must survive user deletion (legal/forensic requirement).
--
-- ADR-4: CreatedAt uses DEFAULT CURRENT_TIMESTAMP authoritative on the
-- server side; the Go writer does NOT pass this column to INSERT.
-- ============================================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id              BIGSERIAL    PRIMARY KEY,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    action          TEXT         NOT NULL,
    resource        TEXT         NOT NULL,
    actor_user_id   BIGINT       NULL,
    actor_ip        INET         NULL,
    correlation_id  TEXT         NULL,
    fields          JSONB        NOT NULL DEFAULT '{}'::jsonb,

    CONSTRAINT chk_audit_logs_action_nonempty   CHECK (length(action) > 0),
    CONSTRAINT chk_audit_logs_resource_nonempty CHECK (length(resource) > 0)
);

-- ADR-5 indexes — DESC on created_at optimizes the dominant frontend
-- query (recent events first); partial index on actor_user_id skips the
-- bulk of rows when actor is unknown (system-emitted events).

CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at
    ON audit_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_user_id
    ON audit_logs (actor_user_id)
    WHERE actor_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_action
    ON audit_logs (action);

CREATE INDEX IF NOT EXISTS idx_audit_logs_resource
    ON audit_logs (resource);

CREATE INDEX IF NOT EXISTS idx_audit_logs_fields_gin
    ON audit_logs USING GIN (fields);

COMMENT ON TABLE audit_logs IS 'Forensic audit trail for compliance — every AuditLogger.LogAuditEvent emission since v0.130.0';
COMMENT ON COLUMN audit_logs.action IS 'Dotted event identifier, e.g. curriculum.created / document.deleted / auth.login';
COMMENT ON COLUMN audit_logs.resource IS 'Aggregate name acted upon, e.g. curriculum / document / session';
COMMENT ON COLUMN audit_logs.actor_user_id IS 'User id who triggered the event; NULL for system-emitted events. No FK to users.id — audit survives user deletion (ADR-6)';
COMMENT ON COLUMN audit_logs.actor_ip IS 'Source IP captured from context if available';
COMMENT ON COLUMN audit_logs.correlation_id IS 'Per-request trace id for cross-event correlation';
COMMENT ON COLUMN audit_logs.fields IS 'Event-specific payload as JSONB — schema-less to absorb future field additions without migration';
