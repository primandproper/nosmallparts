'use client';

import { useState, useEffect } from 'react';

interface UseLiveKitOptions {
  stageId: string;
  participantId: string | null;
  participantName: string;
}

interface UseLiveKitReturn {
  token: string | null;
  url: string | null;
  error: string | null;
}

export function useLiveKit({
  stageId,
  participantId,
  participantName,
}: UseLiveKitOptions): UseLiveKitReturn {
  const [token, setToken] = useState<string | null>(null);
  const [url, setUrl] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!stageId || !participantId || !participantName) return;

    const fetchToken = async () => {
      try {
        const response = await fetch('/api/livekit/token', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            stageId,
            participantId,
            participantName,
          }),
        });

        if (!response.ok) {
          const data = await response.json();
          if (response.status === 503) {
            // LiveKit not configured - this is OK, we'll use mock mode
            setToken(null);
            setUrl('');
            return;
          }
          throw new Error(data.error || 'Failed to get LiveKit token');
        }

        const data = await response.json();
        setToken(data.token);
        setUrl(data.url);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to connect to LiveKit');
      }
    };

    fetchToken();
  }, [stageId, participantId, participantName]);

  return { token, url, error };
}
