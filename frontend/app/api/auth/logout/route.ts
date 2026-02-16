import { NextResponse } from 'next/server';

export interface LogoutResponse {
  success: boolean;
}

export async function POST(): Promise<NextResponse<LogoutResponse>> {
  const res = NextResponse.json({ success: true });
  // Match login cookie options so the browser clears the cookie
  const opts: Parameters<NextResponse['cookies']['set']>[2] = {
    path: '/',
    maxAge: 0,
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
  };
  if (process.env.COOKIE_DOMAIN) {
    opts.domain = process.env.COOKIE_DOMAIN;
  }
  res.cookies.set('nosmallparts_session', '', opts);
  return res;
}
