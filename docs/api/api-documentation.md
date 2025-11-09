# 📖 REST API Документация

## 📋 Обзор API

Микросервисная архитектура с RESTful API для всех компонентов системы. Каждый сервис предоставляет собственный API с единообразной структурой ответов и обработкой ошибок.

## 🌐 Базовая информация

### API Gateway:
- **Base URL**: `https://api.inf-sys.example.com`
- **API Version**: `v1`
- **Protocol**: HTTPS only
- **Content-Type**: `application/json`

### Аутентификация:
```http
Authorization: Bearer <JWT_TOKEN>
X-API-Version: v1
Content-Type: application/json
```

---

## 🔐 Authentication Service API

### Base URL: `/auth`

#### POST `/auth/login`
Аутентификация пользователя

**Request:**
```json
{
  "provider": "google|azure|local",
  "code": "oauth_authorization_code",
  "redirect_uri": "https://app.inf-sys.example.com/callback"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_here",
  "expires_in": 900,
  "token_type": "Bearer",
  "user": {
    "id": "12345",
    "email": "user@example.com",
    "roles": ["методист"],
    "permissions": ["documents:create", "documents:read"]
  }
}
```

#### POST `/auth/refresh`
Обновление токена

#### POST `/auth/logout`
Завершение сессии

#### GET `/auth/me`
Информация о текущем пользователе

---

## 👥 User Service API

### Base URL: `/users`

#### GET `/users`
Получение списка пользователей

**Query Parameters:**
```
?role=методист&department=ИТ&page=1&limit=20&sort=created_at&order=desc
```

**Response:**
```json
{
  "users": [
    {
      "id": "12345",
      "email": "metodist@example.com",
      "first_name": "Иван",
      "last_name": "Петров",
      "roles": ["методист"],
      "department": "ИТ",
      "position": "Старший методист",
      "created_at": "2025-01-01T10:00:00Z",
      "last_login": "2025-01-15T14:30:00Z",
      "is_active": true
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8
  }
}
```

#### GET `/users/{id}`
Получение пользователя по ID

#### POST `/users`
Создание нового пользователя

#### PUT `/users/{id}`
Обновление пользователя

#### DELETE `/users/{id}`
Деактивация пользователя

---

## 📄 Document Service API

### Base URL: `/documents`

#### GET `/documents`
Получение списка документов

**Query Parameters:**
```
?type=curriculum&status=published&author_id=123&created_after=2025-01-01&search=математика
```

**Response:**
```json
{
  "documents": [
    {
      "id": "doc-12345",
      "title": "Учебный план по математике",
      "type": "curriculum",
      "status": "published",
      "author": {
        "id": "user-123",
        "name": "Иван Петров"
      },
      "version": "1.2",
      "created_at": "2025-01-01T10:00:00Z",
      "updated_at": "2025-01-10T15:30:00Z",
      "tags": ["математика", "базовый_курс"],
      "metadata": {
        "department": "Математический факультет",
        "academic_year": "2024-2025",
        "semester": 1
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 89,
    "pages": 5
  }
}
```

#### POST `/documents`
Создание нового документа

**Request:**
```json
{
  "title": "Новый учебный план",
  "type": "curriculum",
  "content": {
    "subjects": [
      {
        "name": "Математический анализ",
        "hours": 120,
        "credits": 4
      }
    ]
  },
  "metadata": {
    "department": "Математический факультет",
    "academic_year": "2024-2025"
  },
  "tags": ["математика", "анализ"]
}
```

#### GET `/documents/{id}`
Получение документа по ID

#### PUT `/documents/{id}`
Обновление документа

#### GET `/documents/{id}/versions`
История версий документа

#### POST `/documents/{id}/versions`
Создание новой версии

---

## 🔄 Workflow Service API

### Base URL: `/workflow`

#### GET `/workflow/processes`
Получение списка процессов

#### POST `/workflow/processes`
Создание нового процесса

#### GET `/workflow/instances`
Активные экземпляры процессов

**Response:**
```json
{
  "instances": [
    {
      "id": "wf-12345",
      "process_id": "curriculum_approval",
      "document_id": "doc-12345",
      "status": "in_progress",
      "current_step": "methodical_review",
      "assignees": ["user-456"],
      "started_at": "2025-01-10T09:00:00Z",
      "deadline": "2025-01-20T17:00:00Z",
      "steps_completed": ["creation", "internal_review"],
      "steps_remaining": ["methodical_review", "final_approval"]
    }
  ]
}
```

#### POST `/workflow/instances/{id}/advance`
Продвижение процесса к следующему этапу

#### POST `/workflow/instances/{id}/reject`
Отклонение и возврат на предыдущий этап

---

## 📅 Schedule Service API

### Base URL: `/schedule`

#### GET `/schedule/events`
Получение событий расписания

**Query Parameters:**
```
?start_date=2025-01-01&end_date=2025-01-31&type=class&group_id=123&teacher_id=456
```

**Response:**
```json
{
  "events": [
    {
      "id": "evt-12345",
      "title": "Математический анализ",
      "type": "class",
      "start_time": "2025-01-15T10:00:00Z",
      "end_time": "2025-01-15T11:30:00Z",
      "location": "Аудитория 101",
      "teacher": {
        "id": "teacher-123",
        "name": "Профессор Иванов"
      },
      "group": {
        "id": "group-456",
        "name": "МТ-21-1"
      },
      "subject": {
        "id": "subj-789",
        "name": "Математический анализ"
      }
    }
  ]
}
```

#### POST `/schedule/events`
Создание нового события

#### PUT `/schedule/events/{id}`
Обновление события

#### DELETE `/schedule/events/{id}`
Удаление события

---

## ✅ Task Service API

### Base URL: `/tasks`

#### GET `/tasks`
Получение списка задач

#### POST `/tasks`
Создание новой задачи

**Request:**
```json
{
  "title": "Подготовить отчет по методической работе",
  "description": "Квартальный отчет с анализом эффективности",
  "type": "report",
  "priority": "high",
  "deadline": "2025-02-01T17:00:00Z",
  "assignees": ["user-123", "user-456"],
  "metadata": {
    "quarter": "Q1",
    "year": 2025,
    "department": "Методический отдел"
  }
}
```

#### PUT `/tasks/{id}/status`
Изменение статуса задачи

---

## 📊 Reporting Service API

### Base URL: `/reports`

#### GET `/reports/templates`
Получение шаблонов отчетов

#### POST `/reports/generate`
Генерация отчета

**Request:**
```json
{
  "template_id": "monthly_methodical_report",
  "parameters": {
    "month": "2025-01",
    "department": "ИТ",
    "include_charts": true
  },
  "format": "pdf"
}
```

#### GET `/reports/{id}/download`
Скачивание сгенерированного отчета

---

## 🔔 Notification Service API

### Base URL: `/notifications`

#### POST `/notifications/send`
Отправка уведомления

#### GET `/notifications/templates`
Получение шаблонов уведомлений

---

## 📁 File Service API

### Base URL: `/files`

#### POST `/files/upload`
Загрузка файла

**Request (multipart/form-data):**
```
file: <binary data>
metadata: {"type": "document", "tags": ["curriculum"]}
```

#### GET `/files/{id}/download`
Скачивание файла

#### GET `/files/{id}/preview`
Предварительный просмотр файла

---

## 🔗 Integration Service API

### Base URL: `/integrations`

#### POST `/integrations/1c/sync`
Синхронизация с 1С

#### GET `/integrations/1c/status`
Статус интеграции с 1С

---

## ❌ Обработка ошибок

### Стандартная структура ошибки:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed for request",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format",
        "code": "INVALID_FORMAT"
      }
    ],
    "request_id": "req-12345-67890",
    "timestamp": "2025-01-15T14:30:00Z"
  }
}
```

### HTTP Status Codes:
| Code | Meaning | Usage |
|------|---------|-------|
| 200 | OK | Успешная операция |
| 201 | Created | Ресурс создан |
| 400 | Bad Request | Ошибка валидации |
| 401 | Unauthorized | Не авторизован |
| 403 | Forbidden | Нет прав доступа |
| 404 | Not Found | Ресурс не найден |
| 409 | Conflict | Конфликт данных |
| 422 | Unprocessable Entity | Семантическая ошибка |
| 429 | Too Many Requests | Rate limit превышен |
| 500 | Internal Server Error | Внутренняя ошибка сервера |

---

## 📄 Pagination и Filtering

### Стандартные параметры пагинации:
```
?page=1&limit=20&sort=created_at&order=desc
```

### Поиск и фильтрация:
```
?search=keyword&filter[status]=published&filter[type]=curriculum&date_from=2025-01-01
```

### Response structure:
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "pages": 8,
    "has_next": true,
    "has_prev": false
  },
  "filters_applied": {
    "status": "published",
    "type": "curriculum"
  }
}
```

---

## 🚦 Rate Limiting

### Лимиты по ролям:
| Роль | Requests/minute | Burst |
|------|----------------|-------|
| Студент | 30 | 5 |
| Преподаватель | 60 | 10 |
| Секретарь | 90 | 15 |
| Методист | 120 | 20 |
| Админ | 300 | 50 |

### Headers:
```http
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1642694400
```

---

## 🔗 Webhook API

### Регистрация webhook:
```json
{
  "url": "https://external-system.com/webhook",
  "events": ["document.created", "workflow.completed"],
  "secret": "webhook_secret_key"
}
```

### Payload формат:
```json
{
  "event": "document.created",
  "timestamp": "2025-01-15T14:30:00Z",
  "data": {
    "document_id": "doc-12345",
    "author_id": "user-123",
    "type": "curriculum"
  },
  "signature": "sha256=abcdef123456..."
}
```
---

**📅 Актуальность документа**  
**Последнее обновление**: 2025-01-15  
**Версия проекта**: 0.1.0  
**Статус**: Актуальный

