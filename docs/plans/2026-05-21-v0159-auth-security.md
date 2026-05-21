# v0.159.0 auth Tier 1 hotfix — 2026-05-21

Closes #279. Part of v1.0.0 Batch 2 closure plan. Auth module FIX-CYCLE (mean 6.4 / min 4 — Security dimension) per `2026-05-20-v1.0.0-batch2-audit.md`.

## Scope

6 ADRs from batch 2 audit + 6 Tier 2 absorbs in single release commit (per `feedback_tier2_absorb_same_release` — ≤6 items <2h ok).

### ADR-1 — Revocation bypass closure (CRITICAL)

**Problem**. `auth_middleware.go:85-125` `JWTMiddlewareWithRevocation` reuses `JWTMiddleware` via `base(c)`. `JWTMiddleware` calls `c.Next()` on line 73 → handler executes → returns → THEN revocation check runs. Side effects committed BEFORE 401 written. Logout effectively no-op.

**Fix**. Inline token extraction + claim parse + revocation check BEFORE `c.Next()`. Single middleware path, no nested chain. Helper `extractAndValidateToken(c, authUseCase)` shared between `JWTMiddleware` (revocation-less) and `JWTMiddlewareWithRevocation` to avoid duplication.

**Test (RED→GREEN pair 1)**.
- RED: integration test through real `gin.Engine` mounting `JWTMiddlewareWithRevocation` + protected handler with side-effect spy (boolean flag). Revoke token JTI before request. Assert spy NOT called.
- GREEN: extract helper + inline check before `c.Next()`.

### ADR-2 — Refresh-token rotation + reuse detection

**Problem**. `auth_usecase.go:340-398` `RefreshToken` issues new pair without blacklist'ing old refresh JTI. Stolen refresh = uncontrolled access until natural expiry (7d).

**Fix**. Extract refresh JTI from claims. Call `revokedTokenRepo.Revoke(ctx, oldJTI, refreshExp)` BEFORE generating new tokens. If `IsRevoked(ctx, oldJTI)` returns true → reuse attempt → revoke ALL of that user's currently-issued tokens (RFC 6749 §10.4) via new `RevokeAllForUser(ctx, userID)` method on the repo.

`AuthUseCase` gains constructor-optional dep `WithRefreshRotation(revokedRepo)` (chainable setter pattern per `feedback_setter_pattern_optional_deps`). Production main wires it; tests bypass when not exercising rotation.

Dead `_ = persistence.NewSessionRepositoryPG(db)` removed (Tier 2 absorb). Session entity + repo + tests deleted (no production consumer after ADR-2).

**Test (RED→GREEN pair 2, table-driven)**.
- Case A: legitimate refresh — new pair issued, old JTI blacklisted.
- Case B: replay of blacklisted JTI — 401 + audit "refresh_token_reuse_detected" + verify all-user-tokens revoked.
- Case C: refresh without rotation dep — backward compat, no blacklist write but still works.

### ADR-3 — Login brute-force lockout + proxy CIDR

**Problem**. `rate_limiting_config.go:24-25` 300+100 RPM IP-only. `rate_limiting.go:167-182 getRealIP` trusts raw `X-Forwarded-For` from any caller. No per-account counter. `main.go:1562` stale comment claims "10 req/min + burst 5".

**Fix**.
- **Per-account lockout**: new `LoginAttemptTracker` (Redis-backed) — `login_fail:{email}` counter, TTL 15min, threshold 5. `AuthUseCase.Login` checks before bcrypt, increments on fail, deletes on success. Lockout returns 429 + audit "login_account_locked" (in addition to existing 401 for wrong password — but locked state returned regardless of password to prevent oracle).
- **Trusted-proxy CIDR**: env `TRUSTED_PROXY_CIDRS` (comma-separated, default empty = trust nothing). `getRealIP` walks `r.Header.Get("X-Forwarded-For")` only if `r.RemoteAddr` is inside trusted CIDR. Otherwise uses `r.RemoteAddr` directly.
- **Stale comment**: fix `main.go:1562` to match actual config "configured via RATE_LIMIT_PUBLIC_RPM/BURST env".

**Test (RED→GREEN pair 3, table-driven)**.
- 5 fails → 6th attempt returns `ErrAccountLocked` regardless of password correctness.
- Successful login resets counter.
- `getRealIP` with untrusted RemoteAddr ignores X-Forwarded-For; trusted RemoteAddr respects it.

### ADR-4 — MFA secret AES-GCM at rest

**Problem**. `user_repository.go:51,138` writes/reads `users.mfa_secret` as plaintext Base32 string.

**Fix**. New `internal/shared/infrastructure/crypto/secretbox.go`:
- `EncryptMFASecret(plaintext string, key []byte) (string, error)` — AES-256-GCM, nonce prepended, base64.StdEncoding output.
- `DecryptMFASecret(ciphertext string, key []byte) (string, error)` — inverse.
- KEK loaded from `MFA_SECRET_ENC_KEY` env (64 hex chars = 32 bytes).

`UserRepositoryPG` wraps `MFASecret.String()` on save, unwraps on load. Constructor accepts `WithMFASecretKEK(key []byte)` setter. When key missing → repo logs warning + persists plaintext (backward-compat for dev environments without KEK; production wiring requires KEK).

Migration `045_mfa_secret_encrypt.sql` — up: rewraps existing plaintext rows on first read (lazy migration via boolean column `mfa_secret_encrypted`). Down: documented no-op (can't recover plaintext from rewrap if KEK rotated).

Alternative considered: separate columns `mfa_secret_ct` + `mfa_secret_nonce`. Rejected — single ciphertext column simpler, GCM nonce can be prefix-stored.

**Test (RED→GREEN pair 4)**.
- Encrypt → decrypt round-trip preserves plaintext.
- Wrong key → decrypt returns error.
- Repo save with KEK present writes ciphertext column distinct from plaintext.
- Repo load handles both encrypted and lazy-migration plaintext rows.

### ADR-5 — Password reset token sha256 before store

**Problem**. `password_reset_token_repository_redis.go:43,58` stores `pwreset:{rawToken} = {userID}`. Redis read access = direct account takeover.

**Fix**. `Store(ctx, rawToken, userID, ttl)` computes `sha256(rawToken)` → key `pwreset:{hex(hash)}`. `LookupUser(ctx, rawToken)` hashes incoming token before GET. `Delete(ctx, rawToken)` hashes before DEL. Token leaving the server (email body) stays raw — only storage transform is hashing.

**Test (RED→GREEN pair 5)**.
- Stored key contains hex hash, not raw token.
- LookupUser(raw) → finds entry; LookupUser(hashed-as-input) → not found (defensive).
- Delete works symmetrically.

### ADR-6 — ConfirmReset atomic via GETDEL

**Problem**. `password_reset_usecase.go:142-173` Save → Delete order. If Delete fails after Save succeeds → password rotated, token still live → replay window.

**Fix**. Add `LookupUserAndConsume(ctx, rawToken) (int64, error)` to `PasswordResetTokenRepository` interface. Redis impl uses GETDEL atomically. `ConfirmReset` calls Consume FIRST (before bcrypt + Save). If Consume succeeds → token is gone irrespective of subsequent Save outcome → no replay window. If Save fails after Consume → user retries reset flow from scratch (acceptable trade-off — same flow as bcrypt failure case).

`VerifyToken` (read-only check before form render) keeps using non-consuming `LookupUser`.

**Test (RED→GREEN pair 6)**.
- Successful confirm → second Consume returns `ErrPasswordResetTokenNotFound`.
- Failed Save after successful Consume → token already consumed, second confirm flow returns InvalidResetToken (no replay).

## Tier 2 absorbs (in release commit, post-reviewer)

1. Delete dead `internal/modules/auth/domain/services/auth_service.go` entirely (HashPassword unused + wrong cost; no other production consumer)
2. Delete dead `internal/modules/auth/domain/entities/session.go` + `internal/modules/auth/application/usecases/session_repository.go` + `internal/modules/auth/infrastructure/persistence/session_repository.go` + tests (ADR-2 unwires; no consumer)
3. `RegisterInput.Role` binding tag → `binding:"omitempty,oneof=student teacher"` (block admin self-registration at boundary)
4. Domain sentinels in `entities/user.go` — `fmt.Errorf` → `errors.New` (DDD gate)
5. Stale comment `cmd/server/main.go:1562` (covered in ADR-3)
6. Remove dead in-memory `RateLimiter` + `RateLimitMiddleware` in `auth_middleware.go:177-259` (production uses Redis-backed shared/middleware/rate_limiting.go)

## Out of scope (explicitly)

- WebAuthn / passkeys (separate epic post-v1.0.0).
- KEK rotation tooling (manual env swap acceptable for diploma scope).
- Per-account rate limit by IP+email composite (current per-account counter sufficient).
- Refresh JTI in DB instead of Redis (Redis blacklist mirrors existing access-token revocation pattern).

## Steps

1. ✅ Verify scope via grep (audit findings still apply 1:1)
2. ✅ GH issue #279 created with 6 ADRs + Tier 2 list + acceptance criteria
3. ✅ Branch `hotfix/issue-279-v0159-auth-security`
4. ✅ Plan doc (this file)
5. RED→GREEN pair 1: ADR-1 revocation bypass closure
6. RED→GREEN pair 2: ADR-2 refresh rotation + reuse detection
7. RED→GREEN pair 3: ADR-3 brute-force lockout + proxy CIDR
8. RED→GREEN pair 4: ADR-4 MFA secret AES-GCM
9. RED→GREEN pair 5: ADR-5 reset token sha256
10. RED→GREEN pair 6: ADR-6 GETDEL atomic consume
11. Reviewer pass `superpowers:code-reviewer` — target SHIP mean ≥8 / min ≥8
12. Tier 2 absorbs in release commit
13. `bash _tools/bump_version.sh 0.159.0`
14. Stage explicit files; release commit; push; `gh pr create`; admin-merge
15. **STICKY #15**: `git tag -a v0.159.0` + `git push origin v0.159.0` + `gh release create v0.159.0` IMMEDIATELY (не "позже backfill")
16. Separate docs PR — `docs/roles-and-flows.md` refresh
17. Memory artifacts — `project_v0159_0_auth_security.md` + MEMORY.md entry + chronicles + handoff

## Risk + mitigations

- **MFA migration on existing prod data**: lazy migration via boolean column avoids forced backfill; first read rewraps. Diploma scope has zero MFA-enabled prod users → near-zero risk.
- **CIDR list misconfiguration**: empty default = trust nothing → most-restrictive default; documented in `.env.example`.
- **GETDEL on Redis < 6.2**: fallback to GET + DEL pipeline transaction (atomic from Redis side). Project uses Redis 7 (see workspace.md) — direct GETDEL safe.
- **ADR-1 helper duplication risk**: shared `extractAndValidateToken(c, uc)` returns `(claims *jwt.MapClaims, ok bool)` — single source of truth.

## Estimates

- ADR-1: 1h (rewrite + integration test)
- ADR-2: 1.5h (rotation + reuse detection + audit emit + dead-code removal)
- ADR-3: 1.5h (LoginAttemptTracker + CIDR parsing + 3 test variants)
- ADR-4: 1h (AES-GCM helper + lazy migration column + repo wrap/unwrap)
- ADR-5: 30min (mechanical sha256 transform)
- ADR-6: 30min (GETDEL Redis call + Consume method)
- Reviewer + fix-cycle: 1-2h
- Release ceremony + docs + memory: 1h

**Total: 7-10h** (above audit estimate; adjusts for 6 ADRs + 6 Tier 2 absorbs in single release).
