// Web implementation of VideoRoom.
//
// We don't use LiveKit's prebuilt <VideoConference /> because we want to
// spotlight whichever participant is currently speaking and overlay their
// current line(s) directly on top of their video tile (teleprompter style).
// Built from LiveKit's lower-level primitives instead.

import { Text, View } from 'react-native';
import {
  getCurrentLines,
  getCurrentLinesForParticipant,
  getCurrentSpeakers,
} from '@nsp/shared';
import type { PlayLine, StageState } from '@nsp/shared';
import { useLiveKitToken } from '../hooks/useLiveKitToken';

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

  if (unavailable) {
    return (
      <View className="flex-1 items-center justify-center rounded-md border border-stage-border p-3">
        <Text className="text-stage-muted text-xs">Video disabled (LiveKit not configured)</Text>
      </View>
    );
  }
  if (loading || !credentials) {
    return (
      <View className="flex-1 items-center justify-center rounded-md border border-stage-border p-3">
        <Text className="text-stage-muted text-xs">
          {error ? `Video unavailable: ${error}` : 'Connecting video…'}
        </Text>
      </View>
    );
  }

  return (
    <WebLiveKitRoom credentials={credentials} state={state} />
  );
}

function WebLiveKitRoom({
  credentials,
  state,
}: {
  credentials: { token: string; url: string };
  state: StageState;
}) {
  const { LiveKitRoom, RoomAudioRenderer } = require('@livekit/components-react');
  require('@livekit/components-styles');
  return (
    <LiveKitRoom
      token={credentials.token}
      serverUrl={credentials.url}
      connect
      // Use the host element as a flex column so the spotlight + strip fill height.
      style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}
    >
      <RoomAudioRenderer />
      <SpotlightStage state={state} />
    </LiveKitRoom>
  );
}

function SpotlightStage({ state }: { state: StageState }) {
  const { useParticipants } = require('@livekit/components-react');
  const lkParticipants: Array<{ identity: string; name?: string }> = useParticipants();

  const speakerAppParticipants = getCurrentSpeakers(state);
  const speakerIds = new Set(speakerAppParticipants.map((p) => p.id));

  const speakers = lkParticipants.filter((p) => speakerIds.has(p.identity));
  const others = lkParticipants.filter((p) => !speakerIds.has(p.identity));

  const fallbackLines: PlayLine[] = speakers.length === 0 ? getCurrentLines(state) : [];

  return (
    <View className="flex-1 min-h-0 gap-2 p-2 md:gap-3 md:p-3">
      {/* Spotlight zone */}
      <View className="flex-1 min-h-0">
        {speakers.length === 0 ? (
          <SpotlightEmpty lines={fallbackLines} />
        ) : (
          <View className="flex-1 min-h-0 flex-col gap-2 md:flex-row md:gap-3">
            {speakers.map((p) => (
              <SpotlightTile
                key={p.identity}
                participant={p}
                lines={getCurrentLinesForParticipant(state, p.identity)}
              />
            ))}
          </View>
        )}
      </View>

      {/* Strip of non-speaking participants */}
      {others.length > 0 ? (
        <View className="flex-row flex-wrap gap-2">
          {others.map((p) => (
            <MiniTile key={p.identity} participant={p} />
          ))}
        </View>
      ) : null}
    </View>
  );
}

function SpotlightTile({
  participant,
  lines,
}: {
  participant: { identity: string; name?: string };
  lines: PlayLine[];
}) {
  const { ParticipantTile } = require('@livekit/components-react');
  return (
    <View className="relative flex-1 min-h-0 overflow-hidden rounded-md border-2 border-stage-gold bg-stage-surface">
      <View className="absolute inset-0">
        <ParticipantTile participant={participant} />
      </View>
      {lines.length > 0 ? <LineOverlay lines={lines} /> : null}
    </View>
  );
}

function SpotlightEmpty({ lines }: { lines: PlayLine[] }) {
  return (
    <View className="flex-1 min-h-0 items-center justify-center rounded-md border border-stage-border bg-stage-surface p-4">
      <Text className="text-xs uppercase tracking-widest text-stage-muted mb-2">
        No speaker
      </Text>
      {lines.map((line, i) => (
        <Text key={i} className="text-center text-lg text-stage-muted">
          {line.text}
        </Text>
      ))}
    </View>
  );
}

function MiniTile({ participant }: { participant: { identity: string; name?: string } }) {
  const { ParticipantTile } = require('@livekit/components-react');
  return (
    <View className="h-20 w-32 overflow-hidden rounded-md border border-stage-border bg-stage-surface md:h-24 md:w-40">
      <ParticipantTile participant={participant} />
    </View>
  );
}

function LineOverlay({ lines }: { lines: PlayLine[] }) {
  // Keyed by line text so the scroll container resets when the line changes,
  // putting the start of the new line back at the top.
  const key = lines.map((l) => l.text).join('|');
  return (
    <View
      key={key}
      className="absolute inset-x-0 bottom-0 max-h-[60%] bg-black/65 px-4 py-3 md:px-6 md:py-4"
      // overflow-y-auto only takes effect on web; native ignores it.
      style={{ overflowY: 'auto' } as object}
    >
      {lines.map((line, i) => (
        <View key={i} className={i > 0 ? 'mt-2' : undefined}>
          {line.direction ? (
            <Text className="italic text-stage-muted">({line.direction})</Text>
          ) : null}
          <Text className="font-serif text-2xl leading-snug text-stage-gold md:text-3xl lg:text-4xl">
            {line.text}
          </Text>
        </View>
      ))}
    </View>
  );
}
