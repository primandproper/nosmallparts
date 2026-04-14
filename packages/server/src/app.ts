import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { auth } from './auth.js';
import type { AuthUser, Session } from './auth.js';
import { playScriptsRouter, tagsRouter } from './routes/play-scripts.js';
import { stagesRouter } from './routes/stages.js';
import { livekitRouter } from './routes/livekit.js';

export type AppVariables = {
  user: AuthUser | null;
  session: Session['session'] | null;
};

export const app = new Hono<{ Variables: AppVariables }>();

const trustedOrigins = (process.env.TRUSTED_ORIGINS || 'http://localhost:8081')
  .split(',')
  .map((s) => s.trim())
  .filter(Boolean);

app.use('*', logger());
app.use(
  '*',
  cors({
    origin: trustedOrigins,
    credentials: true,
    allowHeaders: ['Content-Type', 'Authorization'],
    allowMethods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
    exposeHeaders: ['Set-Cookie'],
  })
);

// Populate user/session on every request for downstream routes.
app.use('*', async (c, next) => {
  const session = await auth.api.getSession({ headers: c.req.raw.headers });
  c.set('user', session?.user ?? null);
  c.set('session', session?.session ?? null);
  return next();
});

// Better Auth mounts all its routes under /api/auth/*.
app.on(['POST', 'GET'], '/api/auth/*', (c) => auth.handler(c.req.raw));

app.get('/api/health', (c) => c.json({ ok: true }));

app.get('/api/me', (c) => {
  const user = c.get('user');
  if (!user) return c.json({ user: null }, 200);
  return c.json({ user });
});

app.route('/api/play-scripts', playScriptsRouter);
app.route('/api/tags', tagsRouter);
app.route('/api/stages', stagesRouter);
app.route('/api/livekit', livekitRouter);
