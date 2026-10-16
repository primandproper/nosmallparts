import { expect, type Page, test } from '@playwright/test';

import { newPage, register, uniqueEmail } from './support';

/** writeScript publishes a two-line script, one line per character, and returns its title. */
async function writeScript(page: Page): Promise<string> {
  const title = `Smoke ${Date.now()}`;

  await page.goto('/scripts/new');
  await page.getByPlaceholder('Title').fill(title);
  await page.getByRole('button', { name: '+ Add character' }).click();
  await page.getByRole('button', { name: '+ Add character' }).click();

  // A new line goes to the first character; the second is moved to the other.
  await page.getByRole('button', { name: '+ Add line' }).click();
  await page.getByPlaceholder('Line text').nth(0).fill('To be, or not to be.');
  await page.getByRole('button', { name: '+ Add line' }).click();
  const secondTurn = page.getByText('Turn 2', { exact: true }).locator('xpath=../..');
  await secondTurn.getByRole('button', { name: 'Character 2' }).click();
  await secondTurn.getByRole('button', { name: 'Character 1' }).click();
  await page.getByPlaceholder('Line text').nth(1).fill('That is the question.');

  await page.getByRole('switch', { name: 'Public' }).check();
  await page.getByRole('button', { name: 'Publish' }).click();
  await expect(page.getByRole('heading', { level: 2, name: title })).toBeVisible();

  return title;
}

const nextLine = (page: Page) => page.getByRole('button', { name: 'Next line' });

test('a host, a performer and a guest put on a play', async ({ browser }) => {
  const host = await newPage(browser);
  const performer = await newPage(browser);
  const guest = await newPage(browser);

  await test.step('the host writes a script and opens a stage with it', async () => {
    await register(host, 'Hosty', uniqueEmail('host'));
    await writeScript(host);

    await host.getByRole('button', { name: 'Start a stage with this script' }).click();
    // The host has never joined a stage, so home asks for a name first.
    await expect(host.getByPlaceholder('Your name')).toHaveValue('Hosty');
    await host.getByRole('button', { name: 'Create stage' }).click();
    await expect(host.getByText('Stage code')).toBeVisible();
  });

  const code = host.url().split('/stage/')[1] ?? '';
  expect(code).toMatch(/^[A-Z0-9]{6}$/);

  await test.step('a signed-in performer joins from the invite link', async () => {
    await register(performer, 'Perry', uniqueEmail('performer'));
    await performer.goto(`/stage/${code}`);
    await performer.getByPlaceholder('Your name').fill('Perry');
    await performer.getByRole('button', { name: 'Join' }).click();
    await expect(performer.getByText('Stage code')).toBeVisible();
  });

  await test.step('a guest joins by typing the code in lower case', async () => {
    await guest.goto('/');
    await guest.getByPlaceholder('Your name').fill('Gus');
    await guest.getByPlaceholder(/^Stage code/).fill(code.toLowerCase());
    await guest.getByRole('button', { name: 'Join stage' }).click();
    await expect(guest.getByText('Stage code')).toBeVisible();
    await expect(host.getByText('Gus', { exact: true })).toBeVisible();
  });

  await test.step('chat reaches everyone, markdown and all', async () => {
    await guest.getByRole('button', { name: 'Open chat' }).click();
    await guest.getByPlaceholder('Message…').fill('break a **leg**');
    await guest.getByRole('button', { name: 'Send message' }).click();

    for (const page of [host, performer]) {
      await page.getByRole('button', { name: 'Open chat' }).click();
      await expect(page.locator('span.font-bold', { hasText: 'leg' })).toBeVisible();
      await page.getByRole('button', { name: 'Close chat' }).click();
    }
    await guest.getByRole('button', { name: 'Close chat' }).click();
  });

  await test.step('roles are claimed, everyone readies, and the play starts', async () => {
    await host.getByRole('button', { name: 'Audience' }).click();
    await performer.getByRole('button', { name: 'Character 1' }).click();
    await guest.getByRole('button', { name: 'Character 2' }).click();
    await performer.getByRole('button', { name: 'Ready up' }).click();
    await guest.getByRole('button', { name: 'Ready up' }).click();

    for (const page of [host, performer, guest]) await expect(page.getByText(/Turn 1 of 2/)).toBeVisible();
  });

  await test.step('only the speaker advances the line', async () => {
    await expect(nextLine(performer)).toBeEnabled();
    await expect(nextLine(guest)).toBeDisabled();
    await expect(nextLine(host)).toBeDisabled();
    await nextLine(performer).click();

    await expect(guest.getByText(/Turn 2 of 2/)).toBeVisible();
    await expect(nextLine(performer)).toBeDisabled();
    await expect(nextLine(guest)).toBeEnabled();
    await nextLine(guest).click();

    for (const page of [host, performer, guest]) await expect(page.getByText('Bravo!')).toBeVisible();
  });

  await test.step('the discussion takes ratings from signed-in non-authors only', async () => {
    await performer.getByRole('button', { name: '★ 5' }).click();
    await expect(performer.getByText('Thanks!')).toBeVisible();
    await expect(guest.getByRole('link', { name: 'Sign in to rate' })).toBeVisible();
    await expect(host.getByText("Authors can't rate their own work.")).toBeVisible();
  });

  await test.step('perform again returns everyone to the lobby', async () => {
    await host.getByRole('button', { name: 'Perform again' }).click();
    for (const page of [host, performer, guest]) await expect(page.getByText('Stage code')).toBeVisible();
  });

  for (const page of [host, performer, guest]) await page.context().close();
});
