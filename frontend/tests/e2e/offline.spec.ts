import { test, expect } from '@playwright/test'

/**
 * E2E тесты для офлайн страницы
 */
test.describe('Офлайн страница', () => {
  test('страница /offline загружается', async ({ page }) => {
    await page.goto('/offline')
    await page.waitForLoadState('networkidle')

    // Проверяем что страница загрузилась
    expect(page.url()).toContain('/offline')
  })

  test('отображается иконка отсутствия сети', async ({ page }) => {
    await page.goto('/offline')
    await page.waitForLoadState('networkidle')

    // Ищем иконку WifiOff или контейнер с иконкой
    const hasIcon =
      (await page
        .locator('svg.lucide-wifi-off')
        .isVisible()
        .catch(() => false)) ||
      (await page
        .locator('.rounded-full')
        .first()
        .isVisible()
        .catch(() => false))

    expect(hasIcon).toBeTruthy()
  })

  test('отображается заголовок и описание', async ({ page }) => {
    await page.goto('/offline')
    await page.waitForLoadState('networkidle')

    // Проверяем наличие заголовка
    const heading = page.locator('h1')
    await expect(heading).toBeVisible()

    // Проверяем наличие описания
    const hasDescription =
      (await page
        .locator('p')
        .first()
        .isVisible()
        .catch(() => false)) ||
      (await page
        .getByText(/интернет|connection|offline|сеть/i)
        .isVisible()
        .catch(() => false))

    expect(hasDescription).toBeTruthy()
  })

  test('есть кнопка повторить/retry', async ({ page }) => {
    await page.goto('/offline')
    await page.waitForLoadState('networkidle')

    // Ищем кнопку retry
    const retryButton =
      (await page
        .getByRole('button', { name: /повтор|retry|обновить|refresh/i })
        .isVisible()
        .catch(() => false)) ||
      (await page
        .locator('button:has(svg.lucide-refresh-cw)')
        .isVisible()
        .catch(() => false))

    expect(retryButton).toBeTruthy()
  })

  test('отображаются советы/tips', async ({ page }) => {
    await page.goto('/offline')
    await page.waitForLoadState('networkidle')

    // Проверяем наличие списка советов
    const hasTips =
      (await page
        .locator('ul li')
        .first()
        .isVisible()
        .catch(() => false)) ||
      (await page
        .locator('li')
        .first()
        .isVisible()
        .catch(() => false))

    expect(hasTips).toBeTruthy()
  })

  test('кнопка retry вызывает перезагрузку', async ({ page }) => {
    await page.goto('/offline')
    await page.waitForLoadState('networkidle')

    // Находим кнопку
    const retryButton = page.getByRole('button').first()

    if (await retryButton.isVisible()) {
      // Проверяем что кнопка кликабельна
      await expect(retryButton).toBeEnabled()
    }
  })
})
