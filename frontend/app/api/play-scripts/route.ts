import { NextRequest, NextResponse } from 'next/server';
import { getBackendCookieForRequest, clearEncryptedBackendCookie } from '@/lib/backend-session-cookie';
import { CastMember, PlayLine } from '@/lib/types';

function getBackendUrl(): string {
  return process.env.BACKEND_HTTP_URL || 'http://localhost:8000';
}

interface CreatePlayScriptBody {
  title: string;
  description: string;
  author?: string;
  isPublic?: boolean;
  cast: CastMember[];
  lines: PlayLine[];
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

// GET /api/play-scripts - List public play scripts (no auth required)
export async function GET() {
  try {
    const backendUrl = getBackendUrl();
    const sessionCookie = await getBackendCookieForRequest();
    const headers: Record<string, string> = {};
    if (sessionCookie) headers['X-Session-Cookie'] = sessionCookie;

    const response = await fetch(`${backendUrl}/api/play-scripts/public`, { headers });

    if (response.status === 401 && sessionCookie) {
      await clearEncryptedBackendCookie();
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      return NextResponse.json(data, { status: response.status });
    }

    const data = (await response.json()) as { data?: BackendPlayScript[] };
    const scripts = (data.data ?? []).map(toFrontendPlay);
    return NextResponse.json(scripts);
  } catch (error: unknown) {
    console.error('Error listing play scripts:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}

// POST /api/play-scripts - Create a play script (auth required)
export async function POST(request: NextRequest) {
  try {
    const sessionCookie = await getBackendCookieForRequest();
    if (!sessionCookie) {
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    const body: CreatePlayScriptBody = await request.json();

    if (!body.title?.trim()) {
      return NextResponse.json({ error: 'Title is required' }, { status: 400 });
    }
    if (!body.cast || body.cast.length === 0) {
      return NextResponse.json({ error: 'At least one cast member is required' }, { status: 400 });
    }
    if (!body.lines || body.lines.length === 0) {
      return NextResponse.json({ error: 'At least one line is required' }, { status: 400 });
    }

    const backendUrl = getBackendUrl();
    const input = {
      title: body.title.trim(),
      description: body.description?.trim() ?? '',
      author: body.author?.trim() ?? '',
      isPublic: body.isPublic ?? false,
      castMembers: body.cast.map((c, i) => ({ id: c.id, name: c.name, sortOrder: i })),
      lines: body.lines.map((line) => ({
        castMemberId: line.castMemberId,
        lineText: line.text,
        lineIndex: line.index,
        direction: line.direction ?? '',
        scrollMode: line.scrollMode ?? 'auto',
        autoScrollWpm: line.autoScrollWpm ?? 0,
      })),
    };

    const response = await fetch(`${backendUrl}/api/play-scripts`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Session-Cookie': sessionCookie,
      },
      body: JSON.stringify(input),
    });

    if (response.status === 401) {
      await clearEncryptedBackendCookie();
      return NextResponse.json({ error: 'Unauthorized' }, { status: 401 });
    }

    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      return NextResponse.json(data, { status: response.status });
    }

    const created = (await response.json()) as BackendPlayScript;
    return NextResponse.json(
      { success: true, playScript: toFrontendPlay(created) },
      { status: 201 }
    );
  } catch (error: unknown) {
    console.error('Error creating play script:', error);
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
