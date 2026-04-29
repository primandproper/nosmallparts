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
import { TeleprompterLine } from './TeleprompterLine';

const DEFAULT_WPM = 150;

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

  const me = state.participants.find((p) => p.id === participantId);
  // Audience watches without publishing; cast members publish video and (unless force-muted) audio.
  const publishVideo = !!me && !me.isAudience;
  const publishAudio = !!me && !me.isAudience && !me.forceMuted;

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
    <WebLiveKitRoom
      credentials={credentials}
      state={state}
      participantId={participantId}
      publishVideo={publishVideo}
      publishAudio={publishAudio}
    />
  );
}

function WebLiveKitRoom({
  credentials,
  state,
  participantId,
  publishVideo,
  publishAudio,
}: {
  credentials: { token: string; url: string };
  state: StageState;
  participantId: string | null;
  publishVideo: boolean;
  publishAudio: boolean;
}) {
  const { LiveKitRoom, RoomAudioRenderer } = require('@livekit/components-react');
  require('@livekit/components-styles');
  return (
    <LiveKitRoom
      token={credentials.token}
      serverUrl={credentials.url}
      connect
      video={publishVideo}
      audio={publishAudio}
      // Use the host element as a flex column so the spotlight + strip fill height.
      style={{ flex: 1, display: 'flex', flexDirection: 'column', minHeight: 0 }}
    >
      <RoomAudioRenderer />
      <SpotlightStage state={state} participantId={participantId} />
    </LiveKitRoom>
  );
}

function SpotlightStage({
  state,
  participantId,
}: {
  state: StageState;
  participantId: string | null;
}) {
  const { useTracks } = require('@livekit/components-react');
  const { Track } = require('livekit-client');
  const tracks: Array<{ participant: { identity: string; name?: string } }> = useTracks(
    [{ source: Track.Source.Camera, withPlaceholder: true }],
    { onlySubscribed: false },
  );

  const speakerAppParticipants = getCurrentSpeakers(state);
  const speakerIds = new Set(speakerAppParticipants.map((p) => p.id));

  const speakerTracks = tracks.filter((t) => speakerIds.has(t.participant.identity));
  const otherTracks = tracks.filter((t) => !speakerIds.has(t.participant.identity));

  const fallbackLines: PlayLine[] = speakerTracks.length === 0 ? getCurrentLines(state) : [];

  return (
    <View className="flex-1 min-h-0 gap-2 p-2 md:gap-3 md:p-3">
      {/* Spotlight zone */}
      <View className="flex-1 min-h-0">
        {speakerTracks.length === 0 ? (
          <SpotlightEmpty lines={fallbackLines} />
        ) : (
          <View className="flex-1 min-h-0 flex-col gap-2 md:flex-row md:gap-3">
            {speakerTracks.map((trackRef) => (
              <SpotlightTile
                key={trackRef.participant.identity}
                trackRef={trackRef}
                lines={getCurrentLinesForParticipant(state, trackRef.participant.identity)}
                isMine={trackRef.participant.identity === participantId}
              />
            ))}
          </View>
        )}
      </View>

      {/* Strip of non-speaking participants */}
      {otherTracks.length > 0 ? (
        <View className="flex-row flex-wrap gap-2">
          {otherTracks.map((trackRef) => (
            <MiniTile key={trackRef.participant.identity} trackRef={trackRef} />
          ))}
        </View>
      ) : null}
    </View>
  );
}

function SpotlightTile({
  trackRef,
  lines,
  isMine,
}: {
  trackRef: object;
  lines: PlayLine[];
  isMine: boolean;
}) {
  const { ParticipantTile } = require('@livekit/components-react');
  return (
    <View className="relative flex-1 min-h-0 overflow-hidden rounded-md border-2 border-stage-gold bg-stage-surface">
      <View className="absolute inset-0">
        <ParticipantTile trackRef={trackRef} />
      </View>
      {lines.length > 0 ? <LineOverlay lines={lines} isMine={isMine} /> : null}
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

function MiniTile({ trackRef }: { trackRef: object }) {
  const { ParticipantTile } = require('@livekit/components-react');
  return (
    <View className="h-20 w-32 overflow-hidden rounded-md border border-stage-border bg-stage-surface md:h-24 md:w-40">
      <ParticipantTile trackRef={trackRef} />
    </View>
  );
}

function LineOverlay({ lines, isMine }: { lines: PlayLine[]; isMine: boolean }) {
  // Keyed by line text so the teleprompter / scroll container resets when the
  // line changes, putting the start of the new line back at the top.
  const key = lines.map((l) => l.text).join('|');

  if (isMine) {
    return (
      <View
        key={key}
        className="absolute inset-x-0 bottom-0 bg-black/65 px-4 py-3 md:px-6 md:py-4"
      >
        {lines.map((line, i) => (
          <View key={i} className={i > 0 ? 'mt-2' : undefined}>
            <TeleprompterLine
              variant="overlay"
              text={line.text}
              direction={line.direction}
              wpm={line.autoScrollWpm ?? DEFAULT_WPM}
              autoScroll={(line.scrollMode ?? 'auto') === 'auto'}
            />
          </View>
        ))}
      </View>
    );
  }

  return (
    <View
      key={key}
      className="absolute inset-x-0 bottom-0 max-h-[60%] bg-black/65 px-4 py-3 md:px-6 md:py-4"
      style={{ overflowY: 'auto' } as object}
    >
      {lines.map((line, i) => (
        <View key={i} className={i > 0 ? 'mt-2' : undefined}>
          {line.direction ? (
            <Text className="italic text-stage-muted text-xs">({line.direction})</Text>
          ) : null}
          <Text className="text-base leading-snug text-stage-gold md:text-lg xl:text-xl 2xl:text-2xl">
            {line.text}
          </Text>
        </View>
      ))}
    </View>
  );
}
