import { Hono } from 'hono';
import { z } from 'zod';
import { isAdminEmail } from '../admin.js';
import type { AppVariables } from '../app.js';
import { requireAuth } from '../middleware/require-auth.js';
import {
  adminDeletePlayScript,
  createPlayScript,
  getPlayScript,
  getPlayScriptOwner,
  getUserRatings,
  listPlayScripts,
  ratePlayScript,
  setPlayScriptEnabled,
  updatePlayScript,
} from '../repos/play-scripts.js';
import { listTags, upsertTagByName } from '../repos/tags.js';

const castMemberSchema = z.object({
  id: z.string().optional(),
  name: z.string().min(1),
  description: z.string().nullish(),
});

const lineSchema = z.object({
  castMemberIds: z.array(z.string().min(1)).min(1),
  text: z.string(),
  index: z.number().int().nonnegative(),
  direction: z.string().nullish(),
  scrollMode: z.enum(['auto', 'manual']).nullish(),
  autoScrollWpm: z.number().int().positive().nullish(),
});

const playScriptSchema = z.object({
  title: z.string().min(1),
  description: z.string().optional(),
  author: z.string().nullish(),
  isPublic: z.boolean().optional(),
  cast: z.array(castMemberSchema),
  lines: z.array(lineSchema),
  tagIds: z.array(z.string()).optional(),
  tagNames: z.array(z.string()).optional(),
});

export const playScriptsRouter = new Hono<{ Variables: AppVariables }>();

// List scripts (public + own; admins see everything including disabled).
playScriptsRouter.get('/', async (c) => {
  const search = c.req.query('search') || undefined;
  const user = c.get('user');
  const plays = await listPlayScripts({
    userId: user?.id ?? null,
    search,
    admin: isAdminEmail(user?.email),
  });
  if (user) {
    const ratings = await getUserRatings(
      user.id,
      plays.map((p) => p.id)
    );
    for (const p of plays) {
      const r = ratings.get(p.id);
      if (r != null) p.userRating = r;
    }
  }
  return c.json({ plays });
});

// Get one script by id. Disabled scripts are invisible to non-admins (§2.4).
playScriptsRouter.get('/:id', async (c) => {
  const id = c.req.param('id');
  const play = await getPlayScript(id);
  if (!play) return c.json({ error: 'not_found' }, 404);
  const user = c.get('user');
  if (play.enabled === false && !isAdminEmail(user?.email)) {
    return c.json({ error: 'not_found' }, 404);
  }
  if (user) {
    const ratings = await getUserRatings(user.id, [id]);
    const r = ratings.get(id);
    if (r != null) play.userRating = r;
  }
  return c.json({ play });
});

// Create.
playScriptsRouter.post('/', requireAuth, async (c) => {
  const user = c.get('user')!;
  const body = await c.req.json();
  const parsed = playScriptSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: 'invalid', details: parsed.error.flatten() }, 400);
  }
  const { tagNames, tagIds, ...rest } = parsed.data;
  const resolvedTagIds = [...(tagIds ?? [])];
  if (tagNames) {
    for (const name of tagNames) {
      if (name.trim()) resolvedTagIds.push(await upsertTagByName(name));
    }
  }
  const play = await createPlayScript(user.id, { ...rest, tagIds: [...new Set(resolvedTagIds)] });
  return c.json({ play }, 201);
});

// Update.
playScriptsRouter.put('/:id', requireAuth, async (c) => {
  const user = c.get('user')!;
  const id = c.req.param('id');
  const body = await c.req.json();
  const parsed = playScriptSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: 'invalid', details: parsed.error.flatten() }, 400);
  }
  const { tagNames, tagIds, ...rest } = parsed.data;
  const resolvedTagIds = [...(tagIds ?? [])];
  if (tagNames) {
    for (const name of tagNames) {
      if (name.trim()) resolvedTagIds.push(await upsertTagByName(name));
    }
  }
  const play = await updatePlayScript(user.id, id, { ...rest, tagIds: [...new Set(resolvedTagIds)] });
  if (!play) return c.json({ error: 'not_found' }, 404);
  return c.json({ play });
});

// Delete — admin-only (§2.4). Even the script's owner gets 403.
playScriptsRouter.delete('/:id', requireAuth, async (c) => {
  const user = c.get('user')!;
  if (!isAdminEmail(user.email)) return c.json({ error: 'forbidden' }, 403);
  const id = c.req.param('id');
  const ok = await adminDeletePlayScript(id);
  if (!ok) return c.json({ error: 'not_found' }, 404);
  return c.json({ ok: true });
});

// Admin toggle for the enabled flag (§2.4).
const enabledSchema = z.object({ enabled: z.boolean() });
playScriptsRouter.patch('/:id/enabled', requireAuth, async (c) => {
  const user = c.get('user')!;
  if (!isAdminEmail(user.email)) return c.json({ error: 'forbidden' }, 403);
  const id = c.req.param('id');
  const body = await c.req.json().catch(() => ({}));
  const parsed = enabledSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: 'invalid' }, 400);
  const ok = await setPlayScriptEnabled(id, parsed.data.enabled);
  if (!ok) return c.json({ error: 'not_found' }, 404);
  return c.json({ ok: true, enabled: parsed.data.enabled });
});

// Rate. Authors cannot rate their own scripts (§2.3).
const rateSchema = z.object({ rating: z.number().int().min(1).max(5) });
playScriptsRouter.post('/:id/rate', requireAuth, async (c) => {
  const user = c.get('user')!;
  const id = c.req.param('id');
  const body = await c.req.json();
  const parsed = rateSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: 'invalid' }, 400);
  const owner = await getPlayScriptOwner(id);
  if (!owner) return c.json({ error: 'not_found' }, 404);
  if (owner === user.id) {
    return c.json({ error: 'self_rating_forbidden', message: "Authors can't rate their own work." }, 403);
  }
  const stats = await ratePlayScript(user.id, id, parsed.data.rating);
  return c.json(stats);
});

// Tag listing.
export const tagsRouter = new Hono<{ Variables: AppVariables }>();
tagsRouter.get('/', async (c) => {
  const tags = await listTags();
  return c.json({ tags });
});
