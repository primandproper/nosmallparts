import { defineConfig, devices } from '@playwright/test';

// `make e2e` builds the binary, starts it on a fresh SQLite database and points
// this suite at it; see scripts/e2e.sh.
const baseURL = process.env.NSP_E2E_BASE_URL ?? 'http://localhost:3101';

export default defineConfig({
  testDir: 'e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['github'], ['list']] : 'list',
  timeout: 90_000,
  expect: { timeout: 10_000 },
  use: {
    baseURL,
    trace: 'retain-on-failure',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
});
