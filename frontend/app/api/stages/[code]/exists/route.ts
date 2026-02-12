import { NextRequest, NextResponse } from 'next/server';
import { getBackendCookieForRequest, clearEncryptedBackendCookie } from '@/lib/backend-session-cookie';

function getBackendUrl(): string {
  return process.env.BACKEND_HTTP_URL || 'http://localhost:8000';
}

interface StageExistsResponse {
  exists: boolean;
}

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ code: string }> }
) {
  try {
    const { code } = await params;

    if (!code?.trim()) {
      return NextResponse.json({ error: 'code is required' }, { status: 400 });
    }

    const sessionCookie = await getBackendCookieForRequest();
    const headers: Record<string, string> = {};
    if (sessionCookie) {
      headers['X-Session-Cookie'] = sessionCookie;
    }

    const backendUrl = getBackendUrl();
    const response = await fetch(`${backendUrl}/api/stages/${encodeURIComponent(code)}/exists`, {
      headers,
    });

    if (response.status === 401 && sessionCookie) {
      await clearEncryptedBackendCookie();
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    if (!response.ok) {
      const data = await response.json();
      return NextResponse.json(data, { status: response.status });
    }

    const data: StageExistsResponse = await response.json();
    return NextResponse.json(data);
  } catch (error: unknown) {
    console.error('[Stages API] Error checking stage existence:', error);

    if (
      error instanceof Error &&
      (error.message.includes('fetch') || error.message.includes('ECONNREFUSED'))
    ) {
      return NextResponse.json({ error: 'Failed to connect to backend service' }, { status: 503 });
    }

    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
