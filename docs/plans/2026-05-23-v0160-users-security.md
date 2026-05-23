# Plan — v0.160.0 users Tier 1 security hotfix

**Date**: 2026-05-23
**Module**: `internal/modules/users`
**Audit verdict**: FIX-CYCLE mean 5.6 / min 3 — see `docs/plans/2026-05-20-v1.0.0-batch2-audit.md` §users
**Batch 2 progress**: 2/5 (after this release; auth ✅ v0.159.0 already shipped)
**Issue**: TBD (will create as `🔒 Security: users module v0.160.0 — profile takeover + departments/positions CRUD + avatar bypass + self/last-admin guard (Tier 1 v1.0.0)`)
**Branch**: `hotfix/issue-XXX-v0160-users-security`

## Scope — 4 ADRs (3 TIER 0 + 1 Tier 1)

### ADR-1 — Profile takeover (TIER 0)

**Threat**: `PUT /api/users/:id/profile` принимает любого caller; нет `actor.ID == target.ID` checks; audit row пишется с `user_id=target` not `actor_user_id` → атакующий untraceable.

**Files**: `user_handler.go:77-111`, `user_usecase.go:102-139`, `routes.go:46`

**Fix**:
- New domain free function `AuthorizeProfileEdit(actor, target) error` returns `ErrProfileEditForbidden` если actor != target AND actor.Role не админ (system_admin override).
- Usecase signature: `UpdateUserProfile(ctx, actorID int64, targetID int64, input *UpdateUserProfileInput)` — actorID required (was implicit).
- Handler reads `actorID := c.Get("user_id")`, passes both.
- Audit row: `actor_user_id=<actor>`, `target_user_id=<target>`, `action=user.profile_updated` (mirror sibling modules).
- Sentinel `ErrProfileEditForbidden` → handler maps к 403.

**RED test**: student A (id=10) PUTs `/api/users/20/profile` → expect 403; audit emitter spy records with `actor_user_id=10, target_user_id=20, denied=true`.

### ADR-2 — Departments + Positions CRUD без role gate (TIER 0)

**Threat**: `main.go:2651-2678` — POST/PUT/DELETE `/api/departments` + `/api/positions` открыты студенту. v0.133.0 split применил `usersAdminMW` только к `/users`; departments/positions groups missed.

**Files**: `main.go:2651-2678` (route registration), `department_handler.go:35-63`, `position_handler.go` (analog)

**Fix**:
- Apply `RequireRole(SystemAdmin, AcademicSecretary)` middleware (per "edit roles" project pattern) к `departmentsGroup` и `positionsGroup`.
- Defense-in-depth: usecase-level role check via existing `IsAdminRole` helper before mutation.
- Read endpoints (`GET`) остаются за `RequireAuth()` (любой authenticated user может видеть list).

**RED test**: student JWT POST `/api/departments` → expect 403 at middleware boundary; admin JWT same → 201.

### ADR-3 — Avatar arbitrary write через UpdateProfile bypass (TIER 0)

**Threat**: `UpdateProfile` принимает `input.Avatar` как arbitrary string (sanitized only через generic `SanitizeString`), persists в `users.avatar`, later signed as MinIO presigned URL — pointing к arbitrary S3 objects. Avatar handler validates prefix, но UpdateProfile bypasses entirely.

**Files**: `user_handler.go:97`, `user_usecase.go:102-139`, `avatar_handler.go:54-183`

**Fix**:
- New domain free function `ValidateAvatarKey(key string, userID int64) error` → checks prefix `avatars/{userID}/` (mirror existing avatar handler logic).
- `UpdateUserProfile` usecase calls `ValidateAvatarKey(input.Avatar, targetID)` если `input.Avatar != ""`.
- Sentinel `ErrInvalidAvatarKey` → handler 400 "Invalid avatar key".
- Empty avatar (`""`) allowed (= clear avatar).

**RED test**: user A (id=10) UpdateProfile with `avatar="avatars/20/evil.png"` → expect 400 ErrInvalidAvatarKey; `avatar="avatars/10/legit.png"` → 200.

### ADR-4 — DeleteUser self-deletion + last-admin removal (Tier 1)

**Threat**: `DELETE /api/users/<own_id>` locks систему если actor — единственный admin. Same для `UpdateUserStatus` блок/деактивация.

**Files**: `user_usecase.go:236-255` (DeleteUser), `user_usecase.go:UpdateUserStatus`

**Fix**:
- `DeleteUser(ctx, actorID, targetID)`: guard 1 — if `actorID == targetID` → `ErrCannotDeleteSelf`. Guard 2 — if `target.Role == SystemAdmin` AND `CountAdmins() == 1` → `ErrLastAdminProtected`.
- `UpdateUserStatus`: same 2 guards для status changes affecting admins (deactivate/block).
- New `UserRepository.CountByRole(ctx, role) (int, error)` method.
- 2 new sentinels: `ErrCannotDeleteSelf`, `ErrLastAdminProtected` → handler 409 Conflict.

**RED test**: admin (id=1, only admin) DELETEs `/api/users/1` → expect 409 ErrCannotDeleteSelf; second admin (id=2), then admin 1 DELETEs `/api/users/2` (count=1 left) — but target != self, so passes guard 1; if target.role==admin && CountByRole(admin)==1 → still need to block. Wait — guard 2 covers "last admin removal", который includes both self-delete-last-admin AND delete-other-but-last-admin.

Actually — guard 1 covers self-delete (more strict, even non-admin). Guard 2 covers "removing the last admin" (admin → 0 admins left).

**Table-driven test**:
1. self-delete admin (only admin): guard 1 fires → 409 ErrCannotDeleteSelf
2. self-delete admin (2 admins): guard 1 fires → 409 ErrCannotDeleteSelf (still self)
3. self-delete student: guard 1 fires → 409 ErrCannotDeleteSelf
4. delete other admin (count=1 after): guard 2 fires → 409 ErrLastAdminProtected
5. delete other admin (count=2 after): passes guards → 200/204
6. delete student: passes guards → 200/204

## Tier 2 absorbs (в release commit per `feedback_tier2_absorb_same_release`)

1. **DIP relocation × 3** — `domain/repositories/{user,department,position}_repository.go` interfaces → `application/usecases/repository_interfaces.go` (consolidated). Sentinels + DTOs stay в domain. Mirror v0.157.1 curriculum.
2. **AuditSink narrow port** — `*logging.AuditLogger` concrete → narrow `AuditSink interface { Emit(ctx, event) }` in usecase pkg. Adapter в main.go. Mirror v0.143 + 9 sibling modules.
3. **Cross-module impl import** — users imports `notifications/application/usecases` directly (если existing). Replace with narrow `NotifyChannel` port + adapter в main.go.
4. **validate→binding tag rename** — sweep all DTOs in `application/dto/*.go`. Gin reads `binding:` only.
5. **2 fire-and-forget without ctx** — find `context.Background()` goroutines in users module, replace with `lifecycleCtx` per shutdown pattern (mirror v0.155 ai scheduler fix).
6. **DeleteUser audit emission** — if DeleteUser swallows errors silently (per audit Tier 2 hint), audit-emit on failure (T1-1 pattern from v0.159.0).

## main.go wiring

- Route group changes: `departmentsGroup.Use(adminEditMW)` + `positionsGroup.Use(adminEditMW)` (or `usersAdminMW` if reused).
- Constructor signature changes: pass narrow `AuditSink` interface to `NewUserUseCase` (was concrete `*logging.AuditLogger`).
- DI adapter: `usersNotifierAdapter` mirror `documentsShareNotifier` pattern from v0.156.0.

## TDD discipline gates (CLAUDE.md mechanical)

- **Each ADR = 2 commits** (`test:` RED + `feat:` GREEN). 4 ADRs × 2 = 8 commits minimum.
- **Table-driven required** для DeleteUser guards (6 cases) and ValidateAvatarKey (4 cases: empty / matching prefix / wrong-user prefix / no prefix).
- **RED stubs**: declare sentinel errors + free functions upfront returning `errors.New("not implemented")` per `feedback_red_commit_compile_via_stub`. Pre-commit hook needs compile-clean.
- **Integration tests through production middleware** — `withAuth(userID, role)` helper, never ad-hoc context.WithValue (per `feedback_handler_context_key_must_match_middleware`).
- **No `&domain.User{...}` outside domain pkg** — use `NewUser` / `ReconstituteUser` constructors.

## Reviewer expectations

- Per audit mean 5.6 / min 3 — expect **FIX-CYCLE round 1** (Tier 0 race + 2-3 Tier 1 absorbs).
- Target: **SHIP round 2 ≥ 8.5/8** (mirror v0.156.0 + v0.159.0 trajectory).
- Acceptable carry-forward T2 (NOT blocking): documentation of edge cases, comment style.

## Release ritual (after SHIP)

1. `bash _tools/bump_version.sh 0.160.0`
2. Explicit stage 8 version files + commit
3. Push + PR + admin-merge
4. STICKY: `git tag -a v0.160.0` + `git push origin v0.160.0` + `gh release create v0.160.0` IMMEDIATELY
5. Docs PR `docs/roles-and-flows.md` refresh (banner 0.159.0 → 0.160.0 + narrative)
6. Memory finalisation
