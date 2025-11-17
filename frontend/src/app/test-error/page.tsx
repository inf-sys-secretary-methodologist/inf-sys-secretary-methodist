'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ErrorBoundary } from '@/components/error'

/**
 * Test page for Error Boundaries
 * This page should be removed in production
 */

// Component that throws during render
function RenderErrorComponent() {
  throw new Error('Test error during component render')
  return <div>This will never render</div>
}

// Component that throws in event handler
function EventErrorComponent() {
  const handleClick = () => {
    throw new Error('Test error in event handler')
  }

  return (
    <Button onClick={handleClick} variant="destructive">
      Throw Event Error (not caught by ErrorBoundary)
    </Button>
  )
}

// Component that throws async error
function AsyncErrorComponent() {
  const handleClick = async () => {
    await new Promise(resolve => setTimeout(resolve, 1000))
    throw new Error('Test async error')
  }

  return (
    <Button onClick={handleClick} variant="destructive">
      Throw Async Error (not caught by ErrorBoundary)
    </Button>
  )
}

export default function TestErrorPage() {
  const [showRenderError, setShowRenderError] = useState(false)

  return (
    <div className="container mx-auto p-8 space-y-8">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold">Error Boundary Test Page</h1>
        <p className="text-muted-foreground">
          Эта страница для тестирования Error Boundaries. Удалите в production.
        </p>
      </div>

      {/* Test 1: Render Error (caught by ErrorBoundary) */}
      <div className="border rounded-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold">
          Test 1: Render Error (caught ✅)
        </h2>
        <p className="text-sm text-muted-foreground">
          Ошибка выбрасывается во время рендеринга компонента
        </p>

        <ErrorBoundary errorMessage="Тестовая ошибка рендеринга">
          {showRenderError ? (
            <RenderErrorComponent />
          ) : (
            <Button
              onClick={() => setShowRenderError(true)}
              variant="destructive"
            >
              Trigger Render Error
            </Button>
          )}
        </ErrorBoundary>

        {showRenderError && (
          <Button onClick={() => setShowRenderError(false)} variant="outline">
            Reset
          </Button>
        )}
      </div>

      {/* Test 2: Event Handler Error (NOT caught by ErrorBoundary) */}
      <div className="border rounded-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold">
          Test 2: Event Handler Error (not caught ❌)
        </h2>
        <p className="text-sm text-muted-foreground">
          Ошибка в event handler не ловится ErrorBoundary. Смотрите консоль.
        </p>

        <ErrorBoundary>
          <EventErrorComponent />
        </ErrorBoundary>

        <p className="text-xs text-yellow-600 dark:text-yellow-400">
          ⚠️ Для event handlers используйте try-catch
        </p>
      </div>

      {/* Test 3: Async Error (NOT caught by ErrorBoundary) */}
      <div className="border rounded-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold">
          Test 3: Async Error (not caught ❌)
        </h2>
        <p className="text-sm text-muted-foreground">
          Асинхронная ошибка не ловится ErrorBoundary. Смотрите консоль.
        </p>

        <ErrorBoundary>
          <AsyncErrorComponent />
        </ErrorBoundary>

        <p className="text-xs text-yellow-600 dark:text-yellow-400">
          ⚠️ Для async кода используйте try-catch
        </p>
      </div>

      {/* Test 4: Multiple ErrorBoundaries */}
      <div className="border rounded-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold">
          Test 4: Multiple ErrorBoundaries (isolated errors)
        </h2>
        <p className="text-sm text-muted-foreground">
          Каждый раздел защищён отдельно. Ошибка в одном не влияет на другие.
        </p>

        <div className="grid grid-cols-3 gap-4">
          <ErrorBoundary errorMessage="Ошибка в секции 1">
            <div className="border p-4 rounded">
              <h3 className="font-semibold mb-2">Section 1</h3>
              <p className="text-sm text-muted-foreground">
                Эта секция работает нормально
              </p>
            </div>
          </ErrorBoundary>

          <ErrorBoundary errorMessage="Ошибка в секции 2">
            <div className="border p-4 rounded">
              <h3 className="font-semibold mb-2">Section 2</h3>
              <RenderErrorComponent />
            </div>
          </ErrorBoundary>

          <ErrorBoundary errorMessage="Ошибка в секции 3">
            <div className="border p-4 rounded">
              <h3 className="font-semibold mb-2">Section 3</h3>
              <p className="text-sm text-muted-foreground">
                Эта секция тоже работает нормально
              </p>
            </div>
          </ErrorBoundary>
        </div>
      </div>

      {/* Test 5: Custom Fallback */}
      <div className="border rounded-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold">
          Test 5: Custom Fallback UI
        </h2>
        <p className="text-sm text-muted-foreground">
          ErrorBoundary с кастомным fallback компонентом
        </p>

        <ErrorBoundary
          fallback={
            <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 p-4 rounded">
              <p className="text-yellow-800 dark:text-yellow-200 font-semibold">
                🎨 Custom Fallback UI
              </p>
              <p className="text-sm text-yellow-700 dark:text-yellow-300 mt-1">
                Это кастомный UI для ошибки
              </p>
            </div>
          }
        >
          <RenderErrorComponent />
        </ErrorBoundary>
      </div>

      {/* Test 6: Test Next.js error.tsx */}
      <div className="border rounded-lg p-6 space-y-4">
        <h2 className="text-xl font-semibold">
          Test 6: Next.js error.tsx
        </h2>
        <p className="text-sm text-muted-foreground">
          Выбросить ошибку, чтобы протестировать app/error.tsx
        </p>

        <Button
          onClick={() => {
            throw new Error('Test Next.js error.tsx boundary')
          }}
          variant="destructive"
        >
          Trigger Next.js Error Boundary
        </Button>

        <p className="text-xs text-yellow-600 dark:text-yellow-400">
          ⚠️ Это покажет страницу app/error.tsx
        </p>
      </div>

      {/* Info Section */}
      <div className="border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 rounded-lg p-6">
        <h2 className="text-xl font-semibold text-blue-900 dark:text-blue-100 mb-2">
          ℹ️ Информация
        </h2>
        <ul className="text-sm text-blue-800 dark:text-blue-200 space-y-1 list-disc list-inside">
          <li>Error Boundaries ловят ошибки только во время рендеринга</li>
          <li>Ошибки в event handlers нужно ловить через try-catch</li>
          <li>Асинхронные ошибки нужно ловить через try-catch</li>
          <li>
            Используйте ErrorBoundary для изоляции ошибок в разных частях UI
          </li>
          <li>В dev mode показываются детали ошибки</li>
          <li>В production показывается только error digest</li>
        </ul>
      </div>
    </div>
  )
}
