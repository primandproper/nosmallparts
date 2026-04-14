// Native implementation of VideoRoom using @livekit/react-native.

import { useEffect, useRef, useState } from 'react';
import { Text, View } from 'react-native';
import { Room, Track } from 'livekit-client';
import { VideoView, registerGlobals } from '@livekit/react-native';
import type { LocalParticipant, RemoteParticipant } from 'livekit-client';
import type { StageState } from '@nsp/shared';
import { useLiveKitToken } from '../hooks/useLiveKitToken';

// WebRTC globals only need to be registered once per app.
let globalsRegistered = false;
function ensureGlobals() {
  if (globalsRegistered) return;
  registerGlobals();
  globalsRegistered = true;
}

interface Props {
  stageId: string;
  participantId: string | null;
  state: StageState;
}

export function VideoRoom({ stageId, participantId, state }: Props) {
  const { credentials, loading, error, unavailable } = useLiveKitToken({
    stageId,
    participantId,
  });
  const roomRef = useRef<Room | null>(null);
  const [participants, setParticipants] = useState<(LocalParticipant | RemoteParticipant)[]>([]);
  const me = state.participants.find((p) => p.id === participantId);
  const isAudience = me?.isAudience ?? false;

  useEffect(() => {
    if (!credentials) return;
    ensureGlobals();
    const room = new Room();
    roomRef.current = room;

    function refresh() {
      const all: (LocalParticipant | RemoteParticipant)[] = [
        room.localParticipant,
        ...Array.from(room.remoteParticipants.values()),
      ];
      setParticipants(all);
    }

    room.on('participantConnected', refresh);
    room.on('participantDisconnected', refresh);
    room.on('trackSubscribed', refresh);
    room.on('trackUnsubscribed', refresh);

    (async () => {
      try {
        await room.connect(credentials.url, credentials.token);
        if (!isAudience) {
          await room.localParticipant.setCameraEnabled(true);
          await room.localParticipant.setMicrophoneEnabled(true);
        }
        refresh();
      } catch (e) {
        console.warn('[livekit] connect failed', e);
      }
    })();

    return () => {
      room.disconnect();
      roomRef.current = null;
    };
  }, [credentials, isAudience]);

  if (unavailable) {
    return (
      <View className="items-center justify-center p-3 border border-stage-border rounded-md">
        <Text className="text-stage-muted text-xs">Video disabled (LiveKit not configured)</Text>
      </View>
    );
  }
  if (loading || !credentials) {
    return (
      <View className="items-center justify-center p-3 border border-stage-border rounded-md">
        <Text className="text-stage-muted text-xs">
          {error ? `Video unavailable: ${error}` : 'Connecting video…'}
        </Text>
      </View>
    );
  }

  return (
    <View className="flex-row flex-wrap gap-2 p-2">
      {participants.map((p) => {
        const videoPub = p.getTrackPublication(Track.Source.Camera);
        return (
          <View
            key={p.identity}
            className="w-40 h-28 rounded-md overflow-hidden border border-stage-border bg-stage-surface"
          >
            {videoPub?.videoTrack ? (
              <VideoView style={{ flex: 1 }} videoTrack={videoPub.videoTrack} />
            ) : (
              <View className="flex-1 items-center justify-center">
                <Text className="text-stage-muted text-xs">{p.name || p.identity}</Text>
              </View>
            )}
          </View>
        );
      })}
    </View>
  );
}
