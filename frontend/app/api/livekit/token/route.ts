import { NextRequest, NextResponse } from 'next/server';
import { getBackendCookieForRequest, clearEncryptedBackendCookie } from '@/lib/backend-session-cookie';

function getBackendUrl(): string {
  const url = process.env.BACKEND_HTTP_URL || 'http://localhost:8000';
  return url;
}

interface TokenRequest {
  stageId: string;
  participantId: string;
  participantName: string;
}

interface TokenResponse {
  token: string;
  url: string;
}

interface ErrorResponse {
  error: string;
}

export async function POST(request: NextRequest) {
  try {
    const body: TokenRequest = await request.json();

    // Validate required fields
    if (!body.stageId?.trim()) {
      return NextResponse.json({ error: 'stageId is required' }, { status: 400 });
    }
    if (!body.participantId?.trim()) {
      return NextResponse.json({ error: 'participantId is required' }, { status: 400 });
    }
    if (!body.participantName?.trim()) {
      return NextResponse.json({ error: 'participantName is required' }, { status: 400 });
    }

    const sessionCookie = await getBackendCookieForRequest();
    if (!sessionCookie) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const backendUrl = getBackendUrl();
    const response = await fetch(`${backendUrl}/api/livekit/token`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-Cookie': sessionCookie,
      },
      body: JSON.stringify({
        stageId: body.stageId,
        participantId: body.participantId,
        participantName: body.participantName,
      }),
    });

    if (response.status === 401) {
      await clearEncryptedBackendCookie();
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    // Handle 503 (LiveKit not configured) specially
    if (response.status === 503) {
      const data: ErrorResponse = await response.json();
      return NextResponse.json(data, { status: 503 });
    }

    if (!response.ok) {
      const data: ErrorResponse = await response.json();
      return NextResponse.json(data, { status: response.status });
    }

    const data: TokenResponse = await response.json();
    console.log('[LiveKit API] Success, returning token');
    return NextResponse.json(data, { status: 200 });
  } catch (error: unknown) {
    console.error('[LiveKit API] Error:', error);

    // Check if it's a connection error to the backend
    if (error instanceof Error) {
      console.error('[LiveKit API] Error message:', error.message);
      if (error.message.includes('fetch') || error.message.includes('ECONNREFUSED')) {
        return NextResponse.json(
          { error: 'Failed to connect to backend service' },
          { status: 503 }
        );
      }
    }

    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
