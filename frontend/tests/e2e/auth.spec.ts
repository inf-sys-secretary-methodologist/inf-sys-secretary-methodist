import { test, expect } from '@playwright/test'

test.describe('Аутентификация', () => {
  test.beforeEach(async ({ page }) => {
    // Очистка cookies перед каждым тестом
    await page.context().clearCookies()
  })

  test('должна отображаться страница входа', async ({ page }) => {
    await page.goto('/login')

    // Проверяем заголовок
    await expect(page.getByRole('heading', { name: /добро пожаловать/i })).toBeVisible()

    // Проверяем наличие полей формы
    await expect(page.getByLabel(/email/i)).toBeVisible()
    await expect(page.getByLabel(/пароль/i)).toBeVisible()

    // Проверяем кнопку входа
    await expect(page.getByRole('button', { name: /войти/i })).toBeVisible()

    // Проверяем ссылку на регистрацию
    await expect(page.getByRole('link', { name: /зарегистрироваться/i })).toBeVisible()
  })

  test('должна показывать ошибку при пустых полях', async ({ page }) => {
    await page.goto('/login')

    // Кликаем на поле email и уходим (blur) для валидации
    await page.getByLabel(/email/i).click()
    await page.getByLabel(/пароль/i).click()
    await page.getByRole('button', { name: /войти/i }).click()

    // Ожидаем появления ошибок валидации (конкретный текст ошибки)
    await expect(page.getByText('Email обязателен')).toBeVisible()
  })

  test('должна показывать ошибку при неверных данных', async ({ page }) => {
    await page.goto('/login')

    // Вводим неверные данные
    await page.getByLabel(/email/i).fill('wrong@email.com')
    await page.getByLabel(/пароль/i).fill('wrongpassword')

    // Отправляем форму
    await page.getByRole('button', { name: /войти/i }).click()

    // Ожидаем ошибку (inline сообщение в форме)
    await expect(
      page.locator('.bg-red-50, .bg-red-900\\/20').getByText(/неверный email или пароль/i)
    ).toBeVisible({ timeout: 10000 })
  })

  test('должна отображаться страница регистрации', async ({ page }) => {
    await page.goto('/register')

    // Проверяем заголовок
    await expect(page.getByRole('heading', { name: /регистрация|создать аккаунт/i })).toBeVisible()

    // Проверяем наличие полей формы
    await expect(page.getByLabel(/email/i)).toBeVisible()
    await expect(page.getByLabel(/пароль/i).first()).toBeVisible()

    // Проверяем ссылку на вход
    await expect(page.getByRole('link', { name: /войти/i })).toBeVisible()
  })

  test('навигация между login и register', async ({ page }) => {
    // Переходим на страницу входа
    await page.goto('/login')

    // Кликаем на ссылку регистрации
    await page.getByRole('link', { name: /зарегистрироваться/i }).click()

    // Проверяем что перешли на страницу регистрации
    await expect(page).toHaveURL(/.*register/)

    // Кликаем на ссылку входа
    await page.getByRole('link', { name: /войти/i }).click()

    // Проверяем что вернулись на страницу входа
    await expect(page).toHaveURL(/.*login/)
  })

  test('должна перенаправлять неавторизованного пользователя', async ({ page }) => {
    // Пытаемся перейти на защищённую страницу
    await page.goto('/dashboard')

    // Должны быть перенаправлены на login или forbidden
    await expect(page).toHaveURL(/.*login|.*forbidden/)
  })
})
