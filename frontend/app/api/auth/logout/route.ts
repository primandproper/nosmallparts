import { NextResponse } from 'next/server';

export interface LogoutResponse {
  success: boolean;
}

export async function POST(): Promise<NextResponse<LogoutResponse>> {
  const res = NextResponse.json({ success: true });
  // Must match path used when setting the cookie
  res.cookies.set('nosmallparts_session', '', { path: '/', maxAge: 0 });
  return res;
}
