# Error Boundary System

Комплексная система обработки ошибок для React приложения на Next.js 15.

## Компоненты

### 1. Global Error Boundaries (Next.js App Router)

#### `app/error.tsx` - Основной Error Boundary

Обрабатывает ошибки в любом сегменте приложения (кроме root layout).

**Функции:**
- ✅ Отображает красивую страницу ошибки
- ✅ Показывает детали ошибки в dev mode
- ✅ Логирует ошибки в консоль
- ✅ Предоставляет кнопку "Попробовать снова"
- ✅ Предоставляет кнопку "На главную"
- ✅ Скрывает чувствительную информацию в production

**Автоматическое использование:**
Этот файл автоматически активируется при ошибках в маршрутах Next.js.

#### `app/global-error.tsx` - Root Layout Error Boundary

Обрабатывает критические ошибки в root layout.

**Функции:**
- ✅ Заменяет всю страницу (включая `<html>` и `<body>`)
- ✅ Обрабатывает ошибки в `layout.tsx` и `template.tsx`
- ✅ Логирует критические ошибки
- ✅ Показывает error digest для поддержки

**Автоматическое использование:**
Активируется только для ошибок в root layout.

### 2. ErrorBoundary Component (Переиспользуемый)

Класс-компонент для ловли ошибок в любом месте React дерева.

#### Базовое использование

```tsx
import { ErrorBoundary } from '@/components/error'

function MyPage() {
  return (
    <ErrorBoundary>
      <MyComponent />
    </ErrorBoundary>
  )
}
```

#### С кастомным сообщением

```tsx
<ErrorBoundary errorMessage="Не удалось загрузить данные пользователя">
  <UserProfile />
</ErrorBoundary>
```

#### С кастомным fallback

```tsx
<ErrorBoundary
  fallback={
    <div className="p-4 text-center">
      <p>Упс! Что-то пошло не так.</p>
      <button onClick={() => window.location.reload()}>
        Перезагрузить
      </button>
    </div>
  }
>
  <ComplexComponent />
</ErrorBoundary>
```

#### С обработчиком ошибок

```tsx
<ErrorBoundary
  onError={(error, errorInfo) => {
    // Отправить в систему мониторинга
    logToSentry(error, errorInfo)

    // Показать уведомление пользователю
    toast.error('Произошла ошибка')
  }}
>
  <CriticalComponent />
</ErrorBoundary>
```

#### Скрыть детали ошибки

```tsx
<ErrorBoundary showDetails={false}>
  <SensitiveComponent />
</ErrorBoundary>
```

## Props

| Prop | Type | Default | Описание |
|------|------|---------|----------|
| `children` | `ReactNode` | - | Дочерние компоненты для защиты |
| `fallback` | `ReactNode` | undefined | Кастомный UI при ошибке |
| `onError` | `(error, errorInfo) => void` | undefined | Колбэк при возникновении ошибки |
| `errorMessage` | `string` | undefined | Кастомное сообщение об ошибке |
| `showDetails` | `boolean` | `true` | Показывать ли детали ошибки в dev mode |

## Примеры использования

### Защита формы

```tsx
import { ErrorBoundary } from '@/components/error'

export function UserForm() {
  return (
    <ErrorBoundary errorMessage="Не удалось загрузить форму">
      <Form>
        <FormFields />
      </Form>
    </ErrorBoundary>
  )
}
```

### Защита раздела страницы

```tsx
import { ErrorBoundary } from '@/components/error'

export function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>

      {/* Каждый раздел защищён отдельно */}
      <ErrorBoundary errorMessage="Не удалось загрузить статистику">
        <Statistics />
      </ErrorBoundary>

      <ErrorBoundary errorMessage="Не удалось загрузить график">
        <Chart />
      </ErrorBoundary>

      <ErrorBoundary errorMessage="Не удалось загрузить список">
        <RecentActivity />
      </ErrorBoundary>
    </div>
  )
}
```

### Защита с интеграцией Sentry

```tsx
import { ErrorBoundary } from '@/components/error'
import * as Sentry from '@sentry/nextjs'

export function ProtectedComponent() {
  return (
    <ErrorBoundary
      onError={(error, errorInfo) => {
        Sentry.captureException(error, {
          contexts: {
            react: {
              componentStack: errorInfo.componentStack,
            },
          },
        })
      }}
    >
      <MyComponent />
    </ErrorBoundary>
  )
}
```

### Nested Error Boundaries

```tsx
import { ErrorBoundary } from '@/components/error'

export function ComplexPage() {
  return (
    <ErrorBoundary errorMessage="Ошибка загрузки страницы">
      <Layout>
        <Sidebar>
          <ErrorBoundary errorMessage="Ошибка загрузки навигации">
            <Navigation />
          </ErrorBoundary>
        </Sidebar>

        <Content>
          <ErrorBoundary errorMessage="Ошибка загрузки контента">
            <MainContent />
          </ErrorBoundary>
        </Content>
      </Layout>
    </ErrorBoundary>
  )
}
```

## Best Practices

### 1. Гранулярная защита

```tsx
// ✅ Good - каждый компонент защищён отдельно
<ErrorBoundary>
  <Header />
</ErrorBoundary>
<ErrorBoundary>
  <Content />
</ErrorBoundary>
<ErrorBoundary>
  <Footer />
</ErrorBoundary>

// ❌ Bad - одна ошибка роняет всю страницу
<ErrorBoundary>
  <Header />
  <Content />
  <Footer />
</ErrorBoundary>
```

### 2. Значимые сообщения об ошибках

```tsx
// ✅ Good - понятное сообщение
<ErrorBoundary errorMessage="Не удалось загрузить список студентов">
  <StudentList />
</ErrorBoundary>

// ❌ Bad - общее сообщение
<ErrorBoundary>
  <StudentList />
</ErrorBoundary>
```

### 3. Логирование ошибок

```tsx
// ✅ Good - ошибки логируются и отправляются в систему мониторинга
<ErrorBoundary
  onError={(error, errorInfo) => {
    logToService(error, errorInfo)
  }}
>
  <CriticalComponent />
</ErrorBoundary>

// ⚠️ Acceptable - только консоль (по умолчанию)
<ErrorBoundary>
  <Component />
</ErrorBoundary>
```

### 4. Кастомный UI для production

```tsx
// ✅ Good - дружелюбный UI в production
<ErrorBoundary
  fallback={
    <div className="text-center p-8">
      <h3>Упс! Что-то пошло не так</h3>
      <p>Мы уже работаем над исправлением</p>
    </div>
  }
>
  <Component />
</ErrorBoundary>
```

## Что НЕ ловится Error Boundaries

Error Boundaries НЕ ловят ошибки в:

- Event handlers (используйте try-catch)
- Асинхронный код (setTimeout, promises)
- Server-side rendering
- Ошибки в самом error boundary

### Примеры

```tsx
// ❌ Error Boundary НЕ поймает эту ошибку
function MyComponent() {
  const handleClick = () => {
    throw new Error('Error in event handler')
  }

  return <button onClick={handleClick}>Click</button>
}

// ✅ Используйте try-catch для event handlers
function MyComponent() {
  const handleClick = () => {
    try {
      // код, который может упасть
      riskyOperation()
    } catch (error) {
      console.error('Error:', error)
      toast.error('Произошла ошибка')
    }
  }

  return <button onClick={handleClick}>Click</button>
}
```

```tsx
// ❌ Error Boundary НЕ поймает асинхронную ошибку
function MyComponent() {
  useEffect(() => {
    fetchData().then(data => {
      throw new Error('Async error')
    })
  }, [])

  return <div>Component</div>
}

// ✅ Используйте try-catch для async/await
function MyComponent() {
  useEffect(() => {
    const loadData = async () => {
      try {
        const data = await fetchData()
        // обработка данных
      } catch (error) {
        console.error('Error:', error)
        setError(error)
      }
    }

    loadData()
  }, [])

  return <div>Component</div>
}
```

## Integration с системами мониторинга

### Sentry

```typescript
// lib/error-tracking.ts
import * as Sentry from '@sentry/nextjs'

export function logErrorToSentry(
  error: Error,
  errorInfo?: React.ErrorInfo
) {
  Sentry.captureException(error, {
    contexts: {
      react: errorInfo ? {
        componentStack: errorInfo.componentStack,
      } : undefined,
    },
  })
}

// Использование
<ErrorBoundary onError={logErrorToSentry}>
  <Component />
</ErrorBoundary>
```

### Custom Analytics

```typescript
// lib/analytics.ts
export function logErrorToAnalytics(
  error: Error,
  errorInfo?: React.ErrorInfo
) {
  analytics.track('Error Occurred', {
    message: error.message,
    stack: error.stack,
    componentStack: errorInfo?.componentStack,
    timestamp: new Date().toISOString(),
  })
}

// Использование
<ErrorBoundary onError={logErrorToAnalytics}>
  <Component />
</ErrorBoundary>
```

## Troubleshooting

### Error Boundary не ловит ошибку

1. Проверьте, что ошибка происходит во время рендеринга
2. Убедитесь, что ErrorBoundary обёрнут вокруг компонента с ошибкой
3. Проверьте, что это не асинхронная ошибка (используйте try-catch)

### Детали ошибки не отображаются

1. Убедитесь, что `NODE_ENV === 'development'`
2. Проверьте prop `showDetails={true}`
3. Проверьте, что ошибка имеет поле `message`

### Кнопка "Попробовать снова" не работает

1. Убедитесь, что состояние компонента сбрасывается
2. Проверьте, что ошибка не возникает в `constructor` или `componentDidMount`
3. Рассмотрите использование `key` prop для принудительного remount

## Заключение

Система Error Boundary обеспечивает:
- ✅ Graceful degradation при ошибках
- ✅ Информативные сообщения об ошибках
- ✅ Возможность recovery без перезагрузки страницы
- ✅ Интеграцию с системами мониторинга
- ✅ Разные уровни гранулярности защиты
- ✅ Production-ready error handling
