import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import type { Play } from '../generated/nosmallparts/scripts/v1/scripts';
import {
  type ChatMessage,
  ClientMessage,
  type RoleOffer,
  type RoleVacancy,
  ServerMessage,
  StagePhase,
  type StageState,
  type SystemMessage,
} from '../generated/nosmallparts/stage/v1/stage';
import { ensureGuestIdentity } from '../lib/identity';
import { accessToken } from '../lib/session';

interface Options {
  stageIdOrCode: string;
  participantName: string;
  enabled?: boolean;
}

/** A transient, dismissable notice (e.g. a rejected action while already joined). */
export interface Toast {
  id: number;
  message: string;
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
  toasts: Toast[];
  dismissToast: (id: number) => void;
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

// The socket's subprotocol, and how an access token rides beside it: a browser
// cannot set an Authorization header on a WebSocket.
const SUBPROTOCOL = 'nsp.v1';
const BEARER_PREFIX = 'bearer.';

const HEARTBEAT_MS = 5000;
// Transient toasts auto-dismiss and never pile up beyond this many at once.
const TOAST_MS = 4000;
const MAX_TOASTS = 3;

function socketURL(): string {
  const scheme = window.location.protocol === 'https:' ? 'wss' : 'ws';
  return `${scheme}://${window.location.host}/ws/stage`;
}

export function usePlayState({ stageIdOrCode, participantName, enabled = true }: Options): UsePlayStateReturn {
  const [stageState, setStageState] = useState<StageState | null>(null);
  const [participantId, setParticipantId] = useState<string | null>(null);
  const [availablePlays, setAvailablePlays] = useState<Play[]>([]);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [systemMessages, setSystemMessages] = useState<SystemMessage[]>([]);
  const [roleOffers, setRoleOffers] = useState<RoleOffer[]>([]);
  const [roleVacancies, setRoleVacancies] = useState<RoleVacancy[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [toasts, setToasts] = useState<Toast[]>([]);
  const [wasReplaced, setWasReplaced] = useState(false);
  const [reconnectAttempt, setReconnectAttempt] = useState(0);

  const wsRef = useRef<WebSocket | null>(null);
  const intentionalCloseRef = useRef(false);
  const heartbeatRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  // True once joined; decides whether a server error is a dismissable toast
  // (in the stage) or the connection error (before joining).
  const joinedRef = useRef(false);
  const toastIdRef = useRef(0);
  const toastTimersRef = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map());

  const dismissToast = useCallback((id: number) => {
    const timer = toastTimersRef.current.get(id);
    if (timer) {
      clearTimeout(timer);
      toastTimersRef.current.delete(id);
    }
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const pushToast = useCallback((message: string) => {
    const id = ++toastIdRef.current;
    setToasts((prev) => [...prev, { id, message }].slice(-MAX_TOASTS));
    const timer = setTimeout(() => {
      toastTimersRef.current.delete(id);
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, TOAST_MS);
    toastTimersRef.current.set(id, timer);
  }, []);

  // Clear any pending toast timers when the hook unmounts.
  useEffect(() => {
    const timers = toastTimersRef.current;
    return () => {
      for (const timer of timers.values()) clearTimeout(timer);
      timers.clear();
    };
  }, []);

  const send = useCallback((msg: ClientMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(ClientMessage.toJSON(msg)));
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

    const handle = (msg: ServerMessage) => {
      if (msg.joined) {
        joinedRef.current = true;
        setParticipantId(msg.joined.participantId);
        setStageState(msg.joined.stageState ?? null);
        if (heartbeatRef.current) clearInterval(heartbeatRef.current);
        heartbeatRef.current = setInterval(() => send({ heartbeat: {} }), HEARTBEAT_MS);
      } else if (msg.stageState) {
        const state = msg.stageState;
        setStageState(state);
        // Vacancies only exist while performing.
        if (state.phase !== StagePhase.STAGE_PHASE_PERFORMING) setRoleVacancies([]);
      } else if (msg.availablePlays) {
        setAvailablePlays(msg.availablePlays.plays);
      } else if (msg.castMembersUpdated) {
        const participants = msg.castMembersUpdated.participants;
        setStageState((prev) => (prev ? { ...prev, participants } : null));
      } else if (msg.lineAdvanced) {
        const currentLineIndex = msg.lineAdvanced.currentLineIndex;
        setStageState((prev) => (prev ? { ...prev, currentLineIndex } : null));
      } else if (msg.playEnded) {
        setStageState((prev) => (prev ? { ...prev, phase: StagePhase.STAGE_PHASE_DISCUSSION } : null));
        setRoleVacancies([]);
      } else if (msg.roleVacated) {
        const vacancy = msg.roleVacated;
        setRoleVacancies((prev) => [...prev.filter((v) => v.castMemberId !== vacancy.castMemberId), vacancy]);
      } else if (msg.roleVacancyCleared) {
        const castMemberId = msg.roleVacancyCleared.castMemberId;
        setRoleVacancies((prev) => prev.filter((v) => v.castMemberId !== castMemberId));
      } else if (msg.countdownStart) {
        setStageState((prev) => (prev ? { ...prev, phase: StagePhase.STAGE_PHASE_COUNTDOWN } : null));
      } else if (msg.chatMessage) {
        const chat = msg.chatMessage;
        setChatMessages((prev) => [...prev, chat]);
      } else if (msg.systemMessage) {
        const system = msg.systemMessage;
        setSystemMessages((prev) => [...prev, system]);
      } else if (msg.roleOffer) {
        const offer = msg.roleOffer;
        setRoleOffers((prev) => (prev.some((o) => o.offerId === offer.offerId) ? prev : [...prev, offer]));
      } else if (msg.roleOfferCleared) {
        const offerId = msg.roleOfferCleared.offerId;
        setRoleOffers((prev) => prev.filter((o) => o.offerId !== offerId));
      } else if (msg.error) {
        // Once joined, an action refused ("Not your line", host-only actions)
        // is transient, so it is a toast rather than the sticky connection
        // error. Refusals before joining (name policy, a full stage, a ban)
        // still drive the join screen.
        if (joinedRef.current) pushToast(msg.error.message);
        else setError(msg.error.message);
      } else if (msg.sessionReplaced) {
        // Another tab joined this stage as the same person. Stop reconnecting,
        // or the two would take turns displacing each other.
        intentionalCloseRef.current = true;
        joinedRef.current = false;
        setWasReplaced(true);
        setError(null);
      }
    };

    const attach = (socket: WebSocket) => {
      socket.onopen = () => {
        setIsConnected(true);
        setError(null);
        send({ join: { stageId: stageIdOrCode, participantName } });
      };

      socket.onmessage = (event) => {
        let msg: ServerMessage;
        try {
          msg = ServerMessage.fromJSON(JSON.parse(String(event.data)));
        } catch {
          return;
        }
        handle(msg);
      };

      socket.onclose = () => {
        joinedRef.current = false;
        setIsConnected(false);
        if (heartbeatRef.current) clearInterval(heartbeatRef.current);
        if (!intentionalCloseRef.current) scheduleReconnect();
      };
    };

    (async () => {
      let token: string | null;
      try {
        // A guest needs their cookie before the socket; a signed-in user needs
        // a current access token for it.
        [, token] = await Promise.all([ensureGuestIdentity(), accessToken()]);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : 'Failed to establish identity');
        scheduleReconnect();
        return;
      }
      if (cancelled) return;

      const protocols = token ? [SUBPROTOCOL, BEARER_PREFIX + token] : [SUBPROTOCOL];
      ws = new WebSocket(socketURL(), protocols);
      wsRef.current = ws;
      attach(ws);
    })();

    return () => {
      cancelled = true;
      intentionalCloseRef.current = true;
      if (heartbeatRef.current) clearInterval(heartbeatRef.current);
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
      ws?.close();
      wsRef.current = null;
    };
  }, [stageIdOrCode, participantName, reconnectAttempt, enabled, send, pushToast]);

  const reconnect = useCallback(() => {
    setError(null);
    setWasReplaced(false);
    setReconnectAttempt((a) => a + 1);
  }, []);

  // The actions are stable for a participant, so a component can depend on
  // one in an effect (the lobby's auto-start does) without re-running it on
  // every render.
  const actions = useMemo(() => {
    // The messages naming a participant name this socket's own, and wait
    // until it has one.
    const self = participantId;

    return {
      toggleCastMember: (castMemberId: string) => {
        if (self) send({ toggleCastMember: { participantId: self, castMemberId } });
      },
      setAudience: (isAudience: boolean) => {
        if (self) send({ setAudience: { participantId: self, isAudience } });
      },
      setReady: (isReady: boolean) => {
        if (self) send({ setReady: { participantId: self, isReady } });
      },
      selectPlay: (playId: string) => send({ selectPlay: { playId } }),
      sendChat: (text: string, castOnly: boolean) => {
        if (self) send({ sendChat: { participantId: self, text, castOnly } });
      },
      startPlay: () => send({ startPlay: {} }),
      nextLine: () => send({ nextLine: {} }),
      resetPlay: () => send({ resetPlay: {} }),
      hostPause: () => send({ hostPause: {} }),
      hostResume: () => send({ hostResume: {} }),
      hostBackLine: () => send({ hostBackLine: {} }),
      hostSkipLine: () => send({ hostSkipLine: {} }),
      hostEndPlay: () => send({ hostEndPlay: {} }),
      hostReassignRole: (targetParticipantId: string, castMemberId: string) =>
        send({ hostReassignRole: { targetParticipantId, castMemberId } }),
      hostUnassignRole: (targetParticipantId: string, castMemberId: string) =>
        send({ hostUnassignRole: { targetParticipantId, castMemberId } }),
      hostTransferHost: (targetParticipantId: string) => send({ hostTransferHost: { targetParticipantId } }),
      hostKick: (targetParticipantId: string) => send({ hostKick: { targetParticipantId } }),
      hostBan: (targetParticipantId: string) => send({ hostBan: { targetParticipantId } }),
      hostMute: (targetParticipantId: string, muted: boolean) => send({ hostMute: { targetParticipantId, muted } }),
      hostToggleDirectorMode: (enabled: boolean) => send({ hostToggleDirectorMode: { enabled } }),
      respondRoleOffer: (offerId: string, accept: boolean) => send({ respondRoleOffer: { offerId, accept } }),
      audienceClaimRole: (castMemberId: string) => send({ audienceClaimRole: { castMemberId } }),
    };
  }, [participantId, send]);

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
    toasts,
    dismissToast,
    wasReplaced,
    reconnect,
    ...actions,
  };
}
