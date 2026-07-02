# Next session prompt — inf-sys-secretary-methodist (ДИПЛОМНЫЙ ПРОЕКТ)

> Отдельно от `NEXT_SESSION_PROMPT.md` (тот — про учебные Go-курсы). Этот файл — про код дипломного проекта.
> main `19398f7f`, VERSION **0.222.0**. **#40 (внешние календари) закрыт E2E, релиз v0.222.0.**
> Осталась **РОВНО ОДНА open issue — #139 (автогенерация расписания)**. После неё продуктовый бэклог пуст.

## ШАГ 0 — прочитать
- `SESSION_START.md` (пересобрать `python3 _tools/build_session_start.py` при надобности) + handoff `.claude/handoffs/2026-07-02_issue-40-ical-feed-релиз.md`.
- Auto-память: [[project_ical_feed_issue40]] (там дорогой урок про мерж stacked-PR), MEMORY.md.

## Контекст: что только что сделано (сессия 2026-07-02)
- **#40 iCal-фид подписки** зашиплен стеком из 5 PR (#486 рендерер RFC 5545 / #491 токен+миграция 054 / #488 usecase / #489 HTTP+DI / #490 frontend), релиз v0.222.0.
- Секретный URL `/api/public/calendar/{token}/feed.ics`, подписка в Google/Outlook/Apple, ноль внешних API.
- Код: `internal/modules/schedule/domain/ical/`, `.../application/usecases/calendar_feed_usecase.go`, `.../interfaces/http/handlers/calendar_feed_handler.go`, `migrations/054_*`, frontend `frontend/src/app/settings/calendar/`.

## ГЛАВНАЯ ЗАДАЧА — #139 автогенерация расписания (в работе, 2 слайса смержены)
CSP-солвер на чистом Go. **Продуктовые решения и полный план — в [[project_schedule_autogen_issue139]] (читать первым).** Дизайн-док: `docs/plans/2026-07-02-schedule-autogen-design.md`.

**Прогресс (сессия 2026-07-02):**
- ✅ Slice 1 `lesson_slots` (каталог пар) — MERGED PR#493.
- ✅ Slice 2a `teaching_load` backend — MERGED PR#494.
- ✅ Slice 2b `teaching_load` frontend + **envelope-fix** — MERGED PR#495, main `8d0e6d13`.
  - ⚠️ Обнаружено: модуль schedule отдавал ГОЛЫЙ JSON вместо конверта `{success,data}` → нагрузка+reference-дропдауны читали пустоту. Починил (response.Success). **Урок: у любого schedule-эндпоинта проверяй обёртку.**
  - **FOLLOW-UP (не сделано):** lesson List/GetTimetable/GetByID/Create/Update + changes ещё голые → `scheduleLessonsApi`-ридеры сломаны (таймтейбл). Доесть конверт по всему lesson_handler отдельным PR (можно перед Slice 4).

**СЛЕДУЮЩЕЕ — Slice 3, CSP-солвер (domain, чистый Go).** Пакет `internal/modules/schedule/domain/solver/`, без I/O, полностью юнит-тестируемый. Готовый дизайн (реализовать по TDD, файл за файлом RED→GREEN):

```
types.go (данные, без логики):
  Variable{ ID int; LoadID,GroupID,TeacherID,DisciplineID,LessonTypeID int64;
            GroupSize int; AllowedRoomTypes []string; WeekType domain.WeekType }
  Value{ Day domain.DayOfWeek; Slot int; RoomID int64 }
  Room{ ID int64; Capacity int; Type string; Available bool }
  Input{ Variables []Variable; Days []domain.DayOfWeek; Slots []int; Rooms []Room; Weights SoftWeights }
  Assignment{ Variable; Value }
  Result{ Assignments []Assignment; Unplaced []Variable }

constraints.go (TDD первым — чистые предикаты):
  parityConflicts(a,b domain.WeekType) bool  // all↔любой=true; odd&odd/even&even=true; odd&even=false
  assignmentsConflict(a1,a2 Assignment) bool // тот же Day+Slot И parityConflicts И (тот же Teacher | Group | Room)

domain.go (H4 — построение доменов):
  buildDomain(v Variable, in Input) []Value // все Day×Slot×Room где room.Available && room.Capacity>=v.GroupSize && roomTypeOK(v.AllowedRoomTypes, room.Type); roomTypeOK: len(allowed)==0(любой) || contains
  // ⚠️ пустой домен переменной = заведомо Unplaced

solver.go (H1-H3, backtracking):
  Solve(in Input) Result
  // MRV: выбирать переменную с наименьшим числом оставшихся валидных значений
  // forward-checking: после присваивания отсеивать конфликтующие значения у соседей; откат если у кого-то домен опустел
  // best-effort: если полное решение не найдено (таймбокс по числу шагов ИЛИ вернуть частичное) → положить нерасставленные в Unplaced, НЕ падать
  // порядок значений — по soft-штрафу (см. soft.go), меньше штраф раньше (эвристика LCV-подобная)

soft.go (мягкие, TDD — проверять что предпочитает лучшее):
  SoftWeights{ GroupGap,TeacherGap,DaySpread,EarlySlot float64 } (дефолты в NewDefaultWeights)
  penalty(candidate Assignment, current []Assignment, w SoftWeights) float64
  // GroupGap: штраф за «окно» в дне у группы (несмежные слоты) ; TeacherGap: то же для препода
  // DaySpread: штраф за перегруз одного дня у группы (равномерность) ; EarlySlot: штраф ~ Slot (ранние лучше)
```

Импортить `schedule/domain` (WeekType/DayOfWeek) — можно (domain→domain). Детерминизм тестов: НЕ использовать map-итерацию для выбора (сортировать), Math.rand запрещён. Порядок: constraints → domain(H4) → solver(H1-H3 полное решение на маленьком входе) → best-effort(unplaced на переполненном входе) → soft(ordering). code-review ≥9. Мердж линейно в main. Правила совместимости room_type↔lesson_type (какой тип занятия→какие типы аудиторий) вычисляет Slice 4 (usecase) и кладёт в `Variable.AllowedRoomTypes` — солвер их только проверяет (остаётся чистым движком).
- **Slice 4 — usecase генерации + HTTP** — собрать teaching_load+lesson_slots+classrooms(caps/type)+groups(size) по семестру, развернуть в переменные, прогнать солвер, вернуть черновик (preview DTO, БЕЗ записи) + отдельный `POST /schedule/generate/apply` пишет через `LessonRepository`. Использовать `lessonSlot.GetByID`/список слотов.
- **Slice 5 — frontend генерации** — кнопка «Сгенерировать расписание» + preview-сетка + «Применить», i18n×4.

TDD по слоям как в #40/Slice1-2. Мержить каждый слайс линейно в main (НЕ стекать).

## Незакрытые долги репозитория
- [[project_rag_searchmode_tdd_followup]] — локальные RAG search_mode/reranker/FTS без тестов (не пушить as-is; доделать по TDD). Бэкапы: `~/rag-local-changes-2026-07-01.patch` + `stash@{0}`.

## Рабочие правила (напоминание)
TDD строго (RED+GREEN, 2 коммита) · DDD (инварианты в domain, repo-интерфейсы в usecase, доменные `var Err…`) · Clean Arch (handler=парс→usecase→map) · i18n×4 · code-review ≥8 перед «сделано» · полный функционал без MVP-cuts · не пушить в main напрямую · **мерж stacked-PR: сперва retarget ВСЕХ детей `--base main`, потом low→high с `git rebase --onto origin/main <ориг-tip-родителя>`** (иначе `--delete-branch` закроет дочерний PR безвозвратно) · релизный bump `_tools/bump_version.sh <ver>` (8 файлов, стейджить `VERSION` заглавными).
