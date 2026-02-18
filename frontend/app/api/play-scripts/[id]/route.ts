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
import { CastMember, PlayLine } from '@/lib/types';
import { Status } from '@grpc/grpc-js/build/src/constants';

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
    isUserScript: true,
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

// GET /api/play-scripts/[id] - Get a single play script
export async function GET(_request: NextRequest, { params }: { params: Promise<{ id: string }> }) {
  try {
    const session = await getSession();
    if (!session?.accessToken) {
      return NextResponse.json({ error: 'Authentication required' }, { status: 401 });
    }

    const { id } = await params;

    if (!id?.trim()) {
      return NextResponse.json({ error: 'id is required' }, { status: 400 });
    }

    const client = getPlayScriptsClient();
    const response = await client.getPlayScript(
      { playScriptId: id },
      { metadata: Metadata({ Authorization: `Bearer ${session.accessToken}` }) }
    );

    if (!response.result) {
      return NextResponse.json({ error: 'Play script not found' }, { status: 404 });
    }

    return NextResponse.json(toFrontendPlay(response.result));
  } catch (error: unknown) {
    console.error('Error getting play script:', error);

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

// DELETE /api/play-scripts/[id] - Archive (delete) a play script
export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getSession();
    if (!session?.accessToken) {
      return NextResponse.json({ error: 'Authentication required' }, { status: 401 });
    }

    const { id } = await params;

    if (!id?.trim()) {
      return NextResponse.json({ error: 'id is required' }, { status: 400 });
    }

    const client = getPlayScriptsClient();
    await client.archivePlayScript(
      { playScriptId: id },
      { metadata: Metadata({ Authorization: `Bearer ${session.accessToken}` }) }
    );

    return NextResponse.json({ success: true });
  } catch (error: unknown) {
    console.error('Error archiving play script:', error);

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
