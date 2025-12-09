# REST API Документация

## Обзор API

Модульная монолитная архитектура с RESTful API. Единый бэкенд на Go предоставляет API для всех модулей системы.

## Базовая информация

- **Base URL**: `http://localhost:8080` (development)
- **API Prefix**: `/api`
- **Protocol**: HTTP/HTTPS
- **Content-Type**: `application/json`

### Аутентификация:
```http
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

---

## Health & Monitoring Endpoints

Доступны без аутентификации.

### GET `/health`
Полная проверка состояния системы (DB + Redis).

**Response:**
```json
{
  "status": "OK",
  "timestamp": "2025-11-29T10:00:00Z",
  "database": {
    "status": "UP",
    "latency_ms": 1.23
  },
  "redis": {
    "status": "UP"
  }
}
```

### GET `/live`
Kubernetes liveness probe.

**Response:**
```json
{
  "status": "UP",
  "timestamp": "2025-11-29T10:00:00Z"
}
```

### GET `/ready`
Kubernetes readiness probe.

**Response:**
```json
{
  "ready": true,
  "timestamp": "2025-11-29T10:00:00Z",
  "checks": {
    "database": {"status": "UP"},
    "redis": {"status": "UP"}
  }
}
```

### GET `/metrics`
Prometheus метрики в формате OpenMetrics.

**Response:** Text/plain с Prometheus метриками.

---

## Authentication API

### Base URL: `/api/auth`

Публичные endpoints с rate limiting (10 req/min + burst 5).

### POST `/api/auth/register`
Регистрация нового пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "name": "Иван Петров"
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "email": "user@example.com",
    "name": "Иван Петров",
    "role": "user",
    "createdAt": "2025-11-29T10:00:00Z"
  }
}
```

**Validation:**
- `email`: required, valid email format
- `password`: required, минимум 8 символов
- `name`: required, 2-100 символов

### POST `/api/auth/login`
Аутентификация пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900
  }
}
```

### POST `/api/auth/refresh`
Обновление access token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900
  }
}
```

---

## User API

### Base URL: `/api`

Требует JWT аутентификации.

### GET `/api/me`
Получение информации о текущем пользователе.

**Response (200):**
```json
{
  "id": 1,
  "email": "user@example.com",
  "name": "Иван Петров",
  "role": "user",
  "createdAt": "2025-11-29T10:00:00Z",
  "updatedAt": "2025-11-29T10:00:00Z"
}
```

---

## Documents API

### Base URL: `/api/documents`

Требует JWT аутентификации. Доступен только при настроенном S3 хранилище.

### POST `/api/documents`
Создание нового документа.

**Request:**
```json
{
  "name": "Учебный план по математике",
  "description": "Учебный план на 2024-2025 год",
  "type_id": 1,
  "category_id": 2
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "Учебный план по математике",
    "description": "Учебный план на 2024-2025 год",
    "type_id": 1,
    "category_id": 2,
    "author_id": 1,
    "created_at": "2025-11-29T10:00:00Z"
  }
}
```

### GET `/api/documents`
Получение списка документов с фильтрацией.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `type_id` | int | Фильтр по типу документа |
| `category_id` | int | Фильтр по категории |
| `author_id` | int | Фильтр по автору |
| `search` | string | Поиск по названию |
| `page` | int | Номер страницы (default: 1) |
| `page_size` | int | Размер страницы (default: 20) |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "items": [
      {
        "id": 1,
        "name": "Учебный план по математике",
        "description": "...",
        "type_id": 1,
        "category_id": 2,
        "author_id": 1,
        "file_path": "documents/123/file.pdf",
        "created_at": "2025-11-29T10:00:00Z",
        "updated_at": "2025-11-29T10:00:00Z"
      }
    ],
    "total": 150,
    "page": 1,
    "page_size": 20
  }
}
```

### GET `/api/documents/:id`
Получение документа по ID.

### PUT `/api/documents/:id`
Обновление документа.

### DELETE `/api/documents/:id`
Удаление документа.

### POST `/api/documents/:id/file`
Загрузка файла к документу.

**Request:** `multipart/form-data`
- `file`: binary

### GET `/api/documents/:id/file`
Скачивание файла документа.

### DELETE `/api/documents/:id/file`
Удаление файла документа.

### GET `/api/document-types`
Получение списка типов документов.

### GET `/api/document-categories`
Получение списка категорий документов.

---

## Notifications API

### Base URL: `/api/notifications`

Требует JWT аутентификации. Доступен только при настроенной интеграции с Composio.

### POST `/api/notifications/send-email`
Отправка email уведомления.

**Request:**
```json
{
  "to": ["recipient@example.com"],
  "cc": ["cc@example.com"],
  "bcc": ["bcc@example.com"],
  "subject": "Важное уведомление",
  "body": "<h1>Привет!</h1><p>Текст письма</p>",
  "is_html": true
}
```

**Validation:**
- `to`: required, 1-50 адресов
- `cc`, `bcc`: optional, до 20 адресов
- `subject`: required, 1-200 символов
- `body`: required, минимум 1 символ

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "Email sent successfully"
  }
}
```

### POST `/api/notifications/send-welcome`
Отправка приветственного email.

**Request:**
```json
{
  "email": "newuser@example.com",
  "name": "Иван Петров"
}
```

---

## Admin API

### Base URL: `/api/admin`

Требует JWT аутентификации и роль `admin`.

### GET `/api/admin/users`
Получение списка пользователей (placeholder).

---

## Files API

### Base URL: `/api/files`

Требует JWT аутентификации. Модуль управления файлами с поддержкой версионирования. Доступен только при настроенном MinIO/S3 хранилище.

### POST `/api/files/upload`
Загрузка нового файла.

**Request:** `multipart/form-data`
- `file`: binary (обязательно)

**Ограничения:**
- Максимальный размер: 100 MB
- Разрешённые типы: PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, TXT, CSV, JPG, JPEG, PNG, GIF, WEBP, ZIP, RAR

**Response (201):**
```json
{
  "success": true,
  "data": {
    "file_id": 1,
    "original_name": "document.pdf",
    "size": 1048576,
    "mime_type": "application/pdf",
    "checksum": "sha256:abc123..."
  }
}
```

### GET `/api/files`
Получение списка файлов с пагинацией.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `page` | int | Номер страницы (default: 1) |
| `limit` | int | Размер страницы (default: 20, max: 100) |
| `uploaded_by` | int | Фильтр по автору загрузки |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "files": [
      {
        "id": 1,
        "original_name": "document.pdf",
        "size": 1048576,
        "mime_type": "application/pdf",
        "checksum": "sha256:abc123...",
        "uploaded_by": 1,
        "document_id": 5,
        "is_temporary": false,
        "created_at": "2025-12-09T10:00:00Z",
        "updated_at": "2025-12-09T10:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "limit": 20,
    "total_pages": 3
  }
}
```

### GET `/api/files/:id`
Получение информации о файле по ID.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "original_name": "document.pdf",
    "size": 1048576,
    "mime_type": "application/pdf",
    "checksum": "sha256:abc123...",
    "uploaded_by": 1,
    "document_id": 5,
    "is_temporary": false,
    "created_at": "2025-12-09T10:00:00Z",
    "updated_at": "2025-12-09T10:00:00Z"
  }
}
```

### GET `/api/files/:id/download`
Получение presigned URL для скачивания файла.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "presigned_url": "https://minio.example.com/bucket/file?...",
    "file_name": "document.pdf",
    "mime_type": "application/pdf",
    "size": 1048576
  }
}
```

### POST `/api/files/:id/attach`
Прикрепление файла к документу, задаче или объявлению.

**Request:**
```json
{
  "document_id": 5
}
```
или
```json
{
  "task_id": 10
}
```
или
```json
{
  "announcement_id": 3
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "original_name": "document.pdf",
    "document_id": 5,
    "is_temporary": false
  }
}
```

### DELETE `/api/files/:id`
Удаление файла (только автор загрузки).

**Response (200):**
```json
{
  "success": true,
  "message": "Файл успешно удалён"
}
```

---

### Версии файлов

### POST `/api/files/:id/versions`
Создание новой версии файла.

**Request:** `multipart/form-data`
- `file`: binary (обязательно)
- `comment`: string (опционально, max 500 символов)

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": 3,
    "version_number": 2,
    "size": 1048576,
    "checksum": "sha256:def456...",
    "comment": "Исправлены опечатки",
    "created_by": 1,
    "created_at": "2025-12-09T11:00:00Z"
  }
}
```

### GET `/api/files/:id/versions`
Получение всех версий файла.

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "version_number": 1,
      "size": 1000000,
      "checksum": "sha256:abc123...",
      "comment": "Первая версия",
      "created_by": 1,
      "created_at": "2025-12-09T10:00:00Z"
    },
    {
      "id": 3,
      "version_number": 2,
      "size": 1048576,
      "checksum": "sha256:def456...",
      "comment": "Исправлены опечатки",
      "created_by": 1,
      "created_at": "2025-12-09T11:00:00Z"
    }
  ]
}
```

### GET `/api/files/:id/versions/:version`
Скачивание конкретной версии файла.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "presigned_url": "https://minio.example.com/bucket/file/v2?...",
    "file_name": "document_v2.pdf",
    "mime_type": "application/pdf",
    "size": 1048576
  }
}
```

---

### Файлы по сущностям

### GET `/api/files/by-document/:document_id`
Получение всех файлов, прикреплённых к документу.

### GET `/api/files/by-task/:task_id`
Получение всех файлов, прикреплённых к задаче.

### GET `/api/files/by-announcement/:announcement_id`
Получение всех файлов, прикреплённых к объявлению.

---

### Администрирование файлов

### POST `/api/files/cleanup`
Очистка устаревших временных файлов (только admin).

**Response (200):**
```json
{
  "success": true,
  "data": {
    "deleted_count": 15,
    "message": "Временные файлы очищены"
  }
}
```

---

## Schedule API (Планируется)

Модуль расписания событий готов, но ещё не подключен к API.

Планируемые endpoints:
- `POST /api/events` - Создание события
- `GET /api/events` - Список событий
- `GET /api/events/:id` - Получение события
- `PUT /api/events/:id` - Обновление события
- `DELETE /api/events/:id` - Удаление события
- `POST /api/events/:id/cancel` - Отмена события
- `POST /api/events/:id/reschedule` - Перенос события
- `POST /api/events/:id/participants` - Добавление участников
- `DELETE /api/events/:id/participants/:user_id` - Удаление участника
- `POST /api/events/:id/respond` - Ответ на приглашение
- `GET /api/events/upcoming` - Предстоящие события
- `GET /api/events/invitations` - Приглашения
- `GET /api/events/range` - События по диапазону дат

---

## Обработка ошибок

### Структура ответа об ошибке:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Ошибка валидации",
    "details": [
      {
        "field": "email",
        "message": "Неверный формат email"
      }
    ]
  }
}
```

### HTTP Status Codes:
| Code | Значение | Использование |
|------|----------|---------------|
| 200 | OK | Успешная операция |
| 201 | Created | Ресурс создан |
| 400 | Bad Request | Ошибка валидации |
| 401 | Unauthorized | Не авторизован |
| 403 | Forbidden | Нет прав доступа |
| 404 | Not Found | Ресурс не найден |
| 429 | Too Many Requests | Rate limit превышен |
| 500 | Internal Server Error | Внутренняя ошибка |

---

## Rate Limiting

### Публичные endpoints (`/api/auth/*`):
- 10 запросов/минуту
- Burst: 5 запросов

### Защищённые endpoints (`/api/*`):
- 60 запросов/минуту
- Burst: 10 запросов

### Headers ответа:
```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1732878000
```

---

## CORS

Настроен через переменные окружения:
- `CORS_ALLOWED_ORIGINS`: Разрешённые origins (default: `http://localhost:3000`)
- `CORS_ALLOWED_METHODS`: Разрешённые методы (default: `GET,POST,PUT,DELETE,OPTIONS`)
- `CORS_ALLOWED_HEADERS`: Разрешённые заголовки (default: `Content-Type,Authorization`)

---

## Примеры использования

### Регистрация и вход:
```bash
# Регистрация
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"Иван"}'

# Вход
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Получение профиля
curl http://localhost:8080/api/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Работа с документами:
```bash
# Список документов
curl http://localhost:8080/api/documents?page=1&page_size=10 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Создание документа
curl -X POST http://localhost:8080/api/documents \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Новый документ","type_id":1,"category_id":1}'

# Загрузка файла
curl -X POST http://localhost:8080/api/documents/1/file \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -F "file=@document.pdf"
```

---

**Последнее обновление**: 2025-12-09
**Версия проекта**: 0.2.0
**Статус**: Актуальный
