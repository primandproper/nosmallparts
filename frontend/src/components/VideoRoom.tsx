// Built from LiveKit's lower-level primitives rather than its prebuilt
// <VideoConference />, to spotlight whoever is speaking and lay their current
// line over their tile, teleprompter style.

import {
  LiveKitRoom,
  ParticipantTile,
  RoomAudioRenderer,
  type TrackReferenceOrPlaceholder,
  useLocalParticipant,
  useTracks,
} from '@livekit/components-react';
import { Track } from 'livekit-client';
import { useEffect } from 'react';

import { type PlayLine, ScrollMode } from '../generated/nosmallparts/scripts/v1/scripts';
import type { StageState } from '../generated/nosmallparts/stage/v1/stage';
import { type LiveKitCredentials, useLiveKitToken } from '../hooks/useLiveKitToken';
import { getCurrentLines, getCurrentLinesForParticipant, getCurrentSpeakers } from '../lib/stage-logic';
import { TeleprompterLine } from './TeleprompterLine';

const DEFAULT_WPM = 150;

interface Props {
  stageId: string;
  participantId: string | null;
  state: StageState;
}

export function VideoRoom({ stageId, participantId, state }: Props) {
  const { credentials, loading, error, unavailable } = useLiveKitToken({ stageId, participantId });

  const me = state.participants.find((p) => p.id === participantId);
  // Mic and camera start off (§7.1); performers opt in from the control bar.
  // The audience gets a subscribe-only token and no controls.
  const canPublish = !!me && !me.isAudience;
  const forceMuted = !!me?.forceMuted;

  if (unavailable) {
    return (
      <div className="flex flex-1 items-center justify-center rounded-md border border-stage-border p-3">
        <p className="text-xs text-stage-muted">Video disabled (LiveKit not configured)</p>
      </div>
    );
  }
  if (loading || !credentials) {
    return (
      <div className="flex flex-1 items-center justify-center rounded-md border border-stage-border p-3">
        <p className="text-xs text-stage-muted">{error ? `Video unavailable: ${error}` : 'Connecting video…'}</p>
      </div>
    );
  }

  return (
    <Room
      credentials={credentials}
      state={state}
      participantId={participantId}
      canPublish={canPublish}
      forceMuted={forceMuted}
    />
  );
}

function Room({
  credentials,
  state,
  participantId,
  canPublish,
  forceMuted,
}: {
  credentials: LiveKitCredentials;
  state: StageState;
  participantId: string | null;
  canPublish: boolean;
  forceMuted: boolean;
}) {
  return (
    <LiveKitRoom
      token={credentials.token}
      serverUrl={credentials.url}
      connect
      // Nothing publishes on its own (§7.1); both stay off until the user opts in.
      video={false}
      audio={false}
      className="flex min-h-0 flex-1 flex-col"
    >
      <RoomAudioRenderer />
      <SpotlightStage state={state} participantId={participantId} />
      {canPublish ? <MediaControls forceMuted={forceMuted} /> : null}
    </LiveKitRoom>
  );
}

function MediaControls({ forceMuted }: { forceMuted: boolean }) {
  const { localParticipant, isMicrophoneEnabled, isCameraEnabled } = useLocalParticipant();

  // Host force-mute (§4.8): the server unpublishes on the SFU. Mirror it here
  // so the track does not come back on reconnect, and lock the toggle.
  useEffect(() => {
    if (forceMuted && isMicrophoneEnabled) void localParticipant.setMicrophoneEnabled(false);
  }, [forceMuted, isMicrophoneEnabled, localParticipant]);

  const micOn = isMicrophoneEnabled && !forceMuted;

  return (
    <div className="flex items-center justify-center gap-2 border-t border-stage-border bg-stage-surface px-3 py-2">
      <button
        type="button"
        onClick={() => void localParticipant.setMicrophoneEnabled(!micOn)}
        disabled={forceMuted}
        aria-label={forceMuted ? 'Microphone muted by host' : micOn ? 'Turn microphone off' : 'Turn microphone on'}
        className={`rounded-md border px-3 py-1.5 text-xs font-semibold ${
          forceMuted
            ? 'border-red-500 text-red-400 opacity-50'
            : micOn
              ? 'border-stage-gold bg-stage-goldDim text-stage-bg'
              : 'border-stage-border text-stage-text'
        }`}
      >
        {forceMuted ? '🔇 Muted by host' : micOn ? '🎙 Mic on' : '🎙 Mic off'}
      </button>
      <button
        type="button"
        onClick={() => void localParticipant.setCameraEnabled(!isCameraEnabled)}
        aria-label={isCameraEnabled ? 'Turn camera off' : 'Turn camera on'}
        className={`rounded-md border px-3 py-1.5 text-xs font-semibold ${
          isCameraEnabled ? 'border-stage-gold bg-stage-goldDim text-stage-bg' : 'border-stage-border text-stage-text'
        }`}
      >
        {isCameraEnabled ? '📷 Camera on' : '📷 Camera off'}
      </button>
    </div>
  );
}

function SpotlightStage({ state, participantId }: { state: StageState; participantId: string | null }) {
  const tracks = useTracks([{ source: Track.Source.Camera, withPlaceholder: true }], { onlySubscribed: false });

  const speakerIds = new Set(getCurrentSpeakers(state).map((p) => p.id));
  const speakerTracks = tracks.filter((t) => speakerIds.has(t.participant.identity));
  const otherTracks = tracks.filter((t) => !speakerIds.has(t.participant.identity));

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-2 p-2 md:gap-3 md:p-3">
      <div className="flex min-h-0 flex-1 flex-col">
        {speakerTracks.length === 0 ? (
          <SpotlightEmpty lines={getCurrentLines(state)} />
        ) : (
          <div className="flex min-h-0 flex-1 flex-col gap-2 md:flex-row md:gap-3">
            {speakerTracks.map((trackRef) => (
              <SpotlightTile
                key={trackRef.participant.identity}
                trackRef={trackRef}
                lines={getCurrentLinesForParticipant(state, trackRef.participant.identity)}
                isMine={trackRef.participant.identity === participantId}
              />
            ))}
          </div>
        )}
      </div>

      {otherTracks.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {otherTracks.map((trackRef) => (
            <div
              key={trackRef.participant.identity}
              className="h-20 w-32 overflow-hidden rounded-md border border-stage-border bg-stage-surface md:h-24 md:w-40"
            >
              <ParticipantTile trackRef={trackRef} />
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}

function SpotlightTile({
  trackRef,
  lines,
  isMine,
}: {
  trackRef: TrackReferenceOrPlaceholder;
  lines: PlayLine[];
  isMine: boolean;
}) {
  return (
    <div className="relative min-h-0 flex-1 overflow-hidden rounded-md border-2 border-stage-gold bg-stage-surface">
      <div className="absolute inset-0">
        <ParticipantTile trackRef={trackRef} />
      </div>
      {lines.length > 0 ? <LineOverlay lines={lines} isMine={isMine} /> : null}
    </div>
  );
}

function SpotlightEmpty({ lines }: { lines: PlayLine[] }) {
  return (
    <div className="flex min-h-0 flex-1 flex-col items-center justify-center rounded-md border border-stage-border bg-stage-surface p-4">
      <p className="mb-2 text-xs uppercase tracking-widest text-stage-muted">No speaker</p>
      {lines.map((line, i) => (
        <p key={i} className="text-center text-lg text-stage-muted">
          {line.text}
        </p>
      ))}
    </div>
  );
}

function LineOverlay({ lines, isMine }: { lines: PlayLine[]; isMine: boolean }) {
  // Keyed by the line text so the scroll position resets to the top when the
  // line changes.
  const key = lines.map((l) => l.text).join('|');

  if (isMine) {
    return (
      <div key={key} className="absolute inset-x-0 bottom-0 bg-black/65 px-4 py-3 md:px-6 md:py-4">
        {lines.map((line, i) => (
          <div key={i} className={i > 0 ? 'mt-2' : undefined}>
            <TeleprompterLine
              variant="overlay"
              text={line.text}
              direction={line.direction}
              wpm={line.autoScrollWpm ?? DEFAULT_WPM}
              autoScroll={line.scrollMode !== ScrollMode.SCROLL_MODE_MANUAL}
            />
          </div>
        ))}
      </div>
    );
  }

  return (
    <div
      key={key}
      className="absolute inset-x-0 bottom-0 max-h-[60%] overflow-y-auto bg-black/65 px-4 py-3 md:px-6 md:py-4"
    >
      {lines.map((line, i) => (
        <div key={i} className={i > 0 ? 'mt-2' : undefined}>
          {line.direction ? <p className="text-xs italic text-stage-muted">({line.direction})</p> : null}
          <p className="text-base leading-snug text-stage-gold md:text-lg xl:text-xl 2xl:text-2xl">{line.text}</p>
        </div>
      ))}
    </div>
  );
}
