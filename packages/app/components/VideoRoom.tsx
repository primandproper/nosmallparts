// Web implementation of VideoRoom using @livekit/components-react.
// React Native gets its own `.native.tsx` variant in this folder.

import { Text, View } from 'react-native';
import type { StageState } from '@nsp/shared';
import { useLiveKitToken } from '../hooks/useLiveKitToken';

interface Props {
  stageId: string;
  participantId: string | null;
  state: StageState;
}

export function VideoRoom({ stageId, participantId }: Props) {
  const { credentials, loading, error, unavailable } = useLiveKitToken({
    stageId,
    participantId,
  });

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

  return <WebLiveKitRoom credentials={credentials} />;
}

// Lazy-load the web LiveKit UI so the component tree stays safe in environments
// that don't need it (native uses VideoRoom.native.tsx and never hits this path).
function WebLiveKitRoom({ credentials }: { credentials: { token: string; url: string } }) {
  const { LiveKitRoom, VideoConference } = require('@livekit/components-react');
  require('@livekit/components-styles');
  return (
    <LiveKitRoom
      token={credentials.token}
      serverUrl={credentials.url}
      connect
      style={{ height: 240, borderRadius: 8, overflow: 'hidden' }}
    >
      <VideoConference />
    </LiveKitRoom>
  );
}
