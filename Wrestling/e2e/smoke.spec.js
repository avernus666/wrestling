const { test, expect } = require('@playwright/test');

test('home page exposes Wrestling navigation and three sections', async ({ page }) => {
  await page.goto('/');
  await expect(page.getByText(/WRESTLING/i).first()).toBeVisible();
  await expect(page.getByRole('button', { name: /стойка/i }).first()).toBeVisible();
  await expect(page.getByRole('button', { name: /партер/i }).first()).toBeVisible();
  await expect(page.getByRole('button', { name: /ОФП/i }).first()).toBeVisible();
});

test('catalog search opens technique details', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('button', { name: /стойка/i }).first().click();
  await expect(page.getByText(/Каталог/i)).toBeVisible();
  await page.getByPlaceholder(/Поиск техники/i).fill('single');
  await expect(page.getByText(/Low single/i).first()).toBeVisible();
  await page.getByRole('button', { name: /Разбор/i }).first().click();
  await expect(page.getByText(/Обучение/i)).toBeVisible();
});

test('mobile navigation remains usable', async ({ page }) => {
  await page.goto('/');
  const menu = page.getByRole('button', { name: /Открыть меню/i });
  if (await menu.count()) {
    await menu.click();
    await expect(page.getByText(/Тренировки/i).last()).toBeVisible();
  }
});

test('health endpoint is reachable when backend is running', async ({ request }) => {
  const response = await request.get('http://127.0.0.1:8080/api/health');
  expect(response.ok()).toBeTruthy();
  expect(await response.json()).toMatchObject({ status: 'ok' });
});
