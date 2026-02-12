import { NextRequest, NextResponse } from 'next/server';
import { getBackendCookieForRequest, clearEncryptedBackendCookie } from '@/lib/backend-session-cookie';
import { CastMember, PlayLine } from '@/lib/types';

function getBackendUrl(): string {
  return process.env.BACKEND_HTTP_URL || 'http://localhost:8000';
}

interface BackendCastMember {
  id: string;
  name: string;
  sortOrder: number;
}

interface BackendLine {
  castMemberId: string;
  lineText: string;
  lineIndex: number;
  direction?: string;
  scrollMode?: string;
  autoScrollWpm?: number;
}

interface BackendPlayScript {
  id: string;
  title: string;
  description: string;
  author?: string;
  castMembers?: BackendCastMember[];
  lines?: BackendLine[];
}

interface FrontendPlay {
  id: string;
  title: string;
  description: string;
  author?: string;
  isUserScript?: boolean;
  cast: CastMember[];
  lines: PlayLine[];
}

function toFrontendPlay(ps: BackendPlayScript): FrontendPlay {
  const cast = (ps.castMembers ?? [])
    .sort((a, b) => a.sortOrder - b.sortOrder)
    .map((cm) => ({ id: cm.id, name: cm.name }));
  const lines: PlayLine[] = (ps.lines ?? []).map((line) => {
    const result: PlayLine = {
      castMemberId: line.castMemberId,
      text: line.lineText,
      index: line.lineIndex,
    };
    if (line.direction) result.direction = line.direction;
    if (line.scrollMode === 'manual') result.scrollMode = 'manual';
    else if (line.scrollMode === 'auto') result.scrollMode = 'auto';
    if (line.autoScrollWpm != null) result.autoScrollWpm = line.autoScrollWpm;
    return result;
  });
  return {
    id: ps.id,
    title: ps.title,
    description: ps.description,
    author: ps.author || undefined,
    isUserScript: true,
    cast,
    lines,
  };
}

// GET /api/play-scripts/[id] - Get a single play script
export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params;

    if (!id?.trim()) {
      return NextResponse.json({ error: 'id is required' }, { status: 400 });
    }

    const sessionCookie = await getBackendCookieForRequest();
    const headers: Record<string, string> = {};
    if (sessionCookie) headers['X-Session-Cookie'] = sessionCookie;

    const backendUrl = getBackendUrl();
    const response = await fetch(`${backendUrl}/api/play-scripts/${encodeURIComponent(id)}`, {
      headers,
    });

    if (response.status === 401 && sessionCookie) {
      await clearEncryptedBackendCookie();
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    if (response.status === 404) {
      return NextResponse.json({ error: 'Play script not found' }, { status: 404 });
    }

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      return NextResponse.json(data, { status: response.status });
    }

    const play = (await response.json()) as BackendPlayScript;
    return NextResponse.json(toFrontendPlay(play));
  } catch (error: unknown) {
    console.error('Error getting play script:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}

// PATCH /api/play-scripts/[id] - Update a play script (auth required)
export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params;

    if (!id?.trim()) {
      return NextResponse.json({ error: 'id is required' }, { status: 400 });
    }

    const sessionCookie = await getBackendCookieForRequest();
    if (!sessionCookie) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const body = await request.json();
    const backendUrl = getBackendUrl();
    const response = await fetch(`${backendUrl}/api/play-scripts/${encodeURIComponent(id)}`, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-Cookie': sessionCookie,
      },
      body: JSON.stringify(body),
    });

    if (response.status === 401) {
      await clearEncryptedBackendCookie();
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      return NextResponse.json(data, { status: response.status });
    }

    const updated = (await response.json()) as BackendPlayScript;
    return NextResponse.json(toFrontendPlay(updated));
  } catch (error: unknown) {
    console.error('Error updating play script:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}

// DELETE /api/play-scripts/[id] - Archive a play script (auth required)
export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const { id } = await params;

    if (!id?.trim()) {
      return NextResponse.json({ error: 'id is required' }, { status: 400 });
    }

    const sessionCookie = await getBackendCookieForRequest();
    if (!sessionCookie) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const backendUrl = getBackendUrl();
    const response = await fetch(`${backendUrl}/api/play-scripts/${encodeURIComponent(id)}`, {
      method: 'DELETE',
      headers: { 'X-Session-Cookie': sessionCookie },
    });

    if (response.status === 401) {
      await clearEncryptedBackendCookie();
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      return NextResponse.json(data, { status: response.status });
    }

    return NextResponse.json({ success: true });
  } catch (error: unknown) {
    console.error('Error archiving play script:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
