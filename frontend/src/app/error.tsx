'use client'

import { useEffect } from 'react'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  useEffect(() => {
    console.error('Application error:', error)
  }, [error])

  return (
    <div className="min-h-screen flex items-center justify-center bg-background p-4">
      <div className="max-w-md w-full space-y-6 text-center">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white">
            Что-то пошло не так
          </h1>
          <p className="text-gray-600 dark:text-gray-300">
            Произошла непредвиденная ошибка. Пожалуйста, попробуйте еще раз.
          </p>
        </div>

        {error.message && (
          <div className="p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800">
            <p className="text-sm text-red-800 dark:text-red-200 font-mono">
              {error.message}
            </p>
          </div>
        )}

        <div className="flex gap-4 justify-center">
          <button
            onClick={reset}
            className="px-6 py-3 rounded-lg font-medium transition-all duration-300 bg-gray-900 text-white hover:bg-gray-800 hover:scale-105 active:scale-95 shadow-lg hover:shadow-xl"
          >
            Попробовать снова
          </button>
          <button
            onClick={() => (window.location.href = '/')}
            className="px-6 py-3 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-gray-800 text-gray-900 dark:text-white hover:bg-gray-100 dark:hover:bg-gray-700 border border-gray-200 dark:border-gray-600 hover:scale-105 active:scale-95 shadow-lg hover:shadow-xl"
          >
            На главную
          </button>
        </div>
      </div>
    </div>
  )
}
