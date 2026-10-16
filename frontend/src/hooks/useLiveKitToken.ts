import { Code } from '@primandproper/platform-client';
import { useEffect, useState } from 'react';

import { hasCode, messageOf } from '../lib/errors';
import { type LiveKitCredentials, mintVideoToken } from '../lib/rpc';

interface Params {
  stageId: string;
  participantId: string | null;
  enabled?: boolean;
}

export type { LiveKitCredentials };

/**
 * Fetches a LiveKit access token from the server. With no LiveKit server
 * configured, `credentials` stays null and `unavailable` is true, and the
 * VideoRoom renders a friendly placeholder.
 */
export function useLiveKitToken({ stageId, participantId, enabled = true }: Params) {
  const [credentials, setCredentials] = useState<LiveKitCredentials | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [unavailable, setUnavailable] = useState(false);

  useEffect(() => {
    if (!enabled || !participantId) {
      setCredentials(null);
      return;
    }
    let cancelled = false;
    (async () => {
      setLoading(true);
      setError(null);
      setUnavailable(false);
      try {
        const creds = await mintVideoToken(stageId, participantId);
        if (!cancelled) setCredentials(creds);
      } catch (e) {
        if (cancelled) return;
        if (hasCode(e, Code.FAILED_PRECONDITION)) setUnavailable(true);
        else setError(messageOf(e, 'Failed to join the video room.'));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [stageId, participantId, enabled]);

  return { credentials, loading, error, unavailable };
}
