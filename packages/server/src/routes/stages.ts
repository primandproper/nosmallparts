import { Hono } from 'hono';
import { z } from 'zod';
import type { AppVariables } from '../app.js';
import { requireAuth } from '../middleware/require-auth.js';
import { rateLimitByIp } from '../rate-limit.js';
import { isPlayScriptEnabled } from '../repos/play-scripts.js';
import { createStage, getStageByCode } from '../repos/stages.js';

export const stagesRouter = new Hono<{ Variables: AppVariables }>();

const createSchema = z.object({
  playScriptId: z.string().optional().nullable(),
});

// Creating a stage requires a signed-in user (§1.4) — guests get 401.
stagesRouter.post('/', requireAuth, async (c) => {
  const body = await c.req.json().catch(() => ({}));
  const parsed = createSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: 'invalid' }, 400);
  const playScriptId = parsed.data.playScriptId ?? null;
  if (playScriptId && !(await isPlayScriptEnabled(playScriptId))) {
    return c.json({ error: 'script_unavailable', message: 'Script unavailable.' }, 400);
  }
  const stage = await createStage(playScriptId);
  return c.json(stage, 201);
});

// 30 requests / minute / IP (§12.5) — documented in README.
const existsRateLimit = rateLimitByIp({ limit: 30, windowMs: 60_000 });

stagesRouter.get('/:code/exists', existsRateLimit, async (c) => {
  const code = c.req.param('code').toUpperCase();
  const stage = await getStageByCode(code);
  if (!stage) return c.json({ exists: false }, 200);
  return c.json({ exists: true, id: stage.id, code: stage.code });
});
