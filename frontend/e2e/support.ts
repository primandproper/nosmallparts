import { expect, type Browser, type Page } from '@playwright/test';

let counter = 0;

/** uniqueEmail is an address no other test, in this run or a retry, has used. */
export function uniqueEmail(prefix: string): string {
  counter += 1;
  return `${prefix}-${Date.now()}-${process.pid}-${counter}@example.com`;
}

export const PASSWORD = 'correct horse battery staple';

/** register creates an account from the register page and waits to be signed in. */
export async function register(page: Page, name: string, email: string): Promise<void> {
  await page.goto('/register');
  await page.getByPlaceholder('Name').fill(name);
  await page.getByPlaceholder('Email').fill(email);
  await page.getByPlaceholder(/^Password/).fill(PASSWORD);
  await page.getByRole('button', { name: 'Create account' }).click();
  await expectSignedIn(page, email);
}

/** expectSignedIn waits for the home page to name who is signed in. */
export async function expectSignedIn(page: Page, email: string): Promise<void> {
  await expect(page.getByText(`Signed in as ${email}`)).toBeVisible();
}

/** newPage opens a page in a context of its own: another person, another browser. */
export async function newPage(browser: Browser): Promise<Page> {
  const context = await browser.newContext({ permissions: ['clipboard-read', 'clipboard-write'] });
  return context.newPage();
}
