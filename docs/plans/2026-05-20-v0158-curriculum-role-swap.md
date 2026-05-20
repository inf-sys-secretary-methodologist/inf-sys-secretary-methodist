# v0.158.0 — Curriculum role swap (academic_secretary author / methodist approver)

Business-logic fix per diploma role matrix: academic secretary (Волкова) is the curriculum author end-to-end (plan + sections + discipline items + submit); methodist is the approver (approve/reject pending_approval); system_admin retains an emergency override on both sides. Previous wiring (methodist author, system_admin approver) was incorrect per the project's intended role taxonomy.

## Scope

Backend whitelist swap + frontend permission matrix update + test maintenance. No DB schema changes (existing `created_by` rows continue to point к whatever user created them — historical data preserved; UI displays author by user id, role is per-user not per-row). No release behavior changes beyond authorization.

## ADRs

### ADR-1 — Backend canWrite / canApprove whitelist swap

`internal/modules/curriculum/interfaces/http/handlers/curriculum_handler.go`:

- `canWrite(role)` was `methodist || system_admin` → now `academic_secretary || system_admin`. Governs CREATE / UPDATE / DELETE / SUBMIT on curriculum + applies к section + discipline_item + bulk endpoints (same package, same helper).
- `canApprove(role)` was `system_admin` → now `methodist || system_admin`. Governs APPROVE / REJECT endpoints.

Error messages updated к match new whitelist ("only academic_secretary or system_admin may create curricula" etc.; "only methodist or system_admin may approve/reject curricula").

### ADR-2 — Route group middleware expansion

`cmd/server/main.go`:

- `adminCurriculumGroup` renamed → `approverCurriculumGroup`; `RequireRole(SystemAdmin)` expanded → `RequireRole(Methodist, SystemAdmin)`.
- Defense in depth: route gate + handler-level `canApprove` whitelist still align.

### ADR-3 — Frontend permission matrix swap

`frontend/src/lib/auth/permissions.ts`:

- `PERMISSION_MATRIX[METHODIST].CURRICULUM`: FULL → LIMITED (read + approve only; cannot CREATE/UPDATE/DELETE per `ACTION_MIN_LEVEL`).
- `PERMISSION_MATRIX[ACADEMIC_SECRETARY].CURRICULUM`: LIMITED → FULL.
- `can()` APPROVE branch: was `role === SYSTEM_ADMIN` → now `role === METHODIST || role === SYSTEM_ADMIN`. Mirror к backend `canApprove`.
- `CURRICULUM_WRITE_ROLES`: `[SYSTEM_ADMIN, METHODIST]` → `[SYSTEM_ADMIN, ACADEMIC_SECRETARY]`.

### ADR-4 — Frontend admin-approve page + navigation role expansion (reviewer round-1 absorb)

Caught by `superpowers:code-reviewer` round-1: backend whitelist swap left two UI surfaces still gated for `system_admin` only, defeating the new role matrix end-to-end.

- `frontend/src/app/admin/curriculum/approve/page.tsx:45,49` — guard `user?.role === 'system_admin'` → `isApprover` constant matching backend `canApprove` (methodist || system_admin). Without this fix, methodist cannot reach the approval queue UI.
- `frontend/src/config/navigation.ts:301-305` — `roles: [UserRole.SYSTEM_ADMIN]` → `[UserRole.SYSTEM_ADMIN, UserRole.METHODIST]`. Without this fix, methodist sees no menu link к approver queue.
- `frontend/src/app/admin/curriculum/approve/__tests__/page.test.tsx` — rejection list rewritten к `['academic_secretary', 'teacher', 'student']`; new admit-test для `['methodist', 'system_admin']`. Locks in dual-role allowlist.
- `frontend/src/config/__tests__/navigation.test.ts` — two test cases updated: nav entry roles array assertion + hidden-from-roles list (drops methodist).

### ADR-5 — Backend doc-string + Swagger annotation refresh (reviewer round-1 absorb)

- `internal/modules/curriculum/interfaces/http/handlers/curriculum_handler.go:281-285,329-335` — Approve / Reject handler comments + `@Summary "(admin only)"` → "(methodist or system_admin)".
- `internal/modules/curriculum/application/usecases/approve_curriculum_usecase.go:25-33` — usecase docstring updated к "Approver-only" wording mentioning v0.158.0 dual-role.
- `internal/modules/curriculum/application/usecases/reject_curriculum_usecase.go:28-34` — same; "the methodist may revise" → "the author (academic_secretary per v0.158.0+) may revise"; "admin's reason" → "approver's reason".

### ADR-6 — Sub-handler RejectsNonWriteRoles backfill (reviewer round-1 absorb)

Section + DisciplineItem + Bulk handlers had only single-actor `Student403` tests, no coverage that methodist is rejected as a writer post-swap. Without backfill, future regression admitting methodist as section-writer would not be caught.

- `section_handler_test.go`: `Create_Student403` → `Create_RejectsNonWriteRoles` (table-driven `["methodist", "teacher", "student", "unknown"]`).
- `discipline_item_handler_test.go`: same shape; fake doesn't track invocation flag so test asserts 403 only.
- `bulk_discipline_items_handler_test.go`: `StudentRole_Returns403` → `RejectsNonWriteRoles` (same table).

### ADR-7 — Test maintenance (not new behavioral coverage)

Test files updated к reflect new role semantics:

- `curriculum_approve_handler_test.go` / `curriculum_reject_handler_test.go`: `HappyPath_Admin*` → `HappyPath_AuthorizedRoles` (table-driven с methodist + system_admin); `RejectsNonAdminRoles` → `RejectsNonApproverRoles` (drops methodist from rejection list).
- `curriculum_handler_test.go` / `curriculum_submit_handler_test.go` / `curriculum_update_handler_test.go`: `RejectsNonWriteRoles` swaps `academic_secretary` (now author) ↔ `methodist` (now approver, not author) in the rejection list.
- Bulk sed `"methodist"` → `"academic_secretary"` в всех handler tests где роль выступает как author / writer (bulk / discipline_item / section / curriculum get/list/update/submit/handler/curriculum tests).
- Frontend: `permissions.test.ts`: 3 test cases rewritten (methodist full→limited; secretary limited→full; approve allowed roles list updated). `permission-matrix-integration.test.ts`: 2 cases rewritten (matrix integration mirrors per-role unit tests). `app/curriculum/__tests__/page.test.tsx`: Create button visibility swaps methodist↔academic_secretary; default test actor switched к academic_secretary.

## Out of scope

- Audit log historical rows — old data shows methodist as `created_by`; that's historical fact, not a bug.
- Migration to backfill `created_by` к secretary — would falsify history; rejected.
- Section + DisciplineItem repository / aggregate boundary changes — sub-aggregates inherit role gates from parent through shared `canWrite` helper (handlers package); no domain-layer changes required.

## Verification

- `go build ./...` green
- `go test ./...` full repo sweep green
- `golangci-lint run ./internal/modules/curriculum/...` 0 issues
- AmE/BrE grep clean
- Frontend: `npx jest` 216 suites / 3271 tests pass
- TypeScript: `tsc --noEmit` clean

## Why this ships separately from batch 2 security hotfix line

Batch 2 audit (`docs/plans/2026-05-20-v1.0.0-batch2-audit.md`) identified auth / users / files / messaging / announcements security work as v0.158.0..v0.162.0 release order. This release pre-empts batch 2 because it's a **business-logic correctness fix from the user / diplomant directly** (научрук's defense scenario requires the role matrix к match the project spec). Auth security hotfix becomes v0.159.0; batch 2 cascade shifts +1.

## Definition of done

- [ ] `go build ./...` green
- [ ] `go test ./...` green
- [ ] `golangci-lint` curriculum 0 issues
- [ ] AmE/BrE grep clean
- [ ] Frontend `npx jest` all green
- [ ] Reviewer pass (`superpowers:code-reviewer`) — single-pass SHIP expected (mechanical whitelist swap + test maintenance; zero new abstractions)
- [ ] Version bump 0.157.1 → 0.158.0
- [ ] Release commit, PR, admin-merge
- [ ] Docs PR — roles-and-flows refresh
- [ ] Memory artifacts (`project_v0158_0_curriculum_role_swap.md` + MEMORY.md + chronicles + handoff)
