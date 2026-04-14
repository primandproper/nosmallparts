import { useCallback, useEffect, useRef, useState } from 'react';
import type { ChatMessage, Play, ServerMessage, StageState } from '@nsp/shared';
import { getWsUrl } from '../lib/config';

interface Options {
  stageIdOrCode: string;
  participantName: string;
  enabled?: boolean;
}

export interface UsePlayStateReturn {
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

const HEARTBEAT_MS = 5000;

export function usePlayState({
  stageIdOrCode,
  participantName,
  enabled = true,
}: Options): UsePlayStateReturn {
  const [stageState, setStageState] = useState<StageState | null>(null);
  const [participantId, setParticipantId] = useState<string | null>(null);
  const [availablePlays, setAvailablePlays] = useState<Play[]>([]);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reconnectAttempt, setReconnectAttempt] = useState(0);

  const wsRef = useRef<WebSocket | null>(null);
  const intentionalCloseRef = useRef(false);
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const send = useCallback((msg: object) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  useEffect(() => {
    if (!enabled || !stageIdOrCode || !participantName) return;

    intentionalCloseRef.current = false;
    const ws = new WebSocket(getWsUrl('/ws/stage'));
    wsRef.current = ws;

    const scheduleReconnect = () => {
      if (intentionalCloseRef.current) return;
      const delay = Math.min(1000 * 2 ** reconnectAttempt, 30000);
      reconnectTimeoutRef.current = setTimeout(() => {
        reconnectTimeoutRef.current = null;
        setError(null);
        setReconnectAttempt((a) => a + 1);
      }, delay);
    };

    ws.onopen = () => {
      setIsConnected(true);
      setError(null);
      ws.send(
        JSON.stringify({
          type: 'join',
          stageId: stageIdOrCode,
          participantName,
        })
      );
    };

    ws.onmessage = (event) => {
      let msg: ServerMessage;
      try {
        msg = JSON.parse(event.data);
      } catch {
        return;
      }
      switch (msg.type) {
        case 'joined':
          setParticipantId(msg.participantId);
          setStageState(msg.stageState);
          if (heartbeatRef.current) clearInterval(heartbeatRef.current);
          heartbeatRef.current = setInterval(() => {
            if (wsRef.current?.readyState === WebSocket.OPEN) {
              wsRef.current.send(JSON.stringify({ type: 'heartbeat' }));
            }
          }, HEARTBEAT_MS);
          break;
        case 'stageState':
          setStageState(msg.state);
          break;
        case 'availablePlays':
          setAvailablePlays(msg.plays);
          break;
        case 'castMembersUpdated':
          setStageState((prev) => (prev ? { ...prev, participants: msg.participants } : null));
          break;
        case 'lineAdvanced':
          setStageState((prev) =>
            prev ? { ...prev, currentLineIndex: msg.currentLineIndex } : null
          );
          break;
        case 'playEnded':
          setStageState((prev) => (prev ? { ...prev, phase: 'discussion' } : null));
          break;
        case 'countdownStart':
          setStageState((prev) => (prev ? { ...prev, phase: 'countdown' } : null));
          break;
        case 'playStarted':
          // full stageState follows
          break;
        case 'chatMessage':
          setChatMessages((prev) => [...prev, msg.message]);
          break;
        case 'error':
          setError(msg.message);
          break;
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      if (!intentionalCloseRef.current) scheduleReconnect();
    };

    ws.onerror = () => {
      setIsConnected(false);
      if (!intentionalCloseRef.current) scheduleReconnect();
    };

    return () => {
      intentionalCloseRef.current = true;
      if (heartbeatRef.current) clearInterval(heartbeatRef.current);
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
      ws.close();
      wsRef.current = null;
    };
  }, [stageIdOrCode, participantName, reconnectAttempt, enabled]);

  const toggleCastMember = useCallback(
    (castMemberId: string) => {
      if (!participantId) return;
      send({ type: 'toggleCastMember', participantId, castMemberId });
    },
    [participantId, send]
  );
  const setDirector = useCallback(
    (isDirector: boolean) => {
      if (!participantId) return;
      send({ type: 'setDirector', participantId, isDirector });
    },
    [participantId, send]
  );
  const setAudience = useCallback(
    (isAudience: boolean) => {
      if (!participantId) return;
      send({ type: 'setAudience', participantId, isAudience });
    },
    [participantId, send]
  );
  const setReady = useCallback(
    (isReady: boolean) => {
      if (!participantId) return;
      send({ type: 'setReady', participantId, isReady });
    },
    [participantId, send]
  );
  const selectPlay = useCallback(
    (playId: string) => {
      send({ type: 'selectPlay', playId });
    },
    [send]
  );
  const sendChat = useCallback(
    (text: string, castOnly: boolean) => {
      if (!participantId) return;
      send({ type: 'sendChat', participantId, text, castOnly });
    },
    [participantId, send]
  );
  const startPlay = useCallback(() => send({ type: 'startPlay' }), [send]);
  const nextLine = useCallback(() => send({ type: 'nextLine' }), [send]);
  const resetPlay = useCallback(() => send({ type: 'resetPlay' }), [send]);

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
