/**
 * WebSocket handler for stage real-time communication.
 * Receives messages from browsers, calls backend via gRPC, broadcasts to participants.
 */

import { WebSocket } from 'ws';
import { getStagesClient } from './grpc-client';
import { type StagePhase, StagePhase as StagePhaseEnum } from '../generated/stages/stages_messages';

// Connection map: stageId -> participantId -> { ws, participantId }
const connections = new Map<string, Map<string, { ws: WebSocket; participantId: string }>>();

function getStagePhaseString(phase: StagePhase): string {
  switch (phase) {
    case 1:
      return 'lobby';
    case 2:
      return 'countdown';
    case 3:
      return 'performing';
    case 4:
      return 'discussion';
    default:
      return 'lobby';
  }
}

/** Convert gRPC StageState to frontend-expected shape (phase as string) */
function toFrontendStageState(st: { phase?: StagePhase } | null | undefined): unknown {
  if (!st) return null;
  return {
    ...st,
    phase: getStagePhaseString((st.phase ?? 0) as StagePhase),
  };
}

function broadcastToStage(stageId: string, msg: object, excludeParticipantId?: string) {
  const stageConns = connections.get(stageId);
  if (!stageConns) return;
  const payload = JSON.stringify(msg);
  Array.from(stageConns.entries()).forEach(([pid, { ws }]) => {
    if (pid !== excludeParticipantId && ws.readyState === WebSocket.OPEN) {
      ws.send(payload);
    }
  });
}

function sendToClient(ws: WebSocket, msg: object) {
  if (ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify(msg));
  }
}

function sendError(ws: WebSocket, message: string) {
  sendToClient(ws, { type: 'error', message });
}

export function handleStageWebSocket(ws: WebSocket, _req: { url?: string }) {
  let stageId: string | null = null;
  let participantId: string | null = null;

  const cleanup = () => {
    if (stageId && participantId) {
      const stageConns = connections.get(stageId);
      if (stageConns) {
        stageConns.delete(participantId);
        if (stageConns.size === 0) connections.delete(stageId);
      }
      getStagesClient()
        .leaveStage({ stageId, participantId })
        .then(() => getStagesClient().getStageState({ stageId: stageId! }))
        .then((res) => {
          const state = toFrontendStageState(res.stageState);
          broadcastToStage(stageId!, { type: 'stageState', state });
        })
        .catch(() => {});
    }
  };

  ws.on('close', cleanup);

  ws.on('message', async (data: Buffer) => {
    try {
      const msg = JSON.parse(data.toString()) as { type: string; [k: string]: unknown };
      const client = getStagesClient();

      switch (msg.type) {
        case 'join': {
          const stageIdParam = String(msg.stageId || '');
          const participantName = String(msg.participantName || '');
          const res = await client.joinStage({
            stageId: stageIdParam,
            participantName,
          });
          stageId = res.stageState?.id ?? stageIdParam;
          participantId = res.participantId;

          if (!stageId || !participantId) {
            sendError(ws, 'Failed to join stage');
            return;
          }

          if (!connections.has(stageId)) {
            connections.set(stageId, new Map());
          }
          connections.get(stageId)!.set(participantId, { ws, participantId });

          const state = toFrontendStageState(res.stageState);
          sendToClient(ws, {
            type: 'joined',
            participantId: res.participantId,
            stageState: state,
          });
          sendToClient(ws, { type: 'availablePlays', plays: res.plays ?? [] });
          broadcastToStage(stageId, { type: 'stageState', state }, participantId);
          break;
        }

        case 'heartbeat': {
          if (!stageId) return;
          const res = await client.getStageState({ stageId });
          const state = toFrontendStageState(res.stageState);
          sendToClient(ws, { type: 'stageState', state });
          break;
        }

        case 'toggleCastMember': {
          if (!stageId || !participantId) return;
          const res = await client.toggleRole({
            stageId,
            participantId: String(msg.participantId || ''),
            castMemberId: String(msg.castMemberId || ''),
          });
          const state = toFrontendStageState(res.stageState);
          broadcastToStage(stageId, {
            type: 'castMembersUpdated',
            participants: res.participants ?? [],
          });
          broadcastToStage(stageId, { type: 'stageState', state });
          break;
        }

        case 'setDirector':
        case 'setAudience':
        case 'setReady': {
          if (!stageId) return;
          const res = await client.updateParticipant({
            stageId,
            participantId: String(msg.participantId || ''),
            isDirector: msg.type === 'setDirector' ? (msg.isDirector as boolean) : undefined,
            isAudience: msg.type === 'setAudience' ? (msg.isAudience as boolean) : undefined,
            isReady: msg.type === 'setReady' ? (msg.isReady as boolean) : undefined,
          });
          const state = toFrontendStageState(res.stageState);
          broadcastToStage(stageId, { type: 'stageState', state });
          break;
        }

        case 'selectPlay': {
          if (!stageId) return;
          const res = await client.updateStage({
            stageId,
            playScriptId: String(msg.playId || ''),
          });
          const state = toFrontendStageState(res.stageState);
          broadcastToStage(stageId, { type: 'stageState', state });
          break;
        }

        case 'sendChat': {
          if (!stageId) return;
          const res = await client.sendChat({
            stageId,
            participantId: String(msg.participantId || ''),
            text: String(msg.text || ''),
            castOnly: Boolean(msg.castOnly),
          });
          const chatMsg = res.message;
          if (chatMsg) {
            const castOnly = Boolean(chatMsg.castOnly);
            if (castOnly) {
              const stateRes = await client.getStageState({ stageId });
              const participants = stateRes.stageState?.participants ?? [];
              const nonAudienceIds = new Set(
                participants.filter((p) => !p.isAudience).map((p) => p.id)
              );
              const payload = JSON.stringify({ type: 'chatMessage', message: chatMsg });
              const stageConns = connections.get(stageId);
              if (stageConns) {
                Array.from(stageConns.entries()).forEach(([pid, { ws: w }]) => {
                  if (nonAudienceIds.has(pid) && w.readyState === WebSocket.OPEN) {
                    w.send(payload);
                  }
                });
              }
            } else {
              broadcastToStage(stageId, { type: 'chatMessage', message: chatMsg });
            }
          }
          break;
        }

        case 'startPlay': {
          if (!stageId) return;
          await client.startPlay({ stageId });
          const res = await client.getStageState({ stageId });
          const state = toFrontendStageState(res.stageState);
          broadcastToStage(stageId, { type: 'countdownStart', duration: 3 });
          broadcastToStage(stageId, { type: 'stageState', state });

          setTimeout(async () => {
            await client.updateStage({
              stageId: stageId!,
              phase: StagePhaseEnum.STAGE_PHASE_PERFORMING,
              currentLineIndex: 0,
            });
            const res2 = await client.getStageState({ stageId: stageId! });
            broadcastToStage(stageId!, { type: 'playStarted' });
            broadcastToStage(stageId!, {
              type: 'stageState',
              state: toFrontendStageState(res2.stageState),
            });
          }, 3000);
          break;
        }

        case 'nextLine': {
          if (!stageId) return;
          const res = await client.nextLine({ stageId });
          const state = toFrontendStageState(res.stageState);
          if (res.playEnded) {
            broadcastToStage(stageId, { type: 'playEnded', nextPlay: res.nextPlay });
          } else if (res.lineAdvanced) {
            broadcastToStage(stageId, {
              type: 'lineAdvanced',
              currentLineIndex: res.currentLineIndex,
            });
          }
          broadcastToStage(stageId, { type: 'stageState', state });
          break;
        }

        case 'resetPlay': {
          if (!stageId) return;
          const res = await client.resetPlay({ stageId });
          const state = toFrontendStageState(res.stageState);
          broadcastToStage(stageId, { type: 'stageState', state });
          break;
        }

        default:
          sendError(ws, 'Unknown message type');
      }
    } catch (err) {
      console.error('[WS Stage] Error:', err);
      const message = err instanceof Error ? err.message : 'Internal error';
      sendError(ws, message);
    }
  });
}
