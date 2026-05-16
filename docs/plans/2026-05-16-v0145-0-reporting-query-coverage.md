# v0.145.0 — reporting/infrastructure/query backfill (Phase 6 #196 sprint 4/15)

**Status**: planning
**Branch**: `feature/issue-196-reporting-query-coverage` (off `main` 0d754547, no commits yet)
**Issue**: #196 (backend coverage 75.0% → 90%, multi-release sprint)

## Scope

Pure backfill для `internal/modules/reporting/infrastructure/query/dynamic_query_builder.go` — single file, 667 LoC, 12 функций все at 0%. Biggest single-package leverage per handoff senior pick.

### What gets added

`dynamic_query_builder_test.go` в том же пакете. 4 logical blocks:

| Block | Surface | Test approach |
|-------|---------|---------------|
| B1 | `formatValue`, `sanitizeFilename`, `truncateString`, `GetAvailableFields` | table-driven, pure functions |
| B2 | `buildWhereClause` | table-driven, 15 operators + 2 array branches (`In`/`NotIn` empty array degenerate) |
| B3 | `Execute` (+ `NewDynamicQueryBuilder` + `configureDataSources` as side effects) | sqlmock — happy / unsupported source / no valid fields / count error / query error / Columns error / scan error / aggregation paths / groupBy + orderBy paths |
| B4 | `Export` + `exportCSV` + `exportXLSX` + `exportPDF` | direct construction of `ReportExecutionResult` + bytes inspection (CSV: parseable; XLSX: open round-trip; PDF: non-empty `%PDF-` magic prefix); error branches: empty columns в PDF, unsupported format |

### What stays out of scope

- `reporting/application/usecases/*.go` coverage gaps — separate sprint slot.
- `reporting/interfaces/http/*.go` handler coverage — concrete-dep coupling blocker (waiting на DIP refactor mirror'ed via v0.141.0 template).
- Cross-package integration through `*sql.DB` — tested через sqlmock не real PG.
- Behavior changes к dynamic_query_builder.go — pure backfill, zero production-code edit.

## ADRs

### ADR-1 — Honest `test: backfill` labels

**Decision**: Все test-commits labeled `test(reporting/query): backfill coverage for X`, не RED→GREEN pairs.

**Rationale**: Code already exists. CLAUDE.md TDD gate explicit: «Покрытие уже написанного кода называть честно: `test: backfill coverage for X`, НЕ выдавать за TDD». Precedent v0.140.0 + v0.143.0.

### ADR-2 — Commit split by surface

| Commit | Type | Surface |
|--------|------|---------|
| C1 | `test(reporting/query):` | pure helpers + `GetAvailableFields` (B1) |
| C2 | `test(reporting/query):` | `buildWhereClause` table-driven (B2) |
| C3 | `test(reporting/query):` | `Execute` sqlmock-based (B3) |
| C4 | `test(reporting/query):` | `Export` + 3 exporters (B4) |
| C5 | `chore(release): 0.145.0` | version bump + CHANGELOG + roles-and-flows banner + Tier 1/2 absorb если есть |

Each commit independently runnable (`go test ./internal/modules/reporting/infrastructure/query/...`). Splitting упрощает rollback если reviewer flags single block.

### ADR-3 — sqlmock + `regexp.QuoteMeta` для query anchoring

**Decision**: Mock SQL queries через `sqlmock.ExpectQuery(regexp.QuoteMeta(expectedSQL)).WithArgs(...)`. WithArgs pin для всех filter values per `feedback_sqlmock_withargs_for_mutation_resistance`.

**Rationale**: Hardcoded substring matchers self-confirm green when production query changes (Execute builds query dynamically — небольшие mutations should break tests). Full `QuoteMeta` + `WithArgs` gives mutation-resistance.

### ADR-4 — Export format verification — round-trip vs magic-byte

**Decision**:
- CSV: parse output bytes через `csv.NewReader`, assert exact rows.
- XLSX: open через `excelize.OpenReader(bytes.NewReader(...))`, assert cell values.
- PDF: assert bytes prefix `%PDF-` + non-empty length (gofpdf binary output — full structural validation overkill для coverage backfill).

**Rationale**: CSV/XLSX format inputs are admin-controlled — strict round-trip catches column-mapping bugs. PDF is binary stream — magic-byte + length sufficient для coverage signal без pulling pdf parser dep.

### ADR-5 — Reviewer gate

**Target**: mean ≥ 9.0 / min ≥ 8.0 (project planka). Single `superpowers:code-reviewer` round перед release commit. Tier 1 absorb mandatory; Tier 2 absorb in release commit per `feedback_tier2_absorb_same_release`.

### ADR-6 — Coverage delta expectation

**Estimate**: +2-3pp global (75.0% → ~77.5%) per handoff senior pick. Package-local 0% → ≥85% target (12 functions × table-driven ≥3 cases per function).

**Verification**: `go test -coverprofile=/tmp/cov.out ./... && go tool cover -func=/tmp/cov.out | grep "reporting/infrastructure/query\|total"` после C4.

## Pre-flight verification

- ✅ `git status` clean on `feature/issue-196-reporting-query-coverage`
- ✅ Baseline 75.0% measured ($ go tool cover -func=/tmp/cov_session_start.out | tail -1)
- ✅ All packages green (`go test ./...` exit 0)
- ✅ `dynamic_query_builder_test.go` не existing (no merge conflict)
