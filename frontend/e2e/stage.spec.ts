import { test, expect } from '@playwright/test';

test.describe('Stage Page', () => {
  test.beforeEach(async ({ page }) => {
    // Set player name in session storage before navigating
    await page.goto('/');
    await page.evaluate(() => {
      sessionStorage.setItem('playerName', 'Test Player');
    });
  });

  test('redirects to home if no player name set', async ({ page }) => {
    // Clear session storage and navigate directly to a stage
    await page.evaluate(() => {
      sessionStorage.removeItem('playerName');
    });
    await page.goto('/stage/TESTSTAGE');
    // Should redirect to home
    await expect(page).toHaveURL('/');
  });

  test('displays stage code in header', async ({ page }) => {
    await page.goto('/stage/MYSTAGE');
    await expect(page.getByText('MYSTAGE')).toBeVisible();
  });

  test('has copy invite link button', async ({ page }) => {
    await page.goto('/stage/TESTCODE');
    await expect(page.getByRole('button', { name: /copy invite link/i })).toBeVisible();
  });
});
