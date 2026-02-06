import { NextRequest, NextResponse } from 'next/server';

// Get the Go backend URL from environment variable
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

    const backendUrl = getBackendUrl();
    const response = await fetch(`${backendUrl}/api/stages/${encodeURIComponent(code)}/exists`);

    if (!response.ok) {
      const data = await response.json();
      return NextResponse.json(data, { status: response.status });
    }

    const data: StageExistsResponse = await response.json();
    return NextResponse.json(data);
  } catch (error: unknown) {
    console.error('[Stages API] Error checking stage existence:', error);

    if (error instanceof Error && (error.message.includes('fetch') || error.message.includes('ECONNREFUSED'))) {
      return NextResponse.json(
        { error: 'Failed to connect to backend service' },
        { status: 503 }
      );
    }

    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
