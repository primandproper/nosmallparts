import { v4 as uuidv4 } from 'uuid';
import { WebSocket } from 'ws';
import { StageState, Participant, Play, ServerMessage, getLineIndices } from './shared/types';
import { getAllPlays, getPlayById } from './shared/plays';

interface Stage {
  state: StageState;
  connections: Map<string, WebSocket>;
}

const stages = new Map<string, Stage>();

// User-created scripts storage
const userScripts = new Map<string, Play>();

// Script CRUD functions
export function getAllScripts(): Play[] {
  const samplePlays = getAllPlays();
  const userPlays = Array.from(userScripts.values());
  return [...samplePlays, ...userPlays];
}

export function getScriptById(id: string): Play | undefined {
  // Check user scripts first, then sample plays
  return userScripts.get(id) || getPlayById(id);
}

export function createScript(play: Omit<Play, 'id'>): Play {
  const id = `user-${uuidv4()}`;
  const newPlay: Play = { ...play, id };
  userScripts.set(id, newPlay);
  return newPlay;
}

export function updateScript(id: string, play: Partial<Play>): Play | null {
  const existing = userScripts.get(id);
  if (!existing) return null;

  const updated = { ...existing, ...play, id }; // Preserve original ID
  userScripts.set(id, updated);
  return updated;
}

export function deleteScript(id: string): boolean {
  // Only allow deleting user scripts, not sample plays
  if (!id.startsWith('user-')) return false;
  return userScripts.delete(id);
}

export function isUserScript(id: string): boolean {
  return id.startsWith('user-');
}

function generateStageCode(): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
  let code = '';
  for (let i = 0; i < 6; i++) {
    code += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return code;
}

export function createStage(): StageState {
  const stageId = generateStageCode();
  const state: StageState = {
    id: stageId,
    participants: [],
    selectedPlay: null,
    currentLineIndex: 0,
    phase: 'lobby',
  };

  stages.set(stageId, {
    state,
    connections: new Map(),
  });

  return state;
}

export function getStage(stageId: string): Stage | undefined {
  return stages.get(stageId);
}

export function getStageState(stageId: string): StageState | undefined {
  return stages.get(stageId)?.state;
}

export function stageExists(stageId: string): boolean {
  return stages.has(stageId);
}

export function addStageConnection(stageId: string, participantId: string, ws: WebSocket): void {
  const stage = stages.get(stageId);
  if (stage) {
    stage.connections.set(participantId, ws);
  }
}

export function removeStageConnection(stageId: string, participantId: string): void {
  const stage = stages.get(stageId);
  if (stage) {
    stage.connections.delete(participantId);
    // Remove participant from state
    stage.state.participants = stage.state.participants.filter((p) => p.id !== participantId);

    // Clean up empty stages
    if (stage.connections.size === 0) {
      stages.delete(stageId);
    }
  }
}

export function broadcastToStage(stageId: string, message: ServerMessage): void {
  const stage = stages.get(stageId);
  if (stage) {
    console.log(`[Broadcast] ${stageId}: ${message.type} to ${stage.connections.size} connections`);
    const data = JSON.stringify(message);
    stage.connections.forEach((ws, participantId) => {
      if (ws.readyState === WebSocket.OPEN) {
        console.log(`  -> Sending to ${participantId}`);
        ws.send(data);
      } else {
        console.log(`  -> Skipping ${participantId} (not open)`);
      }
    });
  }
}

// Broadcast to cast members only (non-audience participants)
export function broadcastToCast(stageId: string, message: ServerMessage): void {
  const stage = stages.get(stageId);
  if (stage) {
    const castParticipantIds = stage.state.participants
      .filter((p) => !p.isAudience)
      .map((p) => p.id);

    console.log(
      `[BroadcastToCast] ${stageId}: ${message.type} to ${castParticipantIds.length} cast members`
    );
    const data = JSON.stringify(message);
    stage.connections.forEach((ws, participantId) => {
      if (castParticipantIds.includes(participantId) && ws.readyState === WebSocket.OPEN) {
        console.log(`  -> Sending to cast member ${participantId}`);
        ws.send(data);
      }
    });
  }
}

export function sendToParticipant(
  stageId: string,
  participantId: string,
  message: ServerMessage
): void {
  const stage = stages.get(stageId);
  if (stage) {
    const ws = stage.connections.get(participantId);
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message));
    }
  }
}

export function addParticipantToStage(stageId: string, name: string): Participant | null {
  const stage = stages.get(stageId);
  if (!stage) {
    console.log(`[AddParticipant] Stage ${stageId} not found!`);
    return null;
  }

  const participant: Participant = {
    id: uuidv4(),
    name,
    castMemberIds: [],
    isDirector: false,
    isReady: false,
    isAudience: false,
  };

  stage.state.participants.push(participant);
  console.log(
    `[AddParticipant] ${name} joined ${stageId}. Total participants: ${stage.state.participants.length}`
  );
  return participant;
}

export function toggleCastMember(
  stageId: string,
  participantId: string,
  castMemberId: string
): boolean {
  const stage = stages.get(stageId);
  if (!stage) return false;

  const participant = stage.state.participants.find((p) => p.id === participantId);
  if (!participant) return false;

  const index = participant.castMemberIds.indexOf(castMemberId);
  if (index === -1) {
    // Claiming a role - if they're a director, remove director status first
    if (participant.isDirector) {
      participant.isDirector = false;
      console.log(`[ToggleCastMember] ${participant.name} gave up director role to claim a part`);
    }

    // Remove this cast member from any other participant first
    stage.state.participants.forEach((p) => {
      const idx = p.castMemberIds.indexOf(castMemberId);
      if (idx !== -1) {
        p.castMemberIds.splice(idx, 1);
      }
    });
    // Add to this participant
    participant.castMemberIds.push(castMemberId);
  } else {
    // Remove from this participant
    participant.castMemberIds.splice(index, 1);
  }

  return true;
}

export function setParticipantDirector(
  stageId: string,
  participantId: string,
  isDirector: boolean
): boolean {
  const stage = stages.get(stageId);
  if (!stage) return false;

  const participant = stage.state.participants.find((p) => p.id === participantId);
  if (!participant) return false;

  // If becoming a director, clear any assigned roles (can't be both actor and director)
  if (isDirector && participant.castMemberIds.length > 0) {
    console.log(
      `[SetDirector] ${participant.name} gave up roles to become director:`,
      participant.castMemberIds
    );
    participant.castMemberIds = [];
    participant.isReady = false; // Reset ready status since roles changed
  }

  participant.isDirector = isDirector;
  return true;
}

export function setParticipantReady(
  stageId: string,
  participantId: string,
  isReady: boolean
): boolean {
  const stage = stages.get(stageId);
  if (!stage) return false;

  const participant = stage.state.participants.find((p) => p.id === participantId);
  if (!participant) return false;

  participant.isReady = isReady;
  return true;
}

export function setParticipantAudience(
  stageId: string,
  participantId: string,
  isAudience: boolean
): boolean {
  const stage = stages.get(stageId);
  if (!stage) return false;

  const participant = stage.state.participants.find((p) => p.id === participantId);
  if (!participant) return false;

  // If becoming audience, clear any roles, director status, and ready status
  if (isAudience) {
    if (participant.castMemberIds.length > 0 || participant.isDirector) {
      console.log(
        `[SetAudience] ${participant.name} joining audience, clearing roles/director status`
      );
    }
    participant.castMemberIds = [];
    participant.isDirector = false;
    participant.isReady = false;
  }

  participant.isAudience = isAudience;
  return true;
}

export function selectPlayForStage(stageId: string, playId: string): Play | null {
  const stage = stages.get(stageId);
  if (!stage) return null;

  // Check both sample plays and user scripts
  const play = getScriptById(playId);
  if (!play) return null;

  stage.state.selectedPlay = play;
  // Clear cast member assignments when play changes
  stage.state.participants.forEach((p) => {
    p.castMemberIds = [];
    p.isReady = false;
  });

  return play;
}

export function startStagePlay(stageId: string): boolean {
  const stage = stages.get(stageId);
  if (!stage || !stage.state.selectedPlay) return false;

  // Check if all cast members are assigned
  const assignedCastMemberIds = new Set(stage.state.participants.flatMap((p) => p.castMemberIds));

  const allAssigned = stage.state.selectedPlay.cast.every((castMember) =>
    assignedCastMemberIds.has(castMember.id)
  );

  // Check if all participants are ready
  const allReady = stage.state.participants.every((p) => p.isReady);

  if (!allAssigned || !allReady) return false;

  stage.state.phase = 'countdown';
  stage.state.currentLineIndex = 0;

  // After countdown, start performing
  setTimeout(() => {
    if (stage.state.phase === 'countdown') {
      stage.state.phase = 'performing';
      broadcastToStage(stageId, { type: 'playStarted' });
      broadcastToStage(stageId, { type: 'stageState', state: stage.state });
    }
  }, 3000);

  return true;
}

export function advanceStageLine(stageId: string): boolean {
  const stage = stages.get(stageId);
  if (!stage || !stage.state.selectedPlay || stage.state.phase !== 'performing') return false;

  const lineIndices = getLineIndices(stage.state.selectedPlay);
  const nextLineIndex = stage.state.currentLineIndex + 1;

  if (nextLineIndex >= lineIndices.length) {
    // Play ended
    stage.state.phase = 'discussion';
    broadcastToStage(stageId, { type: 'playEnded' });
  } else {
    stage.state.currentLineIndex = nextLineIndex;
    broadcastToStage(stageId, { type: 'lineAdvanced', currentLineIndex: nextLineIndex });
  }

  return true;
}

export function resetStagePlay(stageId: string): boolean {
  const stage = stages.get(stageId);
  if (!stage) return false;

  stage.state.phase = 'lobby';
  stage.state.currentLineIndex = 0;
  stage.state.participants.forEach((p) => {
    p.isReady = false;
  });

  return true;
}

export function getAvailablePlaysForStage(): Play[] {
  return getAllScripts();
}
