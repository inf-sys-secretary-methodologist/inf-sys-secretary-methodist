# B1a — Section aggregate (v0.128.0 — v0.128.3)

**Initiative:** Curriculum sprint 3 — двухуровневая иерархия `Curriculum → Sections → DisciplineItems`. Prerequisite для B1b "Bulk-edit РПД" (Phase 4 row #59).

**Scope full initiative:** 4 thin vertical slices, каждый shippable + reviewable. Backend-first; frontend в финальном слайсе (v0.128.3).

**Mirror:** curriculum module v0.116.0 (aggregate root patterns) + assignments line (cross-aggregate guard в usecase).

---

## ADR-1: Aggregate boundary = Beta (3 separate ARs)

`Curriculum`, `Section`, `DisciplineItem` — independent aggregate roots, references по FK. Cross-AR invariants (sum of credits, hours totals) — domain service / read-model validation, не constructor invariant.

**Reasoning:** bulk-edit (B1b) — dominant load pattern. Loading entire Curriculum (50-450 nested rows) для редактирования одной cell = O(N) waste. Vaughn Vernon "small aggregates rule": ARs должны быть transactionally consistent на каждом mutation, не больше. Sections + items не нужны для curriculum-level invariants (status, ownership, year, code).

**Trade-off:** cross-AR consistency через usecase orchestration (fetch curriculum.status в Section CRUD usecase для guard), не через aggregate-internal invariant. Acceptable — invariant "section editable iff curriculum draft/returned" — workflow rule, не data integrity rule.

**Rejected:**
- **Alpha** (single Curriculum AR with inner entities) — large aggregate, lock contention, slow loads.
- **Gamma** (Curriculum AR + section/item navigable) — hybrid усложняет mental model, не оправдан scale-wise (read-model query достаточен).

## ADR-2: Lifecycle inheritance, без own status

`Section` + `DisciplineItem` НЕ имеют own `status` column. Editability inherits curriculum.status:

| curriculum.status | sections + items |
|---|---|
| `draft` / (returned-from-rejection) | editable by author methodist + admin |
| `pending_approval` | frozen (всем) |
| `approved` / `archived` | frozen (всем) |

**Guard в usecase**, не в handler (CLAUDE.md gate). На каждом mutation usecase fetches curriculum, проверяет `curriculum.Status().CanEdit()`, returns `ErrCannotEditSection` (sentinel) если frozen. Cleaner чем дублировать status в каждой section row.

**Reasoning:** alternative — копировать status в section column + cascade triggers — redundant data, sync drift risk. Single-source-of-truth = curriculum.status.

## ADR-3: Optimistic locking foundation в v0.128.0

`version INT NOT NULL DEFAULT 0` column на обеих таблицах (sections + items). UPDATE: `WHERE id = ? AND version = ?` → `RowsAffected() == 0` → `ErrSectionVersionConflict` → handler 409 Conflict.

**Reasoning:** retrofit'ить optimistic lock после ship'а = миграция + усложнение test surface. 1 column at table-create cost — free. B1b bulk-edit потребует conflict detection обязательно. Foundation сейчас → consumed позже.

**Pattern:** RowsAffected check, не SELECT FOR UPDATE (pessimistic lock). Pessimistic = single-writer bottleneck; optimistic = scalable, conflict only on actual race.

## ADR-4: Schema (migrations 034 + 035)

Migration **034** (v0.128.0): `curriculum_sections` — простая структура.

```sql
CREATE TABLE curriculum_sections (
  id BIGSERIAL PRIMARY KEY,
  curriculum_id BIGINT NOT NULL REFERENCES curricula(id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,
  description TEXT,
  order_index INT NOT NULL DEFAULT 0,
  version INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT chk_curriculum_sections_title_nonempty CHECK (length(trim(title)) > 0),
  CONSTRAINT chk_curriculum_sections_description_length
    CHECK (description IS NULL OR length(description) <= 4096),
  CONSTRAINT chk_curriculum_sections_order_index_nonneg CHECK (order_index >= 0),
  CONSTRAINT chk_curriculum_sections_version_nonneg CHECK (version >= 0)
);
CREATE INDEX idx_curriculum_sections_curriculum_id ON curriculum_sections(curriculum_id);

CREATE TRIGGER tr_curriculum_sections_updated_at
  BEFORE UPDATE ON curriculum_sections
  FOR EACH ROW
  EXECUTE FUNCTION update_attendance_updated_at();
```

Migration **035** (v0.128.1): `curriculum_section_items` — rich invariants. Detailed schema deferred to v0.128.1 plan.

**No UNIQUE on (curriculum_id, order_index)**: bulk reorder с unique = deferred-constraint headache. Stable sort `ORDER BY order_index, created_at` — deterministic display даже при duplicates. Reorder UI sets explicit ordering.

**Hard-delete, не soft-delete**: undo via UI confirm dialog (precedent — confirmation pattern в codebase). Audit_sink ловит `section.deleted` event для forensics. Soft-delete только если product требует "корзину" (out-of-scope для v0.128.x).

## ADR-5: ControlForm = typed Value Object (v0.128.1)

Применяется при добавлении DisciplineItem (v0.128.1). Объявление здесь для full picture:

```go
type ControlForm string
const (
    ControlFormZachet              ControlForm = "zachet"
    ControlFormExam                ControlForm = "exam"
    ControlFormCourseProject       ControlForm = "course_project"
    ControlFormDifferentialZachet  ControlForm = "differential_zachet"
)
func (c ControlForm) Validate() error
```

**Reasoning:** РФ academic standard — 4 формы контроля. Per CLAUDE.md ubiquitous-language gate: `string` с magic values → typed enum.

**Semester range:** 1..12. Покрывает бакалавр (8) + магистратуру (4 more). CHECK constraint `BETWEEN 1 AND 12`.

## ADR-6: Authorization mirror к curriculum

`Section.AuthorizeEdit(actorID, isAdmin, curStatus, curCreatedBy)` — entity stays pure of Curriculum knowledge per ADR-1 Beta. Usecase fetches curriculum, передаёт status + createdBy как primitives.

Roles:
- `methodist` — правит свои curricula → правит sections внутри своих
- `system_admin` + `academic_secretary` — override на чужие
- `teacher` + `student` — read-only (если route разрешает) или 403

Handler reads `c.Get("role")` (production middleware contract — pre-commit hook не ловит этот класс; integration test через `withAuth` обязательно per v0.126.0/v0.126.1 lesson).

## ADR-7: Release split = 4 thin slices

Полный scope (Section + Item + bulk-edit + frontend) ≈ 10-12h. Bundling в один release → big PR → reviewer round-1 finds many issues → fix-cycle 2-3 rounds. Senior approach = small reviewable slices.

| Release | Scope | Estimate |
|---|---|---|
| **v0.128.0** | `Section` domain + repo + 5 CRUD usecases + handlers + migration 034 + DI wiring | ~3h |
| **v0.128.1** | `DisciplineItem` domain (richer invariants — hours, credits, ControlForm) + repo + 5 CRUD + handlers + migration 035 | ~3-4h |
| **v0.128.2** | Bulk-edit endpoint `POST /api/sections/:id/items/bulk` с transactional commit-or-rollback + optimistic lock enforcement | ~2-3h |
| **v0.128.3** | Frontend bulk-edit UI (table view, multi-row select, conflict 409 UI) + i18n × 4 | ~3-4h |

Каждый release shippable, defendable перед научруком, reviewable per ось.

> **ADR-9 amendment (2026-05-09, post-v0.128.1)** — Re-shape к 5-release initiative. v0.128.1 reviewer был skipped с honest disclaimer; retroactive `superpowers:code-reviewer` round дал mean 8.71 / min 8 (strict gate FIX-CYCLE по mean<9) с 3 Tier 1 findings. Senior decision: spin separate **v0.128.2 = Review Hardening** release (closes Tier 1 findings + retroactive reviewer pass + honest disclaimer closure) перед бульк-edit. Initiative shifts:
>
> | Release | Scope | Status |
> |---|---|---|
> | v0.128.0 | Section foundation | shipped 2026-05-09 (reviewer SHIP 9.0/9 round-2) |
> | v0.128.1 | DisciplineItem Layer 2 | shipped 2026-05-09 (reviewer skipped → retroactive 8.71/8) |
> | **v0.128.2** | **Review hardening** (3 Tier 1 fixes: ListBySection 404, audit reason backfill, shared withAuth) | **shipped 2026-05-09 (reviewer SHIP 9.14/9 round-1)** |
> | v0.128.3 | Bulk-edit transactional endpoint (was v0.128.2) | pending |
> | v0.128.4 | Frontend bulk-edit UI + i18n × 4 (was v0.128.3) | pending |
>
> **Pattern**: strict reviewer gate (mean ≥9) trumps absorb-in-next-release pragmatism для drift-risk releases. Spin focused review-hardening release когда reviewer mean<9 strict — closes honest disclaimer чисто, keeps next feature release scope-coherent (single concern). Reviewer-endorsed absorb pattern OK для same-release commits, не для cross-release fixes из retroactive rounds.

## ADR-8: Repository interface placement (mirror curriculum)

Project pattern (curriculum module): broad interface в `domain/repositories/<x>_repository.go` (canonical port + sentinel errors); each usecase declares **narrow port inline** для interface segregation + testability.

Senior reading CLAUDE.md "interfaces в usecase package": `domain/repositories/` IS the package "поставщик контракта"; usecase package owns narrow ports for testing — DIP at use-case level. Mirror exact existing pattern, не deviate.

```go
// domain/repositories/section_repository.go — broad
type SectionRepository interface {
    Save(ctx, *Section) error
    GetByID(ctx, id) (*Section, error)
    ListByCurriculumID(ctx, curriculumID) ([]*Section, error)
    Update(ctx, *Section) error
    Delete(ctx, id) error
}
var ErrSectionNotFound = errors.New(...)
var ErrSectionVersionConflict = errors.New(...)

// application/usecases/create_section_usecase.go — narrow
type createSectionRepo interface { Save(ctx, *Section) error }
type curriculumLookup interface { GetByID(ctx, id) (*Curriculum, error) }
```

---

## v0.128.0 — Section CRUD (this release scope)

### Files added

```
internal/modules/curriculum/domain/entities/section.go                                  (~250 LoC)
internal/modules/curriculum/domain/entities/section_test.go                             (~300 LoC)
internal/modules/curriculum/domain/repositories/section_repository.go                   (~80 LoC)
internal/modules/curriculum/infrastructure/persistence/section_repository_pg.go         (~250 LoC)
internal/modules/curriculum/infrastructure/persistence/section_repository_pg_test.go    (~400 LoC)
internal/modules/curriculum/application/usecases/create_section_usecase.go              (~80 LoC)
internal/modules/curriculum/application/usecases/create_section_usecase_test.go         (~150 LoC)
internal/modules/curriculum/application/usecases/get_section_usecase.go                 (~50 LoC)
internal/modules/curriculum/application/usecases/get_section_usecase_test.go            (~100 LoC)
internal/modules/curriculum/application/usecases/list_sections_usecase.go               (~60 LoC)
internal/modules/curriculum/application/usecases/list_sections_usecase_test.go          (~100 LoC)
internal/modules/curriculum/application/usecases/update_section_usecase.go              (~100 LoC)
internal/modules/curriculum/application/usecases/update_section_usecase_test.go         (~200 LoC)
internal/modules/curriculum/application/usecases/delete_section_usecase.go              (~80 LoC)
internal/modules/curriculum/application/usecases/delete_section_usecase_test.go         (~150 LoC)
internal/modules/curriculum/interfaces/http/handlers/section_handler.go                 (~250 LoC)
internal/modules/curriculum/interfaces/http/handlers/section_handler_test.go            (~400 LoC)
migrations/034_create_curriculum_sections.up.sql                                        (~50 LoC)
migrations/034_create_curriculum_sections.down.sql                                      (~5 LoC)
docs/plans/2026-05-09-v0128-section-aggregate.md                                        (this doc)
```

### Files modified

```
cmd/server/main.go             — DI wiring + route registration
docs/swagger/docs.go           — auto-generated по swag annotations
docs/swagger/swagger.json      — auto-generated
docs/swagger/swagger.yaml      — auto-generated
CHANGELOG.md                   — v0.128.0 entry
docs/roles-and-flows.md        — banner + entry
version                        — 0.128.0
frontend/VERSION               — 0.128.0
frontend/package.json          — 0.128.0
frontend/package-lock.json     — 0.128.0
```

### Routes

```
POST   /api/curricula/:id/sections   — create  (methodist + admin/secretary)
GET    /api/curricula/:id/sections   — list    (any non-student)
GET    /api/sections/:id             — get     (any non-student)
PUT    /api/sections/:id             — update  (author methodist + admin/secretary)
DELETE /api/sections/:id             — delete  (author methodist + admin/secretary)
```

### TDD pairs (5 RED→GREEN = 10 commits)

1. **Pair 1** — domain construction: `NewSection` invariants + `Reconstitute` + getters + `ErrInvalidSection` sentinel + `Section` opaque struct
2. **Pair 2** — domain mutation: `UpdateBasics(title, description, order_index, now)` + `AuthorizeEdit(actorID, isAdmin, curStatus, curCreatedBy)` + sentinels (`ErrSectionScopeForbidden`, `ErrCannotEditSection`)
3. **Pair 3** — persistence: broad `SectionRepository` interface + PG impl 5 methods + optimistic lock на Update + migration 034 + sqlmock tests с `WithArgs` pin (нullable description ≥3-variant gate)
4. **Pair 4** — application: 5 usecases с narrow ports + audit emitter + `curriculumLookup` cross-aggregate port; cover gates (status frozen, scope forbidden, not found, version conflict)
5. **Pair 5** — HTTP: 5 handlers + per-port narrow interfaces + `errors.Is` mapping (404/403/409/422) + integration tests via `withAuth` helper + DI wiring

### Test invariants pinned

| Layer | Invariant | Test |
|---|---|---|
| Domain | NewSection title trimmed-non-empty | `TestNewSection_EmptyTitleRejected` |
| Domain | order_index ≥ 0 | `TestNewSection_NegativeOrderIndexRejected` |
| Domain | description ≤ 4096 chars | `TestNewSection_DescriptionTooLong` |
| Domain | curriculum_id > 0 | `TestNewSection_InvalidCurriculumIDRejected` |
| Domain | created_by > 0 | `TestNewSection_InvalidCreatedByRejected` |
| Domain | UpdateBasics atomic on failure | `TestSection_UpdateBasics_AtomicOnFailure` |
| Domain | AuthorizeEdit status gate before ownership | `TestSection_AuthorizeEdit_StatusBeforeOwnership` |
| Domain | AuthorizeEdit admin override | `TestSection_AuthorizeEdit_AdminOverride` |
| Persistence | Save sets generated id | `TestSectionRepoPG_Save_SetsID` |
| Persistence | Update returns ErrSectionVersionConflict on stale version | `TestSectionRepoPG_Update_VersionConflict` |
| Persistence | Update bumps version on success | `TestSectionRepoPG_Update_VersionIncrement` |
| Persistence | Delete cascade тест из миграции | `TestSectionRepoPG_Delete_CurriculumCascade` |
| Use case | Create denied when curriculum frozen | `TestCreateSection_CurriculumFrozenDenied` |
| Use case | Update author-only enforced when not admin | `TestUpdateSection_NonAuthorMethodistDenied` |
| HTTP | Update on non-existent section → 404 | `TestSectionHandler_Update_404` |
| HTTP | Update with stale version → 409 | `TestSectionHandler_Update_409Conflict` |
| HTTP | Authorization via `c.Get("role")` (production middleware contract) | `TestSectionHandler_RoleKeyContract` |

### Out-of-scope for v0.128.0

- DisciplineItem entity + persistence — v0.128.1
- Bulk-edit endpoint — v0.128.2
- Frontend page — v0.128.3
- Cascade на curriculum.Approve / .Reject — handled implicitly by ADR-2 (sections frozen when curriculum frozen, no separate transition needed)
- Soft-delete / restore — out of scope all v0.128.x
- Idempotency keys — out of scope (admin internal tool, retries acceptable)
- Pagination на ListSectionsByCurriculumID — sections per curriculum < 30 typical

### Risk mitigations

| Risk | Mitigation |
|---|---|
| Cross-aggregate consistency drift (curriculum deleted while sections live) | FK CASCADE delete на schema level |
| Mirror role-key bug (v0.126.0 class) | Integration tests через `withAuth` helper (not ad-hoc); pin к production middleware key |
| Reviewer round-1 catches invariant gaps | Expected — feature work с new aggregate; trust fix-cycle |
| Migration 034 collision | Verified: 033 last applied; 034 free slot |
| TDD discipline под scope pressure | 5 pairs = 10 commits; each pair scoped к layer; pair-internal bundling acceptable per CLAUDE.md (same behavior unit) |

### Reviewer triangulation gate

Mean ≥ 9 / min ≥ 8 across 7 axes (TDD / DDD / CA / Tests / Cohesion / Hygiene / Security). Round-1 expected to surface 2-4 findings (new aggregate surface). Round-2 ship target.

### Release loop (8 шагов)

1. `_tools/bump_version.sh 0.128.0`
2. CHANGELOG entry (formatted as v0.127.0)
3. `docs/roles-and-flows.md` banner update + "Изменения в 0.128.0" entry
4. `chore(release): 0.128.0` commit + `git tag -a v0.128.0`
5. **PAUSE** для approval (Red Line)
6. `git push origin main` + `git push origin v0.128.0` + `gh release create`
7. `gh run list --branch main --limit 6` — fix red checks в той же сессии
8. `.claude/handoffs/2026-05-09_v0128-0-section-domain.md` + chronicles ADR/Lesson + auto-memory update + ROADMAP refresh

---

## Open questions deferred to v0.128.1+

- Cross-aggregate validation для credits sum (curriculum.total_credits == sum(items.credits)) — read-model query, не invariant. Decide where to surface: API field on Curriculum.Get или separate `/api/curricula/:id/validation` endpoint. Defer until product owner asks.
- Bulk operations API shape — `POST /api/sections/:id/items/bulk` with array body vs `PATCH` with diff payload. Decide в v0.128.2 plan.
- Frontend column ordering (sections × items 2D table) — UX call в v0.128.3 brainstorm session.
