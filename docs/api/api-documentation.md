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

**Последнее обновление**: 2025-11-29
**Версия проекта**: 0.1.0
**Статус**: Актуальный
