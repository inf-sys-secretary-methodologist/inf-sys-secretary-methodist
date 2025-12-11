# Миграции базы данных

Этот каталог содержит миграции базы данных для проекта "Информационная система секретаря-методиста".

## Обзор

Проект использует [golang-migrate](https://github.com/golang-migrate/migrate) для управления миграциями БД. Миграции написаны на SQL и следуют последовательной нумерации.

## Файлы миграций

| Миграция | Описание |
|----------|----------|
| `001_create_users_table` | Управление пользователями с ролями (admin, secretary, methodist, teacher, student) |
| `002_create_sessions_table` | JWT refresh token сессии |
| `003_create_documents_schema` | Управление документами с workflow (10 таблиц) |
| `004_create_schedule_schema` | Система расписания учебного заведения (13 таблиц) |
| `005_create_tasks_schema` | Управление задачами и поручениями (11 таблиц) |
| `006_create_reports_schema` | Система отчётности и аналитики (9 таблиц) |
| `007_create_events_schema` | Система мероприятий и событий |
| `008_create_announcements_schema` | Система объявлений |
| `009_create_users_module_schema` | Расширение модуля пользователей (подразделения, должности) |
| `010_create_files_schema` | Интеграция с MinIO хранилищем |
| `011_add_fulltext_search` | Полнотекстовый поиск документов (PostgreSQL FTS) |
| `012_create_document_public_links` | Публичные ссылки для документов |
| `013_enhance_document_versioning` | Расширенное версионирование документов (Issue #12) |

**Итого**: 50+ таблиц + 1 служебная таблица `schema_migrations`

## Предварительные требования

1. **PostgreSQL** 12+ установлен и запущен
2. **Docker** для запуска миграций через контейнер migrate
3. **Переменные окружения** настроены в файле `.env`:
   ```bash
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=inf_sys_db
   DB_USER=postgres
   DB_PASSWORD=postgres
   ```

## Настройка базы данных

### Создание базы данных

База данных создаётся автоматически при запуске `docker compose up`. Если нужно создать вручную:

```bash
# Создать production базу
docker exec postgres-dev psql -U postgres -c "CREATE DATABASE inf_sys_db;"

# Создать тестовую базу
docker exec postgres-dev psql -U postgres -c "CREATE DATABASE inf_sys_db_test;"
```

## Запуск миграций

Проект использует `justfile` для автоматизации задач. Доступные команды миграций:

### Применить все миграции

```bash
just migrate-up
```

Применяет все отложенные миграции по порядку.

### Откатить все миграции

```bash
just migrate-down
```

⚠️ **Внимание**: Это удалит все таблицы и данные.

### Откатить одну миграцию

```bash
just migrate-down-one
```

Откатывает только последнюю миграцию.

### Откатить N миграций

```bash
just migrate-down-n 2
```

Откатывает указанное количество миграций.

### Перейти к конкретной версии

```bash
just migrate-goto 3
```

Мигрирует вверх или вниз до указанной версии.

### Проверить текущую версию

```bash
just migrate-version
```

Показывает текущую версию миграции в БД.

### Форсировать версию (восстановление)

```bash
just migrate-force 3
```

Принудительно устанавливает версию миграции без применения. Полезно для восстановления из "грязного" состояния.

### Создать новую миграцию

```bash
just migrate-create add_users_avatar
```

Создаёт пару новых файлов миграции:
- `007_add_users_avatar.up.sql`
- `007_add_users_avatar.down.sql`

## Ручной запуск (без justfile)

Если предпочитаете запускать миграции вручную через Docker:

```bash
# Применить все миграции
docker run --rm -v "$(pwd)/migrations:/migrations" --network host \
  migrate/migrate:latest \
  -path=/migrations \
  -database "postgres://postgres:postgres@localhost:5432/inf_sys_db?sslmode=disable" \
  up

# Откатить все миграции
docker run --rm -v "$(pwd)/migrations:/migrations" --network host \
  migrate/migrate:latest \
  -path=/migrations \
  -database "postgres://postgres:postgres@localhost:5432/inf_sys_db?sslmode=disable" \
  down

# Проверить версию
docker run --rm -v "$(pwd)/migrations:/migrations" --network host \
  migrate/migrate:latest \
  -path=/migrations \
  -database "postgres://postgres:postgres@localhost:5432/inf_sys_db?sslmode=disable" \
  version
```

## Документация схем

### Модуль Documents (003)

Полное управление жизненным циклом документов с поддержкой workflow:

- **document_types**: Типы документов с кодами (memo, order_main, order_hr, order_admin, directive, business_letter, protocol, contract, job_instruction)
- **document_categories**: Иерархическая категоризация
- **documents**: Основная таблица с workflow статусами (черновик → зарегистрирован → маршрутизация → согласование → утверждён/отклонён → исполнение → исполнен → архив)
- **document_versions**: Контроль версий документов
- **document_routes**: Workflow согласования с поддержкой параллельного/последовательного согласования
- **document_permissions**: Детальное управление доступом
- **document_relations**: Связи между документами (ответ, вложение, ссылка, заменяет, исправление)
- **document_tags** + **document_tag_relations**: Система тегов
- **document_history**: Полный аудит с действиями пользователей, IP, user agent

**Ключевые возможности**:
- Движок workflow с многоэтапным согласованием
- Мягкое удаление (deleted_at)
- JSONB для гибких метаданные
- Шаблоны нумерации документов (например, "ПР-{YYYY}-{###}")
- Политики хранения
- Полный аудит-трейл

### Расширенное версионирование (013)

Миграция `013_enhance_document_versioning` добавляет:

- **Расширение `document_versions`**: Новые колонки для хранения полного снимка документа (title, subject, status, mime_type, metadata, storage_key)
- **Таблица `document_version_diffs`**: Кэш сравнений между версиями для ускорения diff-операций
- **Триггер `create_document_version_on_update`**: Автоматическое создание версий при изменении ключевых полей документа

**Автоматическое версионирование**: При обновлении документа PostgreSQL триггер автоматически сохраняет предыдущее состояние в `document_versions`, если изменились поля: title, subject, content, file_path, status.

### Модуль Schedule (004)

Система расписания для образовательного учреждения:

- **academic_years**: Учебные годы (2023/2024)
- **semesters**: Осенний/весенний семестры с датами
- **faculties**: Институты с деканами
- **departments**: Кафедры с заведующими
- **specialties**: Образовательные программы (например, 38.03.01)
- **student_groups**: Группы типа "ЭК-201" с курсом и куратором
- **disciplines**: Дисциплины с кредитами и часами (лекции/практика/лабы)
- **curriculum_plans**: Учебные планы
- **curriculum_disciplines**: Дисциплины в планах по семестрам
- **classrooms**: Аудитории с вместимостью, типом, оборудованием (JSONB)
- **lesson_types**: Лекция, Практика, Лабораторная
- **schedule_lessons**: Расписание с временными слотами и типом недели
- **schedule_changes**: Отмены, переносы, замены преподавателей/аудиторий

**Ключевые возможности**:
- Иерархическая структура организации
- Поддержка обнаружения конфликтов
- Поддержка типов недель (все/чётные/нечётные)
- Отслеживание нагрузки преподавателей
- Учёт оборудования аудиторий в JSONB

### Модуль Tasks (005)

Управление задачами и поручениями:

- **projects**: Группировка проектов
- **tasks**: Основная таблица задач с workflow статусами, приоритетом, прогрессом
- **task_watchers**: Пользователи, следящие за обновлениями задач
- **task_attachments**: Файловые вложения
- **task_comments**: Древовидные комментарии (с parent_comment_id)
- **task_checklists** + **task_checklist_items**: Отслеживание подзадач
- **task_dependencies**: Связи задач (finish_to_start, start_to_start, finish_to_finish, start_to_finish)
- **task_history**: Отслеживание изменений на уровне полей
- **task_reminders**: Запланированные напоминания
- **task_templates**: Переиспользуемые шаблоны задач

**Ключевые возможности**:
- Поддержка диаграмм Ганта через зависимости
- Отслеживание времени (оценочное vs фактическое часы)
- Процент прогресса (0-100)
- Связи с документами
- Массив тегов для категоризации

### Модуль Reports (006)

Система отчётности и аналитики:

- **report_types**: Типы с категориями (академические, административные, финансовые, методические)
- **report_parameters**: Динамические определения параметров для генерации
- **reports**: Сгенерированные отчёты с workflow статусами и хранением файлов
- **report_access**: Детальные права доступа (read, write, approve, publish)
- **report_comments**: Комментарии к отчётам
- **report_generation_log**: Отслеживание генерации с длительностью и ошибками
- **report_subscriptions**: Подписки на автоматическую доставку
- **report_templates**: HTML/LaTeX шаблоны
- **report_history**: Аудит-лог действий с отчётами

**Ключевые возможности**:
- Периодические отчёты (ежедневно, еженедельно, ежемесячно, ежеквартально, ежегодно)
- Множественные форматы вывода (PDF, XLSX, DOCX, HTML)
- Параметры хранятся в JSONB
- Агрегированные данные в JSONB
- Workflow рецензирования
- Управление публичным/приватным доступом

**Seed данные**: 8 предопределённых типов отчётов (успеваемость студентов, посещаемость, нагрузка преподавателей, загрузка аудиторий и др.)

## Тестирование

### Настройка тестовой БД

```bash
just setup-test-db
```

Это:
1. Создаст базу `inf_sys_db_test`
2. Применит все миграции к тестовой базе

### Удаление тестовой БД

```bash
just drop-test-db
```

Полностью удаляет тестовую базу данных.

## Лучшие практики миграций

1. **Всегда создавайте и up, и down миграции**
   - Каждое изменение схемы должно быть обратимым
   - Тестируйте откат перед коммитом

2. **Делайте миграции идемпотентными**
   - Используйте `IF NOT EXISTS` для CREATE
   - Используйте `IF EXISTS` для DROP
   - Безопасно обрабатывайте миграции данных

3. **Не изменяйте существующие миграции**
   - После развёртывания в продакшн миграции неизменяемы
   - Создавайте новые миграции для изменений схемы

4. **Тщательно тестируйте миграции**
   - Применяйте и откатывайте локально
   - Тестируйте на тестовой БД
   - Проверяйте целостность данных

5. **Делайте миграции маленькими и сфокусированными**
   - Одно логическое изменение на миграцию
   - Легче отлаживать и откатывать

6. **Используйте транзакции**
   - Большинство DDL команд в PostgreSQL транзакционны
   - Миграции автоматически откатятся при ошибке

7. **Документируйте сложные миграции**
   - Добавляйте комментарии в SQL файлы
   - Обновляйте этот README с важными изменениями

## Решение проблем

### Грязное состояние миграции

Если миграция упала на середине, БД может быть в "грязном" состоянии:

```bash
# Проверить текущую версию и грязное состояние
just migrate-version

# Форсировать версию до последнего известного хорошего состояния
just migrate-force 2

# Затем исправить миграцию и попробовать снова
just migrate-up
```

### Проблемы с подключением

Если миграции падают с ошибками подключения:

1. Проверить, что PostgreSQL запущен:
   ```bash
   docker ps | grep postgres-dev
   ```

2. Проверить учётные данные в `.env` файле

3. Протестировать подключение вручную:
   ```bash
   docker exec postgres-dev psql -U postgres -d inf_sys_db -c "SELECT version();"
   ```

### Таймаут блокировки

Если миграции зависают, другой процесс может держать блокировку:

```sql
-- Проверить блокировки
SELECT * FROM pg_locks WHERE NOT granted;

-- Убить блокирующий процесс
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE pid = <blocking_pid>;
```

## Продакшн развёртывание

1. **Резервное копирование БД перед миграциями**:
   ```bash
   docker exec postgres-dev pg_dump -U postgres inf_sys_db > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Запустить миграции в окне обслуживания**

3. **Мониторить прогресс миграций**:
   ```bash
   just migrate-version
   ```

4. **Иметь готовый план отката**:
   ```bash
   # Если что-то пошло не так
   just migrate-down-one
   ```

## CI/CD интеграция

Миграции должны запускаться автоматически в CI/CD пайплайне:

```yaml
# Пример GitHub Actions workflow
- name: Run migrations
  run: |
    just migrate-up
  env:
    DB_HOST: ${{ secrets.DB_HOST }}
    DB_PORT: 5432
    DB_NAME: inf_sys_db
    DB_USER: ${{ secrets.DB_USER }}
    DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
```

## Дополнительные ресурсы

- [Документация golang-migrate](https://github.com/golang-migrate/migrate)
- [Лучшие практики PostgreSQL](https://wiki.postgresql.org/wiki/Don%27t_Do_This)
- [Документация проекта](../docs/project-overview.md)
- [Спецификации документов](../docs/)

## Структура файлов миграций

Каждая миграция состоит из двух файлов:

- `NNN_name.up.sql` - применение изменений
- `NNN_name.down.sql` - откат изменений

Пример:
```
migrations/
├── 001_create_users_table.up.sql
├── 001_create_users_table.down.sql
├── 002_create_sessions_table.up.sql
├── 002_create_sessions_table.down.sql
...
```

## Проверка применённых миграций

Посмотреть все таблицы в БД:

```bash
docker exec postgres-dev psql -U postgres -d inf_sys_db -c "\dt"
```

Посмотреть версию миграции:

```bash
docker exec postgres-dev psql -U postgres -d inf_sys_db -c "SELECT version, dirty FROM schema_migrations;"
```

Посмотреть seed данные (например, типы отчётов):

```bash
docker exec postgres-dev psql -U postgres -d inf_sys_db -c "SELECT id, name, code, category FROM report_types ORDER BY id;"
```
