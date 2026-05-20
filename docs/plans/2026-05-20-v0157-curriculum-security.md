# v0.157.0 — Curriculum Module Tier 1 Hotfix

Closes #269. Final v1.0.0 batch 1 fix-cycle (`docs/plans/2026-05-20-v1.0.0-batch1-audit.md`).

## Scope

Curriculum module audit verdict: **FIX-CYCLE** (mean 8.33/10, min 7.5/10). Strongest module из batch 1 — only 2 Tier 1 issues. Section + DisciplineItem aggregates already exemplary; Curriculum aggregate root needs к catch up.

1. Repository interfaces в `domain/repositories/` (DIP violation) — 4 interface files; 52 import sites; tests + production
2. `curricula` table без optimistic lock (lost-update race) — migration + entity + repo + Reconstitute + tests

## ADRs locked upfront

### ADR-1: Relocate repository interfaces from `domain/repositories/` к `application/usecases/`

CLAUDE.md gate: "Repository interfaces — в пакете-потребителе (`usecase/`), НЕ в `domain/`. DIP по классике."

Move:
- `CurriculumRepository` (curriculum_repository.go)
- `SectionRepository` (section_repository.go)
- `DisciplineItemRepository` (discipline_item_repository.go)
- `BulkUnitOfWork` (bulk_unit_of_work.go)

Sentinels (`ErrCurriculumNotFound`, `ErrCurriculumCodeExists`, `ErrSectionNotFound`, etc.) stay в `domain/repositories/` — they're domain values, not just interface contracts. Persistence + usecase still chain через errors.Is.

Implementation:
- New file `internal/modules/curriculum/application/usecases/repository_interfaces.go` (or 4 separate files mirroring originals)
- Update 52 import sites across module
- Tests reference new package path
- Sentinels stay where they are (`domain/repositories`) — no callers break since errors.Is operates on values

### ADR-2: Curricula optimistic lock — apply Section / DisciplineItem precedent

Existing pattern в curriculum_sections (migration 034) + curriculum_section_items (migration 035):
- `version INT NOT NULL DEFAULT 0` column
- CHECK `version >= 0`
- Repo SELECT loads version into entity
- Repo Update: `WHERE id = $N AND version = $N+1` + `SET version = version + 1`
- rows=0 returns concurrency conflict sentinel
- Entity has Version() accessor + Reconstitute(... version) param

Mirror exactly для curricula:

1. **migration 044**: `ALTER TABLE curricula ADD COLUMN version INT NOT NULL DEFAULT 0` + CHECK + COMMENT
2. **Curriculum entity**: add `version int` private field + `Version() int` accessor + bump в Reconstitute signature
3. **CurriculumRepositoryPG.Get / GetByCode / List**: SELECT additionally scans `version` field
4. **CurriculumRepositoryPG.Update**: UPDATE clause appends `, version = version + 1 WHERE id = $N AND version = $N+1`; rows=0 returns `ErrConcurrencyConflict` (new sentinel в domain/repositories OR existing aggregate-level pattern)
5. **Reconstitute(...)** signature gains version param (last param) — minimal call-site impact
6. **Test**: sqlmock-based test для race scenario — load entity @v0, second Update returns rows=0 → ErrConcurrencyConflict
7. **rows=0 + version_in_db_higher** disambiguation: keep ErrCurriculumNotFound if row missing, ErrConcurrencyConflict if row exists с different version — requires either SELECT-after-failure OR distinguishable error from DB. Section precedent uses single sentinel ErrSectionConcurrencyConflict для both (simpler); follow same convention.

## TDD pairs (rough plan; refine in execution)

| # | RED | GREEN |
|---|-----|-------|
| 1 | TestCurriculumRepositoryPG_Update_StaleVersion_ReturnsConflictSentinel | migration 044 + Curriculum.version field + Update WHERE clause + sentinel |
| 2 | TestCurriculumRepositoryPG_Get_LoadsVersion | SELECT scans version; Reconstitute updated |
| 3 | (config-fix) Interface relocation | Move 4 files; update 52 import sites — no behavior change |

## Acceptance criteria

- [ ] All 4 repo interfaces relocated к application/usecases package
- [ ] migration 044 deployed; curricula.version column populated с DEFAULT 0
- [ ] Curriculum entity has Version() accessor + Reconstitute signature updated
- [ ] CurriculumRepositoryPG.Update returns ErrConcurrencyConflict on stale-version race
- [ ] sqlmock-based race scenario test green
- [ ] Existing tests still pass (relocation = mechanical import change)
- [ ] errors.Is chains preserved для all sentinels
- [ ] Reviewer verdict ≥9/min ≥8
- [ ] CI green; 8 version files bumped → 0.157.0
- [ ] CHANGELOG entry; plan ADR doc referenced в release commit

## Out of scope (deferred к later patches)

- Curriculum Tier 2/3 items (not present in audit — module is strong elsewhere)
- Section + DisciplineItem repository interface relocation (already correct в narrow-port style для audit_sink etc.)
- Migration 044 backfill для existing curricula rows — `DEFAULT 0` handles new column on existing rows naturally
- Frontend impact: no UI changes; conflict error mapped к existing 409 handling (mirror Section bulk-edit)

## Carry-forward references

- v0.154 plan: `docs/plans/2026-05-20-v0154-reporting-security.md`
- v0.155 plan: `docs/plans/2026-05-20-v0155-ai-security.md`
- v0.156 plan: `docs/plans/2026-05-20-v0156-documents-security.md`
- Section optimistic-lock precedent: migration 034 + curriculum/infrastructure/persistence/section_repository_pg.go
- DisciplineItem precedent: migration 035 + discipline_item_repository_pg.go
- v0.128.0 — Section aggregate B1a foundation (sets the pattern): `memory/project_v0128_0_section_aggregate.md`
- DBTX interface refactor (precedent): `memory/feedback_dbtx_refactor_for_tx_reuse.md`
- Reviewer single-pass SHIP preconditions: `memory/feedback_single_pass_reviewer_ship.md`
