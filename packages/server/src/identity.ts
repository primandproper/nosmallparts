import type { Context } from 'hono';
import { getCookie, setCookie } from 'hono/cookie';
import { ulid } from 'ulid';
import { auth } from './auth.js';
import type { AppVariables } from './app.js';

export const GUEST_COOKIE = 'nsp_guest_id';
const GUEST_COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 180;

export interface Identity {
  userId: string | null;
  guestId: string | null;
}

function isValidGuestId(s: string | undefined | null): s is string {
  return typeof s === 'string' && /^[0-9A-HJKMNP-TV-Z]{26}$/.test(s);
}

export async function resolveIdentity(c: Context<{ Variables: AppVariables }>): Promise<Identity> {
  const existingUser = c.get('user');
  if (existingUser) {
    return { userId: existingUser.id, guestId: null };
  }

  // For WS upgrades on native, the client can't set an Authorization header,
  // so it passes the better-auth bearer as `?token=`. Synthesize a headers
  // object and let better-auth resolve it.
  const tokenFromQuery = c.req.query('token');
  if (tokenFromQuery) {
    const headers = new Headers(c.req.raw.headers);
    headers.set('Authorization', `Bearer ${tokenFromQuery}`);
    const session = await auth.api.getSession({ headers }).catch(() => null);
    if (session?.user) {
      return { userId: session.user.id, guestId: null };
    }
  }

  const session = await auth.api.getSession({ headers: c.req.raw.headers }).catch(() => null);
  if (session?.user) {
    return { userId: session.user.id, guestId: null };
  }

  const guestFromQuery = c.req.query('guest');
  if (isValidGuestId(guestFromQuery)) {
    return { userId: null, guestId: guestFromQuery };
  }
  const fromCookie = getCookie(c, GUEST_COOKIE);
  if (isValidGuestId(fromCookie)) {
    return { userId: null, guestId: fromCookie };
  }
  return { userId: null, guestId: null };
}

export function mintGuestId(c: Context): string {
  const guestId = ulid();
  setCookie(c, GUEST_COOKIE, guestId, {
    httpOnly: true,
    sameSite: 'Lax',
    secure: process.env.NODE_ENV === 'production',
    maxAge: GUEST_COOKIE_MAX_AGE_SECONDS,
    path: '/',
  });
  return guestId;
}
