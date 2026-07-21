import { beforeAll, describe, expect, it } from 'vitest';
import { app } from '../src/app.js';
import {
  extractCookies,
  jsonInit,
  SAMPLE_SCRIPT_INPUT,
  signUpUser,
  type TestUser,
} from './helpers.js';

let owner: TestUser;
let rater: TestUser;
let admin: TestUser;

beforeAll(async () => {
  owner = await signUpUser(app);
  rater = await signUpUser(app);
  admin = await signUpUser(app, { email: 'admin@example.com' });
});

async function createScript(cookie: string, overrides: Record<string, unknown> = {}) {
  const res = await app.request(
    '/api/play-scripts',
    jsonInit({ ...SAMPLE_SCRIPT_INPUT, ...overrides }, cookie)
  );
  expect(res.status).toBe(201);
  const { play } = (await res.json()) as { play: { id: string; enabled?: boolean } };
  return play;
}

describe('identity (§1.1)', () => {
  it('mints a guest cookie for a fresh visitor', async () => {
    const res = await app.request('/api/identity');
    expect(res.status).toBe(200);
    const body = (await res.json()) as { userId: string | null; guestId: string | null };
    expect(body.userId).toBeNull();
    expect(body.guestId).toMatch(/^[0-9A-HJKMNP-TV-Z]{26}$/);
    expect(extractCookies(res)).toContain('nsp_guest_id=');
  });

  it('keeps the same guest id across requests with the cookie', async () => {
    const first = await app.request('/api/identity');
    const cookie = extractCookies(first);
    const { guestId } = (await first.json()) as { guestId: string };
    const second = await app.request('/api/identity', { headers: { Cookie: cookie } });
    const body = (await second.json()) as { guestId: string };
    expect(body.guestId).toBe(guestId);
  });

  it('treats a tampered non-ULID cookie as no identity and mints a new one', async () => {
    const res = await app.request('/api/identity', {
      headers: { Cookie: 'nsp_guest_id=not-a-ulid' },
    });
    const body = (await res.json()) as { guestId: string };
    expect(body.guestId).toMatch(/^[0-9A-HJKMNP-TV-Z]{26}$/);
    expect(body.guestId).not.toBe('not-a-ulid');
  });

  it('reports userId for a signed-in user', async () => {
    const res = await app.request('/api/identity', { headers: { Cookie: owner.cookie } });
    const body = (await res.json()) as { userId: string | null };
    expect(body.userId).toBe(owner.id);
  });
});

describe('auth (§1.2)', () => {
  it('register with an existing email fails without creating a partial account', async () => {
    const res = await app.request('/api/auth/sign-up/email', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Origin: 'http://localhost:8081' },
      body: JSON.stringify({ email: owner.email, password: 'another-password-1', name: 'Dup' }),
    });
    expect(res.status).toBeGreaterThanOrEqual(400);
  });

  it('register with a weak password fails validation', async () => {
    const res = await app.request('/api/auth/sign-up/email', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Origin: 'http://localhost:8081' },
      body: JSON.stringify({ email: 'weak@test.example', password: 'short', name: 'Weak' }),
    });
    expect(res.status).toBeGreaterThanOrEqual(400);
  });

  it('login with wrong creds yields no session', async () => {
    const res = await app.request('/api/auth/sign-in/email', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Origin: 'http://localhost:8081' },
      body: JSON.stringify({ email: owner.email, password: 'wrong-password-1' }),
    });
    expect(res.status).toBeGreaterThanOrEqual(400);
  });

  it('login with correct creds sets a session cookie', async () => {
    const res = await app.request('/api/auth/sign-in/email', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Origin: 'http://localhost:8081' },
      body: JSON.stringify({ email: owner.email, password: owner.password }),
    });
    expect(res.status).toBe(200);
    const cookie = extractCookies(res);
    const me = await app.request('/api/me', { headers: { Cookie: cookie } });
    const body = (await me.json()) as { user: { id: string } | null };
    expect(body.user?.id).toBe(owner.id);
  });

  it('sign out clears the session; protected routes 401', async () => {
    const u = await signUpUser(app);
    const out = await app.request('/api/auth/sign-out', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Cookie: u.cookie,
        Origin: 'http://localhost:8081',
      },
      body: '{}',
    });
    expect(out.status).toBe(200);
    const clearedCookie = extractCookies(out);
    const res = await app.request(
      '/api/play-scripts',
      jsonInit(SAMPLE_SCRIPT_INPUT, clearedCookie)
    );
    expect(res.status).toBe(401);
  });
});

describe('auth gates (§1.5)', () => {
  it('POST /api/play-scripts requires user', async () => {
    const res = await app.request('/api/play-scripts', jsonInit(SAMPLE_SCRIPT_INPUT));
    expect(res.status).toBe(401);
  });

  it('POST /api/stages requires user (guest gets 401)', async () => {
    const res = await app.request('/api/stages', jsonInit({}));
    expect(res.status).toBe(401);
  });

  it('POST /api/play-scripts/:id/rate requires user', async () => {
    const play = await createScript(owner.cookie);
    const res = await app.request(`/api/play-scripts/${play.id}/rate`, jsonInit({ rating: 5 }));
    expect(res.status).toBe(401);
  });

  it('PUT of a script you do not own returns 404', async () => {
    const play = await createScript(owner.cookie);
    const res = await app.request(`/api/play-scripts/${play.id}`, {
      ...jsonInit({ ...SAMPLE_SCRIPT_INPUT, title: 'Hijacked' }, rater.cookie),
      method: 'PUT',
    });
    expect(res.status).toBe(404);
  });
});

describe('scripts CRUD (§2.2)', () => {
  it('creates a script with cast, lines and tags; defaults to enabled', async () => {
    const play = await createScript(owner.cookie, { tagNames: ['Comedy', 'Short'] });
    expect(play.enabled).toBe(true);
    const res = await app.request(`/api/play-scripts/${play.id}`, {
      headers: { Cookie: owner.cookie },
    });
    const body = (await res.json()) as {
      play: { cast: unknown[]; lines: unknown[]; tags: { name: string }[] };
    };
    expect(body.play.cast).toHaveLength(2);
    expect(body.play.lines).toHaveLength(3);
    expect(body.play.tags.map((t) => t.name).sort()).toEqual(['Comedy', 'Short']);
  });

  it('edit own script persists title/cast/lines and preserves ordering', async () => {
    const play = await createScript(owner.cookie);
    const res = await app.request(`/api/play-scripts/${play.id}`, {
      ...jsonInit(
        {
          ...SAMPLE_SCRIPT_INPUT,
          title: 'Renamed',
          lines: [
            { castMemberIds: ['b'], text: 'New first', index: 0 },
            { castMemberIds: ['a'], text: 'New second', index: 1 },
          ],
        },
        owner.cookie
      ),
      method: 'PUT',
    });
    expect(res.status).toBe(200);
    const { play: updated } = (await res.json()) as {
      play: { title: string; lines: { text: string; index: number }[] };
    };
    expect(updated.title).toBe('Renamed');
    expect(updated.lines.map((l) => l.text)).toEqual(['New first', 'New second']);
    expect(updated.lines.map((l) => l.index)).toEqual([0, 1]);
  });

  it('anonymous visitors see only public enabled scripts', async () => {
    const pub = await createScript(owner.cookie, { title: 'PublicOne', isPublic: true });
    const priv = await createScript(owner.cookie, { title: 'PrivateOne', isPublic: false });
    const res = await app.request('/api/play-scripts');
    const { plays } = (await res.json()) as { plays: { id: string }[] };
    const ids = plays.map((p) => p.id);
    expect(ids).toContain(pub.id);
    expect(ids).not.toContain(priv.id);
  });

  it('?search= filters case-insensitively across title/description/author', async () => {
    await createScript(owner.cookie, { title: 'ZebraQuest Saga' });
    const res = await app.request('/api/play-scripts?search=zebraquest', {
      headers: { Cookie: owner.cookie },
    });
    const { plays } = (await res.json()) as { plays: { title: string }[] };
    expect(plays.length).toBeGreaterThanOrEqual(1);
    expect(plays.every((p) => JSON.stringify(p).toLowerCase().includes('zebraquest'))).toBe(true);
  });
});

describe('ratings (§2.3)', () => {
  it('rates 1–5, averages, re-rates without double counting, blocks self-rating', async () => {
    const play = await createScript(owner.cookie);

    // Owner cannot rate own script.
    const self = await app.request(
      `/api/play-scripts/${play.id}/rate`,
      jsonInit({ rating: 5 }, owner.cookie)
    );
    expect(self.status).toBe(403);

    // Someone else rates.
    const r1 = await app.request(
      `/api/play-scripts/${play.id}/rate`,
      jsonInit({ rating: 4 }, rater.cookie)
    );
    expect(r1.status).toBe(200);
    const stats1 = (await r1.json()) as { averageRating: number; ratingCount: number };
    expect(stats1.ratingCount).toBe(1);
    expect(stats1.averageRating).toBe(4);

    // Re-rating replaces, does not double count.
    const r2 = await app.request(
      `/api/play-scripts/${play.id}/rate`,
      jsonInit({ rating: 2 }, rater.cookie)
    );
    const stats2 = (await r2.json()) as { averageRating: number; ratingCount: number };
    expect(stats2.ratingCount).toBe(1);
    expect(stats2.averageRating).toBe(2);

    // userRating reflected in the play listing for the rater.
    const list = await app.request('/api/play-scripts', { headers: { Cookie: rater.cookie } });
    const { plays } = (await list.json()) as { plays: { id: string; userRating?: number }[] };
    expect(plays.find((p) => p.id === play.id)?.userRating).toBe(2);
  });

  it('rejects out-of-range ratings with 400', async () => {
    const play = await createScript(owner.cookie);
    for (const rating of [0, 6, 3.5]) {
      const res = await app.request(
        `/api/play-scripts/${play.id}/rate`,
        jsonInit({ rating }, rater.cookie)
      );
      expect(res.status).toBe(400);
    }
  });
});

describe('admin + enabled flag (§2.4)', () => {
  it('non-admin DELETE (even the owner) gets 403; admin DELETE succeeds', async () => {
    const play = await createScript(owner.cookie);
    const asOwner = await app.request(`/api/play-scripts/${play.id}`, {
      method: 'DELETE',
      headers: { Cookie: owner.cookie },
    });
    expect(asOwner.status).toBe(403);
    const asAdmin = await app.request(`/api/play-scripts/${play.id}`, {
      method: 'DELETE',
      headers: { Cookie: admin.cookie },
    });
    expect(asAdmin.status).toBe(200);
    const gone = await app.request(`/api/play-scripts/${play.id}`, {
      headers: { Cookie: admin.cookie },
    });
    expect(gone.status).toBe(404);
  });

  it('disable hides the script from non-admins (owner included); re-enable restores', async () => {
    const play = await createScript(owner.cookie, { title: 'ToDisable', isPublic: true });

    const patch = await app.request(`/api/play-scripts/${play.id}/enabled`, {
      ...jsonInit({ enabled: false }, admin.cookie),
      method: 'PATCH',
    });
    expect(patch.status).toBe(200);

    // Hidden from anonymous list and from the owner's own list.
    for (const cookie of [undefined, owner.cookie]) {
      const res = await app.request('/api/play-scripts', {
        headers: cookie ? { Cookie: cookie } : {},
      });
      const { plays } = (await res.json()) as { plays: { id: string }[] };
      expect(plays.map((p) => p.id)).not.toContain(play.id);
    }

    // Admin still sees it, flagged disabled.
    const adminList = await app.request('/api/play-scripts', {
      headers: { Cookie: admin.cookie },
    });
    const adminPlays = (await adminList.json()) as { plays: { id: string; enabled?: boolean }[] };
    const found = adminPlays.plays.find((p) => p.id === play.id);
    expect(found?.enabled).toBe(false);

    // Direct GET is a 404 for non-admins.
    const direct = await app.request(`/api/play-scripts/${play.id}`, {
      headers: { Cookie: owner.cookie },
    });
    expect(direct.status).toBe(404);

    // New stages can't be created against it.
    const stage = await app.request(
      '/api/stages',
      jsonInit({ playScriptId: play.id }, owner.cookie)
    );
    expect(stage.status).toBe(400);

    // Re-enable restores listing.
    await app.request(`/api/play-scripts/${play.id}/enabled`, {
      ...jsonInit({ enabled: true }, admin.cookie),
      method: 'PATCH',
    });
    const restored = await app.request('/api/play-scripts');
    const restoredPlays = (await restored.json()) as { plays: { id: string }[] };
    expect(restoredPlays.plays.map((p) => p.id)).toContain(play.id);
  });

  it('PATCH enabled is admin-only, case-insensitive on ADMIN_EMAILS', async () => {
    const play = await createScript(owner.cookie);
    const nonAdmin = await app.request(`/api/play-scripts/${play.id}/enabled`, {
      ...jsonInit({ enabled: false }, owner.cookie),
      method: 'PATCH',
    });
    expect(nonAdmin.status).toBe(403);

    // 'Boss@Example.COM' in env matches boss@example.com sign-up.
    const boss = await signUpUser(app, { email: 'boss@example.com' });
    const asBoss = await app.request(`/api/play-scripts/${play.id}/enabled`, {
      ...jsonInit({ enabled: true }, boss.cookie),
      method: 'PATCH',
    });
    expect(asBoss.status).toBe(200);
  });
});

describe('tags (§2.5)', () => {
  it('lists tags and merges duplicates case-insensitively', async () => {
    await createScript(owner.cookie, { tagNames: ['Drama'] });
    await createScript(owner.cookie, { tagNames: [' drama ', 'DRAMA'] });
    const res = await app.request('/api/tags');
    const { tags } = (await res.json()) as { tags: { name: string }[] };
    const dramaTags = tags.filter((t) => t.name.toLowerCase() === 'drama');
    expect(dramaTags).toHaveLength(1);
  });
});

describe('stage creation (§3.1)', () => {
  it('creates a stage for a signed-in user with an unambiguous code', async () => {
    const res = await app.request('/api/stages', jsonInit({}, owner.cookie));
    expect(res.status).toBe(201);
    const stage = (await res.json()) as { id: string; code: string };
    expect(stage.code).toMatch(/^[A-HJ-KM-NP-Z2-9]{6}$/);
    expect(stage.code).not.toMatch(/[01OIL]/);

    // Case-insensitive lookup.
    const lower = await app.request(`/api/stages/${stage.code.toLowerCase()}/exists`);
    const body = (await lower.json()) as { exists: boolean };
    expect(body.exists).toBe(true);
  });

  it('nonexistent code reports exists:false', async () => {
    const res = await app.request('/api/stages/ZZZZ99/exists');
    const body = (await res.json()) as { exists: boolean };
    expect(body.exists).toBe(false);
  });
});
