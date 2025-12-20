'use client'

import { useEffect, useState } from 'react'
import { AlertTriangle } from 'lucide-react'

type Locale = 'ru' | 'en' | 'fr' | 'ar'

// Translations for global error page (must be inline since providers are not available)
const translations: Record<
  Locale,
  {
    title: string
    description: string
    details: string
    supportCode: string
    retry: string
    reload: string
  }
> = {
  ru: {
    title: 'Критическая ошибка',
    description:
      'Приложение столкнулось с критической ошибкой. Пожалуйста, перезагрузите страницу.',
    details: 'Детали ошибки:',
    supportCode: 'Код ошибки для службы поддержки:',
    retry: 'Попробовать снова',
    reload: 'Перезагрузить страницу',
  },
  en: {
    title: 'Critical Error',
    description: 'The application encountered a critical error. Please reload the page.',
    details: 'Error details:',
    supportCode: 'Error code for support:',
    retry: 'Try Again',
    reload: 'Reload Page',
  },
  fr: {
    title: 'Erreur critique',
    description: "L'application a rencontré une erreur critique. Veuillez recharger la page.",
    details: "Détails de l'erreur :",
    supportCode: 'Code erreur pour le support :',
    retry: 'Réessayer',
    reload: 'Recharger la page',
  },
  ar: {
    title: 'خطأ حرج',
    description: 'واجه التطبيق خطأ حرجًا. يرجى إعادة تحميل الصفحة.',
    details: 'تفاصيل الخطأ:',
    supportCode: 'رمز الخطأ للدعم:',
    retry: 'حاول مرة أخرى',
    reload: 'إعادة تحميل الصفحة',
  },
}

function getLocaleFromCookie(): Locale {
  if (typeof document === 'undefined') return 'ru'
  const match = document.cookie.match(/NEXT_LOCALE=([^;]+)/)
  const locale = match?.[1] as Locale | undefined
  return locale && locale in translations ? locale : 'ru'
}

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
  const [locale, setLocale] = useState<Locale>('ru')
  const t = translations[locale]
  const isRtl = locale === 'ar'

  useEffect(() => {
    setLocale(getLocaleFromCookie())
  }, [])

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
    <html lang={locale} dir={isRtl ? 'rtl' : 'ltr'}>
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
                {t.title}
              </h1>
              <p className="text-gray-600 dark:text-gray-300">{t.description}</p>
            </div>

            {/* Error Details (Development only) */}
            {process.env.NODE_ENV === 'development' && error.message && (
              <div className="p-4 rounded-lg bg-red-50 dark:bg-red-900/10 border border-red-200 dark:border-red-800 text-left">
                <p className="text-sm font-semibold text-red-800 dark:text-red-200 mb-2">
                  {t.details}
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
                <p className="text-sm text-gray-700 dark:text-gray-300">{t.supportCode}</p>
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
                {t.retry}
              </button>
              <button
                onClick={() => (window.location.href = '/')}
                className="px-6 py-3 rounded-lg font-medium transition-all duration-300 bg-white dark:bg-gray-800 text-gray-900 dark:text-white border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 shadow-lg hover:shadow-xl"
              >
                {t.reload}
              </button>
            </div>
          </div>
        </div>
      </body>
    </html>
  )
}
