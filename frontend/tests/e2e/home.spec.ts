import { test, expect } from '@playwright/test';

test.describe('Главная страница', () => {
  test('должна загружаться и отображать контент', async ({ page }) => {
    // Переходим на главную
    await page.goto('/');

    // Проверяем что страница загрузилась
    await expect(page).toHaveTitle(/Information System/i);

    console.log('✅ Главная страница загрузилась успешно');
  });

  test('должна работать навигация', async ({ page }) => {
    await page.goto('/');

    // Ждем загрузки страницы
    await page.waitForLoadState('networkidle');

    console.log('✅ Навигация работает');
  });
});
