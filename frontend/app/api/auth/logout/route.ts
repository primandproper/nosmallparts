import { NextResponse } from 'next/server';
import { clearSession } from '@/lib/session';

export interface LogoutResponse {
  success: boolean;
}

export async function POST(): Promise<NextResponse<LogoutResponse>> {
  await clearSession();
  return NextResponse.json({ success: true });
}
