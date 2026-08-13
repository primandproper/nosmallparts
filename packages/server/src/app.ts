import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { serveStatic } from '@hono/node-server/serve-static';
import { auth } from './auth.js';
import type { AuthUser, Session } from './auth.js';
import { playScriptsRouter, tagsRouter } from './routes/play-scripts.js';
import { stagesRouter } from './routes/stages.js';
import { livekitRouter } from './routes/livekit.js';
import { mintGuestId, resolveIdentity } from './identity.js';

export type AppVariables = {
  user: AuthUser | null;
  session: Session['session'] | null;
};

export const app = new Hono<{ Variables: AppVariables }>();

const trustedOrigins = (process.env.TRUSTED_ORIGINS || 'http://localhost:8081')
  .split(',')
  .map((s) => s.trim())
  .filter(Boolean);

// Logging hygiene (§5.5, §12.3): stage codes must not appear in INFO logs.
// Mask the :code segment of stage lookups and the /stage/:code app route.
function redactStageCodes(str: string): string {
  return str
    .replace(/(\/api\/stages\/)[A-Za-z0-9]{4,12}(?=\/|\s|$|\?)/g, '$1:code')
    .replace(/(\/stage\/)[A-Za-z0-9]{4,12}(?=\/|\s|$|\?)/g, '$1:code');
}

app.use('*', logger((str, ...rest) => console.log(redactStageCodes(str), ...rest)));
app.use(
  '*',
  cors({
    origin: trustedOrigins,
    credentials: true,
    allowHeaders: ['Content-Type', 'Authorization'],
    allowMethods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS'],
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

app.get('/api/identity', async (c) => {
  const identity = await resolveIdentity(c);
  if (identity.userId) {
    return c.json({ userId: identity.userId, guestId: null });
  }
  const guestId = identity.guestId ?? mintGuestId(c);
  return c.json({ userId: null, guestId });
});

app.route('/api/play-scripts', playScriptsRouter);
app.route('/api/tags', tagsRouter);
app.route('/api/stages', stagesRouter);
app.route('/api/livekit', livekitRouter);

const webDir = process.env.WEB_STATIC_DIR;
if (webDir) {
  const isAppRoute = (path: string) => !path.startsWith('/api/') && !path.startsWith('/ws/');
  const staticHandler = serveStatic({ root: webDir });
  app.use('/*', async (c, next) => (isAppRoute(c.req.path) ? staticHandler(c, next) : next()));
  app.get('/*', async (c) => {
    if (!isAppRoute(c.req.path)) return c.notFound();
    try {
      const html = await readFile(join(webDir, 'index.html'), 'utf8');
      return c.html(html);
    } catch {
      return c.notFound();
    }
  });
}
