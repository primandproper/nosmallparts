import { NextRequest, NextResponse } from 'next/server';

export interface RegisterRequest {
  username: string;
  password: string;
  emailAddress: string;
  firstName: string;
  lastName: string;
  birthday?: string;
  accountName?: string;
  invitationToken?: string;
  invitationId?: string;
  acceptedTos: boolean;
  acceptedPrivacyPolicy: boolean;
}

export interface RegisterResponse {
  success: boolean;
  error?: string;
  userId?: string;
  twoFactorQrCode?: string;
  twoFactorSecret?: string;
}

function getBackendUrl(): string {
  return process.env.BACKEND_HTTP_URL || 'http://localhost:8000';
}

export async function POST(request: NextRequest): Promise<NextResponse<RegisterResponse>> {
  try {
    const body: RegisterRequest = await request.json();

    // Validate required fields
    if (!body.username || !body.password || !body.emailAddress) {
      return NextResponse.json(
        { success: false, error: 'Username, password, and email are required' },
        { status: 400 }
      );
    }

    if (!body.firstName || !body.lastName) {
      return NextResponse.json(
        { success: false, error: 'First name and last name are required' },
        { status: 400 }
      );
    }

    if (!body.acceptedTos || !body.acceptedPrivacyPolicy) {
      return NextResponse.json(
        { success: false, error: 'You must accept the Terms of Service and Privacy Policy' },
        { status: 400 }
      );
    }

    const backendUrl = getBackendUrl();
    const res = await fetch(`${backendUrl}/api/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: body.username,
        password: body.password,
        emailAddress: body.emailAddress,
        firstName: body.firstName,
        lastName: body.lastName,
        birthday: body.birthday ?? undefined,
        accountName: body.accountName || `${body.firstName}'s Account`,
        invitationToken: body.invitationToken || '',
        invitationId: body.invitationId || '',
        acceptedTos: body.acceptedTos,
        acceptedPrivacyPolicy: body.acceptedPrivacyPolicy,
      }),
    });

    const data = await res.json().catch(() => ({}));

    if (!res.ok) {
      return NextResponse.json(
        { success: false, error: (data as { error?: string }).error || 'Registration failed' },
        { status: res.status }
      );
    }

    const created = data as {
      createdUserId?: string;
      twoFactorQrCode?: string;
      twoFactorSecret?: string;
    };
    return NextResponse.json({
      success: true,
      userId: created.createdUserId,
      twoFactorQrCode: created.twoFactorQrCode,
      twoFactorSecret: created.twoFactorSecret,
    });
  } catch (error) {
    console.error('Registration error:', error);
    const errorMessage = error instanceof Error ? error.message : 'Registration failed';
    return NextResponse.json({ success: false, error: errorMessage }, { status: 500 });
  }
}
