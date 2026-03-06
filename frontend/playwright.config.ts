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

  /* Количество воркеров */
  workers: process.env.CI ? 1 : undefined,

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
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },

    /* Тесты для мобильных устройств */
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },
  ],

  /* Запустить локальный dev сервер перед тестами */
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000,
  },
})
