import { test, expect } from '@playwright/test';

test.describe('Home Page', () => {
  test('loads the home page', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('No Small Parts')).toBeVisible();
  });

  test('autofocuses the name input', async ({ page }) => {
    await page.goto('/');
    const nameInput = page.getByPlaceholder('Enter your name...');
    await expect(nameInput).toBeFocused();
  });

  test('pre-fills stage code from URL parameter', async ({ page }) => {
    await page.goto('/?stage=TESTCODE');
    const codeInput = page.getByPlaceholder('ABC123');
    await expect(codeInput).toHaveValue('TESTCODE');
  });

  test('pre-fills stage code from code URL parameter', async ({ page }) => {
    await page.goto('/?code=MYCODE123');
    const codeInput = page.getByPlaceholder('ABC123');
    await expect(codeInput).toHaveValue('MYCODE123');
  });

  test('shows Join Stage button when code is entered', async ({ page }) => {
    await page.goto('/');
    const codeInput = page.getByPlaceholder('ABC123');
    await codeInput.fill('ABCDEF');
    await expect(page.getByRole('button', { name: /join stage/i })).toBeVisible();
  });

  test('shows Start Stage button when no code is entered', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('button', { name: /start stage/i })).toBeVisible();
  });

  test('validates stage code minimum length', async ({ page }) => {
    await page.goto('/');
    const codeInput = page.getByPlaceholder('ABC123');
    await codeInput.fill('ABC');
    await expect(page.getByText('Stage code must be at least 6 characters')).toBeVisible();
  });

  test('extracts stage code from pasted URL', async ({ page }) => {
    await page.goto('/');
    const codeInput = page.getByPlaceholder('ABC123');
    await codeInput.fill('http://localhost:3000/?stage=EXTRACTED');
    await expect(codeInput).toHaveValue('EXTRACTED');
  });
});
