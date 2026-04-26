/**
 * WebSocket handler for real-time stage sessions.
 * Maintains an in-memory connection map, persists state via Drizzle, and
 * broadcasts updates to all connected participants.
 */
import type { Context } from 'hono';
import type { WSContext, WSEvents } from 'hono/ws';
import { ulid } from 'ulid';
import type { AppVariables } from '../app.js';
import { resolveIdentity, type Identity } from '../identity.js';
import { livekitRemoveParticipant, livekitSetMuted } from '../livekit-admin.js';
import { listPlayScripts } from '../repos/play-scripts.js';
import {
  addBan,
  backLine,
  clearParticipantRoles,
  clearStageHost,
  findActiveParticipantByIdentity,
  getParticipant,
  getStageById,
  getStageByIdOrCode,
  getStageHost,
  getStageState,
  identityMatches,
  incrementLine,
  isBanned,
  joinStage,
  leaveStage,
  markParticipantLeft,
  participantDisplayName,
  pickSuccessorIdentity,
  reassignRole,
  resetStage,
  setDirectorMode,
  setForceMuted,
  setPausedByHost,
  setParticipantDisplayName,
  setPlayhead,
  setSelectedPlay,
  setStageHost,
  setStagePhase,
  toggleParticipantRole,
  unassignRole,
  updateParticipantFlags,
  type HostIdentity,
} from '../repos/stages.js';

const HOST_GRACE_MS = 15_000;
const MAX_PARTICIPANTS_PER_STAGE = 12;

type Conn = { ws: WSContext; participantId: string; identity: Identity; displaced?: boolean };

// stageId -> participantId -> connection
const connections = new Map<string, Map<string, Conn>>();

// stageId -> pending succession timer (set while host is disconnected within grace).
const hostGraceTimers = new Map<string, ReturnType<typeof setTimeout>>();

// Mid-play role reassignment offers that are awaiting accept/decline.
// stageId -> offerId -> offer
type PendingOffer = {
  offerId: string;
  castMemberId: string;
  castMemberName: string;
  targetParticipantId: string;
  createdAt: number;
};
const pendingOffers = new Map<string, Map<string, PendingOffer>>();

function addPendingOffer(stageId: string, offer: PendingOffer) {
  if (!pendingOffers.has(stageId)) pendingOffers.set(stageId, new Map());
  pendingOffers.get(stageId)!.set(offer.offerId, offer);
}

function removePendingOffer(stageId: string, offerId: string): PendingOffer | undefined {
  const map = pendingOffers.get(stageId);
  if (!map) return undefined;
  const offer = map.get(offerId);
  if (offer) map.delete(offerId);
  if (map.size === 0) pendingOffers.delete(stageId);
  return offer;
}

function clearOffersForParticipant(stageId: string, participantId: string): PendingOffer[] {
  const map = pendingOffers.get(stageId);
  if (!map) return [];
  const dropped: PendingOffer[] = [];
  for (const [id, offer] of map) {
    if (offer.targetParticipantId === participantId) {
      dropped.push(offer);
      map.delete(id);
    }
  }
  if (map.size === 0) pendingOffers.delete(stageId);
  return dropped;
}

function cancelHostGrace(stageId: string) {
  const t = hostGraceTimers.get(stageId);
  if (t) {
    clearTimeout(t);
    hostGraceTimers.delete(stageId);
  }
}

function isIdentityConnected(stageId: string, identity: HostIdentity): boolean {
  const stage = connections.get(stageId);
  if (!stage) return false;
  for (const { identity: connId } of stage.values()) {
    if (identityMatches(connId, identity)) return true;
  }
  return false;
}

async function isHostIdentity(stageId: string, identity: Identity): Promise<boolean> {
  const host = await getStageHost(stageId);
  if (!host) return false;
  return identityMatches(host, identity);
}

function getConn(stageId: string, participantId: string): Conn | undefined {
  return connections.get(stageId)?.get(participantId);
}

function registerConn(stageId: string, participantId: string, ws: WSContext, identity: Identity) {
  if (!connections.has(stageId)) connections.set(stageId, new Map());
  connections.get(stageId)!.set(participantId, { ws, participantId, identity });
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

function broadcastSystem(stageId: string, text: string) {
  broadcastToStage(stageId, {
    type: 'systemMessage',
    message: { id: ulid(), text, timestamp: Date.now() },
  });
}

async function scheduleHostSuccession(stageId: string, leavingHost: HostIdentity, leavingHostName: string) {
  cancelHostGrace(stageId);
  const timer = setTimeout(() => {
    hostGraceTimers.delete(stageId);
    void promoteSuccessor(stageId, leavingHost, leavingHostName);
  }, HOST_GRACE_MS);
  hostGraceTimers.set(stageId, timer);
}

async function promoteSuccessor(stageId: string, leavingHost: HostIdentity, leavingHostName: string) {
  // If the leaving host has reconnected during grace, nothing to do.
  if (isIdentityConnected(stageId, leavingHost)) return;

  const stageConns = connections.get(stageId);
  const pids = stageConns ? Array.from(stageConns.keys()) : [];
  const successor = await pickSuccessorIdentity(stageId, pids, leavingHost);
  if (!successor) {
    // No one left — clear host, stage is effectively idle.
    await clearStageHost(stageId);
    return;
  }
  await setStageHost(stageId, successor);

  let successorName = 'someone';
  if (stageConns) {
    for (const { participantId, identity } of stageConns.values()) {
      if (identityMatches(identity, successor)) {
        successorName = (await participantDisplayName(participantId)) ?? successorName;
        break;
      }
    }
  }
  broadcastSystem(stageId, `${leavingHostName} left. ${successorName} is now host.`);
  await broadcastCurrentState(stageId);
}

export function createStageWsHandler(): (c: Context<{ Variables: AppVariables }>) => Promise<WSEvents> {
  return async (c) => {
    const identity = await resolveIdentity(c);
    let stageId: string | null = null;
    let participantId: string | null = null;
    let authorized = false;
    let myConn: Conn | null = null;

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
          if (msg.type === 'join') {
            if (!identity.userId && !identity.guestId) {
              sendError(ws, 'Identity required. Call /api/identity before connecting.');
              ws.close();
              return;
            }
            const idOrCode = String(msg.stageId || '');
            const name = String(msg.participantName || '').trim() || 'Performer';
            const stage = await getStageByIdOrCode(idOrCode);
            if (!stage) {
              sendError(ws, 'Stage not found');
              return;
            }
            if (await isBanned(stage.id, identity)) {
              sendError(ws, 'You have been removed from this stage.');
              ws.close();
              return;
            }
            // Reuse an existing participant row if this identity is already on
            // the stage (e.g. a second browser tab). Displace any prior live
            // connection so the same person doesn't appear twice.
            const existing = await findActiveParticipantByIdentity(stage.id, identity);
            if (existing) {
              const prev = getConn(stage.id, existing.id);
              if (prev) {
                prev.displaced = true;
                sendToClient(prev.ws, { type: 'sessionReplaced' });
                try { prev.ws.close(); } catch { /* ignore */ }
              }
              if (existing.displayName !== name) {
                await setParticipantDisplayName(existing.id, name);
              }
              stageId = stage.id;
              participantId = existing.id;
            } else {
              const currentCount = connections.get(stage.id)?.size ?? 0;
              if (currentCount >= MAX_PARTICIPANTS_PER_STAGE) {
                sendError(ws, `Stage is full (${MAX_PARTICIPANTS_PER_STAGE} max).`);
                ws.close();
                return;
              }
              const joined = await joinStage(stage.id, name, identity);
              stageId = joined.stageId;
              participantId = joined.participantId;
            }
            authorized = true;
            registerConn(stageId, participantId, ws, identity);
            myConn = getConn(stageId, participantId) ?? null;

            // Host assignment: if stage has no host, this joiner becomes host.
            // If they ARE the host identity returning during grace, cancel succession.
            const currentHost = await getStageHost(stageId);
            let hostAnnouncement: string | null = null;
            if (!currentHost) {
              await setStageHost(stageId, identity);
              hostAnnouncement = `${name} is host.`;
            } else if (identityMatches(currentHost, identity)) {
              if (hostGraceTimers.has(stageId)) {
                cancelHostGrace(stageId);
                hostAnnouncement = `${name} (host) reconnected.`;
              }
            }

            const state = await getStageState(stageId);
            if (!state) {
              sendError(ws, 'Failed to load stage state');
              return;
            }
            sendToClient(ws, { type: 'joined', participantId, stageState: state });
            const plays = await listPlayScripts({ userId: identity.userId ?? null });
            sendToClient(ws, { type: 'availablePlays', plays });
            broadcastToStage(stageId, { type: 'stageState', state }, { excludeParticipantId: participantId });
            if (hostAnnouncement) broadcastSystem(stageId, hostAnnouncement);
            return;
          }

          if (!authorized || !stageId || !participantId) {
            sendError(ws, 'Must join before sending messages');
            return;
          }

          // Messages that carry a participantId are locked to the connection's own.
          // Host-on-others actions (reassign, kick, etc.) arrive as distinct message types
          // in a later chunk and are gated separately.
          const lockedSelfPid = participantId;

          switch (msg.type) {
            case 'heartbeat': {
              const state = await getStageState(stageId);
              if (state) sendToClient(ws, { type: 'stageState', state });
              break;
            }

            case 'toggleCastMember': {
              // In director mode, the host cannot take a role.
              const stage = await getStageById(stageId);
              if (
                stage?.directorMode &&
                (await isHostIdentity(stageId, identity))
              ) {
                sendError(ws, 'Exit director mode to take a role.');
                break;
              }
              await toggleParticipantRole(stageId, lockedSelfPid, String(msg.castMemberId));
              await broadcastCurrentState(stageId);
              break;
            }

            case 'setAudience':
            case 'setReady': {
              const patch: { isAudience?: boolean; isReady?: boolean } = {};
              if (msg.type === 'setAudience') patch.isAudience = Boolean(msg.isAudience);
              if (msg.type === 'setReady') patch.isReady = Boolean(msg.isReady);
              await updateParticipantFlags(lockedSelfPid, patch);
              await broadcastCurrentState(stageId);
              break;
            }

            case 'selectPlay': {
              const playId = String(msg.playId || '') || null;
              await setSelectedPlay(stageId, playId);
              await broadcastCurrentState(stageId);
              break;
            }

            case 'sendChat': {
              const chat = {
                id: ulid(),
                senderId: lockedSelfPid,
                senderName: await participantNameFor(stageId, lockedSelfPid),
                text: String(msg.text || ''),
                castOnly: Boolean(msg.castOnly),
                timestamp: Date.now(),
              };
              if (chat.castOnly) {
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
              const stage = await getStageById(stageId);
              if (stage?.pausedByHost) {
                sendError(ws, 'Play is paused by host.');
                break;
              }
              if (stage?.directorMode && !(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Director controls line advance.');
                break;
              }
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
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Only the host can reset the play.');
                break;
              }
              await resetStage(stageId);
              await setPausedByHost(stageId, false);
              await broadcastCurrentState(stageId);
              break;
            }

            // ===== Host actions =====
            case 'hostPause':
            case 'hostResume': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const paused = msg.type === 'hostPause';
              await setPausedByHost(stageId, paused);
              await broadcastCurrentState(stageId);
              broadcastSystem(stageId, paused ? 'Host paused the play.' : 'Host resumed the play.');
              break;
            }

            case 'hostBackLine': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const stage = await getStageById(stageId);
              if (!stage || stage.phase !== 'performing') {
                sendError(ws, 'Not currently performing.');
                break;
              }
              const idx = await backLine(stageId);
              broadcastToStage(stageId, { type: 'lineAdvanced', currentLineIndex: idx });
              await broadcastCurrentState(stageId);
              broadcastSystem(stageId, 'Host went back a line.');
              break;
            }

            case 'hostSkipLine': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const stage = await getStageById(stageId);
              if (!stage || stage.phase !== 'performing') {
                sendError(ws, 'Not currently performing.');
                break;
              }
              const { currentLineIndex, playEnded } = await incrementLine(stageId);
              if (playEnded) {
                broadcastToStage(stageId, { type: 'playEnded' });
              } else {
                broadcastToStage(stageId, { type: 'lineAdvanced', currentLineIndex });
              }
              await broadcastCurrentState(stageId);
              broadcastSystem(stageId, 'Host skipped a line.');
              break;
            }

            case 'hostEndPlay': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              await resetStage(stageId);
              await setPausedByHost(stageId, false);
              await broadcastCurrentState(stageId);
              broadcastSystem(stageId, 'Host ended the play.');
              break;
            }

            case 'hostReassignRole': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const targetPid = String(msg.targetParticipantId);
              const castMemberId = String(msg.castMemberId);
              const target = await getParticipant(targetPid);
              if (!target || target.stageId !== stageId) {
                sendError(ws, 'Target participant not found.');
                break;
              }
              const state = await getStageState(stageId);
              const cast = state?.selectedPlay?.cast.find((c) => c.id === castMemberId);
              if (!cast) {
                sendError(ws, 'Cast member not found in current play.');
                break;
              }
              const wasPerforming = state?.phase === 'performing';

              // Lobby reassignment: switch immediately.
              if (!wasPerforming) {
                await reassignRole(stageId, targetPid, castMemberId);
                await broadcastCurrentState(stageId);
                broadcastSystem(
                  stageId,
                  `Host assigned ${cast.name} to ${target.displayName}.`
                );
                break;
              }

              // Mid-play: pause and send the target an offer they can accept/decline.
              const targetConn = getConn(stageId, targetPid);
              if (!targetConn) {
                sendError(ws, 'Target is no longer connected.');
                break;
              }
              await setPausedByHost(stageId, true);
              const offer: PendingOffer = {
                offerId: ulid(),
                castMemberId,
                castMemberName: cast.name,
                targetParticipantId: targetPid,
                createdAt: Date.now(),
              };
              addPendingOffer(stageId, offer);
              sendToClient(targetConn.ws, {
                type: 'roleOffer',
                offer: {
                  offerId: offer.offerId,
                  castMemberId: offer.castMemberId,
                  castMemberName: offer.castMemberName,
                },
              });
              await broadcastCurrentState(stageId);
              broadcastSystem(
                stageId,
                `Host offered ${cast.name} to ${target.displayName}. Play is paused.`
              );
              break;
            }

            case 'respondRoleOffer': {
              const offerId = String(msg.offerId);
              const offer = pendingOffers.get(stageId)?.get(offerId);
              if (!offer) {
                sendError(ws, 'Offer not found or already resolved.');
                break;
              }
              if (offer.targetParticipantId !== participantId) {
                sendError(ws, 'Not your offer to respond to.');
                break;
              }
              removePendingOffer(stageId, offerId);
              broadcastToStage(stageId, { type: 'roleOfferCleared', offerId });
              const targetName = (await participantDisplayName(participantId)) ?? 'Someone';
              if (msg.accept) {
                await reassignRole(stageId, participantId, offer.castMemberId);
                await broadcastCurrentState(stageId);
                broadcastSystem(
                  stageId,
                  `${targetName} accepted ${offer.castMemberName}. Resume when ready.`
                );
              } else {
                broadcastSystem(
                  stageId,
                  `${targetName} declined ${offer.castMemberName}. Play stays paused.`
                );
              }
              break;
            }

            case 'hostUnassignRole': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              await unassignRole(String(msg.targetParticipantId), String(msg.castMemberId));
              await broadcastCurrentState(stageId);
              broadcastSystem(stageId, 'Host unassigned a role.');
              break;
            }

            case 'hostTransferHost': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const targetPid = String(msg.targetParticipantId);
              const targetConn = getConn(stageId, targetPid);
              if (!targetConn) {
                sendError(ws, 'Target is no longer connected.');
                break;
              }
              const targetIdentity = targetConn.identity;
              if (!targetIdentity.userId && !targetIdentity.guestId) {
                sendError(ws, 'Target has no identity.');
                break;
              }
              cancelHostGrace(stageId);
              await setStageHost(stageId, targetIdentity);
              const targetName = (await participantDisplayName(targetPid)) ?? 'Someone';
              await broadcastCurrentState(stageId);
              broadcastSystem(stageId, `Host transferred to ${targetName}.`);
              break;
            }

            case 'hostKick':
            case 'hostBan': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const targetPid = String(msg.targetParticipantId);
              if (targetPid === participantId) {
                sendError(ws, "Can't kick yourself.");
                break;
              }
              const target = await getParticipant(targetPid);
              if (!target || target.stageId !== stageId) {
                sendError(ws, 'Target not found.');
                break;
              }
              const targetName = target.displayName;
              const targetIdentity: HostIdentity = {
                userId: target.userId ?? null,
                guestId: target.guestId ?? null,
              };
              if (msg.type === 'hostBan') {
                await addBan(stageId, targetIdentity);
              }
              await markParticipantLeft(targetPid);
              const targetConn = getConn(stageId, targetPid);
              if (targetConn) {
                sendToClient(targetConn.ws, {
                  type: 'error',
                  message: msg.type === 'hostBan' ? 'You were banned by the host.' : 'You were removed by the host.',
                });
                try {
                  targetConn.ws.close();
                } catch {
                  /* ignore */
                }
                removeConn(stageId, targetPid);
              }
              await livekitRemoveParticipant(stageId, targetPid);
              await broadcastCurrentState(stageId);
              broadcastSystem(
                stageId,
                msg.type === 'hostBan'
                  ? `${targetName} was banned by the host.`
                  : `${targetName} was removed by the host.`
              );
              break;
            }

            case 'hostToggleDirectorMode': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const stage = await getStageById(stageId);
              if (!stage) break;
              if (stage.phase !== 'lobby') {
                sendError(ws, 'Director mode can only be toggled from the lobby.');
                break;
              }
              const enabled = Boolean(msg.enabled);
              await setDirectorMode(stageId, enabled);
              if (enabled) {
                // Host exits the cast when entering director mode.
                const stagePids = connections.get(stageId);
                if (stagePids) {
                  for (const { participantId: pid, identity: connId } of stagePids.values()) {
                    if (identityMatches(connId, identity)) {
                      await clearParticipantRoles(pid);
                    }
                  }
                }
              }
              await broadcastCurrentState(stageId);
              broadcastSystem(
                stageId,
                enabled ? 'Director mode enabled. Host will advance lines.' : 'Director mode disabled.'
              );
              break;
            }

            case 'hostMute': {
              if (!(await isHostIdentity(stageId, identity))) {
                sendError(ws, 'Host only.');
                break;
              }
              const targetPid = String(msg.targetParticipantId);
              const muted = Boolean(msg.muted);
              const target = await getParticipant(targetPid);
              if (!target || target.stageId !== stageId) {
                sendError(ws, 'Target not found.');
                break;
              }
              await setForceMuted(targetPid, muted);
              await livekitSetMuted(stageId, targetPid, muted);
              await broadcastCurrentState(stageId);
              broadcastSystem(
                stageId,
                muted
                  ? `Host muted ${target.displayName}.`
                  : `Host unmuted ${target.displayName}.`
              );
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
          // If this socket was displaced by a newer connection from the same
          // identity (e.g. another tab), the participant is still present
          // under the same id — don't run the leave/host-grace flow.
          if (myConn?.displaced) return;
          const sid = stageId;
          const pid = participantId;
          const leavingIdentity = { ...identity };
          const leavingName = (await participantDisplayName(pid)) ?? 'Host';
          removeConn(sid, pid);
          await leaveStage(sid, pid).catch(() => {});

          // Auto-decline any pending role offers held by the leaving participant.
          const dropped = clearOffersForParticipant(sid, pid);
          for (const offer of dropped) {
            broadcastToStage(sid, { type: 'roleOfferCleared', offerId: offer.offerId });
            broadcastSystem(
              sid,
              `${leavingName} disconnected before responding to ${offer.castMemberName}. Play stays paused.`
            );
          }

          await broadcastCurrentState(sid);

          // If the departing connection was the host's identity AND no other
          // connection for that identity remains, start a grace timer before
          // transferring host.
          const host = await getStageHost(sid).catch(() => null);
          if (
            host &&
            identityMatches(host, leavingIdentity) &&
            !isIdentityConnected(sid, leavingIdentity)
          ) {
            await scheduleHostSuccession(sid, leavingIdentity, leavingName);
          }
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
