# v0.146.0 — Security cluster (CodeQL SQL injection + leaked Google OAuth + postcss XSS)

**Status**: planning
**Branch**: `feature/issue-security-v0146-0` (off main `ab07059a`)
**Trigger**: GitHub Security tab — 2 Dependabot + 2 Code scanning + 2 Secret scanning OPEN alerts

## Scope

Single security release closing all 3 categories of GitHub Security alerts. Mirror к v0.128.10 (CodeQL sweep) + v0.128.8 (gRPC + axios security cluster).

### Findings closed

| Category | Count | Severity | Details |
|---------|-------|----------|---------|
| Code scanning | 2 | HIGH | `go/sql-injection` в `event_repository_pg.go:301` + `document_repository_pg.go:300` — `filter.OrderBy` echoed в `fmt.Sprintf` ORDER BY clause. Whitelist уже validates вход (existing `ErrInvalidOrderBy`), но CodeQL data-flow tracer не recognizes map-key lookup как sanitizer. |
| Secret scanning | 2 | CRITICAL (publicly leaked) | Real Google OAuth Client ID + Client Secret committed в `docs/integrations/composio-gmail.md:156-157` (commit `aaf2edcb`, since 2026-03-04). User confirmed manual rotation в Google Cloud Console. |
| Dependabot | 1 | MEDIUM | `postcss < 8.5.10` XSS via Unescaped `</style>` в CSS Stringify output. Direct `postcss@8.5.14` уже patched, vulnerable copy через `next@16.2.6 → postcss@8.4.31` transitive. |

### Out of scope

- Removing leaked secrets from git history — requires `git filter-repo` / BFG, destructive, blast radius on public repo. User manually rotates instead.
- Push protection rule activation (separate GitHub admin config) — confirmed already on per `feedback_github_security_settings_overhaul`.
- Other Code scanning rule findings — none reported open (2/2 closed).

## ADRs

### ADR-1 — Refactor whitelist к `map[string]string` (canonical SQL value)

**Decision**: Change `validEventOrderBy` / `validDocumentOrderBy` type из `map[string]struct{}` to `map[string]string`. Map value = canonical SQL clause. `orderBy, ok := map[input]; if !ok return ErrInvalidOrderBy`; use `orderBy` (map value) downstream — NOT echo `filter.OrderBy` (user input).

**Rationale**: CodeQL data-flow analyzer flags `fmt.Sprintf("... ORDER BY %s ...", orderBy)` where `orderBy = filter.OrderBy` even with prior map-key check, because static analysis cannot recognize the lookup-then-echo pattern as a sanitizer. Refactoring к value-from-static-map breaks the user-input → SQL flow at the analysis level. Existing validation behavior preserved (any non-whitelisted input still errors); only the "what flows into Sprintf" data path hardened.

### ADR-2 — Honest commit labels

| Commit | Type | Surface |
|--------|------|---------|
| C1 | `refactor(schedule):` | `validEventOrderBy` → `map[string]string`; use map value в List `fmt.Sprintf` |
| C2 | `refactor(documents):` | `validDocumentOrderBy` → `map[string]string`; same pattern |
| C3 | `docs(integrations):` | Replace leaked Google OAuth client ID + secret с `YOUR_CLIENT_ID` / `YOUR_CLIENT_SECRET` placeholders |
| C4 | `chore(deps,frontend):` | `package.json` add `overrides.postcss: ^8.5.10`; run `npm install` (regenerate package-lock); verify `npm ls postcss` shows no 8.4.31 |
| C5 | `chore(release): 0.146.0` | 8 version files + CHANGELOG `[0.146.0]` + roles-and-flows banner + Tier 2 absorbs if any |

No TDD pairs — refactor preserves existing test contracts (`ErrInvalidOrderBy` on invalid input). Behavior identical from caller's perspective.

### ADR-3 — Reviewer gate

**Target**: mean ≥ 9.0 / min ≥ 8.0 per project planka. Single reviewer round before release commit. Tier 1 absorb mandatory; Tier 2 absorb in release commit.

### ADR-4 — Manual rotation handoff to user

**Decision**: After release merge, user must:
1. Login к Google Cloud Console (`https://console.cloud.google.com/apis/credentials`)
2. Revoke OAuth Client ID `451773640106-403d63dukqff5qgvusjpub5u6nfhtgr9...`
3. Create new OAuth Client ID with same scopes (`gmail.send`)
4. Update Composio dashboard с new credentials
5. Optionally close secret-scanning alerts 1 + 2 as "revoked" в GitHub Security tab

This step не в scope автоматизации (external system, requires Google account). Plan + handoff document этот flow.

### ADR-5 — postcss override pattern

**Decision**: Add `overrides` (npm 8.3+) к `frontend/package.json`:

```json
{
  "overrides": {
    "postcss": "^8.5.10"
  }
}
```

**Rationale**: Direct dep `postcss@8.5.14` already patched; transitive vulnerable copy via `next@16.2.6 → postcss@8.4.31` cannot be bumped without `next` release. npm `overrides` forces version resolution на all nested copies. Verify через `npm ls postcss` post-install — should show no `8.4.31`.

Alternative considered: pnpm/yarn resolutions — project uses npm. Override is canonical.

### ADR-6 — Verification commands

Pre-release checklist:
- ✅ `go test ./...` exit 0 (existing tests pass post-refactor)
- ✅ `golangci-lint run --config=.github/golangci.yml ./...` 0 issues
- ✅ `grep -rn "GOCSPX\|451773640106" docs/` — no hits
- ✅ `cd frontend && npm ls postcss` — no `8.4.31` entries
- ✅ CodeQL re-scan on PR — alerts auto-closed when fix lands (или manually dismiss as resolved)
- ✅ Dependabot alert 40 — auto-closes когда `npm install` regenerates lock с patched postcss

Post-merge:
- User rotates Google OAuth (manual, per ADR-4)
- Mark secret-scanning alerts 1 + 2 as "revoked" в Security tab
