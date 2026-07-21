import { Hono } from 'hono';
import { AccessToken } from 'livekit-server-sdk';
import { z } from 'zod';
import type { AppVariables } from '../app.js';
import { getStageByIdOrCode, getStageState } from '../repos/stages.js';

export const livekitRouter = new Hono<{ Variables: AppVariables }>();

const tokenSchema = z.object({
  stageId: z.string().min(1),
  participantId: z.string().min(1),
});

livekitRouter.post('/token', async (c) => {
  const apiKey = process.env.LIVEKIT_API_KEY;
  const apiSecret = process.env.LIVEKIT_API_SECRET;
  const wsUrl = process.env.LIVEKIT_URL;

  if (!apiKey || !apiSecret || !wsUrl) {
    return c.json(
      { error: 'livekit_not_configured', message: 'LIVEKIT_API_KEY/SECRET/URL not set' },
      503
    );
  }

  const body = await c.req.json().catch(() => ({}));
  const parsed = tokenSchema.safeParse(body);
  if (!parsed.success) return c.json({ error: 'invalid' }, 400);

  const stage = await getStageByIdOrCode(parsed.data.stageId);
  if (!stage) return c.json({ error: 'stage_not_found' }, 404);
  const state = await getStageState(stage.id);
  const participant = state?.participants.find((p) => p.id === parsed.data.participantId);
  if (!participant) return c.json({ error: 'participant_not_found' }, 404);

  // Audience cannot publish; performers can publish + subscribe.
  const canPublish = !participant.isAudience;

  const at = new AccessToken(apiKey, apiSecret, {
    identity: participant.id,
    name: participant.name,
    // Long TTL so a network-drop reconnect late in a long session still works
    // (stages live at most a few hours; the room is torn down with the stage).
    ttl: 12 * 60 * 60,
  });
  at.addGrant({
    room: stage.id,
    roomJoin: true,
    canPublish,
    canSubscribe: true,
    canPublishData: canPublish,
  });

  const token = await at.toJwt();
  return c.json({ token, url: wsUrl, room: stage.id });
});
