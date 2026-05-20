# v0.154.0 — Reporting Tier 1 Security Hotfix

Closes #260. Part of v1.0.0 batch 1 fix-cycle (см. `docs/plans/2026-05-20-v1.0.0-batch1-audit.md`).

## Scope

Tier 1 security blockers identified в reporting module audit (mean 4.5/10, min 3/10, BLOCK):

1. SQL injection через `field.Alias` JSON input — `dynamic_query_builder.go:170`
2. `/api/custom-reports` group без role gate — privilege escalation surface
3. `simulateGeneration` shipped fake/lying endpoint
4. Background goroutines с `context.Background()` — shutdown leak

## ADRs locked

### ADR-1: Defense-in-depth Alias validation

**Layers**:
1. **Domain (entity)** — `SelectedField.Validate() error` rejects invalid Alias at construction time. `CustomReport.SetFields([]) error` propagates. `var ErrInvalidAlias = errors.New(...)` в `domain/entities/custom_report.go`.
2. **Query builder** — `DynamicQueryBuilder.Execute(...)` re-validates Alias from already-loaded entity (defense-in-depth covers legacy / corrupted DB data or bypass scenarios).
3. **Persistence reconstitution** — `SetFieldsFromJSON([]byte) error` validates после unmarshal; corrupt DB row → repo returns error vs silently feeding injection.

**Why three layers**: domain catches at write, persistence catches at read, query builder catches at execute. Each layer can fail independently — single-layer defense is bypassable.

### ADR-2: Whitelist regex `^[A-Za-z_][A-Za-z0-9_]{0,62}$`

**Reasoning**:
- PG identifier safe (matches SQL identifier rules without needing quoting)
- 63 chars max (PG `NAMEDATALEN` default - 1 for null terminator)
- Allows empty Alias (regex applies only when Alias != "") — Alias is optional per existing schema
- No mixed-case escape funkiness; canonical PG identifiers

**Alternatives rejected**:
- Quote+escape (`"foo""bar"` PG style): more permissive but increases attack surface
- Allow `.` for qualified names: not needed — Field.Name path already gates that

### ADR-3: `SetFields` API breaking change (returns error)

`func (r *CustomReport) SetFields(fields []SelectedField) error` (was: void). 2 callers in module (Create + Update в `custom_report_usecase.go`). All tests updated. Persistence reconstitute (`SetFieldsFromJSON`) becomes `func (...) error` too — already returns error.

### ADR-4: `Generate` endpoint returns 501 Not Implemented

`simulateGeneration` removed completely. `Generate` use case returns `ErrGenerationNotImplemented` sentinel. Handler maps к 501 + structured response. Background goroutines в reporting use case (lines 335, 487) eliminated с removal of simulateGeneration.

**Audit**: every 501 emission logged via `auditLogger.LogAuditEvent(ctx, "report.generate_not_implemented", ...)` для forensics. Will be replaced с real generator (extract Annual's `docxgen/render.go` pattern) в post-v1.0.0 work.

### ADR-5: Role gate at routing level

`cmd/server/main.go customReportsGroup.Use(authMiddleware.RequireRole("methodist", "academic_secretary", "system_admin"))`. Mirrors `reportsGroup.Use(authMiddleware.RequireNonStudent())` precedent (line 2012). Integration test через `httptest` мount production middleware → 403 для student.

## TDD pairs

| # | RED | GREEN |
|---|-----|-------|
| 1 | table-driven `TestSelectedField_Validate_Alias` rejects 10+ injection payloads + accepts safe identifiers | `SelectedField.Validate() error` + `ErrInvalidAlias` sentinel |
| 2 | `TestCustomReport_SetFields_RejectsInvalidAlias` | `SetFields([]SelectedField) error` returns error from Validate; update 2 callers + tests |
| 3 | `TestDynamicQueryBuilder_Execute_DefenseInDepth_RejectsMaliciousAlias` (bypasses domain via direct entity construction) | Re-validate в Execute before SQL gen; return `entities.ErrInvalidAlias` |
| 4 | `TestCustomReportRoutes_StudentDenied` integration через production middleware mount | `main.go customReportsGroup.Use(RequireRole(...))` |
| 5 | `TestReportHandler_Generate_Returns501` + remove simulateGeneration paths | `Generate` returns `ErrGenerationNotImplemented` mapped к 501; audit log entry |

Plus: `SetFieldsFromJSON` validates after unmarshal (covered by Pair 2 if implementation routes through SetFields).

## Acceptance criteria

- [ ] Table-driven test ≥10 SQL injection payloads rejected at domain layer (Alias whitelist)
- [ ] Domain entity `SelectedField.Validate()` + `ErrInvalidAlias` sentinel
- [ ] Defense-in-depth: query builder re-validates before SQL execution
- [ ] `SetFields` returns error; 2 callers updated
- [ ] `/api/custom-reports` group gated к methodist/academic_secretary/system_admin
- [ ] Integration test: student gets 403 POST /api/custom-reports
- [ ] `simulateGeneration` removed entirely
- [ ] `Generate` returns 501 + audit log emission
- [ ] Background goroutines с `context.Background()` eliminated
- [ ] All existing reporting tests still pass after API changes
- [ ] Reviewer pass: mean ≥9, min ≥8
- [ ] CI green; 8 version files bumped 0.153.13 → 0.154.0
- [ ] CHANGELOG.md entry citing #260

## Out of scope (deferred к later v0.x patches)

- Tier 2 items (god interface split, repo interfaces relocation, N+1 pagination fix, etc.)
- Tier 3 items (Cyrillic PDF font, golden tests for exports, etc.)
- Real Generate implementation (post-v1.0.0; extract Annual renderer pattern)

## Pattern references

- Reviewer praised `reports/annual/infrastructure/docxgen/render.go:151-155` paragraph escape pattern — guide для future Generate implementation
- `reports/annual/application/usecases/annual_report_usecase.go:15-32` — narrow port declaration pattern (deferred к broader reporting refactor)
