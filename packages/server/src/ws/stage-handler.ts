/**
 * WebSocket handler for real-time stage sessions.
 * Maintains an in-memory connection map, persists state via Drizzle, and
 * broadcasts updates to all connected participants.
 */
import type { Context } from 'hono';
import type { WSContext, WSEvents } from 'hono/ws';
import { ulid } from 'ulid';
import {
  getPerformers,
  isParticipantSpeaking,
  truncateChatText,
  type StageState,
} from '@nsp/shared';
import type { AppVariables } from '../app.js';
import { getClientIp } from '../client-ip.js';
import { dedupeDisplayName, validateDisplayName } from '../display-name.js';
import { resolveIdentity, type Identity } from '../identity.js';
import { livekitRemoveParticipant, livekitSetMuted } from '../livekit-admin.js';
import { listPlayScripts } from '../repos/play-scripts.js';
import {
  activeDisplayNames,
  addBan,
  addIpBan,
  backLine,
  deleteStageDeep,
  findIdleStageIds,
  getParticipantRoleIds,
  isIpBanned,
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
  rejoinParticipant,
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

// Durations are env-overridable so integration tests can shrink them.
const HOST_GRACE_MS = Number(process.env.NSP_HOST_GRACE_MS || 15_000);
const MAX_PARTICIPANTS_PER_STAGE = 12;
const VACANCY_CLAIM_MS = Number(process.env.NSP_VACANCY_CLAIM_MS || 45_000);
const STAGE_CLEANUP_DELAY_MS = Number(process.env.NSP_STAGE_CLEANUP_DELAY_MS || 5_000);
const STAGE_IDLE_MS = Number(process.env.NSP_STAGE_IDLE_MS || 10 * 60_000);
const STAGE_REAPER_INTERVAL_MS = Number(process.env.NSP_STAGE_REAPER_INTERVAL_MS || 60_000);
const COUNTDOWN_MS = Number(process.env.NSP_COUNTDOWN_MS || 3_000);

type Conn = { ws: WSContext; participantId: string; identity: Identity; ip: string | null; displaced?: boolean };

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

// ===== Mid-play role vacancy (§4.5) =====
// stageId -> castMemberId -> vacancy
type Vacancy = {
  castMemberId: string;
  castMemberName: string;
  previousParticipantId: string;
  claimDeadline: number | null; // null in director mode (no countdown)
  timer: ReturnType<typeof setTimeout> | null;
};
const roleVacancies = new Map<string, Map<string, Vacancy>>();

function getVacancies(stageId: string): Map<string, Vacancy> {
  if (!roleVacancies.has(stageId)) roleVacancies.set(stageId, new Map());
  return roleVacancies.get(stageId)!;
}

function clearVacancy(stageId: string, castMemberId: string): Vacancy | undefined {
  const map = roleVacancies.get(stageId);
  if (!map) return undefined;
  const v = map.get(castMemberId);
  if (v) {
    if (v.timer) clearTimeout(v.timer);
    map.delete(castMemberId);
  }
  if (map.size === 0) roleVacancies.delete(stageId);
  return v;
}

function clearAllVacancies(stageId: string) {
  const map = roleVacancies.get(stageId);
  if (!map) return;
  for (const v of map.values()) {
    if (v.timer) clearTimeout(v.timer);
  }
  roleVacancies.delete(stageId);
}

// ===== Stage cleanup (§10.2) =====
// stageId -> pending empty-stage deletion timer (5s after last connection drops).
const stageCleanupTimers = new Map<string, ReturnType<typeof setTimeout>>();

function cancelStageCleanup(stageId: string) {
  const t = stageCleanupTimers.get(stageId);
  if (t) {
    clearTimeout(t);
    stageCleanupTimers.delete(stageId);
  }
}

function dropStageMemory(stageId: string) {
  connections.delete(stageId);
  pendingOffers.delete(stageId);
  cancelHostGrace(stageId);
  clearAllVacancies(stageId);
  cancelStageCleanup(stageId);
}

async function cleanupStage(stageId: string): Promise<void> {
  // A joiner during the grace window cancels this timer; double-check anyway.
  if ((connections.get(stageId)?.size ?? 0) > 0) return;
  dropStageMemory(stageId);
  await deleteStageDeep(stageId).catch((err) => {
    console.error('[cleanup] failed to delete stage', err);
  });
}

function scheduleStageCleanup(stageId: string) {
  cancelStageCleanup(stageId);
  const timer = setTimeout(() => {
    stageCleanupTimers.delete(stageId);
    void cleanupStage(stageId);
  }, STAGE_CLEANUP_DELAY_MS);
  stageCleanupTimers.set(stageId, timer);
}

/**
 * Background reaper: deletes stages with no live connections whose last
 * activity is older than STAGE_IDLE_MS. Runs against the DB, so it also
 * covers stages orphaned by a server restart.
 */
export async function runStageReaperOnce(): Promise<string[]> {
  const cutoff = new Date(Date.now() - STAGE_IDLE_MS);
  const idleIds = await findIdleStageIds(cutoff);
  const deleted: string[] = [];
  for (const id of idleIds) {
    if ((connections.get(id)?.size ?? 0) > 0) continue;
    dropStageMemory(id);
    await deleteStageDeep(id).catch(() => {});
    deleted.push(id);
  }
  return deleted;
}

let reaperInterval: ReturnType<typeof setInterval> | null = null;

export function startStageReaper(): void {
  if (reaperInterval) return;
  reaperInterval = setInterval(() => {
    void runStageReaperOnce();
  }, STAGE_REAPER_INTERVAL_MS);
  reaperInterval.unref?.();
}

export function stopStageReaper(): void {
  if (reaperInterval) {
    clearInterval(reaperInterval);
    reaperInterval = null;
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

function registerConn(
  stageId: string,
  participantId: string,
  ws: WSContext,
  identity: Identity,
  ip: string | null
) {
  if (!connections.has(stageId)) connections.set(stageId, new Map());
  connections.get(stageId)!.set(participantId, { ws, participantId, identity, ip });
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

/**
 * A role-holder disconnected mid-performance (§4.5). Pause the play and open
 * a vacancy for each role they held. Non-director mode starts a 45s claim
 * countdown for the audience; director mode pauses and waits for the director
 * to run the offer flow.
 */
async function openRoleVacancies(stageId: string, leavingPid: string, leavingName: string) {
  const stage = await getStageById(stageId);
  if (!stage || stage.phase !== 'performing') return;
  const roleIds = await getParticipantRoleIds(leavingPid);
  if (roleIds.length === 0) return;
  const state = await getStageState(stageId);
  const cast = state?.selectedPlay?.cast ?? [];
  const vacatedCast = cast.filter((c) => roleIds.includes(c.id));
  if (vacatedCast.length === 0) return;

  await setPausedByHost(stageId, true);
  const directorMode = !!stage.directorMode;
  for (const cm of vacatedCast) {
    if (getVacancies(stageId).has(cm.id)) continue;
    const claimDeadline = directorMode ? null : Date.now() + VACANCY_CLAIM_MS;
    const timer = directorMode
      ? null
      : setTimeout(() => {
          void expireVacancy(stageId, cm.id);
        }, VACANCY_CLAIM_MS);
    getVacancies(stageId).set(cm.id, {
      castMemberId: cm.id,
      castMemberName: cm.name,
      previousParticipantId: leavingPid,
      claimDeadline,
      timer,
    });
    broadcastToStage(stageId, {
      type: 'roleVacated',
      vacancy: { castMemberId: cm.id, castMemberName: cm.name, claimDeadline },
    });
    broadcastSystem(
      stageId,
      directorMode
        ? `${leavingName} disconnected — ${cm.name} is vacant. The director can reassign the role.`
        : `${leavingName} disconnected — ${cm.name} is vacant. Any audience member can claim it.`
    );
  }
  await broadcastCurrentState(stageId);
}

/** The 45s claim countdown ran out with no claim: auto-end to discussion. */
async function expireVacancy(stageId: string, castMemberId: string) {
  const v = clearVacancy(stageId, castMemberId);
  if (!v) return;
  const stage = await getStageById(stageId);
  if (!stage || stage.phase !== 'performing') return;
  clearAllVacancies(stageId);
  await setStagePhase(stageId, 'discussion');
  await setPausedByHost(stageId, false);
  broadcastToStage(stageId, { type: 'playEnded' });
  broadcastSystem(stageId, `Play ended: ${v.castMemberName} went vacant.`);
  await broadcastCurrentState(stageId);
}

export function createStageWsHandler(): (c: Context<{ Variables: AppVariables }>) => Promise<WSEvents> {
  return async (c) => {
    const identity = await resolveIdentity(c);
    const clientIp = getClientIp(c);
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
            const nameCheck = validateDisplayName(String(msg.participantName || ''));
            if (!nameCheck.ok) {
              sendError(ws, nameCheck.error);
              return;
            }
            let name = nameCheck.name;
            const stage = await getStageByIdOrCode(idOrCode);
            if (!stage) {
              sendError(ws, 'Stage not found');
              return;
            }
            if ((await isBanned(stage.id, identity)) || (await isIpBanned(stage.id, clientIp))) {
              sendError(ws, 'You have been removed from this stage.');
              ws.close();
              return;
            }
            // A successful join cancels any pending empty-stage cleanup (§10.2).
            cancelStageCleanup(stage.id);
            // Reuse an existing participant row if this identity is already on
            // the stage (e.g. a second browser tab). Displace any prior live
            // connection so the same person doesn't appear twice.
            const existing = await findActiveParticipantByIdentity(stage.id, identity);
            let lateJoinAnnouncement: string | null = null;
            if (existing) {
              const prev = getConn(stage.id, existing.id);
              if (prev) {
                prev.displaced = true;
                sendToClient(prev.ws, { type: 'sessionReplaced' });
                try { prev.ws.close(); } catch { /* ignore */ }
              }
              if (existing.displayName !== name) {
                name = dedupeDisplayName(name, await activeDisplayNames(stage.id, existing.id));
                if (existing.displayName !== name) {
                  await setParticipantDisplayName(existing.id, name);
                }
              }
              // Re-admit the participant in case their WS dropped and onClose
              // already set leftAt (which would hide them from stageState).
              await rejoinParticipant(existing.id);
              stageId = stage.id;
              participantId = existing.id;
            } else {
              const currentCount = connections.get(stage.id)?.size ?? 0;
              if (currentCount >= MAX_PARTICIPANTS_PER_STAGE) {
                sendError(ws, `Stage is full (${MAX_PARTICIPANTS_PER_STAGE} max).`);
                ws.close();
                return;
              }
              name = dedupeDisplayName(name, await activeDisplayNames(stage.id));
              // Late joiners land in the audience (§3.9) and stay there until
              // the stage returns to the lobby.
              const lateJoin = stage.phase !== 'lobby';
              const joined = await joinStage(stage.id, name, identity, { isAudience: lateJoin });
              stageId = joined.stageId;
              participantId = joined.participantId;
              if (lateJoin) lateJoinAnnouncement = `${name} joined the audience.`;
            }
            authorized = true;
            registerConn(stageId, participantId, ws, identity, clientIp);
            myConn = getConn(stageId, participantId) ?? null;

            // If this participant left roles vacant mid-play, they reclaim them
            // now — their role rows survived the disconnect (§4.5).
            const stageVacancies = roleVacancies.get(stageId);
            if (stageVacancies) {
              for (const v of Array.from(stageVacancies.values())) {
                if (v.previousParticipantId === participantId) {
                  clearVacancy(stageId, v.castMemberId);
                  broadcastToStage(stageId, { type: 'roleVacancyCleared', castMemberId: v.castMemberId });
                  broadcastSystem(stageId, `${name} reconnected and reclaimed ${v.castMemberName}.`);
                }
              }
            }

            // Host assignment: a host-less stage is claimed by the first
            // signed-in joiner. Guests never host (§1.4). If the returning
            // joiner IS the host identity, cancel succession.
            const currentHost = await getStageHost(stageId);
            let hostAnnouncement: string | null = null;
            if (!currentHost) {
              if (identity.userId) {
                await setStageHost(stageId, identity);
                hostAnnouncement = `${name} is host.`;
              }
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
            // Late joiners see any active vacancy countdown with its real deadline (§9).
            for (const v of roleVacancies.get(stageId)?.values() ?? []) {
              sendToClient(ws, {
                type: 'roleVacated',
                vacancy: {
                  castMemberId: v.castMemberId,
                  castMemberName: v.castMemberName,
                  claimDeadline: v.claimDeadline,
                },
              });
            }
            broadcastToStage(stageId, { type: 'stageState', state }, { excludeParticipantId: participantId });
            if (lateJoinAnnouncement) broadcastSystem(stageId, lateJoinAnnouncement);
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
              if (msg.type === 'setAudience') {
                const toAudience = Boolean(msg.isAudience);
                if (!toAudience) {
                  // Leaving the audience is only possible in the lobby (§3.9).
                  const stage = await getStageById(stageId);
                  if (stage && stage.phase !== 'lobby') {
                    sendError(ws, 'You can join the cast when the stage returns to the lobby.');
                    break;
                  }
                }
                patch.isAudience = toAudience;
              }
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
              // Server-side truncation (§6.3) — client input is not hard-capped.
              const chat = {
                id: ulid(),
                senderId: lockedSelfPid,
                senderName: await participantNameFor(stageId, lockedSelfPid),
                text: truncateChatText(String(msg.text || '')),
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
              const state = await getStageState(sid);
              if (!state) break;
              if (state.phase !== 'lobby') {
                sendError(ws, 'Play already started.');
                break;
              }
              if (!state.selectedPlay) {
                sendError(ws, 'Select a play first.');
                break;
              }
              // In director mode, only the director may start (§3.4).
              if (state.directorMode && !(await isHostIdentity(sid, identity))) {
                sendError(ws, 'Only the director can start the play.');
                break;
              }
              // Preconditions (§3.4): every cast role filled, every
              // non-audience participant ready.
              const problems = startPlayProblems(state);
              if (problems) {
                sendError(ws, `Cannot start: ${problems}`);
                break;
              }
              await setStagePhase(sid, 'countdown', 0);
              const countdownState = await getStageState(sid);
              if (countdownState) {
                broadcastToStage(sid, { type: 'countdownStart', duration: 3 });
                broadcastToStage(sid, { type: 'stageState', state: countdownState });
              }
              setTimeout(async () => {
                // The stage may have been reset or deleted during the countdown.
                const current = await getStageById(sid);
                if (!current || current.phase !== 'countdown') return;
                await setStagePhase(sid, 'performing', 0);
                const perfState = await getStageState(sid);
                if (perfState) {
                  broadcastToStage(sid, { type: 'playStarted' });
                  broadcastToStage(sid, { type: 'stageState', state: perfState });
                }
              }, COUNTDOWN_MS);
              break;
            }

            case 'nextLine': {
              const stage = await getStageById(stageId);
              if (stage?.pausedByHost) {
                sendError(ws, 'Play is paused by host.');
                break;
              }
              const state = await getStageState(stageId);
              const me = state?.participants.find((p) => p.id === participantId);
              if (me?.isAudience) {
                sendError(ws, 'Audience cannot advance lines.');
                break;
              }
              if (stage?.directorMode) {
                if (!(await isHostIdentity(stageId, identity))) {
                  sendError(ws, 'Director controls line advance.');
                  break;
                }
              } else if (state && !isParticipantSpeaking(state, participantId)) {
                // Default mode: only the speaker of the current line advances (§3.6).
                sendError(ws, 'Not your line.');
                break;
              }
              const { currentLineIndex, playEnded } = await incrementLine(stageId);
              if (playEnded) {
                clearAllVacancies(stageId);
                broadcastToStage(stageId, { type: 'playEnded' });
              } else {
                broadcastToStage(stageId, { type: 'lineAdvanced', currentLineIndex });
              }
              await broadcastCurrentState(stageId);
              break;
            }

            case 'resetPlay': {
              const stage = await getStageById(stageId);
              const isHost = await isHostIdentity(stageId, identity);
              // From the discussion phase (play just ended), anyone can request
              // "perform again" — the reset just sends everyone back to lobby
              // with cast assignments preserved. From any other phase, only the
              // host can reset (avoids accidental mid-performance resets).
              if (!isHost && stage?.phase !== 'discussion') {
                sendError(ws, 'Only the host can reset the play.');
                break;
              }
              clearAllVacancies(stageId);
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
                clearAllVacancies(stageId);
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
              clearAllVacancies(stageId);
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
              // Host fast-track bypasses any vacancy countdown for this role (§4.5).
              if (clearVacancy(stageId, castMemberId)) {
                broadcastToStage(stageId, { type: 'roleVacancyCleared', castMemberId });
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

            case 'audienceClaimRole': {
              // During the vacancy countdown, any audience member can claim
              // the vacant role with a single click (§4.5). Non-director only.
              const castMemberId = String(msg.castMemberId);
              const vacancy = roleVacancies.get(stageId)?.get(castMemberId);
              if (!vacancy || vacancy.claimDeadline === null) {
                sendError(ws, 'That role is not open to claims.');
                break;
              }
              const me = await getParticipant(participantId);
              if (!me || !me.isAudience) {
                sendError(ws, 'Only audience members can claim a vacant role.');
                break;
              }
              clearVacancy(stageId, castMemberId);
              await updateParticipantFlags(participantId, { isAudience: false });
              await reassignRole(stageId, participantId, castMemberId);
              broadcastToStage(stageId, { type: 'roleVacancyCleared', castMemberId });
              await broadcastCurrentState(stageId);
              broadcastSystem(
                stageId,
                `${me.displayName} claimed ${vacancy.castMemberName}. Host can resume when ready.`
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
              if (!targetIdentity.userId) {
                // Guests cannot host (§4.2).
                sendError(ws, 'Target must be signed in.');
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
              const targetConn = getConn(stageId, targetPid);
              if (msg.type === 'hostBan') {
                await addBan(stageId, targetIdentity);
                // Record the banned identity's IP so a cleared cookie doesn't
                // get them back in (§4.7). Only known while they're connected.
                if (targetConn?.ip) {
                  await addIpBan(stageId, targetConn.ip);
                }
              }
              await markParticipantLeft(targetPid);
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

          // A role-holder leaving mid-performance opens a vacancy (§4.5).
          await openRoleVacancies(sid, pid, leavingName).catch((err) => {
            console.error('[ws] vacancy flow failed', err);
          });

          // If the departing connection was the host's identity AND no other
          // connection for that identity remains, start a grace timer before
          // transferring host.
          const host = await getStageHost(sid).catch(() => null);
          if (
            host &&
            identityMatches(host, leavingIdentity) &&
            !isIdentityConnected(sid, leavingIdentity)
          ) {
            // The host walked away: auto-clear any offers they had pending (§9).
            const stageOffers = pendingOffers.get(sid);
            if (stageOffers) {
              for (const offer of Array.from(stageOffers.values())) {
                removePendingOffer(sid, offer.offerId);
                broadcastToStage(sid, { type: 'roleOfferCleared', offerId: offer.offerId });
              }
            }
            await scheduleHostSuccession(sid, leavingIdentity, leavingName);
          }

          // Last one out: schedule stage deletion in 5s (§10.2). A joiner
          // arriving during the window cancels it.
          if ((connections.get(sid)?.size ?? 0) === 0) {
            scheduleStageCleanup(sid);
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

/**
 * Human-readable list of unmet start-play preconditions (§3.4), e.g.
 * "3 unfilled roles, 2 not ready". Null when the play can start.
 */
function startPlayProblems(state: StageState): string | null {
  if (!state.selectedPlay) return 'no play selected';
  const problems: string[] = [];
  const performers = getPerformers(state);
  if (performers.length === 0) {
    problems.push('no performers');
  }
  const assigned = new Set(performers.flatMap((p) => p.castMemberIds));
  const unfilled = (state.selectedPlay.cast || []).filter((cm) => !assigned.has(cm.id)).length;
  if (unfilled > 0) problems.push(`${unfilled} unfilled role${unfilled === 1 ? '' : 's'}`);
  const notReady = getPerformers(state).filter((p) => !p.isReady).length;
  if (notReady > 0) problems.push(`${notReady} not ready`);
  return problems.length ? problems.join(', ') : null;
}
