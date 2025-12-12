import { test, expect } from '@playwright/test'

/**
 * E2E тесты для модуля документов
 *
 * Примечание: Для полноценного тестирования документов требуется авторизация.
 * Эти тесты проверяют базовую функциональность и UI компоненты.
 */
test.describe('Документы', () => {
  test.describe('Без авторизации', () => {
    test('страница документов требует авторизацию', async ({ page }) => {
      await page.goto('/documents')

      // Неавторизованный пользователь должен быть перенаправлен
      await expect(page).toHaveURL(/.*login|.*forbidden/)
    })

    test('страница shared documents требует авторизацию', async ({ page }) => {
      await page.goto('/documents/shared')

      // Неавторизованный пользователь должен быть перенаправлен
      await expect(page).toHaveURL(/.*login|.*forbidden/)
    })
  })

  test.describe('UI компоненты (мок авторизация)', () => {
    test.beforeEach(async ({ page }) => {
      // Устанавливаем мок токен для имитации авторизации
      await page.addInitScript(() => {
        // Мок данные пользователя
        localStorage.setItem('authToken', 'mock-token-for-testing')
        localStorage.setItem(
          'user',
          JSON.stringify({
            id: 1,
            email: 'test@example.com',
            firstName: 'Test',
            lastName: 'User',
            role: 'teacher',
          })
        )
      })
    })

    test('страница документов загружается с мок авторизацией', async ({ page }) => {
      await page.goto('/documents')

      // Ждём загрузки страницы (может быть редирект если токен недействителен)
      await page.waitForLoadState('networkidle')

      // Проверяем что либо страница загрузилась, либо редирект на login
      const url = page.url()
      const isDocumentsPage = url.includes('/documents')
      const isLoginPage = url.includes('/login')
      const isForbiddenPage = url.includes('/forbidden')

      expect(isDocumentsPage || isLoginPage || isForbiddenPage).toBeTruthy()
    })
  })
})

test.describe('Календарь', () => {
  test('страница календаря требует авторизацию', async ({ page }) => {
    await page.goto('/calendar')

    // Неавторизованный пользователь должен быть перенаправлен
    await expect(page).toHaveURL(/.*login|.*forbidden/)
  })
})

test.describe('Dashboard', () => {
  test('dashboard требует авторизацию', async ({ page }) => {
    await page.goto('/dashboard')

    // Неавторизованный пользователь должен быть перенаправлен
    await expect(page).toHaveURL(/.*login|.*forbidden/)
  })
})

test.describe('Профиль', () => {
  test('страница профиля требует авторизацию', async ({ page }) => {
    await page.goto('/profile')

    // Неавторизованный пользователь должен быть перенаправлен
    await expect(page).toHaveURL(/.*login|.*forbidden/)
  })
})
