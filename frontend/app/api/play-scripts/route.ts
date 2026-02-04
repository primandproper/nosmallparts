import { NextRequest, NextResponse } from 'next/server';
import { getPlayScriptsClient } from '@/lib/grpc-client';
import { PlayScript } from '@/generated/play_scripts/play_scripts_messages';
import {
  PlayScriptCreationRequestInput,
  PlayScriptCastMemberCreationInput,
  PlayScriptLineCreationInput,
} from '@/generated/play_scripts/play_scripts_service_types';
import { ScrollMode } from '@/generated/play_scripts/play_scripts_messages';
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

export async function POST(request: NextRequest) {
  try {
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

    // Call gRPC service
    const client = getPlayScriptsClient();
    const response = await client.createPlayScript({ input });

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
