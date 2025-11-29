# Authentication API

## Обзор аутентификации

Сервис аутентификации обеспечивает безопасную регистрацию и вход через локальную аутентификацию email/password с использованием JWT токенов.

## Base URL
```
http://localhost:8080/api/auth
```

## Endpoints

### POST `/api/auth/register`
Регистрация нового пользователя с автоматическим входом.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123",
  "name": "Иван Петров",
  "role": "user"
}
```

**Validation:**
- `email`: required, valid email format, уникальный
- `password`: required, минимум 8 символов
- `name`: required, 2-100 символов
- `role`: optional (default: "user")

**Response (201):**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "Иван Петров",
      "role": "user",
      "createdAt": "2025-11-29T10:00:00Z",
      "updatedAt": "2025-11-29T10:00:00Z"
    }
  }
}
```

**Errors:**
- `400` - Неверный формат запроса или ошибка валидации
- `409` - Email уже зарегистрирован

**Примечание:** После успешной регистрации автоматически отправляется приветственное email (если настроен Composio).

---

### POST `/api/auth/login`
Аутентификация пользователя.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

**Validation:**
- `email`: required, valid email format
- `password`: required

**Response (200):**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "email": "user@example.com",
      "name": "Иван Петров",
      "role": "user",
      "createdAt": "2025-11-29T10:00:00Z",
      "updatedAt": "2025-11-29T10:00:00Z"
    }
  }
}
```

**Errors:**
- `400` - Неверный формат запроса
- `401` - Неверные учетные данные

---

### POST `/api/auth/refresh`
Обновление access token с использованием refresh token.

**Request:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200):**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Errors:**
- `400` - Неверный формат запроса
- `401` - Недействительный refresh token

---

### GET `/api/me`
Получение информации о текущем пользователе.

**Headers:**
```http
Authorization: Bearer <JWT_TOKEN>
```

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

**Errors:**
- `401` - Не авторизован или токен истёк

---

## JWT Token

### Структура токена

**Header:**
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
```

**Payload:**
```json
{
  "user_id": 1,
  "role": "user",
  "exp": 1732878000,
  "iat": 1732877100
}
```

### Время жизни токенов

| Тип токена | TTL | Настройка |
|------------|-----|-----------|
| Access Token | 15 минут | `JWT_ACCESS_TTL` |
| Refresh Token | 7 дней | `JWT_REFRESH_TTL` |

---

## Роли пользователей

| Роль | Описание |
|------|----------|
| `user` | Обычный пользователь |
| `admin` | Администратор с полным доступом |

### Проверка роли в коде:
```go
adminGroup := router.Group("/api/admin")
adminGroup.Use(authMiddleware.RequireRole("admin"))
```

---

## Rate Limiting

Публичные auth endpoints защищены rate limiting:
- **Лимит**: 10 запросов в минуту
- **Burst**: 5 запросов

При превышении лимита возвращается `429 Too Many Requests`.

---

## Безопасность

### Хэширование паролей
Пароли хэшируются с использованием bcrypt с cost factor 10.

### Секреты
Необходимо настроить следующие переменные окружения:
```bash
JWT_ACCESS_SECRET=your_secure_access_secret_key
JWT_REFRESH_SECRET=your_secure_refresh_secret_key
```

### Рекомендации
1. Используйте HTTPS в production
2. Установите сильные JWT секреты (минимум 32 символа)
3. Храните токены безопасно на клиенте (httpOnly cookies или secure storage)
4. Реализуйте logout через удаление токенов на клиенте

---

## Frontend Integration

### React пример
```typescript
interface AuthResponse {
  token: string;
  refreshToken: string;
  user: {
    id: number;
    email: string;
    name: string;
    role: string;
  };
}

// Регистрация
const register = async (email: string, password: string, name: string) => {
  const response = await fetch('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password, name })
  });

  if (!response.ok) throw new Error('Registration failed');

  const data = await response.json();
  localStorage.setItem('token', data.data.token);
  localStorage.setItem('refreshToken', data.data.refreshToken);
  return data.data.user;
};

// Вход
const login = async (email: string, password: string) => {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });

  if (!response.ok) throw new Error('Login failed');

  const data = await response.json();
  localStorage.setItem('token', data.data.token);
  localStorage.setItem('refreshToken', data.data.refreshToken);
  return data.data.user;
};

// Обновление токена
const refreshAccessToken = async () => {
  const refreshToken = localStorage.getItem('refreshToken');

  const response = await fetch('/api/auth/refresh', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refreshToken })
  });

  if (!response.ok) {
    // Токен истёк, нужен повторный вход
    localStorage.removeItem('token');
    localStorage.removeItem('refreshToken');
    throw new Error('Session expired');
  }

  const data = await response.json();
  localStorage.setItem('token', data.data.token);
  localStorage.setItem('refreshToken', data.data.refreshToken);
};

// Получение профиля
const getProfile = async () => {
  const token = localStorage.getItem('token');

  const response = await fetch('/api/me', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });

  if (!response.ok) throw new Error('Failed to get profile');
  return response.json();
};
```

---

## Примеры cURL

### Регистрация
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "Иван Петров"
  }'
```

### Вход
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Обновление токена
```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "YOUR_REFRESH_TOKEN"
  }'
```

### Получение профиля
```bash
curl http://localhost:8080/api/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## Коды ошибок

| HTTP Code | Описание |
|-----------|----------|
| 400 | Неверный формат запроса или ошибка валидации |
| 401 | Не авторизован или неверные учётные данные |
| 403 | Доступ запрещён (недостаточно прав) |
| 409 | Конфликт (email уже зарегистрирован) |
| 429 | Превышен лимит запросов |

---

**Последнее обновление**: 2025-11-29
**Версия проекта**: 0.1.0
**Статус**: Актуальный
