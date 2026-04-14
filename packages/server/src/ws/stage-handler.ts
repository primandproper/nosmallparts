/**
 * WebSocket handler for real-time stage sessions.
 * Maintains an in-memory connection map, persists state via Drizzle, and
 * broadcasts updates to all connected participants.
 */
import type { Context } from 'hono';
import type { WSContext, WSEvents } from 'hono/ws';
import { ulid } from 'ulid';
import { listPlayScripts } from '../repos/play-scripts.js';
import {
  getStageByIdOrCode,
  getStageState,
  incrementLine,
  joinStage,
  leaveStage,
  resetStage,
  setSelectedPlay,
  setStagePhase,
  toggleParticipantRole,
  updateParticipantFlags,
} from '../repos/stages.js';

type Conn = { ws: WSContext; participantId: string };

// stageId -> participantId -> connection
const connections = new Map<string, Map<string, Conn>>();

function registerConn(stageId: string, participantId: string, ws: WSContext) {
  if (!connections.has(stageId)) connections.set(stageId, new Map());
  connections.get(stageId)!.set(participantId, { ws, participantId });
}

function removeConn(stageId: string, participantId: string) {
  const stage = connections.get(stageId);
  if (!stage) return;
  stage.delete(participantId);
  if (stage.size === 0) connections.delete(stageId);
}

function broadcastToStage(stageId: string, msg: object, opts: { excludeParticipantId?: string } = {}) {
  const stage = connections.get(stageId);
  if (!stage) return;
  const payload = JSON.stringify(msg);
  for (const [pid, { ws }] of stage) {
    if (pid !== opts.excludeParticipantId) {
      try {
        ws.send(payload);
      } catch {
        // ignore send failures
      }
    }
  }
}

function sendToClient(ws: WSContext, msg: object) {
  try {
    ws.send(JSON.stringify(msg));
  } catch {
    // ignore
  }
}

function sendError(ws: WSContext, message: string) {
  sendToClient(ws, { type: 'error', message });
}

async function broadcastCurrentState(stageId: string, excludeParticipantId?: string) {
  const state = await getStageState(stageId);
  if (!state) return;
  broadcastToStage(stageId, { type: 'stageState', state }, { excludeParticipantId });
}

export function createStageWsHandler(): (c: Context) => Promise<WSEvents> {
  return async () => {
    let stageId: string | null = null;
    let participantId: string | null = null;

    return {
      async onMessage(event, ws) {
        let msg: { type: string; [k: string]: unknown };
        try {
          msg = JSON.parse(
            typeof event.data === 'string' ? event.data : new TextDecoder().decode(event.data as ArrayBuffer)
          );
        } catch {
          sendError(ws, 'Invalid JSON');
          return;
        }

        try {
          switch (msg.type) {
            case 'join': {
              const idOrCode = String(msg.stageId || '');
              const name = String(msg.participantName || '').trim() || 'Performer';
              const stage = await getStageByIdOrCode(idOrCode);
              if (!stage) {
                sendError(ws, 'Stage not found');
                return;
              }
              const joined = await joinStage(stage.id, name, null);
              stageId = joined.stageId;
              participantId = joined.participantId;
              registerConn(stageId, participantId, ws);

              const state = await getStageState(stageId);
              if (!state) {
                sendError(ws, 'Failed to load stage state');
                return;
              }
              sendToClient(ws, { type: 'joined', participantId, stageState: state });
              // Send list of public plays the user can choose from.
              const plays = await listPlayScripts({ userId: null });
              sendToClient(ws, { type: 'availablePlays', plays });
              // Notify others of new participant.
              broadcastToStage(stageId, { type: 'stageState', state }, { excludeParticipantId: participantId });
              break;
            }

            case 'heartbeat': {
              if (!stageId) return;
              const state = await getStageState(stageId);
              if (state) sendToClient(ws, { type: 'stageState', state });
              break;
            }

            case 'toggleCastMember': {
              if (!stageId || !participantId) return;
              await toggleParticipantRole(
                stageId,
                String(msg.participantId),
                String(msg.castMemberId)
              );
              await broadcastCurrentState(stageId);
              break;
            }

            case 'setDirector':
            case 'setAudience':
            case 'setReady': {
              if (!stageId) return;
              const patch: { isDirector?: boolean; isAudience?: boolean; isReady?: boolean } = {};
              if (msg.type === 'setDirector') patch.isDirector = Boolean(msg.isDirector);
              if (msg.type === 'setAudience') patch.isAudience = Boolean(msg.isAudience);
              if (msg.type === 'setReady') patch.isReady = Boolean(msg.isReady);
              await updateParticipantFlags(String(msg.participantId), patch);
              await broadcastCurrentState(stageId);
              break;
            }

            case 'selectPlay': {
              if (!stageId) return;
              const playId = String(msg.playId || '') || null;
              await setSelectedPlay(stageId, playId);
              await broadcastCurrentState(stageId);
              break;
            }

            case 'sendChat': {
              if (!stageId) return;
              const chat = {
                id: ulid(),
                senderId: String(msg.participantId || ''),
                senderName: await participantNameFor(stageId, String(msg.participantId || '')),
                text: String(msg.text || ''),
                castOnly: Boolean(msg.castOnly),
                timestamp: Date.now(),
              };
              if (chat.castOnly) {
                // Only broadcast to non-audience participants.
                const state = await getStageState(stageId);
                if (!state) return;
                const allowed = new Set(state.participants.filter((p) => !p.isAudience).map((p) => p.id));
                const conns = connections.get(stageId);
                if (conns) {
                  const payload = JSON.stringify({ type: 'chatMessage', message: chat });
                  for (const [pid, { ws: w }] of conns) {
                    if (allowed.has(pid)) {
                      try {
                        w.send(payload);
                      } catch {
                        /* ignore */
                      }
                    }
                  }
                }
              } else {
                broadcastToStage(stageId, { type: 'chatMessage', message: chat });
              }
              break;
            }

            case 'startPlay': {
              if (!stageId) return;
              const sid = stageId;
              await setStagePhase(sid, 'countdown', 0);
              const countdownState = await getStageState(sid);
              if (countdownState) {
                broadcastToStage(sid, { type: 'countdownStart', duration: 3 });
                broadcastToStage(sid, { type: 'stageState', state: countdownState });
              }
              setTimeout(async () => {
                await setStagePhase(sid, 'performing', 0);
                const perfState = await getStageState(sid);
                if (perfState) {
                  broadcastToStage(sid, { type: 'playStarted' });
                  broadcastToStage(sid, { type: 'stageState', state: perfState });
                }
              }, 3000);
              break;
            }

            case 'nextLine': {
              if (!stageId) return;
              const { currentLineIndex, playEnded } = await incrementLine(stageId);
              if (playEnded) {
                broadcastToStage(stageId, { type: 'playEnded' });
              } else {
                broadcastToStage(stageId, { type: 'lineAdvanced', currentLineIndex });
              }
              await broadcastCurrentState(stageId);
              break;
            }

            case 'resetPlay': {
              if (!stageId) return;
              await resetStage(stageId);
              await broadcastCurrentState(stageId);
              break;
            }

            default:
              sendError(ws, `Unknown message type: ${msg.type}`);
          }
        } catch (err) {
          console.error('[ws] error handling', msg.type, err);
          sendError(ws, err instanceof Error ? err.message : 'Internal error');
        }
      },

      async onClose() {
        if (stageId && participantId) {
          const sid = stageId;
          const pid = participantId;
          removeConn(sid, pid);
          await leaveStage(sid, pid).catch(() => {});
          await broadcastCurrentState(sid);
        }
      },

      onError(err) {
        console.error('[ws] connection error', err);
      },
    };
  };
}

async function participantNameFor(stageId: string, participantId: string): Promise<string> {
  const state = await getStageState(stageId);
  return state?.participants.find((p) => p.id === participantId)?.name ?? 'Unknown';
}
