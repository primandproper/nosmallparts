import { useEffect, useState } from 'react';
import { api } from '../lib/api';

interface Params {
  stageId: string;
  participantId: string | null;
  enabled?: boolean;
}

export interface LiveKitCredentials {
  token: string;
  url: string;
  room: string;
}

/**
 * Fetches a LiveKit access token from the server. If LiveKit is not
 * configured on the server, `credentials` will be null and `unavailable`
 * will be true — the VideoRoom component should render a friendly placeholder.
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
        const creds = await api<LiveKitCredentials>('/api/livekit/token', {
          method: 'POST',
          body: JSON.stringify({ stageId, participantId }),
        });
        if (!cancelled) setCredentials(creds);
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Failed';
        if (!cancelled) {
          if (msg.includes('503')) {
            setUnavailable(true);
          } else {
            setError(msg);
          }
        }
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
