# План критических задач — TDD + DDD + Clean Architecture

> Создан 2026-04-25. Порядок исполнения: от наименьшей к наибольшей. Каждое поведенческое изменение = 2 коммита (RED → GREEN).

## Порядок задач

| Порядок | # | Задача | Оценка | Тип |
|---------|---|--------|--------|-----|
| 1 | #199 | SECURITY: Self-registration | ~4 часа | Backend (Go) + Frontend (React) |
| 2 | #200 | Tasks Module Frontend | 3-5 дней | Frontend |
| 3 | #201 | Schedule Lessons Frontend | 4-6 дней | Frontend |
| 4 | #196 | Backend Test Coverage 90% | 2-3 недели | Backend tests |

---

## Задача #199 — Self-registration privilege escalation

### DDD-анализ

**Где живёт инвариант?** Бизнес-правило "при self-registration разрешены не все роли" — это инвариант **домена авторизации**. Принадлежит пакету `internal/modules/auth/domain`.

**Где живут ошибки?** Доменные ошибки идут в `internal/modules/auth/domain/entities/user.go` (рядом с `ErrAccountNotActive`).

### Clean Architecture слои

| Слой | Файл | Что делает |
|------|------|------------|
| **Domain** | `auth/domain/role.go` | Метод `RoleType.IsAllowedForSelfRegistration() bool` + `var ErrRoleNotAllowedForSelfRegistration` |
| **Usecase** | `auth/application/usecases/auth_usecase.go` | `Register()` валидирует роль через домен, возвращает `ErrRoleNotAllowedForSelfRegistration` |
| **Handler** | `auth/interfaces/http/handlers/auth_handler.go` | Маппит ошибку в HTTP 403 (без бизнес-логики) |
| **Frontend** | `frontend/src/components/auth/RegisterForm.tsx` | Убирает привилегированные роли из `<select>` |

### TDD-план (по 2 коммита на каждое поведение)

#### Поведение 1: домен знает какие роли допустимы для self-registration

**Шаг 1.1 — RED**
- Файл: `internal/modules/auth/domain/role_test.go`
- Тест: `TestRoleType_IsAllowedForSelfRegistration` (table-driven, 5 ролей)
  - `student` → true
  - `teacher` → true
  - `methodist` → false
  - `academic_secretary` → false
  - `system_admin` → false
- Запуск: `go test ./internal/modules/auth/domain/ -run IsAllowedForSelfRegistration` → FAIL (метода нет)
- Коммит: `test(auth): add failing test for RoleType.IsAllowedForSelfRegistration`

**Шаг 1.2 — GREEN**
- Файл: `internal/modules/auth/domain/role.go`
- Добавить:
  ```go
  func (r RoleType) IsAllowedForSelfRegistration() bool {
      return r == RoleStudent || r == RoleTeacher
  }
  ```
- Запуск: тест проходит
- Коммит: `feat(auth): add IsAllowedForSelfRegistration domain method`

#### Поведение 2: usecase отвергает регистрацию с привилегированной ролью

**Шаг 2.1 — RED**
- Файл: `internal/modules/auth/application/usecases/auth_usecase_test.go`
- Тест: `TestAuthUseCase_Register_RejectsPrivilegedRoles` (table-driven)
  - input role `system_admin` → `errors.Is(err, domain.ErrRoleNotAllowedForSelfRegistration)`
  - input role `methodist` → same
  - input role `academic_secretary` → same
  - input role `student` → no error
  - input role `teacher` → no error
- Запуск: FAIL
- Коммит: `test(auth): add failing test for Register role whitelist`

**Шаг 2.2 — GREEN**
- Файл: `internal/modules/auth/domain/role.go` — добавить:
  ```go
  var ErrRoleNotAllowedForSelfRegistration = errors.New("role not allowed for self-registration")
  ```
- Файл: `internal/modules/auth/application/usecases/auth_usecase.go` — в начале `Register()`:
  ```go
  role := domain.RoleType(input.Role)
  if !role.IsAllowedForSelfRegistration() {
      u.logRegistration(ctx, input.Email, input.Role, false, "role not allowed for self-registration")
      return domain.ErrRoleNotAllowedForSelfRegistration
  }
  ```
- Коммит: `feat(auth): reject self-registration with privileged roles`

#### Поведение 3: handler мапит ошибку в HTTP 403

**Шаг 3.1 — RED**
- Файл: `auth/interfaces/http/handlers/auth_handler_unit_test.go`
- Тест: `TestRegister_ReturnsForbiddenForPrivilegedRole`
  - input role = `system_admin` → response status 403, body содержит код ошибки
- FAIL
- Коммит: `test(auth): add failing test for Register HTTP 403 on privileged role`

**Шаг 3.2 — GREEN**
- Файл: `auth_handler.go` Register() — добавить ветку:
  ```go
  if errors.Is(err, domain.ErrRoleNotAllowedForSelfRegistration) {
      c.JSON(http.StatusForbidden, gin.H{...})
      return
  }
  ```
- Коммит: `feat(auth): map ErrRoleNotAllowedForSelfRegistration to HTTP 403`

#### Поведение 4: frontend RegisterForm не показывает привилегированные роли

**Шаг 4.1 — RED**
- Файл: `frontend/src/components/auth/__tests__/RegisterForm.test.tsx`
- Тест: `does not show privileged roles in select`
  - render → query `<option>` → ожидаем `student`, `teacher` присутствуют, `system_admin`, `methodist`, `academic_secretary` отсутствуют
- FAIL
- Коммит: `test(auth-ui): add failing test for RegisterForm role select whitelist`

**Шаг 4.2 — GREEN**
- Файл: `RegisterForm.tsx` — убрать 3 привилегированные `<option>`
- Коммит: `feat(auth-ui): remove privileged roles from RegisterForm select`

### Что НЕ делаем в #199 (out of scope)

- Admin endpoint `POST /api/admin/users` — отдельная задача (есть в users module уже частично)
- Audit log при попытке атаки — оставим существующий `logRegistration`, который уже пишет fail
- Пенetest — отдельно, не часть кода

### Verification

После всех 8 коммитов:
- `go test ./internal/modules/auth/...` — green
- `npm test -- RegisterForm` — green
- Запустить агента `superpowers:code-reviewer` с промптом: "беспристрастно, оценка 1-10 по TDD/DDD/CA, конкретные файлы+строки"
- Если **каждая ось ≥ 8/10** — заявляем done

---

## Задача #200 — Tasks Module Frontend

### Тип задачи
Чистый фронтенд (Next.js + React). Backend готов. **TDD здесь = component tests на Jest+RTL перед компонентом**, DDD не применяется (это UI-слой), CA = разделение страница/компоненты/hooks/api-клиент.

### Архитектура

```
frontend/src/
├── app/tasks/
│   ├── page.tsx                  # список + фильтры (server-side data fetch)
│   └── [id]/page.tsx             # детальная задача
├── components/tasks/
│   ├── TaskCard.tsx
│   ├── TaskForm.tsx
│   ├── TaskFilters.tsx
│   ├── KanbanColumn.tsx
│   ├── KanbanBoard.tsx
│   └── __tests__/                # RTL тесты
├── hooks/
│   ├── useTasks.ts
│   ├── useTask.ts
│   └── useProjects.ts
├── lib/api/
│   └── tasks.ts                  # API клиент
└── types/
    └── tasks.ts                  # Task, Project, TaskStatus types
```

### TDD-план (frontend)

Каждый компонент — 2 коммита (тест RED → компонент GREEN):

1. `useTasks` hook — тест мока API → реализация
2. `useTask` hook — аналогично
3. `useProjects` hook — аналогично
4. `TaskCard` component — тест рендера → реализация
5. `TaskFilters` — тест фильтрации → реализация
6. `KanbanColumn` — тест drag-drop → реализация
7. `KanbanBoard` — тест общей картины → реализация
8. `/tasks` page — e2e тест → реализация
9. `/tasks/[id]` page — e2e тест → реализация
10. Навигация в sidebar — тест видимости → добавление
11. i18n переводы → ru/en/fr/ar

Итого ~22 коммита.

---

## Задача #201 — Schedule Lessons Frontend

### Архитектура

Аналогично #200, но больше UI-сложности:

```
frontend/src/
├── app/schedule/
│   └── page.tsx                  # сетка + фильтры
├── components/schedule/
│   ├── TimetableGrid.tsx         # главная сетка дни×пары
│   ├── LessonCard.tsx
│   ├── ScheduleFilters.tsx       # группа/преподаватель/аудитория
│   ├── ScheduleChangeForm.tsx    # форма замены
│   ├── ScheduleExporter.tsx      # iCal/PDF
│   └── __tests__/
├── hooks/
│   ├── useScheduleLessons.ts
│   ├── useScheduleChanges.ts
│   └── useGroups.ts              # для фильтра
├── lib/api/schedule.ts
└── lib/export/ical.ts            # iCal генератор
```

### Важные нюансы

- Сетка 6 дней × 8 пар — `<table>` или CSS Grid
- Замены (`schedule_changes`) рендерятся поверх обычных пар с visual indicator
- Мобильный view — переключение на список вместо сетки
- Доступ по ролям: read-only для student/teacher, edit для academic_secretary/methodist/admin

TDD по компонентам, ~25-30 коммитов.

---

## Задача #196 — Backend Test Coverage до 90%

### Тип
**Это НЕ TDD** — код уже написан. Это **backfill coverage** — честно называть это "test: backfill coverage for X" в коммитах.

### Приоритет модулей (по текущему покрытию)

| Модуль | Текущее % | Цель | Сложность |
|--------|:---------:|:----:|:---------:|
| documents/usecases | 33.0% | 90% | high (3 usecase: document, version, sharing) |
| auth/usecases | 33.6% | 90% | medium (1 usecase) |
| reporting/usecases | 38.6% | 90% | high |
| notifications/usecases | 37.4% | 90% | medium |
| messaging/usecases | 59.2% | 90% | low |

Также:
- handlers всех модулей
- repositories (требуют integration tests с testcontainers)

### Подход

Для каждого модуля по очереди:
1. Запустить покрытие: `go test ./internal/modules/X/... -cover -coverprofile=cov.out`
2. Открыть `go tool cover -html=cov.out` — найти красные ветки
3. Написать table-driven тесты для непокрытых веток
4. Использовать gomock для зависимостей
5. Коммит: `test(X): backfill coverage to NN% for usecase Y`

Каждые **+10% общего покрытия бэкенда** — нотификация пользователю (feedback memory).

### Не делаем

- Не пишем тесты для `cmd/server/main.go` (DI wiring, тестируется e2e)
- Не покрываем сгенерированный код (`docs/swagger/`)

---

## Общий verification gate

После завершения каждой задачи:
1. `just test` — все тесты зелёные
2. `just lint` — без ошибок
3. Запустить `superpowers:code-reviewer` с жёстким промптом
4. Все оси (TDD / DDD / CA) ≥ 8/10
5. Только тогда заявить done и переходить к следующей задаче
