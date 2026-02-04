import { NextRequest, NextResponse } from 'next/server';
import { getIdentityClient } from '@/lib/auth-client';

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

    const identityClient = getIdentityClient();

    const response = await identityClient.createUser({
      input: {
        username: body.username,
        password: body.password,
        emailAddress: body.emailAddress,
        firstName: body.firstName,
        lastName: body.lastName,
        birthday: body.birthday ? new Date(body.birthday) : undefined,
        accountName: body.accountName || `${body.firstName}'s Account`,
        invitationToken: body.invitationToken || '',
        invitationId: body.invitationId || '',
        acceptedTos: body.acceptedTos,
        acceptedPrivacyPolicy: body.acceptedPrivacyPolicy,
      },
    });

    if (!response.created) {
      return NextResponse.json(
        { success: false, error: 'Registration failed - no response returned' },
        { status: 500 }
      );
    }

    return NextResponse.json({
      success: true,
      userId: response.created.createdUserId,
      twoFactorQrCode: response.created.twoFactorQrCode,
      twoFactorSecret: response.created.twoFactorSecret,
    });
  } catch (error) {
    console.error('Registration error:', error);
    const errorMessage = error instanceof Error ? error.message : 'Registration failed';
    return NextResponse.json({ success: false, error: errorMessage }, { status: 500 });
  }
}
