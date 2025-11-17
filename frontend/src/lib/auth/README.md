н# Authentication & Authorization System

Система аутентификации и авторизации с поддержкой Role-Based Access Control (RBAC).

## Компоненты системы

### 1. JWT Utilities (`jwt.ts`)

Утилиты для работы с JWT токенами:

```typescript
import { decodeJWT, isTokenExpired, isTokenExpiringSoon } from '@/lib/auth/jwt'

// Декодировать токен
const payload = decodeJWT(token)
console.log(payload.email, payload.role)

// Проверить истечение
if (isTokenExpired(token)) {
  // Токен истёк, нужно обновить
}

// Проверить скоро ли истечёт (< 5 минут)
if (isTokenExpiringSoon(token)) {
  // Время обновить токен
}
```

### 2. Route Configuration (`route-config.ts`)

Конфигурация маршрутов с RBAC:

```typescript
import { hasRouteAccess, getRouteConfig } from '@/lib/auth/route-config'
import { UserRole } from '@/types/auth'

// Проверить доступ к маршруту
const canAccess = hasRouteAccess('/documents', UserRole.METHODIST) // true
const cannotAccess = hasRouteAccess('/admin', UserRole.STUDENT) // false

// Получить конфигурацию маршрута
const config = getRouteConfig('/documents')
console.log(config?.allowedRoles) // [SYSTEM_ADMIN, METHODIST, ACADEMIC_SECRETARY]
```

#### Публичные маршруты
- `/` - главная страница
- `/login` - вход
- `/register` - регистрация
- `/forgot-password` - восстановление пароля
- `/reset-password` - сброс пароля

#### Защищённые маршруты по ролям

| Маршрут | Роли с доступом |
|---------|-----------------|
| `/admin` | SYSTEM_ADMIN |
| `/users` | SYSTEM_ADMIN |
| `/documents` | SYSTEM_ADMIN, METHODIST, ACADEMIC_SECRETARY |
| `/templates` | SYSTEM_ADMIN, METHODIST |
| `/reports` | SYSTEM_ADMIN, METHODIST, ACADEMIC_SECRETARY |
| `/schedule` | SYSTEM_ADMIN, ACADEMIC_SECRETARY, METHODIST |
| `/tasks` | SYSTEM_ADMIN, ACADEMIC_SECRETARY, METHODIST |
| `/students` | SYSTEM_ADMIN, METHODIST, ACADEMIC_SECRETARY, TEACHER |
| `/dashboard` | Все авторизованные |
| `/profile` | Все авторизованные |
| `/settings` | Все авторизованные |

### 3. Middleware (`middleware.ts`)

Next.js middleware для серверной защиты маршрутов:

**Автоматически:**
- ✅ Проверяет аутентификацию
- ✅ Валидирует JWT токен
- ✅ Проверяет истечение токена
- ✅ Проверяет RBAC доступ
- ✅ Редиректит на `/login` если не авторизован
- ✅ Редиректит на `/forbidden` если нет прав
- ✅ Сохраняет intended URL для редиректа после логина
- ✅ Очищает cookie при истечении токена

**Обработка ошибок:**
- Expired token → redirect to `/login?redirect={path}&session_expired=true`
- Invalid token → redirect to `/login?redirect={path}`
- No permission → redirect to `/forbidden`

### 4. withAuth HOC (`components/auth/withAuth.tsx`)

Higher-Order Component для client-side защиты:

#### Базовое использование

```typescript
import { withAuth } from '@/components/auth/withAuth'

// Защитить страницу - любой авторизованный пользователь
function DashboardPage() {
  return <div>Dashboard</div>
}

export default withAuth(DashboardPage)
```

#### Защита с проверкой роли

```typescript
import { withAuth } from '@/components/auth/withAuth'
import { UserRole } from '@/types/auth'

// Только для админов
function AdminPage() {
  return <div>Admin Panel</div>
}

export default withAuth(AdminPage, {
  roles: [UserRole.SYSTEM_ADMIN]
})
```

#### Защита для нескольких ролей

```typescript
import { withAuth } from '@/components/auth/withAuth'
import { UserRole } from '@/types/auth'

// Для админов, методистов и секретарей
function DocumentsPage() {
  return <div>Documents</div>
}

export default withAuth(DocumentsPage, {
  roles: [
    UserRole.SYSTEM_ADMIN,
    UserRole.METHODIST,
    UserRole.ACADEMIC_SECRETARY
  ]
})
```

#### Кастомный loading компонент

```typescript
import { withAuth } from '@/components/auth/withAuth'

function CustomLoader() {
  return <div className="custom-loader">Loading...</div>
}

function MyPage() {
  return <div>My Page</div>
}

export default withAuth(MyPage, {
  LoadingComponent: CustomLoader
})
```

#### Кастомный redirect

```typescript
import { withAuth } from '@/components/auth/withAuth'

function MyPage() {
  return <div>My Page</div>
}

// Redirect to custom page instead of /login
export default withAuth(MyPage, {
  redirectTo: '/auth/signin'
})
```

## Роли пользователей

```typescript
enum UserRole {
  SYSTEM_ADMIN = 'system_admin',           // Полный доступ
  METHODIST = 'methodist',                  // Методист
  ACADEMIC_SECRETARY = 'academic_secretary', // Секретарь
  TEACHER = 'teacher',                      // Преподаватель
  STUDENT = 'student',                      // Студент
}
```

## Примеры использования

### Создание защищённой страницы

```typescript
// app/documents/page.tsx
"use client"

import { withAuth } from '@/components/auth/withAuth'
import { UserRole } from '@/types/auth'

function DocumentsPage() {
  return (
    <div>
      <h1>Documents Management</h1>
      {/* Your content */}
    </div>
  )
}

// Only methodists and secretaries can access
export default withAuth(DocumentsPage, {
  roles: [UserRole.METHODIST, UserRole.ACADEMIC_SECRETARY]
})
```

### Условный рендеринг по ролям

```typescript
"use client"

import { useAuthStore } from '@/stores/authStore'
import { UserRole } from '@/types/auth'

export function Sidebar() {
  const { user } = useAuthStore()

  return (
    <nav>
      <a href="/dashboard">Dashboard</a>
      <a href="/profile">Profile</a>

      {/* Only for admins */}
      {user?.role === UserRole.SYSTEM_ADMIN && (
        <>
          <a href="/admin">Admin Panel</a>
          <a href="/users">Users</a>
        </>
      )}

      {/* Only for methodists and secretaries */}
      {(user?.role === UserRole.METHODIST || user?.role === UserRole.ACADEMIC_SECRETARY) && (
        <a href="/documents">Documents</a>
      )}
    </nav>
  )
}
```

### Проверка прав доступа перед действием

```typescript
"use client"

import { useAuthStore } from '@/stores/authStore'
import { UserRole } from '@/types/auth'
import { hasRouteAccess } from '@/lib/auth/route-config'

export function DocumentActions() {
  const { user } = useAuthStore()

  const canDelete = user && [UserRole.SYSTEM_ADMIN, UserRole.METHODIST].includes(user.role)

  const handleDelete = () => {
    if (!canDelete) {
      alert('У вас нет прав для удаления документов')
      return
    }

    // Perform delete action
  }

  return (
    <div>
      {canDelete && (
        <button onClick={handleDelete}>Delete Document</button>
      )}
    </div>
  )
}
```

## Best Practices

### 1. Всегда используйте withAuth для защищённых страниц

```typescript
// ✅ Good
export default withAuth(MyPage)

// ❌ Bad - страница не защищена
export default MyPage
```

### 2. Указывайте роли когда они нужны

```typescript
// ✅ Good - четко указаны требуемые роли
export default withAuth(AdminPage, {
  roles: [UserRole.SYSTEM_ADMIN]
})

// ⚠️ Acceptable - если доступ нужен всем авторизованным
export default withAuth(DashboardPage)
```

### 3. Не полагайтесь только на client-side защиту

```typescript
// ✅ Good - защита на сервере (middleware) + клиенте (withAuth)
// middleware.ts проверяет автоматически
export default withAuth(SecretPage)

// ❌ Bad - только client-side защита (легко обойти)
function SecretPage() {
  const { isAuthenticated } = useAuthStore()
  if (!isAuthenticated) return null
  return <div>Secret</div>
}
```

### 4. Проверяйте роли для критичных операций

```typescript
// ✅ Good
const handleDeleteUser = async (userId: string) => {
  if (user?.role !== UserRole.SYSTEM_ADMIN) {
    throw new Error('Unauthorized')
  }

  // Also verify on backend!
  await deleteUser(userId)
}

// ❌ Bad - нет проверки роли
const handleDeleteUser = async (userId: string) => {
  await deleteUser(userId) // Easy to bypass!
}
```

## Troubleshooting

### Infinite redirect loop
Убедитесь, что `/login` и `/forbidden` в `publicRoutes` или не требуют авторизации.

### "Session expired" не показывается
Проверьте query параметр `session_expired` в URL и покажите уведомление на странице логина.

### Пользователь видит "403 Forbidden" вместо контента
Проверьте:
1. Правильно ли указана роль в `protectedRoutes`
2. У пользователя есть правильная роль в токене
3. Токен не истёк

### withAuth не работает
Убедитесь, что:
1. Компонент использует `"use client"`
2. AuthStore корректно настроен
3. Токен хранится в cookies с именем `auth-storage`

## Security Notes

⚠️ **Важно**: Client-side проверки (withAuth, useAuthStore) могут быть обойдены опытным пользователем. Всегда:

1. ✅ Используйте middleware для серверной защиты
2. ✅ Проверяйте права на backend при каждом API запросе
3. ✅ Валидируйте JWT на backend
4. ✅ Не храните секреты в client-side коде

Middleware обеспечивает серверную защиту, withAuth - улучшает UX.
