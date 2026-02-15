'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import {
  StageState,
  Play,
  PlayLine,
  CastMember,
  Participant,
  ServerMessage,
  ChatMessage,
  getConfig,
} from '@/lib/types';

// The Go backend sends PlayScript objects over WebSocket with a different shape
// than the frontend Play type. These functions convert between the two formats.

interface BackendCastMember {
  id: string;
  name: string;
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
  // Frontend Play fields (in case data is already in frontend format)
  cast?: CastMember[];
}

function convertBackendPlay(raw: BackendPlayScript): Play {
  // If the data already has 'cast', it's likely already in frontend format
  const cast: CastMember[] = raw.cast
    ? raw.cast
    : (raw.castMembers || [])
        .sort((a, b) => a.sortOrder - b.sortOrder)
        .map((cm) => ({ id: cm.id, name: cm.name }));

  const lines: PlayLine[] = (raw.lines || []).map((line) => {
    // Handle both backend format (lineText/lineIndex) and frontend format (text/index)
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

  const wsRef = useRef<WebSocket | null>(null);
  const isConnectingRef = useRef(false);

  const sendMessage = useCallback((message: object) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      console.log('[WS] Sending:', message);
      wsRef.current.send(JSON.stringify(message));
    } else {
      console.warn('[WS] Cannot send - not connected');
    }
  }, []);

  useEffect(() => {
    if (!stageId || !participantName || isConnectingRef.current) return;

    isConnectingRef.current = true;

    const handleMessage = (event: MessageEvent) => {
      try {
        const message: ServerMessage = JSON.parse(event.data);
        console.log('[WS] Received:', message.type, message);

        switch (message.type) {
          case 'joined':
            {
              const state = convertStageState(message.stageState);
              console.log(
                '[WS] Joined stage! participantId:',
                message.participantId,
                'participants:',
                state.participants.length
              );
              setParticipantId(message.participantId);
              setStageState(state);
            }
            break;
          case 'stageState':
            {
              const state = convertStageState(message.state);
              console.log(
                '[WS] State update, participants:',
                state.participants.map((p: any) => ({
                  name: p.name,
                  castMemberIds: p.castMemberIds,
                }))
              );
              setStageState(state);
            }
            break;
          case 'availablePlays':
            setAvailablePlays((message.plays || []).map((play: any) => convertBackendPlay(play)));
            break;
          case 'castMembersUpdated':
            setStageState((prev) =>
              prev
                ? {
                    ...prev,
                    participants: (message.participants || []).map((p: any) => ({
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
            break;
          case 'playStarted':
            // State will be updated via stageState message
            break;
          case 'lineAdvanced':
            setStageState((prev) =>
              prev ? { ...prev, currentLineIndex: message.currentLineIndex } : null
            );
            break;
          case 'playEnded':
            setStageState((prev) => (prev ? { ...prev, phase: 'discussion' } : null));
            break;
          case 'countdownStart':
            setStageState((prev) => (prev ? { ...prev, phase: 'countdown' } : null));
            break;
          case 'chatMessage':
            setChatMessages((prev) => [...prev, message.message]);
            break;
          case 'error':
            setError(message.message);
            break;
        }
      } catch (err) {
        console.error('Failed to parse message:', err);
      }
    };

    const wsUrl = getConfig().wsUrl;
    console.log('[WS] Connecting to:', wsUrl, 'stageId:', stageId, 'as:', participantName);
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('[WS] Connected! Joining stage:', stageId);
      setIsConnected(true);
      setError(null);
      // Join the stage
      ws.send(
        JSON.stringify({
          type: 'join',
          stageId,
          participantName,
        })
      );
    };

    ws.onmessage = handleMessage;

    ws.onclose = (event: CloseEvent) => {
      console.log('[WS] Stages WebSocket CLOSED', {
        code: event.code,
        reason: event.reason || '(none)',
        wasClean: event.wasClean,
        timestamp: new Date().toISOString(),
      });
      setIsConnected(false);
      isConnectingRef.current = false;
    };

    ws.onerror = () => {
      console.log('[WS] Stages WebSocket ERROR', { timestamp: new Date().toISOString() });
      setError('Connection error');
      isConnectingRef.current = false;
    };

    return () => {
      console.log('[WS] Effect cleanup: closing stages WebSocket (stageId/participant changed or unmount)', {
        timestamp: new Date().toISOString(),
      });
      ws.close();
      wsRef.current = null;
      isConnectingRef.current = false;
    };
  }, [stageId, participantName]);

  const toggleCastMember = useCallback(
    (castMemberId: string) => {
      if (!participantId) return;
      sendMessage({
        type: 'toggleCastMember',
        participantId,
        castMemberId,
      });
    },
    [participantId, sendMessage]
  );

  const setDirector = useCallback(
    (isDirector: boolean) => {
      if (!participantId) return;
      sendMessage({
        type: 'setDirector',
        participantId,
        isDirector,
      });
    },
    [participantId, sendMessage]
  );

  const setReady = useCallback(
    (isReady: boolean) => {
      if (!participantId) return;
      sendMessage({
        type: 'setReady',
        participantId,
        isReady,
      });
    },
    [participantId, sendMessage]
  );

  const setAudience = useCallback(
    (isAudience: boolean) => {
      if (!participantId) return;
      sendMessage({
        type: 'setAudience',
        participantId,
        isAudience,
      });
    },
    [participantId, sendMessage]
  );

  const selectPlay = useCallback(
    (playId: string) => {
      sendMessage({
        type: 'selectPlay',
        playId,
      });
    },
    [sendMessage]
  );

  const sendChat = useCallback(
    (text: string, castOnly: boolean) => {
      if (!participantId) return;
      sendMessage({
        type: 'sendChat',
        participantId,
        text,
        castOnly,
      });
    },
    [participantId, sendMessage]
  );

  const startPlay = useCallback(() => {
    sendMessage({ type: 'startPlay' });
  }, [sendMessage]);

  const nextLine = useCallback(() => {
    sendMessage({ type: 'nextLine' });
  }, [sendMessage]);

  const resetPlay = useCallback(() => {
    sendMessage({ type: 'resetPlay' });
  }, [sendMessage]);

  return {
    stageState,
    participantId,
    availablePlays,
    chatMessages,
    isConnected,
    error,
    toggleCastMember,
    setDirector,
    setAudience,
    setReady,
    selectPlay,
    sendChat,
    startPlay,
    nextLine,
    resetPlay,
  };
}
