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

### Document Sharing API
Подробная документация по API шаринга документов: [documents.md](documents.md#-document-sharing-api-issue-13)

- `POST /api/documents/:id/share` - Шаринг документа пользователю/роли
- `GET /api/documents/:id/permissions` - Получение прав доступа
- `DELETE /api/documents/:id/permissions/:permissionId` - Удаление прав
- `GET /api/documents/shared` - Документы, к которым у меня есть доступ
- `GET /api/documents/my-shared` - Мои документы, которыми я поделился
- `POST /api/documents/:id/public-links` - Создание публичной ссылки
- `GET /api/public/documents/:token` - Доступ по публичной ссылке

---

## Notifications API

### Base URL: `/api/notifications`

Требует JWT аутентификации. Модуль управления уведомлениями с поддержкой in-app уведомлений, email (Composio Gmail) и Telegram.

### Типы уведомлений

| Тип | Описание |
|-----|----------|
| `info` | Информационное уведомление |
| `success` | Успешное выполнение операции |
| `warning` | Предупреждение |
| `error` | Ошибка |
| `reminder` | Напоминание о событии |
| `task` | Уведомление о задаче |
| `document` | Уведомление о документе |
| `event` | Уведомление о событии |
| `system` | Системное уведомление |

### Приоритеты

| Приоритет | Описание |
|-----------|----------|
| `low` | Низкий приоритет |
| `normal` | Обычный приоритет (по умолчанию) |
| `high` | Высокий приоритет |
| `urgent` | Срочное уведомление |

### In-App Notifications

### GET `/api/notifications`
Получение списка уведомлений текущего пользователя с пагинацией и фильтрацией.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `limit` | int | Размер страницы (default: 20, max: 100) |
| `offset` | int | Смещение для пагинации (default: 0) |
| `type` | string | Фильтр по типу: info, success, warning, error, reminder, task, document, event, system |
| `priority` | string | Фильтр по приоритету: low, normal, high, urgent |
| `is_read` | bool | Фильтр по статусу прочтения |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "notifications": [
      {
        "id": 1,
        "user_id": 5,
        "type": "task",
        "priority": "normal",
        "title": "Новая задача",
        "message": "Вам назначена новая задача",
        "link": "/tasks/123",
        "image_url": null,
        "is_read": false,
        "read_at": null,
        "expires_at": null,
        "metadata": {},
        "created_at": "2025-12-13T10:00:00Z",
        "updated_at": "2025-12-13T10:00:00Z"
      }
    ],
    "total": 50,
    "limit": 20,
    "offset": 0
  }
}
```

### GET `/api/notifications/:id`
Получение уведомления по ID.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 5,
    "type": "task",
    "priority": "normal",
    "title": "Новая задача",
    "message": "Вам назначена новая задача",
    "link": "/tasks/123",
    "is_read": false,
    "created_at": "2025-12-13T10:00:00Z"
  }
}
```

### GET `/api/notifications/unread-count`
Получение количества непрочитанных уведомлений.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "count": 5
  }
}
```

### GET `/api/notifications/stats`
Получение статистики уведомлений пользователя.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "total_count": 150,
    "unread_count": 5,
    "today_count": 12,
    "urgent_count": 2,
    "expired_count": 3
  }
}
```

### PUT `/api/notifications/:id/read`
Отметить уведомление как прочитанное.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "Notification marked as read"
  }
}
```

### PUT `/api/notifications/read-all`
Отметить все уведомления как прочитанные.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "All notifications marked as read"
  }
}
```

### DELETE `/api/notifications/:id`
Удаление уведомления.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "Notification deleted"
  }
}
```

### DELETE `/api/notifications`
Удаление всех уведомлений пользователя.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "All notifications deleted"
  }
}
```

### Admin Endpoints

### POST `/api/admin/notifications`
Создание уведомления (только для админов).

**Request:**
```json
{
  "user_id": 5,
  "type": "task",
  "priority": "normal",
  "title": "Новая задача",
  "message": "Вам назначена новая задача",
  "link": "/tasks/123",
  "expires_at": "2025-12-20T23:59:59Z",
  "metadata": {
    "task_id": 123
  }
}
```

**Validation:**
- `user_id`: required
- `type`: required, одно из: info, success, warning, error, reminder, task, document, event, system
- `priority`: optional (default: normal)
- `title`: required, до 500 символов
- `message`: required
- `link`: optional, до 1000 символов
- `expires_at`: optional, ISO 8601 datetime

### POST `/api/admin/notifications/bulk`
Массовое создание уведомлений для нескольких пользователей.

**Request:**
```json
{
  "user_ids": [1, 2, 3, 4, 5],
  "type": "system",
  "priority": "high",
  "title": "Системное обновление",
  "message": "Запланировано техническое обслуживание"
}
```

---

### Notification Preferences

### GET `/api/notifications/preferences`
Получение настроек уведомлений пользователя.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "user_id": 5,
    "email_enabled": true,
    "push_enabled": true,
    "in_app_enabled": true,
    "telegram_enabled": true,
    "slack_enabled": false,
    "quiet_hours_enabled": true,
    "quiet_hours_start": "22:00",
    "quiet_hours_end": "07:00",
    "timezone": "Europe/Moscow",
    "digest_enabled": false,
    "digest_frequency": "daily",
    "digest_time": "09:00",
    "type_preferences": {
      "task": {
        "enabled": true,
        "channels": ["email", "telegram", "in_app"],
        "priority": "normal"
      }
    }
  }
}
```

### PUT `/api/notifications/preferences`
Обновление настроек уведомлений.

**Request:**
```json
{
  "email_enabled": true,
  "telegram_enabled": true,
  "quiet_hours_enabled": true,
  "quiet_hours_start": "23:00",
  "quiet_hours_end": "08:00",
  "timezone": "Europe/Moscow"
}
```

### PUT `/api/notifications/preferences/channel`
Включение/отключение канала уведомлений.

**Request:**
```json
{
  "channel": "telegram",
  "enabled": true
}
```

**Доступные каналы:** `email`, `push`, `in_app`, `telegram`, `slack`

### PUT `/api/notifications/preferences/quiet-hours`
Обновление тихих часов.

**Request:**
```json
{
  "enabled": true,
  "start": "22:00",
  "end": "07:00",
  "timezone": "Europe/Moscow"
}
```

### POST `/api/notifications/preferences/reset`
Сброс настроек уведомлений к значениям по умолчанию.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "Preferences reset to defaults"
  }
}
```

### GET `/api/notifications/timezones`
Получение списка доступных таймзон.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "timezones": [
      "UTC",
      "Europe/Kaliningrad",
      "Europe/Moscow",
      "Europe/Samara",
      "Asia/Yekaterinburg",
      "Asia/Omsk",
      "Asia/Krasnoyarsk",
      "Asia/Irkutsk",
      "Asia/Yakutsk",
      "Asia/Vladivostok",
      "Asia/Magadan",
      "Asia/Kamchatka"
    ]
  }
}
```

---

### Email Notifications (Composio Gmail)

Доступен только при настроенной интеграции с Composio.

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

### Telegram Integration

Интеграция с Telegram для push-уведомлений. Подробнее: [Telegram Bot Integration](../integrations/telegram-bot.md)

### GET `/api/notifications/telegram/status`
Получить статус подключения Telegram.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "connected": true,
    "username": "john_doe",
    "first_name": "John",
    "connected_at": "2025-12-13T10:00:00Z"
  }
}
```

### POST `/api/notifications/telegram/generate-code`
Сгенерировать код для привязки Telegram аккаунта.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "code": "ABC123",
    "expires_at": "2025-12-13T10:05:00Z",
    "bot_link": "https://t.me/your_bot?start=ABC123"
  }
}
```

### DELETE `/api/notifications/telegram/disconnect`
Отключить Telegram аккаунт.

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "Telegram disconnected"
  }
}
```

### POST `/api/notifications/telegram/webhook`
Webhook для входящих сообщений от Telegram (только для режима webhook).

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

## Messaging API

### Base URL: `/api/v1/messaging`

Модуль внутренних сообщений и чатов. Поддерживает прямые сообщения, групповые чаты, real-time обмен через WebSocket.

**Подробная документация:** [messaging.md](messaging.md)

### Основные endpoints:

#### Conversations (Чаты)
- `POST /api/v1/messaging/conversations/direct` - Создать прямой чат
- `POST /api/v1/messaging/conversations/group` - Создать групповой чат
- `GET /api/v1/messaging/conversations` - Список чатов
- `GET /api/v1/messaging/conversations/:id` - Получить чат
- `PUT /api/v1/messaging/conversations/:id` - Обновить чат
- `POST /api/v1/messaging/conversations/:id/participants` - Добавить участников
- `DELETE /api/v1/messaging/conversations/:id/leave` - Покинуть чат

#### Messages (Сообщения)
- `POST /api/v1/messaging/conversations/:id/messages` - Отправить сообщение
- `GET /api/v1/messaging/conversations/:id/messages` - Получить сообщения
- `PUT /api/v1/messaging/messages/:id` - Редактировать сообщение
- `DELETE /api/v1/messaging/messages/:id` - Удалить сообщение
- `POST /api/v1/messaging/conversations/:id/read` - Отметить как прочитанное
- `GET /api/v1/messaging/messages/search` - Поиск сообщений

#### WebSocket
- `ws://localhost:8080/api/v1/messaging/ws?token=<jwt>` - WebSocket подключение

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

## Integration API (1C)

### Base URL: `/api/integration`

Требует JWT аутентификации и роль `admin`. Модуль интеграции с системой 1С. Доступен только при включенной интеграции (`INTEGRATION_1C_ENABLED=true`).

### POST `/api/integration/sync/employees`
Запустить синхронизацию сотрудников из 1С.

**Request Body:**
```json
{
  "force": false,
  "dry_run": false
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "sync_id": 1,
    "entity_type": "employee",
    "status": "in_progress",
    "started_at": "2025-12-17T10:00:00Z"
  }
}
```

### POST `/api/integration/sync/students`
Запустить синхронизацию студентов из 1С.

**Request Body:**
```json
{
  "force": false,
  "dry_run": false
}
```

### GET `/api/integration/sync/logs`
Получить логи синхронизации с пагинацией.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `page` | int | Номер страницы (default: 1) |
| `limit` | int | Размер страницы (default: 20, max: 100) |
| `entity_type` | string | Фильтр по типу (employee, student) |
| `status` | string | Фильтр по статусу (pending, in_progress, completed, failed) |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": 1,
        "entity_type": "employee",
        "direction": "import",
        "status": "completed",
        "total_records": 150,
        "processed_count": 150,
        "success_count": 148,
        "error_count": 2,
        "conflict_count": 5,
        "started_at": "2025-12-17T10:00:00Z",
        "completed_at": "2025-12-17T10:05:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "limit": 20
  }
}
```

### GET `/api/integration/conflicts`
Получить список конфликтов синхронизации.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `resolution` | string | Фильтр по статусу разрешения (pending, use_local, use_external, merge, skip) |

**Response (200):**
```json
{
  "success": true,
  "data": {
    "conflicts": [
      {
        "id": 1,
        "entity_type": "employee",
        "entity_id": "ABC123",
        "conflict_type": "update",
        "conflict_fields": ["email", "phone"],
        "local_data": {"email": "old@local.com"},
        "external_data": {"email": "new@1c.com"},
        "resolution": "pending",
        "created_at": "2025-12-17T10:00:00Z"
      }
    ]
  }
}
```

### POST `/api/integration/conflicts/:id/resolve`
Разрешить конфликт синхронизации.

**Request Body:**
```json
{
  "resolution": "use_external",
  "notes": "Используем данные из 1С"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "message": "Conflict resolved"
  }
}
```

---

## Custom Reports API

### Base URL: `/api/custom-reports`

Требует JWT аутентификации. Модуль создания пользовательских отчётов с поддержкой экспорта в PDF, Excel и CSV. Реализован в рамках GitHub Issue #21 и #167.

### Источники данных (DataSource)

| Источник | Описание |
|----------|----------|
| `documents` | Документы |
| `users` | Пользователи |
| `events` | События |
| `tasks` | Задачи |
| `students` | Студенты |

### POST `/api/custom-reports`
Создание пользовательского отчёта.

**Request:**
```json
{
  "name": "Отчёт по документам",
  "description": "Список документов за месяц",
  "data_source": "documents",
  "fields": [
    {"field_key": "id", "display_name": "ID", "order": 1},
    {"field_key": "name", "display_name": "Название", "order": 2},
    {"field_key": "created_at", "display_name": "Дата создания", "order": 3}
  ],
  "filters": [
    {"field": "created_at", "operator": "gte", "value": "2025-01-01"}
  ],
  "groupings": [],
  "sortings": [
    {"field": "created_at", "direction": "desc"}
  ],
  "is_public": false
}
```

**Response (201):**
```json
{
  "status": "success",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Отчёт по документам",
    "description": "Список документов за месяц",
    "data_source": "documents",
    "fields": [...],
    "filters": [...],
    "groupings": [],
    "sortings": [...],
    "is_public": false,
    "created_by": 1,
    "created_at": "2025-12-23T10:00:00Z",
    "updated_at": "2025-12-23T10:00:00Z"
  }
}
```

### GET `/api/custom-reports`
Получение списка отчётов с пагинацией.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `page` | int | Номер страницы (default: 1) |
| `page_size` | int | Размер страницы (default: 10, max: 100) |
| `data_source` | string | Фильтр по источнику данных |
| `search` | string | Поиск по названию |
| `is_public` | bool | Фильтр по публичности |

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "reports": [...],
    "total": 50,
    "page": 1,
    "page_size": 10,
    "total_pages": 5
  }
}
```

### GET `/api/custom-reports/:id`
Получение отчёта по ID.

### PUT `/api/custom-reports/:id`
Обновление отчёта (только создатель).

**Request:**
```json
{
  "name": "Обновлённое название",
  "description": "Новое описание",
  "is_public": true
}
```

### DELETE `/api/custom-reports/:id`
Удаление отчёта (только создатель).

### POST `/api/custom-reports/:id/execute`
Выполнение отчёта и получение данных.

**Request:**
```json
{
  "page": 1,
  "page_size": 50
}
```

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "columns": [
      {"key": "id", "label": "ID"},
      {"key": "name", "label": "Название"},
      {"key": "created_at", "label": "Дата создания"}
    ],
    "rows": [
      {"id": 1, "name": "Документ 1", "created_at": "2025-12-20T10:00:00Z"},
      {"id": 2, "name": "Документ 2", "created_at": "2025-12-21T11:00:00Z"}
    ],
    "total_count": 150,
    "page": 1,
    "page_size": 50,
    "total_pages": 3
  }
}
```

### POST `/api/custom-reports/:id/export`
Экспорт отчёта в файл.

**Request:**
```json
{
  "format": "xlsx",
  "include_headers": true,
  "page_size": 1000,
  "orientation": "landscape"
}
```

**Форматы экспорта:**
| Формат | Content-Type | Описание |
|--------|--------------|----------|
| `csv` | text/csv | Простой табличный формат |
| `xlsx` | application/vnd.openxmlformats-officedocument.spreadsheetml.sheet | Microsoft Excel |
| `pdf` | application/pdf | PDF документ |

**Response:** Бинарный файл с заголовком `Content-Disposition: attachment; filename="report.xlsx"`

### GET `/api/custom-reports/my`
Получение отчётов текущего пользователя.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `page` | int | Номер страницы (default: 1) |
| `page_size` | int | Размер страницы (default: 10) |

### GET `/api/custom-reports/public`
Получение публичных отчётов.

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `page` | int | Номер страницы (default: 1) |
| `page_size` | int | Размер страницы (default: 10) |

### GET `/api/custom-reports/available-fields/:dataSource`
Получение доступных полей для источника данных.

**Response (200):**
```json
{
  "status": "success",
  "data": {
    "fields": [
      {"id": "id", "name": "id", "label": "ID", "type": "number"},
      {"id": "name", "name": "name", "label": "Название", "type": "string"},
      {"id": "created_at", "name": "created_at", "label": "Дата создания", "type": "date"},
      {"id": "author_id", "name": "author_id", "label": "ID автора", "type": "number"}
    ]
  }
}
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

**Последнее обновление**: 2025-12-23
**Версия проекта**: 0.3.1
**Статус**: Актуальный
