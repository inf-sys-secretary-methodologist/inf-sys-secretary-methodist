import { defineConfig, devices } from '@playwright/test'

/**
 * Конфигурация Playwright для E2E тестирования
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './tests/e2e',

  /* Запускать тесты параллельно */
  fullyParallel: true,

  /* Падать на CI если случайно оставили test.only */
  forbidOnly: !!process.env.CI,

  /* Повторять тесты на CI при падении */
  retries: process.env.CI ? 2 : 0,

  /* На CI '50%' ядер (public-repo runner = 4 vCPU → 2 воркера). 75% (3 воркера)
     давало networkidle-таймауты: 3 воркера + co-located Next-сервер не оставляли
     запаса CPU. 2 воркера держат ядро под сервер/ОС → стабильно и ~2× быстрее
     сериала. Состояние между тестами не шарится (бэкенда нет). Локально — авто. */
  workers: process.env.CI ? '50%' : undefined,

  /* Репортер для результатов */
  reporter: 'html',

  /* Общие настройки для всех проектов */
  use: {
    /* Базовый URL для page.goto('/') */
    baseURL: process.env.BASE_URL || 'http://localhost:3000',

    /* Собирать trace при повторном запуске упавшего теста */
    trace: 'on-first-retry',

    /* Скриншот при падении */
    screenshot: 'only-on-failure',
  },

  /* Конфигурация для разных браузеров */
  projects: process.env.CI
    ? [
        {
          name: 'chromium',
          use: { ...devices['Desktop Chrome'] },
        },
      ]
    : [
        {
          name: 'chromium',
          use: { ...devices['Desktop Chrome'] },
        },
        {
          name: 'webkit',
          use: { ...devices['Desktop Safari'] },
        },
        {
          name: 'Mobile Chrome',
          use: { ...devices['Pixel 5'] },
        },
        {
          name: 'Mobile Safari',
          use: { ...devices['iPhone 12'] },
        },
      ],

  /* Запустить локальный сервер перед тестами */
  webServer: {
    command: process.env.CI ? 'npm run build && npm run start' : 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
})
