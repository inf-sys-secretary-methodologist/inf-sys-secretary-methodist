# 🔐 Authentication API

## 📋 Обзор аутентификации

Сервис аутентификации обеспечивает безопасный вход в систему через OAuth 2.0 провайдеры и управление JWT токенами для всех микросервисов.

## 🌐 Base URL
```
https://api.inf-sys.example.com/auth
```

## 🔑 Поддерживаемые провайдеры

- **Google Workspace** - основной провайдер для сотрудников
- **Microsoft Azure AD** - альтернативный корпоративный провайдер
- **Local Auth** - локальная аутентификация для специальных случаев

---

## 🚀 Endpoints

### POST `/auth/login`
Инициация процесса аутентификации

**Request:**
```json
{
  "provider": "google|azure|local",
  "code": "oauth_authorization_code",
  "redirect_uri": "https://app.inf-sys.example.com/callback",
  "state": "optional_state_parameter"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_here",
  "expires_in": 900,
  "token_type": "Bearer",
  "user": {
    "id": "12345",
    "email": "metodist@university.edu",
    "first_name": "Анна",
    "last_name": "Петрова",
    "roles": ["методист", "преподаватель"],
    "permissions": [
      "documents:create",
      "documents:read",
      "documents:update",
      "schedule:read",
      "reports:generate"
    ],
    "department": "Математический факультет",
    "position": "Старший методист"
  }
}
```

### POST `/auth/refresh`
Обновление access token с использованием refresh token

**Request:**
```json
{
  "refresh_token": "refresh_token_here"
}
```

**Response (200):**
```json
{
  "access_token": "new_jwt_token_here",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

### POST `/auth/logout`
Завершение сессии и аннулирование токенов

**Request:**
```json
{
  "refresh_token": "refresh_token_here"
}
```

**Response (204):**
```
No Content
```

### GET `/auth/me`
Получение информации о текущем пользователе

**Headers:**
```http
Authorization: Bearer <JWT_TOKEN>
```

**Response (200):**
```json
{
  "user": {
    "id": "12345",
    "email": "metodist@university.edu",
    "first_name": "Анна",
    "last_name": "Петрова",
    "roles": ["методист", "преподаватель"],
    "permissions": [
      "documents:create",
      "documents:read",
      "documents:update",
      "schedule:read",
      "reports:generate"
    ],
    "department": "Математический факультет",
    "position": "Старший методист",
    "last_login": "2025-01-15T14:30:00Z",
    "created_at": "2024-09-01T10:00:00Z"
  }
}
```

### POST `/auth/change-password`
Изменение пароля (только для local auth)

**Request:**
```json
{
  "current_password": "current_password",
  "new_password": "new_secure_password",
  "confirm_password": "new_secure_password"
}
```

**Response (200):**
```json
{
  "message": "Password changed successfully"
}
```

---

## 🔐 JWT Token Structure

### Header
```json
{
  "alg": "RS256",
  "typ": "JWT",
  "kid": "key_id_here"
}
```

### Payload
```json
{
  "iss": "https://api.inf-sys.example.com",
  "sub": "12345",
  "aud": "inf-sys-app",
  "exp": 1642694400,
  "iat": 1642693500,
  "roles": ["методист", "преподаватель"],
  "permissions": [
    "documents:create",
    "documents:read",
    "schedule:read"
  ],
  "department": "Математический факультет"
}
```

---

## 🛡️ Роли и разрешения

### Системные роли:
- **админ** - полный доступ ко всей системе
- **методист** - управление учебными планами и методическими материалами
- **секретарь** - работа с документооборотом и расписанием
- **преподаватель** - доступ к расписанию и учебным материалам
- **студент** - просмотр расписания и учебных материалов

### Разрешения (permissions):
```
documents:create, documents:read, documents:update, documents:delete
schedule:create, schedule:read, schedule:update, schedule:delete
reports:generate, reports:read
users:create, users:read, users:update, users:delete
tasks:create, tasks:read, tasks:update, tasks:delete
workflows:manage
integrations:manage
```

---

## 🔒 Безопасность

### Rate Limiting
- **Login attempts**: 5 попыток в 15 минут на IP
- **Token refresh**: 10 запросов в минуту на пользователя
- **Password change**: 3 попытки в час на пользователя

### Token Security
- **Access Token TTL**: 15 минут
- **Refresh Token TTL**: 7 дней
- **Automatic rotation**: Refresh tokens обновляются при каждом использовании
- **Revocation**: Все токены пользователя аннулируются при logout

---

## 🌐 Frontend Integration

### React Hook Example
```typescript
import { useState, useEffect } from 'react';

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  roles: string[];
  permissions: string[];
}

export const useAuth = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const login = async (provider: string, code: string) => {
    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          provider,
          code,
          redirect_uri: window.location.origin + '/callback'
        })
      });

      const data = await response.json();

      // Store tokens
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);

      setUser(data.user);
      return data;
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const logout = async () => {
    try {
      const refreshToken = localStorage.getItem('refresh_token');

      await fetch('/api/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken })
      });
    } finally {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      setUser(null);
    }
  };

  const getCurrentUser = async () => {
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch('/api/auth/me', {
        headers: { Authorization: `Bearer ${token}` }
      });

      if (response.ok) {
        const data = await response.json();
        setUser(data.user);
      }
    } catch (error) {
      console.error('Failed to get current user:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    getCurrentUser();
  }, []);

  return { user, loading, login, logout };
};
```

---

## 🔧 Backend Integration (Go)

### JWT Middleware
```go
package middleware

import (
    "context"
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID      string   `json:"sub"`
    Roles       []string `json:"roles"`
    Permissions []string `json:"permissions"`
    Department  string   `json:"department"`
    jwt.RegisteredClaims
}

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
            return getPublicKey(), nil // Implement your key retrieval
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        claims, ok := token.Claims.(*Claims)
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
            c.Abort()
            return
        }

        // Add user info to context
        c.Set("user_id", claims.UserID)
        c.Set("roles", claims.Roles)
        c.Set("permissions", claims.Permissions)
        c.Set("department", claims.Department)

        c.Next()
    }
}

// Permission check middleware
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        permissions, exists := c.Get("permissions")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "No permissions"})
            c.Abort()
            return
        }

        userPermissions := permissions.([]string)
        for _, p := range userPermissions {
            if p == permission {
                c.Next()
                return
            }
        }

        c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
        c.Abort()
    }
}
```

---

## 🧪 Testing Examples

### Unit Tests (Go)
```go
func TestAuthLogin(t *testing.T) {
    router := setupTestRouter()

    payload := `{
        "provider": "google",
        "code": "test_auth_code",
        "redirect_uri": "http://localhost:3000/callback"
    }`

    req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.Contains(t, response, "access_token")
    assert.Contains(t, response, "user")
}
```

### Integration Tests (Jest)
```typescript
describe('Auth API', () => {
  test('should login successfully', async () => {
    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        provider: 'google',
        code: 'test_code',
        redirect_uri: 'http://localhost:3000/callback'
      })
    });

    expect(response.status).toBe(200);

    const data = await response.json();
    expect(data).toHaveProperty('access_token');
    expect(data).toHaveProperty('user');
    expect(data.user).toHaveProperty('roles');
  });
});
```

---

## ⚡ Performance

### Metrics
- **Login Response Time**: < 500ms (95th percentile)
- **Token Validation**: < 50ms (99th percentile)
- **Concurrent Users**: 1000+ simultaneous authentications

### Caching
- Public keys cached for 1 hour
- User permissions cached for 5 minutes
- Rate limit data cached in Redis

---

## 🚨 Error Codes

| Code | Message | Description |
|------|---------|-------------|
| AUTH_001 | Invalid credentials | Неверные учетные данные |
| AUTH_002 | Token expired | Токен истек |
| AUTH_003 | Invalid token | Недействительный токен |
| AUTH_004 | Insufficient permissions | Недостаточно прав |
| AUTH_005 | Account locked | Аккаунт заблокирован |
| AUTH_006 | Rate limit exceeded | Превышен лимит запросов |
| AUTH_007 | Invalid provider | Неподдерживаемый провайдер |