'use client';

import { useEffect, useRef, useState } from 'react';
import {
  LiveKitRoom,
  GridLayout,
  ParticipantTile,
  RoomAudioRenderer,
  useTracks,
  useParticipants,
  useLocalParticipant,
} from '@livekit/components-react';
import { DisconnectReason, Track, RoomEvent } from 'livekit-client';
import {
  StageState,
  Participant,
  getCastMemberById,
  getCurrentSpeakers,
  getCurrentLines,
  getCastMemberNameForLine,
} from '@/lib/types';

interface VideoRoomProps {
  stageState: StageState;
  participantId: string | null;
  livekitToken: string | null;
  livekitUrl: string | null;
}

// Audience placeholder - shown to audience members instead of video
function AudienceView({ stageState }: { stageState: StageState }) {
  const performers = (stageState.participants || []).filter((p) => !p.isAudience);
  const audienceCount = (stageState.participants || []).filter((p) => p.isAudience).length;

  return (
    <div className="card-elegant rounded-2xl overflow-hidden min-h-[400px] h-[500px] relative">
      <div className="h-full flex flex-col items-center justify-center p-8 text-center">
        <div className="text-6xl mb-4">👁️</div>
        <h3 className="text-[var(--gold)] text-2xl font-[Cormorant_Garamond] mb-2">
          Audience Mode
        </h3>
        <p className="text-[var(--cream)] opacity-60 mb-6">
          Your camera is off. Sit back and enjoy the show!
        </p>
        <div className="text-[var(--cream)] opacity-40 text-sm space-y-1">
          <p>
            🎭 {performers.length} performer{performers.length !== 1 ? 's' : ''} on stage
          </p>
          {audienceCount > 1 && (
            <p>
              👥 {audienceCount - 1} other audience member{audienceCount > 2 ? 's' : ''}
            </p>
          )}
        </div>
      </div>

      {/* Audience mode indicator */}
      <div className="absolute top-2 right-2 px-3 py-1.5 rounded-full text-xs bg-purple-500/20 border border-purple-500/30">
        <span className="text-purple-300">👁️ Watching</span>
      </div>
    </div>
  );
}

function LocalVideo() {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [error, setError] = useState<string | null>(null);
  const [hasPermission, setHasPermission] = useState<boolean | null>(null);

  useEffect(() => {
    let stream: MediaStream | null = null;

    const setupVideo = async () => {
      if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        setError('Camera requires HTTPS');
        setHasPermission(false);
        return;
      }

      try {
        stream = await navigator.mediaDevices.getUserMedia({
          video: true,
          audio: false,
        });
        setHasPermission(true);
        if (videoRef.current) {
          videoRef.current.srcObject = stream;
        }
      } catch (err: unknown) {
        setHasPermission(false);
        if (err instanceof Error) {
          if (err.name === 'NotAllowedError') {
            setError('Camera permission denied');
          } else if (err.name === 'NotFoundError') {
            setError('No camera found');
          } else {
            setError('Could not access camera');
          }
        }
      }
    };

    setupVideo();

    return () => {
      if (stream) {
        stream.getTracks().forEach((track) => track.stop());
      }
    };
  }, []);

  if (error) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-[var(--black-light)]">
        <div className="text-center">
          <div className="text-4xl mb-3 opacity-50">📷</div>
          <div className="text-[var(--cream)] opacity-40 text-sm">{error}</div>
        </div>
      </div>
    );
  }

  if (hasPermission === null) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-[var(--black-light)]">
        <div className="text-center">
          <div className="text-4xl mb-3 animate-pulse">📷</div>
          <div className="text-[var(--cream)] opacity-40 text-sm">Requesting camera...</div>
        </div>
      </div>
    );
  }

  return <video ref={videoRef} autoPlay playsInline muted className="w-full h-full object-cover" />;
}

// LiveKit grid that respects stage state - only shows speakers during performance
function LiveKitGrid({ stageState }: { stageState: StageState }) {
  const allTracks = useTracks(
    [
      { source: Track.Source.Camera, withPlaceholder: true },
      { source: Track.Source.ScreenShare, withPlaceholder: false },
    ],
    { onlySubscribed: false }
  );

  // Filter out audience members from tracks
  const audienceIds = (stageState.participants || []).filter((p) => p.isAudience).map((p) => p.id);
  const tracks = allTracks.filter((track) => !audienceIds.includes(track.participant.identity));

  const isPerforming = stageState.phase === 'performing';
  const currentSpeakers = getCurrentSpeakers(stageState);
  const currentLines = getCurrentLines(stageState);

  if (tracks.length === 0) {
    return (
      <div className="h-full flex items-center justify-center">
        <div className="text-center">
          <div className="text-4xl mb-2 animate-pulse">📹</div>
          <p className="text-[var(--cream)] opacity-40">Connecting to video...</p>
        </div>
      </div>
    );
  }

  // During performance, only show current speakers
  if (isPerforming && currentSpeakers.length > 0) {
    // Find the tracks that belong to current speakers
    // LiveKit participant identity = our participant ID
    const speakerIds = currentSpeakers.map((s) => s.id);
    const speakerTracks = tracks.filter((track) => speakerIds.includes(track.participant.identity));

    // If we can't match tracks to speakers, show all (fallback)
    const tracksToShow = speakerTracks.length > 0 ? speakerTracks : tracks;
    const gridCols = tracksToShow.length === 1 ? 1 : 2;

    return (
      <div
        className="grid gap-4 h-full p-4"
        style={{ gridTemplateColumns: `repeat(${gridCols}, 1fr)` }}
      >
        {tracksToShow.map((track) => {
          // Find the speaker info
          const speaker = currentSpeakers.find((s) => s.id === track.participant.identity);
          const speakerLines = speaker
            ? currentLines.filter((l) => (speaker.castMemberIds || []).includes(l.castMemberId))
            : [];
          const castNames = speakerLines
            .map((l) =>
              stageState.selectedPlay ? getCastMemberNameForLine(stageState.selectedPlay, l) : ''
            )
            .filter(Boolean);

          return (
            <div key={track.participant.identity} className="video-tile speaking relative">
              <ParticipantTile trackRef={track} className="w-full h-full" />
              <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-4">
                <div className="flex items-center gap-2">
                  {castNames.length > 0 && (
                    <span className="text-[var(--gold)] font-medium">{castNames.join(' & ')}</span>
                  )}
                  <span className="text-[var(--cream)] opacity-60 text-sm">
                    ({track.participant.name || track.participant.identity})
                  </span>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    );
  }

  // Lobby/discussion mode - show everyone
  const gridCols = tracks.length === 1 ? 1 : tracks.length <= 4 ? 2 : 3;

  return (
    <div
      className="grid gap-3 h-full p-4"
      style={{ gridTemplateColumns: `repeat(${gridCols}, 1fr)` }}
    >
      {tracks.map((track) => {
        // Try to find this participant in our state to get their role
        const participant = stageState.participants.find(
          (p) => p.id === track.participant.identity
        );
        const castMembers =
          participant && stageState.selectedPlay
            ? (participant.castMemberIds || [])
                .map((id) => getCastMemberById(stageState.selectedPlay!, id))
                .filter(Boolean)
            : [];

        return (
          <div key={track.participant.identity} className="video-tile relative">
            <ParticipantTile trackRef={track} className="w-full h-full" />
            <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-3">
              <div className="flex flex-col gap-0.5">
                <div className="flex items-center justify-between">
                  <span className="text-[var(--cream)] text-sm font-medium">
                    {track.participant.name || track.participant.identity}
                  </span>
                  {castMembers.length > 0 && (
                    <span className="text-[var(--gold)] text-xs">
                      {castMembers.map((cm) => cm!.name).join(', ')}
                    </span>
                  )}
                </div>
                {castMembers.some((cm) => cm!.description) && (
                  <p className="text-[var(--cream)] opacity-50 text-xs">
                    {castMembers
                      .map((cm) => cm!.description)
                      .filter(Boolean)
                      .join(' • ')}
                  </p>
                )}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

// LiveKit-powered video room
function LiveKitVideoRoom({
  stageState,
  livekitToken,
  livekitUrl,
}: {
  stageState: StageState;
  livekitToken: string;
  livekitUrl: string;
}) {
  const [connectionState, setConnectionState] = useState<string>('connecting');

  return (
    <div className="card-elegant rounded-2xl overflow-hidden min-h-[400px] h-[500px] relative">
      <LiveKitRoom
        token={livekitToken}
        serverUrl={livekitUrl}
        connect={true}
        video={true}
        audio={true}
        onConnected={() => {
          console.log('LiveKit connected!');
          setConnectionState('connected');
        }}
        onDisconnected={(reason) => {
          console.log('LiveKit disconnected', {
            reason: reason != null ? (DisconnectReason[reason] ?? reason) : undefined,
            raw: reason,
          });
          setConnectionState('disconnected');
        }}
        onError={(error) => {
          console.error('LiveKit error:', error);
          setConnectionState('error');
        }}
        style={{ height: '100%' }}
      >
        <LiveKitGrid stageState={stageState} />
        <RoomAudioRenderer />
      </LiveKitRoom>

      {/* Connection status indicator */}
      <div className="absolute top-2 right-2 px-2 py-1 rounded text-xs bg-black/50 z-10">
        <span
          className={`inline-block w-2 h-2 rounded-full mr-2 ${
            connectionState === 'connected'
              ? 'bg-green-500'
              : connectionState === 'connecting'
                ? 'bg-yellow-500 animate-pulse'
                : 'bg-red-500'
          }`}
        />
        <span className="text-[var(--cream)] opacity-60">{connectionState}</span>
      </div>
    </div>
  );
}

function MockVideoRoom({
  stageState,
  participantId,
}: {
  stageState: StageState;
  participantId: string | null;
}) {
  // Filter out audience members - they don't show on camera
  const performers = (stageState.participants || []).filter((p) => !p.isAudience);
  const isPerforming = stageState.phase === 'performing';
  const currentSpeakers = getCurrentSpeakers(stageState);
  const currentLines = getCurrentLines(stageState);

  // During performance, only show current speakers (who are never audience)
  const participantsToShow =
    isPerforming && currentSpeakers.length > 0 ? currentSpeakers : performers;

  if (participantsToShow.length === 0) {
    return (
      <div className="card-elegant rounded-2xl overflow-hidden min-h-[400px] h-[500px]">
        <div className="h-full flex items-center justify-center p-4">
          <div className="text-center">
            <div className="text-4xl mb-2">👥</div>
            <p className="text-[var(--cream)] opacity-40">Waiting for participants...</p>
          </div>
        </div>
      </div>
    );
  }

  const gridCols = participantsToShow.length === 1 ? 1 : participantsToShow.length <= 4 ? 2 : 3;

  return (
    <div className="card-elegant rounded-2xl overflow-hidden min-h-[400px] h-[500px] relative">
      <div
        className="grid gap-3 h-full p-4 auto-rows-fr"
        style={{ gridTemplateColumns: `repeat(${gridCols}, 1fr)` }}
      >
        {participantsToShow.map((participant) => {
          const isMe = participant.id === participantId;

          // Get cast member info depending on mode
          let displayNames: string[] = [];
          let castMembers:
            | { id: string; name: string; description?: string }[]
            | undefined;
          if (isPerforming) {
            const speakerLines = currentLines.filter((l) =>
              (participant.castMemberIds || []).includes(l.castMemberId)
            );
            displayNames = speakerLines
              .map((l) =>
                stageState.selectedPlay ? getCastMemberNameForLine(stageState.selectedPlay, l) : ''
              )
              .filter(Boolean);
          } else if (stageState.selectedPlay) {
            castMembers = (participant.castMemberIds || [])
              .map((id) => getCastMemberById(stageState.selectedPlay!, id))
              .filter(Boolean) as { id: string; name: string; description?: string }[];
            displayNames = castMembers.map((cm) => cm.name);
          }

          return (
            <div
              key={participant.id}
              className={`video-tile relative ${isPerforming ? 'speaking' : ''}`}
            >
              {isMe ? (
                <LocalVideo />
              ) : (
                <div className="w-full h-full bg-gradient-to-br from-[var(--black-lighter)] to-[var(--black)] flex items-center justify-center">
                  <div className="text-center">
                    <div className="w-16 h-16 rounded-full bg-[var(--gold)]/20 flex items-center justify-center mx-auto mb-2">
                      <span className="text-2xl text-[var(--gold)]">
                        {participant.name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                    <div className="text-[var(--cream)] opacity-60 text-sm">{participant.name}</div>
                  </div>
                </div>
              )}
              <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-3">
                <div className="flex flex-col gap-0.5">
                  <div className="flex items-center justify-between">
                    <span className="text-[var(--cream)] text-sm font-medium">
                      {participant.name}
                      {isMe && <span className="text-[var(--gold)] opacity-60 ml-1">(you)</span>}
                    </span>
                    {displayNames.length > 0 && (
                      <span className="text-[var(--gold)] text-xs">{displayNames.join(', ')}</span>
                    )}
                  </div>
                  {castMembers?.some((cm) => cm.description) && (
                    <p className="text-[var(--cream)] opacity-50 text-xs">
                      {castMembers
                        .map((cm) => cm.description)
                        .filter(Boolean)
                        .join(' • ')}
                    </p>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Mock mode indicator */}
      <div className="absolute top-2 right-2 px-2 py-1 rounded text-xs bg-black/50">
        <span className="text-[var(--cream)] opacity-60">Mock Mode</span>
      </div>
    </div>
  );
}

export function VideoRoom({ stageState, participantId, livekitToken, livekitUrl }: VideoRoomProps) {
  const hasLiveKit = !!(livekitToken && livekitUrl);
  const mountedAt = useRef(Date.now());

  useEffect(() => {
    console.log('[VideoRoom] Mounted', {
      hasLiveKit,
      participantId,
      stageId: stageState?.id,
      timestamp: new Date().toISOString(),
    });
    return () => {
      console.log('[VideoRoom] Unmounting (duration ms:', Date.now() - mountedAt.current, ')', {
        hasLiveKit,
        timestamp: new Date().toISOString(),
      });
    };
  }, []); // eslint-disable-line react-hooks/exhaustive-deps -- only log mount/unmount
  // Check if current participant is audience
  const currentParticipant = (stageState.participants || []).find((p) => p.id === participantId);
  const isAudience = currentParticipant?.isAudience || false;

  // Audience members see a special view without camera
  if (isAudience) {
    return <AudienceView stageState={stageState} />;
  }

  // Use LiveKit if we have valid credentials
  if (livekitToken && livekitUrl) {
    return (
      <LiveKitVideoRoom
        stageState={stageState}
        livekitToken={livekitToken}
        livekitUrl={livekitUrl}
      />
    );
  }

  // Fallback to mock mode
  return <MockVideoRoom stageState={stageState} participantId={participantId} />;
}
