# Next session prompt — inf-sys-secretary-methodist (ДИПЛОМНЫЙ ПРОЕКТ)

> Отдельно от `NEXT_SESSION_PROMPT.md` (тот — про учебные Go-курсы). Этот файл — про код дипломного проекта.
> main `a31dc259`, VERSION **0.222.0** (Slice 3 — доменный слой, без релиза/миграций).
> Осталась **РОВНО ОДНА open issue — #139 (автогенерация расписания)**, слайсы 1-3/5 смержены. После неё продуктовый бэклог пуст.

## ШАГ 0 — прочитать
- `SESSION_START.md` (пересобрать `python3 _tools/build_session_start.py` при надобности) + handoff `.claude/handoffs/2026-07-02_issue-40-ical-feed-релиз.md`.
- Auto-память: [[project_ical_feed_issue40]] (там дорогой урок про мерж stacked-PR), MEMORY.md.

## Контекст: что только что сделано (сессия 2026-07-02)
- **#40 iCal-фид подписки** зашиплен стеком из 5 PR (#486 рендерер RFC 5545 / #491 токен+миграция 054 / #488 usecase / #489 HTTP+DI / #490 frontend), релиз v0.222.0.
- Секретный URL `/api/public/calendar/{token}/feed.ics`, подписка в Google/Outlook/Apple, ноль внешних API.
- Код: `internal/modules/schedule/domain/ical/`, `.../application/usecases/calendar_feed_usecase.go`, `.../interfaces/http/handlers/calendar_feed_handler.go`, `migrations/054_*`, frontend `frontend/src/app/settings/calendar/`.

## ГЛАВНАЯ ЗАДАЧА — #139 автогенерация расписания (в работе, 2 слайса смержены)
CSP-солвер на чистом Go. **Продуктовые решения и полный план — в [[project_schedule_autogen_issue139]] (читать первым).** Дизайн-док: `docs/plans/2026-07-02-schedule-autogen-design.md`.

**Прогресс:**
- ✅ Slice 1 `lesson_slots` (каталог пар) — MERGED PR#493.
- ✅ Slice 2a `teaching_load` backend — MERGED PR#494.
- ✅ Slice 2b `teaching_load` frontend + **envelope-fix** — MERGED PR#495, main `8d0e6d13`.
  - ⚠️ Урок: модуль schedule отдавал ГОЛЫЙ JSON вместо конверта `{success,data}` → нагрузка+reference-дропдауны читали пустоту. Починил (response.Success).
- ✅ **Slice 3 CSP-солвер (domain, чистый Go) — MERGED, main `a31dc259`.** Пакет `internal/modules/schedule/domain/solver/`: types/constraints(H1-H3 parity)/domainbuild(H4)/soft(penalty parity-aware)/solver(Solve: backtracking+MRV+forward-checking+best-effort greedy). 98.7% покрытие, code-review SHIP 9/9/10/9/9/9. TDD-парами.
  - ⚠️ **Мерж-грабли (важно для Slice 4/5):** одиночный PR был 1020 стр → «Check PR Size» **hard-fail >1000**; ветка `feat/...` не прошла «Validate branch naming» (нужен `feature|task|bugfix|hotfix|config|chore|docs|refactor`/(`issue-<N>[.M]`|`v<x.y.z>`)-desc). Решил 2 ПОСЛЕДОВАТЕЛЬНЫМИ (не стек) PR: #497 foundation (594)→merge main→`git rebase --onto main <cut>`→#498 search (428). **Держи PR <1000 и имя ветки по конвенции.**

**СЛЕДУЮЩЕЕ — Slice 4, usecase генерации + HTTP (preview→apply).**
- Собрать по семестру: `teaching_load` (нагрузки) + `lesson_slots` (сетка пар, `GetByID`/список) + classrooms (caps/type — модуль classrooms) + groups (size — модуль groups) + teachers (users role=teacher).
- **Вычислить `AllowedRoomTypes` по типу занятия** (правило room_type↔lesson_type живёт ЗДЕСЬ, в usecase, не в солвере — солвер уже готов их только проверять). Развернуть нагрузки в `[]solver.Variable` (pairs_per_week раз каждую), классы в `[]solver.Room`.
- Прогнать `solver.Solve(Input)`, вернуть **черновик preview DTO БЕЗ записи** (`POST /api/schedule/generate`). Отдельный `POST /api/schedule/generate/apply` пишет `schedule_lessons` через `LessonRepository`.
- ⚠️ **Follow-up перед этим слайсом:** доесть конверт `{success,data}` по всему lesson_handler (lesson List/GetTimetable/GetByID/Create/Update + changes ещё голые → `scheduleLessonsApi`-ридеры/таймтейбл сломаны). + опц. полное инкрементальное forward-checking в солвере (сейчас per-node O(n²·D), для больших семестров медленно — best-effort greedy спасает, но качество ниже).
- **Slice 5 — frontend генерации** — кнопка «Сгенерировать расписание» + preview-сетка + «Применить», i18n×4.

Солвер (готов) — конвенции: импортит только `schedule/domain`+stdlib; `Solve(in Input) Result{Assignments,Unplaced}`; детерминирован; правила room_type↔lesson_type НЕ в нём.

TDD по слоям как в #40/Slice1-2. Мержить каждый слайс линейно в main (НЕ стекать).

## Незакрытые долги репозитория
- [[project_rag_searchmode_tdd_followup]] — локальные RAG search_mode/reranker/FTS без тестов (не пушить as-is; доделать по TDD). Бэкапы: `~/rag-local-changes-2026-07-01.patch` + `stash@{0}`.

## Рабочие правила (напоминание)
TDD строго (RED+GREEN, 2 коммита) · DDD (инварианты в domain, repo-интерфейсы в usecase, доменные `var Err…`) · Clean Arch (handler=парс→usecase→map) · i18n×4 · code-review ≥8 перед «сделано» · полный функционал без MVP-cuts · не пушить в main напрямую · **мерж stacked-PR: сперва retarget ВСЕХ детей `--base main`, потом low→high с `git rebase --onto origin/main <ориг-tip-родителя>`** (иначе `--delete-branch` закроет дочерний PR безвозвратно) · релизный bump `_tools/bump_version.sh <ver>` (8 файлов, стейджить `VERSION` заглавными).
