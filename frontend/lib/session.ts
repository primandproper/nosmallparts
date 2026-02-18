import { cookies } from 'next/headers';

const SESSION_COOKIE_NAME = 'nosmallparts_session';

export interface SessionData {
  userId: string;
  accountId: string;
  accessToken: string;
  refreshToken: string;
  expiresAt: string;
}

/**
 * Get OAuth2 client credentials from environment variables.
 * Required for making authenticated gRPC/API requests to the backend (e.g. OAuth2 code exchange).
 * Set OAUTH2_CLIENT_ID and OAUTH2_CLIENT_SECRET (see frontend/env.example).
 */
export function getOAuth2ClientCredentials() {
  const clientId = process.env.OAUTH2_CLIENT_ID;
  const clientSecret = process.env.OAUTH2_CLIENT_SECRET;

  if (!clientId || !clientSecret) {
    throw new Error(
      'OAuth2 client credentials not configured. Set OAUTH2_CLIENT_ID and OAUTH2_CLIENT_SECRET (see env.example).'
    );
  }

  return { clientId, clientSecret };
}

/**
 * Store session data in an HTTP-only cookie.
 * This should only be called from server-side code (API routes).
 */
export async function setSession(session: SessionData): Promise<void> {
  const cookieStore = await cookies();

  // Encode session as base64 JSON
  const encoded = Buffer.from(JSON.stringify(session)).toString('base64');

  cookieStore.set(SESSION_COOKIE_NAME, encoded, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    // Set expiry based on token expiration
    expires: new Date(session.expiresAt),
  });
}

/**
 * Get session data from HTTP-only cookie.
 * This should only be called from server-side code.
 */
export async function getSession(): Promise<SessionData | null> {
  const cookieStore = await cookies();
  const cookie = cookieStore.get(SESSION_COOKIE_NAME);

  if (!cookie) {
    return null;
  }

  try {
    const decoded = Buffer.from(cookie.value, 'base64').toString('utf-8');
    return JSON.parse(decoded) as SessionData;
  } catch {
    return null;
  }
}

/**
 * Clear the session cookie.
 */
export async function clearSession(): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.delete(SESSION_COOKIE_NAME);
}

/**
 * Check if the user is authenticated.
 */
export async function isAuthenticated(): Promise<boolean> {
  const session = await getSession();
  if (!session) {
    return false;
  }

  // Check if token has expired
  const expiresAt = new Date(session.expiresAt);
  return expiresAt > new Date();
}
