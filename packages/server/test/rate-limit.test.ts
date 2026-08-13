import { describe, expect, it } from 'vitest';
import { app } from '../src/app.js';

describe('stage code lookup rate limit (§12.5)', () => {
  it('allows 30/min/IP then returns 429 with Retry-After; per-IP isolation', async () => {
    const ip = '192.0.2.55';
    for (let i = 0; i < 30; i++) {
      const res = await app.request('/api/stages/ABCDEF/exists', {
        headers: { 'X-Forwarded-For': ip },
      });
      expect(res.status).toBe(200);
    }
    const blocked = await app.request('/api/stages/ABCDEF/exists', {
      headers: { 'X-Forwarded-For': ip },
    });
    expect(blocked.status).toBe(429);
    expect(Number(blocked.headers.get('Retry-After'))).toBeGreaterThan(0);

    // A different IP is unaffected.
    const other = await app.request('/api/stages/ABCDEF/exists', {
      headers: { 'X-Forwarded-For': '192.0.2.56' },
    });
    expect(other.status).toBe(200);
  });
});
