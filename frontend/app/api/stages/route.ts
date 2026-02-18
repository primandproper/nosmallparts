import { NextResponse } from 'next/server';

// Get the Go backend URL from environment variable
function getBackendUrl(): string {
  return process.env.BACKEND_HTTP_URL || 'http://localhost:8000';
}

interface CreateStageResponse {
  stageId: string;
  code: string;
}

export async function POST() {
  try {
    const backendUrl = getBackendUrl();
    const response = await fetch(`${backendUrl}/api/stages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      const data = await response.json();
      return NextResponse.json(data, { status: response.status });
    }

    const data: CreateStageResponse = await response.json();
    return NextResponse.json(data, { status: 201 });
  } catch (error: unknown) {
    console.error('[Stages API] Error creating stage:', error);

    if (
      error instanceof Error &&
      (error.message.includes('fetch') || error.message.includes('ECONNREFUSED'))
    ) {
      return NextResponse.json({ error: 'Failed to connect to backend service' }, { status: 503 });
    }

    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
