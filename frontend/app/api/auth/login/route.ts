import { NextRequest, NextResponse } from 'next/server';
import { setSession } from '@/lib/session';
import {
  encryptBackendCookiePayload,
  setEncryptedBackendCookie,
} from '@/lib/backend-session-cookie';

const BACKEND_LOGIN_PATH = '/api/login';

function getBackendUrl(): string {
  return process.env.BACKEND_HTTP_URL || 'http://localhost:8000';
}

export interface LoginRequest {
  username: string;
  password: string;
  totpToken?: string;
}

export interface LoginResponse {
  success: boolean;
  error?: string;
  requiresTOTP?: boolean;
  userId?: string;
}

/** Parse the first Set-Cookie header value to get the cookie value (for the given name or first cookie). */
function parseSetCookieValue(setCookieHeader: string): string | null {
  const eq = setCookieHeader.indexOf('=');
  if (eq === -1) return null;
  const rest = setCookieHeader.slice(eq + 1);
  const semicolon = rest.indexOf(';');
  const value = (semicolon === -1 ? rest : rest.slice(0, semicolon)).trim();
  return value || null;
}

export async function POST(request: NextRequest): Promise<NextResponse<LoginResponse>> {
  try {
    const body: LoginRequest = await request.json();

    if (!body.username || !body.password) {
      return NextResponse.json(
        { success: false, error: 'Username and password are required' },
        { status: 400 }
      );
    }

    const backendUrl = getBackendUrl();
    const loginRes = await fetch(`${backendUrl}${BACKEND_LOGIN_PATH}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: body.username,
        password: body.password,
        totpToken: body.totpToken ?? '',
      }),
    });

    if (!loginRes.ok) {
      const errBody = await loginRes.json().catch(() => ({}));
      const message = (errBody as { error?: { message?: string } })?.error?.message;
      if (
        typeof message === 'string' &&
        (message.toLowerCase().includes('totp') || message.toLowerCase().includes('two-factor'))
      ) {
        return NextResponse.json(
          { success: false, error: 'Two-factor authentication required', requiresTOTP: true },
          { status: 401 }
        );
      }
      return NextResponse.json(
        { success: false, error: message ?? 'Invalid credentials' },
        { status: 401 }
      );
    }

    const data = (await loginRes.json()) as {
      data?: { userID?: string; accountID?: string };
    };
    const userId = data?.data?.userID ?? '';
    const accountId = data?.data?.accountID ?? '';
    if (!userId || !accountId) {
      return NextResponse.json(
        { success: false, error: 'Login failed - no user data returned' },
        { status: 401 }
      );
    }

    const getSetCookie =
      typeof loginRes.headers.getSetCookie === 'function'
        ? loginRes.headers.getSetCookie()
        : [];
    const setCookieHeader = getSetCookie[0];
    const backendCookieValue = setCookieHeader ? parseSetCookieValue(setCookieHeader) : null;
    if (!backendCookieValue) {
      return NextResponse.json(
        { success: false, error: 'Login failed - no session cookie from backend' },
        { status: 401 }
      );
    }

    const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000);
    const payload = {
      backendCookie: backendCookieValue,
      exp: Math.floor(expiresAt.getTime() / 1000),
    };
    const encrypted = await encryptBackendCookiePayload(payload);
    await setEncryptedBackendCookie(encrypted, expiresAt);

    await setSession({
      userId,
      accountId,
      accessToken: '',
      refreshToken: '',
      expiresAt: expiresAt.toISOString(),
    });

    return NextResponse.json({
      success: true,
      userId,
    });
  } catch (error) {
    console.error('Login error:', error);
    const errorMessage = error instanceof Error ? error.message : 'Login failed';
    if (
      errorMessage.toLowerCase().includes('totp') ||
      errorMessage.toLowerCase().includes('two-factor')
    ) {
      return NextResponse.json(
        { success: false, error: 'Two-factor authentication required', requiresTOTP: true },
        { status: 401 }
      );
    }
    return NextResponse.json({ success: false, error: errorMessage }, { status: 401 });
  }
}
