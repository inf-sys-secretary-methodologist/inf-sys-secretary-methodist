# Use Case Audit Report

**Дата:** 2026-05-03
**Версия проекта на момент аудита:** 0.105.1
**Целевой релиз исправлений:** 0.102.3
**Источники:** Use Case диаграммы 5 ролей + `docs/roles-and-flows.md` v0.102.2

**Легенда:** ✅ полностью реализовано · ⚠️ частично · ❌ отсутствует или только заглушка

---

## 1. STUDENT

### Базовые UC

| UC | Статус | File:line | Что доделать |
|---|---|---|---|
| 1. Login | ✅ | `internal/modules/auth/interfaces/http/handlers/auth_handler.go:117-156` | — |
| 1. Logout | ❌ | — | **Logout endpoint отсутствует.** Нужно: usecase + handler + invalidate session |
| 1. Recover password | ⚠️ | `internal/modules/notifications/.../notification_handler.go:95` | Есть только `SendPasswordResetEmail` (notification), но **нет полноценного flow**: запрос reset → token → проверка → смена пароля. Нет usecase, нет handlers |
| 2. Dashboard | ✅ | `internal/modules/dashboard/.../dashboard_handler.go:35,65,91` | — |
| 3. Просматривать документы (через ACL) | ✅ | `internal/modules/documents/application/usecases/document_usecase.go:211` + `sharing_usecase.go:224` (HasPermission) | — |
| 4. Открыть вложение | ✅ | `internal/modules/files/application/usecases/file_usecase.go` | — |
| 5. Фильтр документов | ✅ | `document_usecase.go:211, 434` | — |
| 6. Календарь | ✅ | `schedule/.../event_handler.go:230,264,312` | — |
| 7. Сообщения (WS) | ✅ | `messaging/infrastructure/websocket/hub.go` + `messaging_handler.go:336-407` | — |
| 8. AI RAG-чат + цитаты | ✅ | `ai/.../ai_handler.go:53` + `chat_usecase.go:123,1500` (FormatRAGContext) | — |
| 9. Редактировать профиль | ✅ | `users/.../user_handler.go:76` + `user_usecase.go:103` | — |
| 10. Просматривать задания | ✅ | `tasks/.../task_handler.go:89,152` | — |
| 11. Просматривать объявления | ✅ | `announcements/.../announcement_handler.go:109,174,242` | — |
| 12. Уведомления (in-app+TG+email+WebPush) | ✅ | `notifications/interfaces/http/handlers/*` (4 канала) | — |
| 13. Личные настройки | ✅ | `notifications/.../preferences_handler.go:38,54` | — |

### Запреты — **КРИТИЧЕСКИЕ ДЫРЫ**

| Запрет | Статус | Замечание |
|---|---|---|
| НЕ создаёт документы | ❌ | Permission guard в Create-handler **не проверяет роль** — студент технически может вызвать create |
| НЕ видит отчёты/аналитику | ❌ | Reports/Analytics handlers **не блокируют student** |
| НЕ управляет пользователями | ⚠️ | Гард есть в `permission.go`, но не везде применяется в handlers |

### Shared UC

| UC | Статус | File:line |
|---|---|---|
| Сессии (httpOnly cookie + БД) | ✅ | `auth/.../middleware/auth_middleware.go` |
| PermissionMatrix loader | ⚠️ | per-document ACL есть, **глобальный matrix loader для role-action отсутствует** |
| ACL документа | ✅ | `documents/.../sharing_usecase.go:224` |
| RAG-поиск + цитаты | ✅ | `ai/.../chat_usecase.go:123` |
| Валидация форм (server-side) | ✅ | `internal/shared/infrastructure/validation/validator.go` |
| Audit log | ⚠️ | Логгер есть, **не везде вызывается** (особенно documents, users, integration) |

---

## 2. TEACHER

| UC | Статус | File:line | Замечания |
|---|---|---|---|
| 1. Login/logout/recover | ⚠️ | (см. студент) | Те же дыры в logout/recover |
| 2. Dashboard | ✅ | `frontend/src/app/dashboard/` | — |
| 3. Создать документ (validate, ACL, version, audit) | ✅ | `documents/.../document_usecase.go:50` | — |
| 4. Редактировать свой документ | ⚠️ | `document_usecase.go:127` | **Ownership check неполный** — нет явной проверки `document.AuthorID == current_user.id` |
| 5. Открыть документ | ✅ | `document_usecase.go:117` | — |
| 6. Запустить маршрут согласования | ❌ | `internal/modules/workflow/` пустая | **Workflow модуль не реализован** (issue #41 — заглушка по плану, OK) |
| 7. Создать задание | ✅ | `tasks/.../task_usecase.go:52` | — |
| 8. Проверить задание (оценка + notify) | ⚠️ | `task_usecase.go:105` | **SaveGrade usecase отсутствует** — Update задания есть, но логика выставления оценки нет |
| 9. Отчёт по своим группам (limited) | ⚠️ | `reporting/.../report_usecase.go:55,115` | **Scope-фильтр «только свои группы» не реализован** — преподаватель может видеть отчёты по чужим группам |
| 10. Создать событие | ✅ | `schedule/.../event_usecase.go:43` | — |
| 11. Сообщения | ✅ | `messaging/.../messaging_usecase.go:85` | — |
| 12. AI RAG | ✅ | `ai/.../chat_usecase.go:86` | — |
| 13-14. Профиль + настройки | ✅ | `users/.../user_usecase.go:103` + `frontend/src/app/settings/` | — |
| 16. Вложения к документу (#202) | ✅ | `document_usecase.go:78` (SetFile) | — |
| 17. Применить шаблон (limited) | ⚠️ | `template_usecase.go` | **Permission check для фильтра шаблонов отсутствует** — teacher видит все шаблоны (включая чужие методические) |
| 18. Поделиться документом | ✅ | `sharing_usecase.go` | — |
| 19. Произвольный согласующий | ❌ | `workflow/` пустой | Зависит от #6 |
| 20. ЭП #140 | ❌ | заглушка | **OK** — по плану заглушка |
| 21. Reminder deadline на задание | ⚠️ | `event_usecase.go:98` | Reminder есть **только для events**, нет `SetReminder` для tasks |
| 22. Фильтр на отчёт | ✅ | `frontend/src/app/reports/` | — |
| 23. Экспорт CSV/XLSX (limited) | ⚠️ | `reporting/...` | Export endpoint есть, **role-based scope не реализован** — преподаватель экспортирует чужие отчёты |
| 24. Recurrence event | ✅ | `event_usecase.go:66` | — |
| 25. Пригласить участников | ✅ | `event_usecase.go:76` | — |
| 26. Голосовой ввод #23 | ❌ | заглушка | **OK** — по плану заглушка |

### Запреты teacher

| Запрет | Статус |
|---|---|
| НЕ видит чужие отчёты | ⚠️ scope не реализован |
| НЕ управляет curriculum | ✅ `permission.go:102` (ActionApprove: Denied) |
| НЕ создаёт пользователей | ✅ `permission.go:97` |
| НЕ утверждает учебные планы | ✅ `permission.go:106` |
| НЕ имеет system settings | ✅ — вне PermissionMatrix |

**Итог teacher:** 14 ✅ / 8 ⚠️ / 4 ❌. Заглушки #6, #19, #20, #26 — OK по плану.

---

## 3. ACADEMIC_SECRETARY

| UC | Статус | File:line | Замечания |
|---|---|---|---|
| 1. Login | ✅ | `auth/...` | — |
| 2. Управлять документами CRUD | ⚠️ | `documents/.../document_usecase.go` | **`ResourceDocuments` ОТСУТСТВУЕТ в PermissionMatrix** (`permission.go:7`) — документы не контролируются через матрицу |
| 3. Управлять шаблонами | ✅ | `template_usecase.go` | — |
| 4. Управлять событиями full | ✅ | `schedule/.../event_usecase.go` | — |
| 5. Создать отчёт full | ✅ | `reporting/.../report_usecase.go` | — |
| 6. Аналитика студентов | ✅ | `analytics/.../analytics_usecase.go:39` | — |
| 7. Анализ посещаемости | ✅ | `cmd/server/main.go:445` (RiskRecalcScheduler 03:00) | — |
| 8. Сообщения WS | ✅ | `messaging/.../messaging_usecase.go:30` | — |
| 9. AI RAG | ✅ | `ai/.../chat_usecase.go:30` | — |
| 10. Профиль | ✅ | `users/.../user_usecase.go:49` | — |
| 11. Просматривать пользователей (limited) | ⚠️ | `permission.go:67-72` | Permission ОК, но **нет явной фильтрации по группам в handler** |
| 12. Личные настройки | ✅ | `cmd/server/main.go:1283` | — |
| 13. Вложения | ✅ | `document_handler.go:92` | — |
| 14. Применить шаблон | ✅ | `cmd/server/main.go:1500` | — |
| 15. Запустить согласование #41 | ❌ | заглушка | **OK** |
| 16. Авто-расписание #139 | ❌ | заглушка | **OK** |
| 17. iCal/Google #40 | ❌ | заглушка | **OK** |
| 18. Пригласить участников | ✅ | `event_usecase.go` | — |
| 19. Экспорт CSV/XLSX/PDF | ✅ | `analytics/.../analytics_handler.go` | — |
| 20. Drill-down по студенту | ✅ | `analytics_handler.go:69` | — |
| **21. Алерт по пропускам через n8n** | ⚠️ | `workflows/absence-alert.json` + `cmd/server/main.go:452` | **КРИТИЧНО: workflow JSON существует, но НЕ подключен.** `RiskRecalcScheduler` вызывает `NotificationUseCase`, **не n8n webhook**. `WebhookEventHandler.pathMap` (`internal/shared/infrastructure/n8n/event_handler.go:22-26`) обрабатывает только document events. Нужно добавить `risk-alert-detected` в pathMap и вызвать `TriggerAsync` в `riskAlertFunc` |
| 22. Фильтр по группам | ⚠️ | `users/.../user_usecase.go:70` | UserFilter есть, но guard `AccessLimited` явно не фильтрует по группам |
| 23. Голосовой ввод #23 | ❌ | заглушка | **OK** |

### Запреты secretary

| Запрет | Статус |
|---|---|
| НЕ создаёт пользователей | ✅ `permission.go:68` |
| НЕ управляет curriculum | ✅ `permission.go:74-76` |
| НЕ подписывает задания | ✅ `permission.go:85` |
| НЕ имеет system settings | ✅ |
| **ResourceDocuments в матрице** | ❌ **отсутствует — нужно добавить** |

---

## 4. METHODIST

| UC | Статус | File:line | Замечания |
|---|---|---|---|
| 1. Login/logout | ✅ | `auth/.../middleware/auth_middleware.go:19` | (logout — общая дыра) |
| **2. Curriculum CRUD** | ❌ | модуль отсутствует | **КРИТИЧНО: нет `internal/modules/curriculum/`**, не в routes, не в frontend nav. Permission matrix описана (`ActionCreate/Update: AccessFull`), но реализация нулевая |
| 3. Approve curriculum | ❌ | `internal/modules/workflow/` (0 файлов) | Заглушка по плану (`#41`), но **блокирует UC #2** |
| 4. Manage documents full CRUD | ✅ | `documents/.../document_handler.go` (14 handlers) | — |
| 5. Templates manage | ✅ | `documents/.../template_handler.go` | — |
| 6. Reports full | ✅ | `reporting/.../report_usecase.go` | — |
| 7. Analytics read | ✅ | `analytics/.../analytics_handler.go` | — |
| 8. Messaging WS | ✅ | `messaging/infrastructure/websocket/hub.go` | — |
| 9. AI расширенный | ✅ | `ai/.../` | (нет ролевых ограничений в коде — все роли equal в AI) |
| 10. Edit profile | ✅ | `users/.../profile_handler.go` | — |
| 11. Personal settings | ✅ | `frontend/src/config/navigation.ts:242` | — |
| 12. View users limited | ✅ | `users/.../user_handler.go` | — |

### Запреты methodist

| Запрет | Статус |
|---|---|
| НЕ создаёт пользователей | ✅ |
| НЕ интеграции/system settings | ✅ — вне permission matrix, не в nav (`navigation.ts:230-238`) |
| ⚠️ Permission guards не везде применяются в handlers | ⚠️ только 1 `RequireRole("admin")` найден, для остального опираемся на matrix без middleware |

**Итог methodist:** 11/12 базовых UC работают. Главный блокер — **отсутствие curriculum модуля** (отдельная большая фича, не вписывается в патч 0.102.3).

---

## 5. SYSTEM_ADMIN

| UC | Статус | File:line | Замечания |
|---|---|---|---|
| 1. Login (с MFA) | ⚠️ | `auth/.../auth_handler.go` | **MFA не реализован**, только базовый login |
| 2. Управление пользователями (CRUD) | ✅ | `users/.../user_usecase.go` | UpdateUserRole, UpdateUserStatus работают; **нет frontend `/admin/users`** |
| 3. Управление ролями + permissions | ✅ | `permission.go:7-154` + `RequireRole` middleware | **Нет UI редактирования матрицы** (read-only из кода) |
| 4a. Settings: General | ⚠️ | `frontend/src/app/admin/settings/` | Только 3 страницы (appearance/automation/notifications), **нет общих настроек** (SMTP host etc) |
| 4b. Settings: Security | ❌ | — | **Не найдено**: 2FA, rate limiting, IP whitelist |
| 4c. Settings: Branding | ⚠️ | `appearance/page.tsx:34-59` | UI есть, но **input'ы disabled**, нет API сохранения |
| 5a. Интеграции 1C | ✅ | `integration/.../sync_handler.go:27-73` | StartSync/ListSyncLogs/GetSyncLog работают |
| 5b. n8n workflows | ⚠️ | `automation/page.tsx:38-51` | **Список workflows жёстко захардкожен**, нет UI управления |
| 5c. Composio | ❌ | — | **Нет admin UI** для Gmail/Telegram bot |
| 5d. Sentry | ❌ | — | Инициализируется в main.go, **нет admin UI** |
| 5e. Web Push VAPID | ⚠️ | `notifications/page.tsx` | Страница есть, **VAPID хранится в env, нет UI** |
| 6. Audit log просмотр | ⚠️ | `internal/shared/infrastructure/logging/security_logger.go:177-213` | Logger есть, пишется в reporting/schedule/tasks; **нет `/admin/audit-logs` UI**, не везде вызывается |
| 7. Backup/Restore | ❌ | — | **Не найдено** ни UI ни API |
| 8. Мониторинг | ✅ | `docs/roles-and-flows.md:104-109` (Grafana) | Подключено, **нет ссылок в admin UI** |
| 9. Documents full | ✅ | matrix AccessFull | — |
| 10. Аналитика read all | ✅ | — | — |
| 11. Override других ролей | ✅ | `permission.go:6-154` | RoleSystemAdmin = AccessFull везде |
| 12. Brand customization | ⚠️ | (см. 4c) | UI без сохранения |

### КРИТИЧЕСКАЯ УЯЗВИМОСТЬ

> `RequireRole("admin")` middleware **НЕ применяется** на `/api/integration/sync/*` — sync 1C может запустить **любой авторизованный пользователь**.

Файл: `cmd/server/main.go` (поиск регистрации sync routes).

---

# Сводный список доработок

## Критические (security/blocker) — обязательно в 0.102.3

1. **Permission guards для student** — заблокировать создание документов, отчётов, аналитики на handler-уровне (не только в permission matrix)
2. **`ResourceDocuments` добавить в PermissionMatrix** (`permission.go`)
3. **`RequireRole("admin")` на `/api/integration/sync/*`** — закрыть уязвимость
4. **Logout endpoint** — отсутствует полностью
5. **Password recovery flow** — есть только email-сервис, нет usecase + handlers
6. **n8n absence alert connection** — workflow JSON есть, не подключен в `RiskRecalcScheduler`. Add `risk-alert-detected` в `WebhookEventHandler.pathMap` + вызвать в `riskAlertFunc`

## Средние — желательно в 0.102.3

7. **Ownership check на edit document** (teacher) — явная проверка `AuthorID`
8. **Scope-фильтр teacher reports «только свои группы»** + такой же фильтр на export
9. **SaveGrade usecase** в tasks — выставление оценки за задание
10. **SetReminder для tasks** (deadline reminder)
11. **Permission filter для templates** — teacher видит только не-методические
12. **Group filter в users list** для secretary

## Отдельные крупные фичи — НЕ в 0.102.3, выделить в отдельные релизы

13. **Curriculum модуль** (для methodist) — отдельный релиз, неделя+ работы
14. **Workflow модуль (#41)** — заглушка по спеке, OK оставить
15. **MFA для admin** — отдельная security-фича
16. **Backup/Restore admin UI** — отдельная фича
17. **`/admin/audit-logs` UI** — отдельная фича
18. **Composio/Sentry admin UI** — отдельная фича
19. **VAPID/branding API сохранения** — отдельная фича
20. **/admin/users frontend** — отдельная мелкая фича

## Заглушки — оставить как есть (по плану задачи)

- `#23` Голосовой ввод AI
- `#40` iCal/Google export
- `#41` Workflow approval
- `#139` Auto-schedule CSP
- `#140` ЭП

## Audit log

- Дописать вызовы LogAuditEvent в `documents`, `users`, `integration` usecases (не везде сейчас)

---

# Что войдёт в релиз 0.102.3

Только пункты 1–12 (критические + средние). Пункты 13–20 — отдельными релизами по согласованию с пользователем (см. план тасков).
