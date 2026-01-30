import { test, expect } from '@playwright/test'

test.describe('Навигация', () => {
  test('главная страница загружается', async ({ page }) => {
    await page.goto('/')

    // Проверяем что страница загрузилась (title на русском)
    await expect(page).toHaveTitle(/Секретарь-Методист|Информационная система/i)

    // Ожидаем загрузки контента
    await page.waitForLoadState('networkidle')
  })

  test('переход на страницу входа', async ({ page }) => {
    await page.goto('/')

    // Ищем ссылку на вход
    const loginLink = page.getByRole('link', { name: /войти|вход|login/i })

    if (await loginLink.isVisible()) {
      await loginLink.click()
      await expect(page).toHaveURL(/.*login/)
    } else {
      // Если ссылки нет на главной, переходим напрямую
      await page.goto('/login')
      await expect(page).toHaveURL(/.*login/)
    }
  })

  test('страница 404 для несуществующих маршрутов', async ({ page }) => {
    await page.goto('/non-existent-page-12345')

    // Ждём загрузки
    await page.waitForLoadState('networkidle')

    // Проверяем отображение страницы 404 или редирект
    const content = await page.content()
    const has404Content =
      content.includes('404') ||
      content.includes('не найден') ||
      content.includes('Not Found') ||
      content.includes('not-found')

    expect(has404Content).toBeTruthy()
  })

  test('страница forbidden загружается', async ({ page }) => {
    await page.goto('/forbidden')

    // Ждём загрузки
    await page.waitForLoadState('networkidle')

    // Проверяем что страница forbidden отображается
    const content = await page.content()
    const hasForbiddenContent =
      content.includes('доступ') ||
      content.includes('Forbidden') ||
      content.includes('forbidden') ||
      content.includes('запрещ') ||
      content.includes('403')

    expect(hasForbiddenContent).toBeTruthy()
  })
})
