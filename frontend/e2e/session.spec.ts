import { expect, test } from '@playwright/test';

import { expectSignedIn, register, uniqueEmail } from './support';

// The client refreshes an access token within 30 seconds of its expiry, and the
// server issues them for 15 minutes; past that, every call refreshes first.
const PAST_ACCESS_TOKEN_EXPIRY = '16:00';

test('a reload stays signed in', async ({ page }) => {
  const email = uniqueEmail('reload');
  await register(page, 'Reloader', email);

  await page.reload();
  await expectSignedIn(page, email);
});

test('two tabs stay signed in through a refresh', async ({ page, context }) => {
  const email = uniqueEmail('tabs');
  await page.clock.install();
  await register(page, 'Two Tabs', email);

  const second = await context.newPage();
  await second.clock.install();
  await second.goto('/');
  await expectSignedIn(second, email);

  // Both tabs now hold an expired access token and share one refresh cookie.
  // Presenting it twice at once would be reuse, which ends the login, so the
  // two refreshes have to take turns.
  await page.clock.fastForward(PAST_ACCESS_TOKEN_EXPIRY);
  await second.clock.fastForward(PAST_ACCESS_TOKEN_EXPIRY);
  const refreshes = [page, second].map((p) => p.waitForRequest((req) => req.url().endsWith('/ExchangeRefreshToken')));
  await Promise.all([
    page.getByRole('link', { name: 'Browse scripts' }).click(),
    second.getByRole('link', { name: 'Browse scripts' }).click(),
  ]);
  await Promise.all(refreshes);
  await expect(page.getByPlaceholder(/^Search/)).toBeVisible();
  await expect(second.getByPlaceholder(/^Search/)).toBeVisible();

  await page.goto('/');
  await expectSignedIn(page, email);
  await second.goto('/');
  await expectSignedIn(second, email);
});

test('a reload while offline offers a retry, and recovers', async ({ page }) => {
  const email = uniqueEmail('offline');
  await register(page, 'Offline', email);

  // The page itself loads; the server behind it does not answer. A browser
  // fully offline could not reload the page at all.
  await page.route('**/grpc/**', (route) => route.abort('internetdisconnected'));
  await page.reload();

  const retry = page.getByRole('button', { name: /Couldn't reach the server/ });
  await expect(retry).toBeVisible();
  await expect(page.getByRole('link', { name: 'Sign in' })).toHaveCount(0);

  await page.unroute('**/grpc/**');
  await retry.click();
  await expectSignedIn(page, email);
});
