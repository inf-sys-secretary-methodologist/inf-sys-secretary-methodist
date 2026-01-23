import * as Sentry from '@sentry/nextjs'

Sentry.init({
  dsn: process.env.NEXT_PUBLIC_SENTRY_DSN,

  // Процент трассировки (1.0 = 100% в dev, меньше в prod)
  tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,

  // Session Replay
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,

  // Отключить в development если нужно
  enabled: process.env.NODE_ENV === 'production',

  // Интеграции
  integrations: [
    Sentry.replayIntegration({
      maskAllText: true,
      blockAllMedia: true,
    }),
  ],

  // Фильтрация ошибок
  ignoreErrors: [
    // Игнорировать ошибки отмены запросов
    'AbortError',
    'cancelled',
    // Игнорировать ошибки расширений браузера
    /chrome-extension/,
    /moz-extension/,
  ],
})
