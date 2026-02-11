import * as Sentry from '@sentry/nextjs'

export async function register() {
  if (process.env.NEXT_RUNTIME === 'nodejs') {
    // Серверная конфигурация (объединяет server и edge)
    Sentry.init({
      dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,

      // Трассировка на сервере
      tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

      // Отключить в development
      enabled: process.env.NODE_ENV === 'production',
    })
  }

  if (process.env.NEXT_RUNTIME === 'edge') {
    // Edge конфигурация
    Sentry.init({
      dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,

      tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

      enabled: process.env.NODE_ENV === 'production',
    })
  }
}
