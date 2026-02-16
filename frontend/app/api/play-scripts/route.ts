import { NextRequest, NextResponse } from 'next/server';
import { Metadata } from 'nice-grpc';
import { getPlayScriptsClient } from '@/lib/grpc-client';
import { getSession } from '@/lib/session';
import {
  PlayScript,
  PlayScriptCastMember,
  PlayScriptLine,
  ScrollMode,
} from '@/generated/play_scripts/play_scripts_messages';
import {
  PlayScriptCreationRequestInput,
  PlayScriptCastMemberCreationInput,
  PlayScriptLineCreationInput,
} from '@/generated/play_scripts/play_scripts_service_types';
import { CastMember, PlayLine } from '@/lib/types';
import { Status } from '@grpc/grpc-js/build/src/constants';

// Request body shape from the frontend
interface CreatePlayScriptBody {
  title: string;
  description: string;
  author?: string;
  isPublic?: boolean;
  cast: CastMember[];
  lines: PlayLine[];
}

// Frontend Play shape (what the scripts pages expect)
interface FrontendPlay {
  id: string;
  title: string;
  description: string;
  author?: string;
  isUserScript?: boolean;
  cast: CastMember[];
  lines: PlayLine[];
}

// Convert a gRPC PlayScript to the frontend Play format
function toFrontendPlay(ps: PlayScript): FrontendPlay {
  return {
    id: ps.id,
    title: ps.title,
    description: ps.description,
    author: ps.author || undefined,
    isUserScript: true, // All scripts from the backend are user-created
    cast: (ps.castMembers || [])
      .sort((a: PlayScriptCastMember, b: PlayScriptCastMember) => a.sortOrder - b.sortOrder)
      .map((cm: PlayScriptCastMember) => ({
        id: cm.id,
        name: cm.name,
        description: cm.description || undefined,
      })),
    lines: (ps.lines || []).map((line: PlayScriptLine) => {
      const result: PlayLine = {
        castMemberId: line.castMemberId,
        text: line.lineText,
        index: line.lineIndex,
      };
      if (line.direction) {
        result.direction = line.direction;
      }
      if (line.scrollMode === ScrollMode.SCROLL_MODE_MANUAL) {
        result.scrollMode = 'manual';
      } else if (line.scrollMode === ScrollMode.SCROLL_MODE_AUTO) {
        result.scrollMode = 'auto';
      }
      if (line.autoScrollWpm) {
        result.autoScrollWpm = line.autoScrollWpm;
      }
      return result;
    }),
  };
}

// Convert frontend scroll mode string to protobuf enum
function toScrollModeEnum(mode?: 'auto' | 'manual'): ScrollMode {
  switch (mode) {
    case 'manual':
      return ScrollMode.SCROLL_MODE_MANUAL;
    case 'auto':
      return ScrollMode.SCROLL_MODE_AUTO;
    default:
      return ScrollMode.SCROLL_MODE_UNSPECIFIED; // Will default to AUTO on backend
  }
}

// Convert frontend CastMember to protobuf creation input
function toCastMemberInput(cast: CastMember, index: number): PlayScriptCastMemberCreationInput {
  return PlayScriptCastMemberCreationInput.create({
    id: cast.id,
    name: cast.name,
    description: cast.description?.trim() || '',
    sortOrder: index,
  });
}

// Convert frontend PlayLine to protobuf creation input
function toLineInput(line: PlayLine): PlayScriptLineCreationInput {
  return PlayScriptLineCreationInput.create({
    castMemberId: line.castMemberId,
    lineText: line.text,
    lineIndex: line.index,
    direction: line.direction || '',
    scrollMode: toScrollModeEnum(line.scrollMode),
    autoScrollWpm: line.autoScrollWpm || 0,
  });
}

// Map gRPC status codes to HTTP status codes
function grpcStatusToHttp(code: number): number {
  switch (code) {
    case Status.OK:
      return 200;
    case Status.INVALID_ARGUMENT:
      return 400;
    case Status.UNAUTHENTICATED:
      return 401;
    case Status.PERMISSION_DENIED:
      return 403;
    case Status.NOT_FOUND:
      return 404;
    case Status.ALREADY_EXISTS:
      return 409;
    case Status.RESOURCE_EXHAUSTED:
      return 429;
    case Status.INTERNAL:
      return 500;
    case Status.UNAVAILABLE:
      return 503;
    default:
      return 500;
  }
}

// GET /api/play-scripts - List public play scripts
export async function GET() {
  try {
    const session = await getSession();
    if (!session?.accessToken) {
      return NextResponse.json({ error: 'Authentication required' }, { status: 401 });
    }

    const client = getPlayScriptsClient();
    const response = await client.getPublicPlayScripts(
      { filter: undefined },
      { metadata: Metadata({ Authorization: `Bearer ${session.accessToken}` }) }
    );

    const scripts = (response.results || []).map(toFrontendPlay);
    return NextResponse.json(scripts);
  } catch (error: unknown) {
    console.error('Error listing play scripts:', error);

    if (error && typeof error === 'object' && 'code' in error) {
      const grpcError = error as { code: number; message?: string };
      const httpStatus = grpcStatusToHttp(grpcError.code);
      return NextResponse.json(
        { error: grpcError.message || 'gRPC error occurred' },
        { status: httpStatus }
      );
    }

    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}

// POST /api/play-scripts - Create a play script
export async function POST(request: NextRequest) {
  try {
    const session = await getSession();
    if (!session?.accessToken) {
      return NextResponse.json({ error: 'Authentication required' }, { status: 401 });
    }

    const body: CreatePlayScriptBody = await request.json();

    // Validate required fields
    if (!body.title?.trim()) {
      return NextResponse.json({ error: 'Title is required' }, { status: 400 });
    }
    if (!body.cast || body.cast.length === 0) {
      return NextResponse.json({ error: 'At least one cast member is required' }, { status: 400 });
    }
    if (!body.lines || body.lines.length === 0) {
      return NextResponse.json({ error: 'At least one line is required' }, { status: 400 });
    }

    // Convert to protobuf input format
    const input: PlayScriptCreationRequestInput = PlayScriptCreationRequestInput.create({
      title: body.title.trim(),
      description: body.description?.trim() || '',
      author: body.author?.trim() || '',
      isPublic: body.isPublic ?? false,
      castMembers: body.cast.map((c, i) => toCastMemberInput(c, i)),
      lines: body.lines.map(toLineInput),
    });

    // Call gRPC service with auth
    const client = getPlayScriptsClient();
    const response = await client.createPlayScript(
      { input },
      { metadata: Metadata({ Authorization: `Bearer ${session.accessToken}` }) }
    );

    // Convert response to JSON
    const created = response.created;
    if (!created) {
      return NextResponse.json({ error: 'Failed to create play script' }, { status: 500 });
    }

    return NextResponse.json(
      {
        success: true,
        playScript: PlayScript.toJSON(created),
      },
      { status: 201 }
    );
  } catch (error: unknown) {
    console.error('Error creating play script:', error);

    // Handle gRPC errors
    if (error && typeof error === 'object' && 'code' in error) {
      const grpcError = error as { code: number; message?: string };
      const httpStatus = grpcStatusToHttp(grpcError.code);
      return NextResponse.json(
        { error: grpcError.message || 'gRPC error occurred' },
        { status: httpStatus }
      );
    }

    return NextResponse.json({ error: 'Internal server error' }, { status: 500 });
  }
}
