import { Hono } from 'hono';
import { z } from 'zod';
import type { AppVariables } from '../app.js';
import { createStage, getStageByCode } from '../repos/stages.js';

export const stagesRouter = new Hono<{ Variables: AppVariables }>();

const createSchema = z.object({
  playScriptId: z.string().optional().nullable(),
});

stagesRouter.post('/', async (c) => {
  const body = await c.req.json().catch(() => ({}));
  const parsed = createSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: 'invalid' }, 400);
  const stage = await createStage(parsed.data.playScriptId ?? null);
  return c.json(stage, 201);
});

stagesRouter.get('/:code/exists', async (c) => {
  const code = c.req.param('code').toUpperCase();
  const stage = await getStageByCode(code);
  if (!stage) return c.json({ exists: false }, 200);
  return c.json({ exists: true, id: stage.id, code: stage.code });
});
