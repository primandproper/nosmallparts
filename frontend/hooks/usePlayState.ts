'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import Pusher, { Channel, PresenceChannel } from 'pusher-js';
import {
  StageState,
  Play,
  PlayLine,
  CastMember,
  Participant,
  ChatMessage,
} from '@/lib/types';
import * as stageApi from '@/lib/stage-api';

// The Go backend sends PlayScript objects with a different shape
// than the frontend Play type. These functions convert between the two formats.

interface BackendCastMember {
  id: string;
  name: string;
  description?: string;
  sortOrder: number;
}

interface BackendLine {
  castMemberId: string;
  lineText: string;
  lineIndex: number;
  direction?: string;
  scrollMode?: string;
  autoScrollWpm?: number;
}

interface BackendPlayScript {
  id: string;
  title: string;
  description: string;
  author?: string;
  castMembers?: BackendCastMember[];
  lines?: BackendLine[];
  cast?: CastMember[];
}

function convertBackendPlay(raw: BackendPlayScript): Play {
  const cast: CastMember[] = raw.cast
    ? raw.cast
    : (raw.castMembers || [])
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map((cm) => ({
          id: cm.id,
          name: cm.name,
          ...(cm.description && { description: cm.description }),
        }));

  const lines: PlayLine[] = (raw.lines || []).map((line) => {
    const result: PlayLine = {
      castMemberId: line.castMemberId,
      text: (line as any).text || line.lineText || '',
      index: (line as any).index ?? line.lineIndex ?? 0,
    };
    if (line.direction) {
      result.direction = line.direction;
    }
    if (line.scrollMode === 'manual') {
      result.scrollMode = 'manual';
    } else if (line.scrollMode === 'auto') {
      result.scrollMode = 'auto';
    }
    if (line.autoScrollWpm) {
      result.autoScrollWpm = line.autoScrollWpm;
    }
    return result;
  });

  return {
    id: raw.id,
    title: raw.title,
    description: raw.description,
    author: raw.author,
    cast,
    lines,
  };
}

function convertStageState(raw: any): StageState {
  return {
    id: raw.id,
    code: raw.code || '',
    phase: raw.phase,
    currentLineIndex: raw.currentLineIndex ?? 0,
    participants: (raw.participants || []).map((p: any) => ({
      id: p.id,
      name: p.name,
      castMemberIds: p.castMemberIds || [],
      isDirector: p.isDirector ?? false,
      isReady: p.isReady ?? false,
      isAudience: p.isAudience ?? false,
    })),
    selectedPlay: raw.selectedPlay ? convertBackendPlay(raw.selectedPlay) : null,
  };
}

interface UsePlayStateOptions {
  stageId: string;
  participantName: string;
}

interface UsePlayStateReturn {
  stageState: StageState | null;
  participantId: string | null;
  availablePlays: Play[];
  chatMessages: ChatMessage[];
  isConnected: boolean;
  error: string | null;
  toggleCastMember: (castMemberId: string) => void;
  setDirector: (isDirector: boolean) => void;
  setAudience: (isAudience: boolean) => void;
  setReady: (isReady: boolean) => void;
  selectPlay: (playId: string) => void;
  sendChat: (text: string, castOnly: boolean) => void;
  startPlay: () => void;
  nextLine: () => void;
  resetPlay: () => void;
}

const POLL_INTERVAL_MS = 10_000;

export function usePlayState({
  stageId,
  participantName,
}: UsePlayStateOptions): UsePlayStateReturn {
  const [stageState, setStageState] = useState<StageState | null>(null);
  const [participantId, setParticipantId] = useState<string | null>(null);
  const [availablePlays, setAvailablePlays] = useState<Play[]>([]);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const pusherRef = useRef<Pusher | null>(null);
  const stageChannelRef = useRef<PresenceChannel | null>(null);
  const castChannelRef = useRef<Channel | null>(null);
  const pollIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const stageIdRef = useRef<string>('');
  const participantIdRef = useRef<string | null>(null);

  // Keep refs in sync
  useEffect(() => {
    participantIdRef.current = participantId;
  }, [participantId]);

  useEffect(() => {
    stageIdRef.current = stageId;
  }, [stageId]);

  // Join stage and set up Pusher
  useEffect(() => {
    if (!stageId || !participantName) return;

    let cancelled = false;

    async function init() {
      try {
        const result = await stageApi.joinStage(stageId, participantName);
        if (cancelled) return;

        const pid = result.participantId;
        setParticipantId(pid);
        participantIdRef.current = pid;

        const state = convertStageState(result.stageState);
        setStageState(state);

        if (result.availablePlays) {
          setAvailablePlays(result.availablePlays.map((p: any) => convertBackendPlay(p)));
        }

        // Set up Pusher if we have credentials
        if (result.pusherKey) {
          const pusher = new Pusher(result.pusherKey, {
            cluster: result.pusherCluster,
            authEndpoint: `${getApiUrlForPusher()}/api/pusher/auth`,
            auth: {
              params: { participant_id: pid },
            },
          });

          pusherRef.current = pusher;

          pusher.connection.bind('connected', () => {
            setIsConnected(true);
            setError(null);
          });
          pusher.connection.bind('disconnected', () => setIsConnected(false));
          pusher.connection.bind('error', () => setIsConnected(false));

          // Subscribe to presence channel
          const stageChannel = pusher.subscribe(`presence-stage-${stageId}`) as PresenceChannel;
          stageChannelRef.current = stageChannel;

          stageChannel.bind('stage-state', (data: any) => {
            setStageState(convertStageState(data));
          });

          stageChannel.bind('countdown-start', () => {
            setStageState((prev) => (prev ? { ...prev, phase: 'countdown' } : null));
          });

          stageChannel.bind('cast-members-updated', (data: any) => {
            setStageState((prev) =>
              prev
                ? {
                    ...prev,
                    participants: (data.participants || []).map((p: any) => ({
                      id: p.id,
                      name: p.name,
                      castMemberIds: p.castMemberIds || [],
                      isDirector: p.isDirector ?? false,
                      isReady: p.isReady ?? false,
                      isAudience: p.isAudience ?? false,
                    })),
                  }
                : null
            );
          });

          stageChannel.bind('chat-message', (data: any) => {
            if (data.message) {
              setChatMessages((prev) => [...prev, data.message]);
            }
          });

          stageChannel.bind('play-started', () => {
            // State will be updated via stage-state event
          });

          stageChannel.bind('line-advanced', (data: any) => {
            setStageState((prev) =>
              prev ? { ...prev, currentLineIndex: data.currentLineIndex } : null
            );
          });

          stageChannel.bind('play-ended', () => {
            setStageState((prev) => (prev ? { ...prev, phase: 'discussion' } : null));
          });

          // Subscribe to cast channel if not audience
          const currentParticipant = state.participants.find((p) => p.id === pid);
          if (!currentParticipant?.isAudience) {
            const castCh = pusher.subscribe(`private-cast-${stageId}`);
            castChannelRef.current = castCh;
            castCh.bind('chat-message', (data: any) => {
              if (data.message) {
                setChatMessages((prev) => [...prev, data.message]);
              }
            });
          }
        } else {
          // No Pusher credentials - still mark as connected for local dev
          setIsConnected(true);
        }

        // Start polling for state reconciliation
        pollIntervalRef.current = setInterval(async () => {
          try {
            const stateData = await stageApi.getStageState(stageId);
            setStageState(convertStageState(stateData));
          } catch {
            // Ignore polling errors
          }
        }, POLL_INTERVAL_MS);
      } catch (err) {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : 'Failed to join stage');
      }
    }

    init();

    return () => {
      cancelled = true;

      if (pollIntervalRef.current) {
        clearInterval(pollIntervalRef.current);
        pollIntervalRef.current = null;
      }

      // Leave stage
      if (participantIdRef.current && stageId) {
        stageApi.leaveStage(stageId, participantIdRef.current).catch(() => {});
      }

      // Disconnect Pusher
      if (pusherRef.current) {
        pusherRef.current.disconnect();
        pusherRef.current = null;
      }
      stageChannelRef.current = null;
      castChannelRef.current = null;
    };
  }, [stageId, participantName]);

  // Manage cast channel subscription based on audience status
  useEffect(() => {
    if (!pusherRef.current || !stageState || !participantId) return;

    const currentParticipant = stageState.participants.find((p) => p.id === participantId);
    if (!currentParticipant) return;

    if (currentParticipant.isAudience && castChannelRef.current) {
      pusherRef.current.unsubscribe(`private-cast-${stageIdRef.current}`);
      castChannelRef.current = null;
    } else if (!currentParticipant.isAudience && !castChannelRef.current && pusherRef.current) {
      const castCh = pusherRef.current.subscribe(`private-cast-${stageIdRef.current}`);
      castChannelRef.current = castCh;
      castCh.bind('chat-message', (data: any) => {
        if (data.message) {
          setChatMessages((prev) => [...prev, data.message]);
        }
      });
    }
  }, [stageState?.participants, participantId]);

  const toggleCastMember = useCallback(
    (castMemberId: string) => {
      if (!participantId || !stageIdRef.current) return;
      stageApi.toggleCastMember(stageIdRef.current, participantId, castMemberId).catch(console.error);
    },
    [participantId]
  );

  const setDirectorCb = useCallback(
    (isDirector: boolean) => {
      if (!participantId || !stageIdRef.current) return;
      stageApi.setDirector(stageIdRef.current, participantId, isDirector).catch(console.error);
    },
    [participantId]
  );

  const setReadyCb = useCallback(
    (isReady: boolean) => {
      if (!participantId || !stageIdRef.current) return;
      stageApi.setReady(stageIdRef.current, participantId, isReady).catch(console.error);
    },
    [participantId]
  );

  const setAudienceCb = useCallback(
    (isAudience: boolean) => {
      if (!participantId || !stageIdRef.current) return;
      stageApi.setAudience(stageIdRef.current, participantId, isAudience).catch(console.error);
    },
    [participantId]
  );

  const selectPlayCb = useCallback(
    (playId: string) => {
      if (!stageIdRef.current) return;
      stageApi.selectPlay(stageIdRef.current, playId).catch(console.error);
    },
    []
  );

  const sendChatCb = useCallback(
    (text: string, castOnly: boolean) => {
      if (!participantId || !stageIdRef.current) return;
      stageApi.sendChat(stageIdRef.current, participantId, text, castOnly).catch(console.error);
    },
    [participantId]
  );

  const startPlayCb = useCallback(() => {
    if (!stageIdRef.current) return;
    stageApi.startPlay(stageIdRef.current).catch(console.error);
  }, []);

  const nextLineCb = useCallback(() => {
    if (!stageIdRef.current) return;
    stageApi.nextLine(stageIdRef.current).catch(console.error);
  }, []);

  const resetPlayCb = useCallback(() => {
    if (!stageIdRef.current) return;
    stageApi.resetPlay(stageIdRef.current).catch(console.error);
  }, []);

  return {
    stageState,
    participantId,
    availablePlays,
    chatMessages,
    isConnected,
    error,
    toggleCastMember,
    setDirector: setDirectorCb,
    setAudience: setAudienceCb,
    setReady: setReadyCb,
    selectPlay: selectPlayCb,
    sendChat: sendChatCb,
    startPlay: startPlayCb,
    nextLine: nextLineCb,
    resetPlay: resetPlayCb,
  };
}

function getApiUrlForPusher(): string {
  const envUrl = typeof process !== 'undefined' ? process.env.NEXT_PUBLIC_API_URL : undefined;
  if (envUrl) return envUrl;

  if (typeof window === 'undefined') return 'http://localhost:8000';

  const { protocol, hostname } = window.location;
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return `${protocol}//${hostname}:8000`;
  }
  return `${protocol}//api.${hostname.replace(/^www\./, '')}`;
}
