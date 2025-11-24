'use client'

import { useEffect } from 'react'
import { AlertTriangle } from 'lucide-react'

/**
 * Global error boundary for root layout errors
 * This component catches errors that occur in the root layout
 * Must include <html> and <body> tags as it replaces the entire page
 */
export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    // Log critical error to console
    console.error('Critical application error (global-error):', {
      message: error.message,
      digest: error.digest,
      stack: error.stack,
      timestamp: new Date().toISOString(),
    })

    // TODO: Send to error tracking service (e.g., Sentry) with high priority
    // logCriticalError(error)
  }, [error])

  return (
    <html lang="ru">
      <body>
        <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 p-4">
          <div className="max-w-md w-full space-y-6 text-center">
            {/* Critical Error Icon */}
            <div className="flex justify-center">
              <div className="rounded-full bg-red-100 dark:bg-red-900/20 p-6">
                <AlertTriangle className="h-16 w-16 text-red-600 dark:text-red-400" />
              </div>
            </div>

            {/* Error Title */}
            <div className="space-y-2">
              <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">
                Критическая ошибка
              </h1>
              <p className="text-gray-600 dark:text-gray-300">
                Приложение столкнулось с критической ошибкой. Пожалуйста, перезагрузите страницу.
              </p>
            </div>

            {/* Error Details (Development only) */}
            {process.env.NODE_ENV === 'development' && error.message && (
              <div className="p-4 rounded-lg bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-800 text-left">
                <p className="text-sm font-semibold text-red-800 dark:text-red-200 mb-2">
                  Детали ошибки:
                </p>
                <p className="text-xs text-red-700 dark:text-red-300 font-mono break-words">
                  {error.message}
                </p>
                {error.digest && (
                  <p className="text-xs text-gray-600 dark:text-gray-400 mt-2">
                    Error ID: {error.digest}
                  </p>
                )}
              </div>
            )}

            {/* Production error digest */}
            {process.env.NODE_ENV === 'production' && error.digest && (
              <div className="p-4 rounded-lg bg-gray-100 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 text-left">
                <p className="text-sm text-gray-700 dark:text-gray-300">
                  Код ошибки для службы поддержки:
                </p>
                <p className="text-xs font-mono text-gray-900 dark:text-white mt-1">
                  {error.digest}
                </p>
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex flex-col sm:flex-row gap-3 justify-center">
              <button
                onClick={reset}
                className="px-6 py-3 rounded-lg font-medium transition-all duration-300 bg-gray-900 dark:bg-white text-white dark:text-gray-900 hover:bg-gray-800 dark:hover:bg-gray-100 shadow-lg hover:shadow-xl"
              >
                Попробовать снова
              </button>
              <button
                onClick={() => (window.location.href = '/')}
                className="px-6 py-3 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-gray-800 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 shadow-lg hover:shadow-xl"
              >
                Перезагрузить страницу
              </button>
            </div>
          </div>
        </div>
      </body>
    </html>
  )
}
