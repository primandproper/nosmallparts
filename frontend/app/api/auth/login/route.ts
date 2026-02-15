import { NextRequest, NextResponse } from 'next/server';
import { getAuthClient } from '@/lib/auth-client';

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

export async function POST(request: NextRequest): Promise<NextResponse<LoginResponse>> {
  try {
    const body: LoginRequest = await request.json();

    if (!body.username || !body.password) {
      return NextResponse.json(
        { success: false, error: 'Username and password are required' },
        { status: 400 }
      );
    }

    const authClient = getAuthClient();

    const response = await authClient.loginForToken({
      input: {
        username: body.username,
        password: body.password,
        totpToken: body.totpToken || '',
      },
    });

    if (!response.result) {
      return NextResponse.json(
        { success: false, error: 'Login failed - no result returned' },
        { status: 401 }
      );
    }

    const { userId, accountId, accessToken, refreshToken, expiresUtc } = response.result;

    const sessionData = {
      userId,
      accountId,
      accessToken,
      refreshToken,
      expiresAt:
        expiresUtc?.toISOString() || new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
    };

    // Set cookie on the response object so it reaches the browser (cookies().set() alone may not)
    const res = NextResponse.json({ success: true, userId });
    const encoded = Buffer.from(JSON.stringify(sessionData)).toString('base64');

    // Secure for production (HTTPS)
    const isSecure = process.env.NODE_ENV === 'production';

    const cookieOptions: Parameters<NextResponse['cookies']['set']>[2] = {
      httpOnly: true,
      secure: isSecure,
      sameSite: 'lax',
      path: '/',
      expires: new Date(sessionData.expiresAt),
    };
    if (process.env.COOKIE_DOMAIN) {
      cookieOptions.domain = process.env.COOKIE_DOMAIN;
    }
    res.cookies.set('nosmallparts_session', encoded, cookieOptions);
    return res;
  } catch (error) {
    console.error('Login error:', error);

    // Check if this is a TOTP required error
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
