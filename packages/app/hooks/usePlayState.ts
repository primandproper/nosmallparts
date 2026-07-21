import { useCallback, useEffect, useRef, useState } from 'react';
import type {
  ChatMessage,
  Play,
  RoleOffer,
  RoleVacancy,
  ServerMessage,
  StageState,
  SystemMessage,
} from '@nsp/shared';
import { getWsUrl } from '../lib/config';
import { ensureIdentity, getSessionToken } from '../lib/identity';

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
  systemMessages: SystemMessage[];
  roleOffers: RoleOffer[];
  roleVacancies: RoleVacancy[];
  isConnected: boolean;
  error: string | null;
  wasReplaced: boolean;
  reconnect: () => void;
  toggleCastMember: (castMemberId: string) => void;
  setAudience: (isAudience: boolean) => void;
  setReady: (isReady: boolean) => void;
  selectPlay: (playId: string) => void;
  sendChat: (text: string, castOnly: boolean) => void;
  startPlay: () => void;
  nextLine: () => void;
  resetPlay: () => void;
  // Host actions
  hostPause: () => void;
  hostResume: () => void;
  hostBackLine: () => void;
  hostSkipLine: () => void;
  hostEndPlay: () => void;
  hostReassignRole: (targetParticipantId: string, castMemberId: string) => void;
  hostUnassignRole: (targetParticipantId: string, castMemberId: string) => void;
  hostTransferHost: (targetParticipantId: string) => void;
  hostKick: (targetParticipantId: string) => void;
  hostBan: (targetParticipantId: string) => void;
  hostMute: (targetParticipantId: string, muted: boolean) => void;
  hostToggleDirectorMode: (enabled: boolean) => void;
  respondRoleOffer: (offerId: string, accept: boolean) => void;
  audienceClaimRole: (castMemberId: string) => void;
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
  const [systemMessages, setSystemMessages] = useState<SystemMessage[]>([]);
  const [roleOffers, setRoleOffers] = useState<RoleOffer[]>([]);
  const [roleVacancies, setRoleVacancies] = useState<RoleVacancy[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [wasReplaced, setWasReplaced] = useState(false);
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
    let ws: WebSocket | null = null;
    let cancelled = false;

    const scheduleReconnect = () => {
      if (intentionalCloseRef.current) return;
      const delay = Math.min(1000 * 2 ** reconnectAttempt, 30000);
      reconnectTimeoutRef.current = setTimeout(() => {
        reconnectTimeoutRef.current = null;
        setError(null);
        setReconnectAttempt((a) => a + 1);
      }, delay);
    };

    (async () => {
      let identity: Awaited<ReturnType<typeof ensureIdentity>>;
      let token: string | null;
      try {
        [identity, token] = await Promise.all([ensureIdentity(), getSessionToken()]);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : 'Failed to establish identity');
        scheduleReconnect();
        return;
      }
      if (cancelled) return;

      const params = new URLSearchParams();
      if (identity.guestId) params.set('guest', identity.guestId);
      if (token) params.set('token', token);
      const qs = params.toString();
      ws = new WebSocket(getWsUrl('/ws/stage') + (qs ? `?${qs}` : ''));
      wsRef.current = ws;
      attachSocketHandlers(ws);
    })();

    function attachSocketHandlers(ws: WebSocket) {
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
            // Vacancies only exist while performing.
            if (msg.state.phase !== 'performing') setRoleVacancies([]);
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
            setRoleVacancies([]);
            break;
          case 'roleVacated':
            setRoleVacancies((prev) => [
              ...prev.filter((v) => v.castMemberId !== msg.vacancy.castMemberId),
              msg.vacancy,
            ]);
            break;
          case 'roleVacancyCleared':
            setRoleVacancies((prev) => prev.filter((v) => v.castMemberId !== msg.castMemberId));
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
          case 'systemMessage':
            setSystemMessages((prev) => [...prev, msg.message]);
            break;
          case 'roleOffer':
            setRoleOffers((prev) =>
              prev.some((o) => o.offerId === msg.offer.offerId) ? prev : [...prev, msg.offer]
            );
            break;
          case 'roleOfferCleared':
            setRoleOffers((prev) => prev.filter((o) => o.offerId !== msg.offerId));
            break;
          case 'error':
            setError(msg.message);
            break;
          case 'sessionReplaced':
            // Another tab/device joined this stage with the same identity.
            // Stop reconnecting so we don't ping-pong displacing each other.
            intentionalCloseRef.current = true;
            setWasReplaced(true);
            setError(null);
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
    }

    return () => {
      cancelled = true;
      intentionalCloseRef.current = true;
      if (heartbeatRef.current) clearInterval(heartbeatRef.current);
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
      ws?.close();
      wsRef.current = null;
    };
  }, [stageIdOrCode, participantName, reconnectAttempt, enabled]);

  const reconnect = useCallback(() => {
    setError(null);
    setWasReplaced(false);
    setReconnectAttempt((a) => a + 1);
  }, []);

  const toggleCastMember = useCallback(
    (castMemberId: string) => {
      if (!participantId) return;
      send({ type: 'toggleCastMember', participantId, castMemberId });
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
  const hostPause = useCallback(() => send({ type: 'hostPause' }), [send]);
  const hostResume = useCallback(() => send({ type: 'hostResume' }), [send]);
  const hostBackLine = useCallback(() => send({ type: 'hostBackLine' }), [send]);
  const hostSkipLine = useCallback(() => send({ type: 'hostSkipLine' }), [send]);
  const hostEndPlay = useCallback(() => send({ type: 'hostEndPlay' }), [send]);
  const hostReassignRole = useCallback(
    (targetParticipantId: string, castMemberId: string) =>
      send({ type: 'hostReassignRole', targetParticipantId, castMemberId }),
    [send]
  );
  const hostUnassignRole = useCallback(
    (targetParticipantId: string, castMemberId: string) =>
      send({ type: 'hostUnassignRole', targetParticipantId, castMemberId }),
    [send]
  );
  const hostTransferHost = useCallback(
    (targetParticipantId: string) =>
      send({ type: 'hostTransferHost', targetParticipantId }),
    [send]
  );
  const hostKick = useCallback(
    (targetParticipantId: string) => send({ type: 'hostKick', targetParticipantId }),
    [send]
  );
  const hostBan = useCallback(
    (targetParticipantId: string) => send({ type: 'hostBan', targetParticipantId }),
    [send]
  );
  const hostMute = useCallback(
    (targetParticipantId: string, muted: boolean) =>
      send({ type: 'hostMute', targetParticipantId, muted }),
    [send]
  );
  const hostToggleDirectorMode = useCallback(
    (enabled: boolean) => send({ type: 'hostToggleDirectorMode', enabled }),
    [send]
  );
  const respondRoleOffer = useCallback(
    (offerId: string, accept: boolean) =>
      send({ type: 'respondRoleOffer', offerId, accept }),
    [send]
  );
  const audienceClaimRole = useCallback(
    (castMemberId: string) => send({ type: 'audienceClaimRole', castMemberId }),
    [send]
  );

  return {
    stageState,
    participantId,
    availablePlays,
    chatMessages,
    systemMessages,
    roleOffers,
    roleVacancies,
    isConnected,
    error,
    wasReplaced,
    reconnect,
    toggleCastMember,
    setAudience,
    setReady,
    selectPlay,
    sendChat,
    startPlay,
    nextLine,
    resetPlay,
    hostPause,
    hostResume,
    hostBackLine,
    hostSkipLine,
    hostEndPlay,
    hostReassignRole,
    hostUnassignRole,
    hostTransferHost,
    hostKick,
    hostBan,
    hostMute,
    hostToggleDirectorMode,
    respondRoleOffer,
    audienceClaimRole,
  };
}
