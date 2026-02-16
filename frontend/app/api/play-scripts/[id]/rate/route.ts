import { NextRequest, NextResponse } from 'next/server';
import { Metadata } from 'nice-grpc';
import { getPlayScriptsClient } from '@/lib/grpc-client';
import { getSession } from '@/lib/session';
import { Status } from '@grpc/grpc-js/build/src/constants';

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
    default:
      return 500;
  }
}

// POST /api/play-scripts/[id]/rate - Rate a play script (1-5)
export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  try {
    const session = await getSession();
    if (!session?.accessToken) {
      return NextResponse.json({ error: 'Authentication required' }, { status: 401 });
    }

    const { id: playScriptId } = await params;

    if (!playScriptId?.trim()) {
      return NextResponse.json({ error: 'Play script ID is required' }, { status: 400 });
    }

    const body = await request.json();
    const rating = typeof body.rating === 'number' ? body.rating : parseInt(String(body.rating), 10);

    if (Number.isNaN(rating) || rating < 1 || rating > 5) {
      return NextResponse.json(
        { error: 'Rating must be a number between 1 and 5' },
        { status: 400 }
      );
    }

    const client = getPlayScriptsClient();
    await client.ratePlayScript(
      { playScriptId, rating },
      { metadata: Metadata({ Authorization: `Bearer ${session.accessToken}` }) }
    );

    return NextResponse.json({ success: true });
  } catch (error: unknown) {
    console.error('Error rating play script:', error);

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
