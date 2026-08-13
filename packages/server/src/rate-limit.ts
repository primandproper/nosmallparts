import type { MiddlewareHandler } from 'hono';
import { getClientIp } from './client-ip.js';

interface Window {
  count: number;
  windowStart: number;
}

/**
 * Simple fixed-window in-memory rate limiter keyed by client IP.
 * Suitable for single-process deploys (which is all v1 supports).
 */
export function rateLimitByIp(opts: { limit: number; windowMs: number }): MiddlewareHandler {
  const windows = new Map<string, Window>();

  return async (c, next) => {
    const ip = getClientIp(c) ?? 'unknown';
    const now = Date.now();
    let w = windows.get(ip);
    if (!w || now - w.windowStart >= opts.windowMs) {
      w = { count: 0, windowStart: now };
      windows.set(ip, w);
    }
    w.count += 1;
    if (w.count > opts.limit) {
      const retryAfterSec = Math.max(1, Math.ceil((w.windowStart + opts.windowMs - now) / 1000));
      c.header('Retry-After', String(retryAfterSec));
      return c.json({ error: 'rate_limited' }, 429);
    }
    // Opportunistic cleanup so the map doesn't grow forever.
    if (windows.size > 10_000) {
      for (const [k, v] of windows) {
        if (now - v.windowStart >= opts.windowMs) windows.delete(k);
      }
    }
    return next();
  };
}
