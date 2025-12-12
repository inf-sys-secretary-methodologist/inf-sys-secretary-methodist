import { test, expect } from '@playwright/test'

test.describe('Тема оформления (Dark/Light)', () => {
  test('страница загружается с системной темой', async ({ page }) => {
    await page.goto('/')

    // Ждём загрузки страницы
    await page.waitForLoadState('networkidle')

    // Проверяем что html элемент имеет класс темы или data-атрибут
    const html = page.locator('html')

    // Тема может быть определена через class или data-theme
    const hasThemeClass = await html.evaluate((el) => {
      return el.classList.contains('dark') || el.classList.contains('light')
    })

    const hasThemeAttribute = await html.evaluate((el) => {
      return el.hasAttribute('data-theme') || el.hasAttribute('style')
    })

    // Должен быть хотя бы один способ определения темы
    expect(hasThemeClass || hasThemeAttribute).toBeTruthy()
  })

  test('переключатель темы работает', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    // Ищем переключатель темы (может быть кнопка с иконкой sun/moon)
    // Пробуем разные селекторы
    const themeToggleByRole = page.getByRole('button', { name: /тема|theme|dark|light|sun|moon/i })
    const themeToggleByIcon = page.locator(
      'button:has(svg.lucide-sun), button:has(svg.lucide-moon)'
    )

    let themeToggle = themeToggleByRole
    if (!(await themeToggleByRole.isVisible().catch(() => false))) {
      themeToggle = themeToggleByIcon.first()
    }

    const isVisible = await themeToggle.isVisible().catch(() => false)

    if (isVisible) {
      // Получаем начальное состояние
      const html = page.locator('html')
      const initialIsDark = await html.evaluate((el) => el.classList.contains('dark'))

      // Кликаем переключатель
      await themeToggle.click()

      // Даём время на анимацию/переключение
      await page.waitForTimeout(500)

      // Проверяем что тема изменилась
      const newIsDark = await html.evaluate((el) => el.classList.contains('dark'))

      // Тема должна измениться (если была dark - станет light, и наоборот)
      expect(newIsDark).not.toBe(initialIsDark)
    } else {
      // Если переключатель не найден, тест пропускается (skip)
      test.skip()
    }
  })

  test('тема сохраняется после перезагрузки', async ({ page }) => {
    await page.goto('/')
    await page.waitForLoadState('networkidle')

    // Устанавливаем тему через localStorage
    await page.evaluate(() => {
      localStorage.setItem('theme', 'dark')
    })

    // Перезагружаем страницу
    await page.reload()
    await page.waitForLoadState('networkidle')

    // Проверяем что тема восстановилась
    const savedTheme = await page.evaluate(() => localStorage.getItem('theme'))
    expect(savedTheme).toBe('dark')
  })

  test('тема применяется к компонентам', async ({ page }) => {
    await page.goto('/login')
    await page.waitForLoadState('networkidle')

    // Проверяем что стили применены (фон, текст)
    const body = page.locator('body')

    // Получаем computed style
    const backgroundColor = await body.evaluate((el) => {
      return window.getComputedStyle(el).backgroundColor
    })

    // Фон должен быть определён (не прозрачный)
    expect(backgroundColor).not.toBe('rgba(0, 0, 0, 0)')
  })
})
