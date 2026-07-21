import type { Context } from 'hono';

/**
 * Resolve the client IP for ban / rate-limit purposes.
 *
 * Behind a reverse proxy we trust `X-Forwarded-For` and take the LAST entry
 * (the hop appended by the proxy closest to us — i.e. the last public hop).
 * Without the header we fall back to the raw socket address exposed by
 * @hono/node-server. If the proxy is misconfigured and the header is missing
 * in production, everyone shares the proxy's socket IP — see README for the
 * loud-failure note.
 */
export function getClientIp(c: Context): string | null {
  const xff = c.req.header('x-forwarded-for');
  if (xff) {
    const parts = xff
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
    const last = parts[parts.length - 1];
    if (last) return last;
  }
  // @hono/node-server exposes the Node IncomingMessage on c.env.incoming.
  const incoming = (c.env as { incoming?: { socket?: { remoteAddress?: string } } } | undefined)
    ?.incoming;
  return incoming?.socket?.remoteAddress ?? null;
}
