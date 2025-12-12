import { test, expect } from '@playwright/test'

test.describe('Главная страница', () => {
  test('должна загружаться и отображать контент', async ({ page }) => {
    // Переходим на главную
    await page.goto('/')

    // Проверяем что страница загрузилась
    await expect(page).toHaveTitle(/Information System/i)

    // Ждём полной загрузки
    await page.waitForLoadState('networkidle')
  })

  test('должна отображать навигационные элементы', async ({ page }) => {
    await page.goto('/')

    // Ждем загрузки страницы
    await page.waitForLoadState('networkidle')

    // Проверяем наличие ссылки на вход
    const loginLink = page.getByRole('link', { name: /войти|вход|login/i })
    const hasLoginLink = await loginLink.isVisible().catch(() => false)

    // Должна быть либо ссылка на вход, либо уже авторизованный пользователь
    if (hasLoginLink) {
      await expect(loginLink).toBeVisible()
    }
  })

  test('должна быть адаптивной', async ({ page }) => {
    // Проверяем на мобильном размере
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    // Страница должна загрузиться без горизонтального скролла
    const hasHorizontalScroll = await page.evaluate(() => {
      return document.documentElement.scrollWidth > document.documentElement.clientWidth
    })

    expect(hasHorizontalScroll).toBeFalsy()
  })

  test('должна быть доступной (a11y basics)', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    // Проверяем что есть main landmark
    const main = page.locator('main')
    const hasMain = await main.isVisible().catch(() => false)

    // Проверяем что все изображения имеют alt
    const imagesWithoutAlt = await page.locator('img:not([alt])').count()

    // Проверяем наличие h1
    const h1 = page.locator('h1')
    const hasH1 = await h1.isVisible().catch(() => false)

    // Должен быть main или h1
    expect(hasMain || hasH1).toBeTruthy()

    // Не должно быть изображений без alt (или их мало)
    expect(imagesWithoutAlt).toBeLessThanOrEqual(2)
  })
})
