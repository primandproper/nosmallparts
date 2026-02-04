import { NextResponse } from 'next/server';
import { getSession, isAuthenticated } from '@/lib/session';

export interface SessionResponse {
  authenticated: boolean;
  userId?: string;
  accountId?: string;
}

export async function GET(): Promise<NextResponse<SessionResponse>> {
  const authenticated = await isAuthenticated();

  if (!authenticated) {
    return NextResponse.json({ authenticated: false });
  }

  const session = await getSession();

  return NextResponse.json({
    authenticated: true,
    userId: session?.userId,
    accountId: session?.accountId,
  });
}
