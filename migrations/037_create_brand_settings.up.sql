-- ============================================================================
-- BRAND_SETTINGS — editable system brand singleton (v0.136.0)
-- ============================================================================
-- Initiative: docs/plans/2026-05-14-v0136-0-branding-backend.md (local-only).
--
-- ADR-1: full module under internal/modules/branding/ (not under
-- internal/shared/admin/) because branding has domain invariants —
-- hex color format, URL scheme whitelist, app name length bounds —
-- that need a domain entity with constructor validation.
--
-- ADR-3: singleton pattern. The table holds exactly one row enforced
-- by CHECK (id = 1). A versioned config table is overkill — brand
-- is one current state, not a series. The DEFAULT row is seeded in
-- this migration so the first GET never returns 404 and the
-- frontend always has something to render.
--
-- ADR-4: validation lives in the Go domain entity (hex regex + URL
-- scheme whitelist + length bounds). The SQL CHECKs below are
-- defense-in-depth for length only — the regex / scheme guards
-- would be brittle as PG constraints and are not duplicated.
-- ============================================================================

CREATE TABLE IF NOT EXISTS brand_settings (
    id              INT          PRIMARY KEY DEFAULT 1
                    CONSTRAINT chk_brand_settings_singleton CHECK (id = 1),
    app_name        TEXT         NOT NULL,
    tagline         TEXT         NOT NULL DEFAULT '',
    logo_url        TEXT         NOT NULL DEFAULT '',
    favicon_url     TEXT         NOT NULL DEFAULT '',
    primary_color   TEXT         NOT NULL DEFAULT '',
    secondary_color TEXT         NOT NULL DEFAULT '',
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT chk_brand_settings_app_name_length
        CHECK (length(app_name) BETWEEN 1 AND 100),
    CONSTRAINT chk_brand_settings_tagline_length
        CHECK (length(tagline) <= 200)
);

-- Seed the singleton row. ON CONFLICT DO NOTHING makes the
-- migration idempotent across re-runs and across environments
-- that may have applied a hot-fix seed already.
INSERT INTO brand_settings (id, app_name)
VALUES (1, 'Информационная система академического секретаря/методиста')
ON CONFLICT (id) DO NOTHING;
