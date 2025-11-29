# 📅 Calendar & Schedule API

## 📋 Обзор

API календаря и планирования обеспечивает управление событиями, встречами, дедлайнами с поддержкой повторяющихся событий (recurring events), участников и напоминаний.

## 🌐 Base URL
```
https://api.inf-sys.example.com/api/v1/events
```

## 📚 Типы событий

| Тип | Описание |
|-----|----------|
| `meeting` | Встреча |
| `deadline` | Дедлайн |
| `task` | Задача |
| `reminder` | Напоминание |
| `holiday` | Праздник/выходной |
| `personal` | Личное событие |

## 📊 Статусы событий

| Статус | Описание |
|--------|----------|
| `scheduled` | Запланировано |
| `ongoing` | В процессе |
| `completed` | Завершено |
| `cancelled` | Отменено |
| `postponed` | Отложено |

## 🔄 Правила повторения (Recurrence)

Формат совместим с RFC 5545 RRULE:

| Параметр | Описание | Пример |
|----------|----------|--------|
| `frequency` | Частота: daily, weekly, monthly, yearly | `"weekly"` |
| `interval` | Интервал повторения | `2` (каждые 2 недели) |
| `count` | Количество повторений | `10` |
| `until` | Дата окончания повторений | `"2025-12-31T23:59:59Z"` |
| `by_weekday` | Дни недели: MO, TU, WE, TH, FR, SA, SU | `["MO", "WE", "FR"]` |
| `by_monthday` | Дни месяца (1-31) | `[1, 15]` |
| `by_month` | Месяцы (1-12) | `[1, 6, 12]` |
| `week_start` | Начало недели | `"MO"` |

---

## 🚀 Endpoints

### POST `/events`

Создание нового события

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Совещание по проекту",
  "description": "Обсуждение статуса разработки",
  "event_type": "meeting",
  "start_time": "2025-01-20T10:00:00Z",
  "end_time": "2025-01-20T11:30:00Z",
  "all_day": false,
  "timezone": "Europe/Moscow",
  "location": "Переговорная #3",
  "participant_ids": [2, 3, 5],
  "color": "#4CAF50",
  "priority": 4,
  "is_recurring": true,
  "recurrence_rule": {
    "frequency": "weekly",
    "interval": 1,
    "by_weekday": ["MO", "WE"],
    "count": 10
  },
  "reminders": [
    {"reminder_type": "in_app", "minutes_before": 15},
    {"reminder_type": "email", "minutes_before": 60}
  ]
}
```

**Response (201):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Совещание по проекту",
    "description": "Обсуждение статуса разработки",
    "event_type": "meeting",
    "status": "scheduled",
    "start_time": "2025-01-20T10:00:00Z",
    "end_time": "2025-01-20T11:30:00Z",
    "all_day": false,
    "timezone": "Europe/Moscow",
    "location": "Переговорная #3",
    "organizer_id": 1,
    "organizer_name": "Иван Петров",
    "participants": [
      {"user_id": 2, "user_name": "Анна Смирнова", "response_status": "pending", "role": "required"},
      {"user_id": 3, "user_name": "Дмитрий Козлов", "response_status": "pending", "role": "required"},
      {"user_id": 5, "user_name": "Елена Иванова", "response_status": "pending", "role": "required"}
    ],
    "is_recurring": true,
    "recurrence_rule": {
      "frequency": "weekly",
      "interval": 1,
      "by_weekday": ["MO", "WE"],
      "count": 10,
      "week_start": "MO"
    },
    "color": "#4CAF50",
    "priority": 4,
    "reminders": [
      {"id": 1, "reminder_type": "in_app", "minutes_before": 15, "is_sent": false},
      {"id": 2, "reminder_type": "email", "minutes_before": 60, "is_sent": false}
    ],
    "created_at": "2025-01-15T08:00:00Z",
    "updated_at": "2025-01-15T08:00:00Z"
  },
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### GET `/events`

Получение списка событий с фильтрацией

**Query Parameters:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `organizer_id` | int | ID организатора |
| `participant_id` | int | ID участника |
| `event_type` | string | Тип события |
| `status` | string | Статус события |
| `start_from` | string | Начало периода (RFC3339) |
| `start_to` | string | Конец периода (RFC3339) |
| `search` | string | Поиск по названию и описанию |
| `is_recurring` | bool | Только повторяющиеся события |
| `page` | int | Номер страницы (default: 1) |
| `page_size` | int | Размер страницы (default: 20, max: 100) |
| `order_by` | string | Сортировка (default: "start_time ASC") |

**Example:**
```
GET /events?event_type=meeting&start_from=2025-01-01T00:00:00Z&start_to=2025-01-31T23:59:59Z&page=1&page_size=20
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "events": [...],
    "total": 45,
    "page": 1,
    "page_size": 20,
    "total_pages": 3
  },
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### GET `/events/{id}`

Получение события по ID

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Совещание по проекту",
    ...
  },
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

**Response (404):**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Событие не найдено"
  },
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### PUT `/events/{id}`

Обновление события (только для организатора)

**Request Body:**
```json
{
  "title": "Обновленное название",
  "description": "Новое описание",
  "start_time": "2025-01-20T11:00:00Z",
  "end_time": "2025-01-20T12:30:00Z",
  "location": "Переговорная #5",
  "status": "scheduled",
  "priority": 5
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Обновленное название",
    ...
  },
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### DELETE `/events/{id}`

Удаление события (soft delete, только для организатора)

**Response (200):**
```json
{
  "success": true,
  "data": null,
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### GET `/events/range`

Получение событий по диапазону дат

**Query Parameters:**

| Параметр | Тип | Обязательный | Описание |
|----------|-----|--------------|----------|
| `start` | string | да | Начало периода (RFC3339) |
| `end` | string | да | Конец периода (RFC3339) |

**Example:**
```
GET /events/range?start=2025-01-01T00:00:00Z&end=2025-01-31T23:59:59Z
```

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "Совещание",
      "start_time": "2025-01-15T10:00:00Z",
      ...
    }
  ],
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### GET `/events/upcoming`

Получение предстоящих событий для текущего пользователя

**Query Parameters:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `limit` | int | Количество событий (default: 10, max: 50) |

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "Совещание через час",
      "start_time": "2025-01-15T11:00:00Z",
      ...
    }
  ],
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### GET `/events/invitations`

Получение ожидающих приглашений для текущего пользователя

**Response (200):**
```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "title": "Приглашение на встречу",
      "organizer_name": "Анна Смирнова",
      "start_time": "2025-01-20T14:00:00Z",
      ...
    }
  ],
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### POST `/events/{id}/cancel`

Отмена события (только для организатора)

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "status": "cancelled",
    ...
  },
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### POST `/events/{id}/reschedule`

Перенос события на другое время (только для организатора)

**Request Body:**
```json
{
  "start_time": "2025-01-25T10:00:00Z",
  "end_time": "2025-01-25T11:30:00Z"
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "start_time": "2025-01-25T10:00:00Z",
    "end_time": "2025-01-25T11:30:00Z",
    "status": "scheduled",
    ...
  },
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

## 👥 Управление участниками

### POST `/events/{id}/participants`

Добавление участников в событие (только для организатора)

**Request Body:**
```json
{
  "user_ids": [4, 6, 7],
  "role": "optional"
}
```

**Роли участников:**

| Роль | Описание |
|------|----------|
| `required` | Обязательный участник |
| `optional` | Необязательный участник |
| `resource` | Ресурс (переговорная, оборудование) |

**Response (200):**
```json
{
  "success": true,
  "data": null,
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### DELETE `/events/{id}/participants/{user_id}`

Удаление участника из события

Организатор может удалить любого участника. Участник может удалить только себя.

**Response (200):**
```json
{
  "success": true,
  "data": null,
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

### POST `/events/{id}/respond`

Ответ на приглашение (принять/отклонить)

**Request Body:**
```json
{
  "status": "accepted"
}
```

**Статусы ответа:**

| Статус | Описание |
|--------|----------|
| `accepted` | Принял приглашение |
| `declined` | Отклонил приглашение |
| `tentative` | Возможно приму участие |

**Response (200):**
```json
{
  "success": true,
  "data": null,
  "meta": {
    "timestamp": "2025-01-15T08:00:00Z"
  }
}
```

---

## 🔔 Напоминания

### Типы напоминаний

| Тип | Описание |
|-----|----------|
| `email` | Email уведомление |
| `push` | Push-уведомление |
| `in_app` | Уведомление в приложении |
| `telegram` | Telegram уведомление |

Напоминания создаются автоматически при создании события (по умолчанию: 15 мин, 1 час, 1 день до события) или указываются явно в запросе.

---

## ⚠️ Коды ошибок

| Код | HTTP Status | Описание |
|-----|-------------|----------|
| `BAD_REQUEST` | 400 | Неверный формат запроса |
| `UNAUTHORIZED` | 401 | Требуется авторизация |
| `FORBIDDEN` | 403 | Недостаточно прав (не организатор) |
| `NOT_FOUND` | 404 | Событие не найдено |
| `INTERNAL_ERROR` | 500 | Внутренняя ошибка сервера |

---

## 📝 Примеры использования

### Создание еженедельной встречи

```bash
curl -X POST https://api.inf-sys.example.com/api/v1/events \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Еженедельный стендап",
    "event_type": "meeting",
    "start_time": "2025-01-20T09:00:00Z",
    "end_time": "2025-01-20T09:30:00Z",
    "location": "Zoom",
    "is_recurring": true,
    "recurrence_rule": {
      "frequency": "weekly",
      "interval": 1,
      "by_weekday": ["MO"],
      "until": "2025-06-30T23:59:59Z"
    },
    "participant_ids": [2, 3, 4, 5]
  }'
```

### Получение событий на неделю

```bash
curl -X GET "https://api.inf-sys.example.com/api/v1/events/range?start=2025-01-20T00:00:00Z&end=2025-01-26T23:59:59Z" \
  -H "Authorization: Bearer <token>"
```

### Принятие приглашения

```bash
curl -X POST https://api.inf-sys.example.com/api/v1/events/5/respond \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"status": "accepted"}'
```
