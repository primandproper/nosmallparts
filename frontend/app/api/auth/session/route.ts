import { Metadata } from 'nice-grpc';
import { NextResponse } from 'next/server';
import { getAuthClient } from '@/lib/auth-client';
import { getSession, isAuthenticated } from '@/lib/session';

export interface SessionResponse {
  authenticated: boolean;
  userId?: string;
  accountId?: string;
  firstName?: string;
}

export async function GET(): Promise<NextResponse<SessionResponse>> {
  const authenticated = await isAuthenticated();

  if (!authenticated) {
    return NextResponse.json({ authenticated: false });
  }

  const session = await getSession();
  if (!session?.accessToken) {
    return NextResponse.json({
      authenticated: true,
      userId: session?.userId,
      accountId: session?.accountId,
    });
  }

  let firstName: string | undefined;
  try {
    const authClient = getAuthClient();
    const response = await authClient.getSelf(
      {},
      { metadata: Metadata({ Authorization: `Bearer ${session.accessToken}` }) }
    );
    firstName = response.result?.firstName;
  } catch {
    // Continue without firstName if GetSelf fails
  }

  return NextResponse.json({
    authenticated: true,
    userId: session.userId,
    accountId: session.accountId,
    ...(firstName && { firstName }),
  });
}
