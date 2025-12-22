# 💬 Messaging API

API для модуля внутренних сообщений и чатов.

## 📋 Обзор

Модуль messaging предоставляет функциональность для:
- Прямых сообщений (1-на-1)
- Групповых чатов
- Real-time обмена сообщениями через WebSocket
- Прикрепления файлов
- Поиска по сообщениям

## 🔑 Аутентификация

Все эндпоинты требуют JWT токен в заголовке:
```
Authorization: Bearer <token>
```

## 📡 Эндпоинты

### Conversations (Чаты)

#### Создать прямой чат

```http
POST /api/v1/messaging/conversations/direct
```

**Request Body:**
```json
{
  "participant_id": 123
}
```

**Response (201 Created):**
```json
{
  "id": 1,
  "type": "direct",
  "participants": [
    {
      "user_id": 1,
      "name": "John Doe",
      "avatar_url": "/avatars/1.jpg",
      "role": "owner",
      "joined_at": "2025-01-15T10:00:00Z"
    },
    {
      "user_id": 123,
      "name": "Jane Smith",
      "avatar_url": "/avatars/123.jpg",
      "role": "member",
      "joined_at": "2025-01-15T10:00:00Z"
    }
  ],
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

#### Создать групповой чат

```http
POST /api/v1/messaging/conversations/group
```

**Request Body:**
```json
{
  "name": "Project Team",
  "participant_ids": [123, 456, 789]
}
```

**Response (201 Created):**
```json
{
  "id": 2,
  "type": "group",
  "name": "Project Team",
  "participants": [...],
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

#### Получить список чатов

```http
GET /api/v1/messaging/conversations
```

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `type` | string | Фильтр по типу: `direct`, `group` |
| `limit` | int | Количество записей (default: 20) |
| `offset` | int | Смещение для пагинации |

**Response (200 OK):**
```json
{
  "conversations": [
    {
      "id": 1,
      "type": "direct",
      "name": null,
      "participants": [...],
      "last_message": {
        "id": 100,
        "content": "Hello!",
        "sender_id": 123,
        "created_at": "2025-01-15T12:00:00Z"
      },
      "unread_count": 3,
      "created_at": "2025-01-15T10:00:00Z",
      "updated_at": "2025-01-15T12:00:00Z"
    }
  ],
  "total": 15,
  "limit": 20,
  "offset": 0
}
```

#### Получить чат по ID

```http
GET /api/v1/messaging/conversations/:id
```

**Response (200 OK):**
```json
{
  "id": 1,
  "type": "direct",
  "participants": [...],
  "last_message": {...},
  "unread_count": 0,
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T12:00:00Z"
}
```

#### Обновить групповой чат

```http
PUT /api/v1/messaging/conversations/:id
```

**Request Body:**
```json
{
  "name": "Updated Team Name"
}
```

#### Добавить участников в групповой чат

```http
POST /api/v1/messaging/conversations/:id/participants
```

**Request Body:**
```json
{
  "user_ids": [999, 888]
}
```

#### Покинуть чат

```http
DELETE /api/v1/messaging/conversations/:id/leave
```

**Response (204 No Content)**

---

### Messages (Сообщения)

#### Отправить сообщение

```http
POST /api/v1/messaging/conversations/:id/messages
```

**Request Body:**
```json
{
  "content": "Hello, team!",
  "reply_to_id": null,
  "attachments": [
    {
      "file_id": "abc123",
      "file_name": "document.pdf",
      "file_type": "application/pdf",
      "file_size": 1024000,
      "url": "/files/abc123"
    }
  ]
}
```

**Response (201 Created):**
```json
{
  "id": 101,
  "conversation_id": 1,
  "sender_id": 1,
  "sender": {
    "id": 1,
    "name": "John Doe",
    "avatar_url": "/avatars/1.jpg"
  },
  "content": "Hello, team!",
  "reply_to_id": null,
  "reply_to": null,
  "attachments": [...],
  "is_edited": false,
  "created_at": "2025-01-15T12:30:00Z",
  "updated_at": "2025-01-15T12:30:00Z"
}
```

#### Получить сообщения чата

```http
GET /api/v1/messaging/conversations/:id/messages
```

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `limit` | int | Количество записей (default: 50) |
| `before` | int | ID сообщения для пагинации назад |
| `after` | int | ID сообщения для пагинации вперед |

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": 100,
      "conversation_id": 1,
      "sender_id": 123,
      "sender": {...},
      "content": "Hello!",
      "attachments": [],
      "is_edited": false,
      "created_at": "2025-01-15T12:00:00Z",
      "updated_at": "2025-01-15T12:00:00Z"
    }
  ],
  "has_more": true
}
```

#### Редактировать сообщение

```http
PUT /api/v1/messaging/messages/:id
```

**Request Body:**
```json
{
  "content": "Updated message content"
}
```

**Response (200 OK):**
```json
{
  "id": 100,
  "content": "Updated message content",
  "is_edited": true,
  "updated_at": "2025-01-15T12:35:00Z"
}
```

#### Удалить сообщение

```http
DELETE /api/v1/messaging/messages/:id
```

**Response (204 No Content)**

#### Отметить сообщения как прочитанные

```http
POST /api/v1/messaging/conversations/:id/read
```

**Request Body:**
```json
{
  "message_id": 100
}
```

**Response (200 OK):**
```json
{
  "read_count": 5
}
```

#### Поиск сообщений

```http
GET /api/v1/messaging/messages/search
```

**Query Parameters:**
| Параметр | Тип | Описание |
|----------|-----|----------|
| `q` | string | Поисковый запрос (обязательный) |
| `conversation_id` | int | Ограничить поиск конкретным чатом |
| `limit` | int | Количество результатов (default: 20) |

**Response (200 OK):**
```json
{
  "messages": [
    {
      "id": 50,
      "conversation_id": 1,
      "content": "...matched text...",
      "sender": {...},
      "created_at": "2025-01-10T09:00:00Z"
    }
  ],
  "total": 5
}
```

---

## 🔌 WebSocket API

### Подключение

```
ws://localhost:8080/api/v1/messaging/ws?token=<jwt_token>
```

### События от сервера

#### Новое сообщение
```json
{
  "type": "new_message",
  "payload": {
    "conversation_id": 1,
    "message": {
      "id": 102,
      "sender_id": 123,
      "content": "New message!",
      "created_at": "2025-01-15T13:00:00Z"
    }
  }
}
```

#### Сообщение отредактировано
```json
{
  "type": "message_edited",
  "payload": {
    "conversation_id": 1,
    "message_id": 100,
    "content": "Edited content",
    "updated_at": "2025-01-15T13:05:00Z"
  }
}
```

#### Сообщение удалено
```json
{
  "type": "message_deleted",
  "payload": {
    "conversation_id": 1,
    "message_id": 100
  }
}
```

#### Пользователь печатает
```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 1,
    "user_id": 123,
    "user_name": "Jane Smith"
  }
}
```

#### Сообщения прочитаны
```json
{
  "type": "messages_read",
  "payload": {
    "conversation_id": 1,
    "user_id": 123,
    "last_read_message_id": 100
  }
}
```

### События от клиента

#### Индикатор набора
```json
{
  "type": "typing",
  "payload": {
    "conversation_id": 1
  }
}
```

#### Подписка на чат
```json
{
  "type": "subscribe",
  "payload": {
    "conversation_id": 1
  }
}
```

#### Отписка от чата
```json
{
  "type": "unsubscribe",
  "payload": {
    "conversation_id": 1
  }
}
```

---

## 🔒 Коды ошибок

| Код | Описание |
|-----|----------|
| 400 | Bad Request - Неверные параметры запроса |
| 401 | Unauthorized - Отсутствует или недействительный токен |
| 403 | Forbidden - Нет доступа к ресурсу |
| 404 | Not Found - Чат или сообщение не найдено |
| 409 | Conflict - Прямой чат уже существует |
| 422 | Unprocessable Entity - Ошибка валидации |
| 500 | Internal Server Error - Внутренняя ошибка сервера |

**Пример ответа с ошибкой:**
```json
{
  "error": {
    "code": "CONVERSATION_NOT_FOUND",
    "message": "Conversation with ID 999 not found"
  }
}
```

---

## 📝 Примеры использования

### cURL: Создание прямого чата

```bash
curl -X POST http://localhost:8080/api/v1/messaging/conversations/direct \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"participant_id": 123}'
```

### cURL: Отправка сообщения

```bash
curl -X POST http://localhost:8080/api/v1/messaging/conversations/1/messages \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hello, World!"}'
```

### JavaScript: WebSocket подключение

```javascript
const ws = new WebSocket(`ws://localhost:8080/api/v1/messaging/ws?token=${token}`);

ws.onopen = () => {
  console.log('Connected to messaging');

  // Подписка на чат
  ws.send(JSON.stringify({
    type: 'subscribe',
    payload: { conversation_id: 1 }
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  switch (data.type) {
    case 'new_message':
      console.log('New message:', data.payload.message);
      break;
    case 'typing':
      console.log(`${data.payload.user_name} is typing...`);
      break;
  }
};

// Отправка индикатора набора
function sendTyping(conversationId) {
  ws.send(JSON.stringify({
    type: 'typing',
    payload: { conversation_id: conversationId }
  }));
}
```

---

**📅 Актуальность документа**
**Последнее обновление**: 2025-12-22
**Версия API**: v1
**Статус**: Актуальный
