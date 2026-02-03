'use client';

import { useState, useEffect, useRef, useCallback } from 'react';
import { StageState, Play, Participant, ServerMessage, ChatMessage, getConfig } from '@/lib/types';

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
            console.log(
              '[WS] Joined stage! participantId:',
              message.participantId,
              'participants:',
              message.stageState.participants.length
            );
            setParticipantId(message.participantId);
            setStageState(message.stageState);
            break;
          case 'stageState':
            console.log(
              '[WS] State update, participants:',
              message.state.participants.map((p: any) => ({
                name: p.name,
                castMemberIds: p.castMemberIds,
              }))
            );
            setStageState(message.state);
            break;
          case 'availablePlays':
            setAvailablePlays(message.plays);
            break;
          case 'castMembersUpdated':
            setStageState((prev) =>
              prev ? { ...prev, participants: message.participants } : null
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

    ws.onclose = () => {
      setIsConnected(false);
      isConnectingRef.current = false;
    };

    ws.onerror = () => {
      setError('Connection error');
      isConnectingRef.current = false;
    };

    return () => {
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
