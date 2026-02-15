// Shared types for the play performance app

export interface CastMember {
  id: string;
  name: string;
  description?: string;
}

export interface PlayLine {
  castMemberId: string;
  text: string;
  index: number;
  direction?: string; // Stage direction/emotion, e.g., "skeptically", "with tears"
  scrollMode?: 'auto' | 'manual'; // For long monologues - auto scrolls at WPM, manual allows user control
  autoScrollWpm?: number; // Words per minute for auto-scroll (default: 150)
}

export interface Play {
  id: string;
  title: string;
  description: string;
  author?: string;
  cast: CastMember[];
  lines: PlayLine[];
}

export interface Participant {
  id: string;
  name: string;
  castMemberIds: string[];
  isDirector: boolean;
  isReady: boolean;
  isAudience: boolean;
}

export type StagePhase = 'lobby' | 'countdown' | 'performing' | 'discussion';
// Keep RoomPhase as alias for backward compatibility
export type RoomPhase = StagePhase;

export interface StageState {
  id: string;
  code: string;
  participants: Participant[];
  selectedPlay: Play | null;
  currentLineIndex: number;
  phase: StagePhase;
}

// Keep RoomState as alias for backward compatibility
export type RoomState = StageState;

// Message types for WebSocket communication
export interface ChatMessage {
  id: string;
  senderId: string;
  senderName: string;
  text: string;
  castOnly: boolean;
  timestamp: number;
}

export type ClientMessage =
  | { type: 'join'; stageId: string; participantName: string }
  | { type: 'toggleCastMember'; participantId: string; castMemberId: string }
  | { type: 'setDirector'; participantId: string; isDirector: boolean }
  | { type: 'setAudience'; participantId: string; isAudience: boolean }
  | { type: 'selectPlay'; playId: string }
  | { type: 'setReady'; participantId: string; isReady: boolean }
  | { type: 'sendChat'; participantId: string; text: string; castOnly: boolean }
  | { type: 'startPlay' }
  | { type: 'nextLine' }
  | { type: 'resetPlay' };

export type ServerMessage =
  | { type: 'stageState'; state: StageState }
  | { type: 'joined'; participantId: string; stageState: StageState }
  | { type: 'error'; message: string }
  | { type: 'availablePlays'; plays: Play[] }
  | { type: 'countdownStart'; duration: number }
  | { type: 'castMembersUpdated'; participants: Participant[] }
  | { type: 'chatMessage'; message: ChatMessage }
  | { type: 'playStarted' }
  | { type: 'lineAdvanced'; currentLineIndex: number }
  | { type: 'playEnded' };

// Utility functions
export function getCastMemberById(play: Play, castMemberId: string): CastMember | undefined {
  return (play.cast || []).find((c) => c.id === castMemberId);
}

export function getCastMemberNameForLine(play: Play, line: PlayLine): string {
  const castMember = getCastMemberById(play, line.castMemberId);
  return castMember?.name || 'Unknown';
}

export function getPerformers(state: StageState): Participant[] {
  return (state.participants || []).filter((p) => !p.isAudience);
}

export function getAudience(state: StageState): Participant[] {
  return (state.participants || []).filter((p) => p.isAudience);
}

export function areAllCastMembersAssigned(state: StageState): boolean {
  if (!state.selectedPlay) return false;

  // Only count performers (non-audience) for role assignments
  const performers = getPerformers(state);
  const assignedCastMemberIds = new Set(performers.flatMap((p) => p.castMemberIds));

  return (state.selectedPlay.cast || []).every((castMember) =>
    assignedCastMemberIds.has(castMember.id)
  );
}

export function areAllParticipantsReady(state: StageState): boolean {
  // Only performers need to be ready, not audience
  const performers = getPerformers(state);
  return performers.length > 0 && performers.every((p) => p.isReady || p.isDirector);
}

export function canStartPlay(state: StageState): boolean {
  return areAllCastMembersAssigned(state) && areAllParticipantsReady(state);
}

export function getLineIndices(play: Play): number[] {
  const indices = new Set((play.lines || []).map((l) => l.index));
  return Array.from(indices).sort((a, b) => a - b);
}

export function getLinesAtIndex(play: Play, index: number): PlayLine[] {
  return (play.lines || []).filter((l) => l.index === index);
}

export function getCurrentLines(state: StageState): PlayLine[] {
  if (!state.selectedPlay || state.phase !== 'performing') return [];

  const indices = getLineIndices(state.selectedPlay);
  if (state.currentLineIndex >= indices.length) return [];

  const currentIndex = indices[state.currentLineIndex];
  return getLinesAtIndex(state.selectedPlay, currentIndex);
}

export function getCurrentSpeakerCastMemberIds(state: StageState): string[] {
  const currentLines = getCurrentLines(state);
  return currentLines.map((l) => l.castMemberId);
}

export function getCurrentSpeakers(state: StageState): Participant[] {
  const speakerCastMemberIds = getCurrentSpeakerCastMemberIds(state);
  return (state.participants || []).filter((p) =>
    (p.castMemberIds || []).some((id) => speakerCastMemberIds.includes(id))
  );
}

export function isParticipantSpeaking(state: StageState, participantId: string): boolean {
  const speakers = getCurrentSpeakers(state);
  return speakers.some((s) => s.id === participantId);
}

export function getCurrentLinesForParticipant(
  state: StageState,
  participantId: string
): PlayLine[] {
  const currentLines = getCurrentLines(state);
  const participant = (state.participants || []).find((p) => p.id === participantId);
  if (!participant) return [];

  return currentLines.filter((l) => (participant.castMemberIds || []).includes(l.castMemberId));
}

export function getNextLinesForParticipant(
  state: StageState,
  participantId: string
): { lines: PlayLine[]; turnsAway: number } | null {
  if (!state.selectedPlay || state.phase !== 'performing') return null;

  const participant = (state.participants || []).find((p) => p.id === participantId);
  if (!participant || (participant.castMemberIds || []).length === 0) return null;

  const indices = getLineIndices(state.selectedPlay);

  // Look for the next turn where this participant has lines
  for (let i = state.currentLineIndex + 1; i < indices.length; i++) {
    const linesAtIndex = getLinesAtIndex(state.selectedPlay, indices[i]);
    const participantLines = linesAtIndex.filter((l) =>
      (participant.castMemberIds || []).includes(l.castMemberId)
    );

    if (participantLines.length > 0) {
      return {
        lines: participantLines,
        turnsAway: i - state.currentLineIndex,
      };
    }
  }

  return null;
}

export function getTotalLineIndices(play: Play): number {
  return getLineIndices(play).length;
}

// Legacy function - keeping for compatibility but now uses line indices
export function getCurrentSpeaker(state: StageState): Participant | null {
  const speakers = getCurrentSpeakers(state);
  return speakers[0] || null;
}

export function getNextLineForParticipant(
  state: StageState,
  participantId: string
): { line: PlayLine; linesAway: number } | null {
  const result = getNextLinesForParticipant(state, participantId);
  if (!result || result.lines.length === 0) return null;
  return { line: result.lines[0], linesAway: result.turnsAway };
}

// Frontend-specific configuration
export interface StageConfig {
  wsUrl: string;
}

// Keep RoomConfig as alias for backward compatibility
export type RoomConfig = StageConfig;

function getWsUrl(): string {
  // Allow override via environment variable
  const envUrl =
    typeof process !== 'undefined' ? process.env.NEXT_PUBLIC_BACKEND_WS_URL : undefined;
  if (envUrl) {
    return envUrl;
  }

  if (typeof window === 'undefined') {
    return 'ws://localhost:8000/ws/stages';
  }

  // Use ws:// for http:// and wss:// for https://
  const { protocol, hostname } = window.location;
  const wsProtocol = protocol === 'https:' ? 'wss:' : 'ws:';
  return `${wsProtocol}//${hostname}:8000/ws/stages`;
}

export function getConfig(): StageConfig {
  return {
    wsUrl: getWsUrl(),
  };
}

// localStorage utilities for draft scripts
const DRAFTS_STORAGE_KEY = 'nosmallparts_drafts';

function generateDraftId(): string {
  return `draft-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

export function getDrafts(): Play[] {
  if (typeof window === 'undefined') return [];

  try {
    const stored = localStorage.getItem(DRAFTS_STORAGE_KEY);
    if (!stored) return [];
    return JSON.parse(stored);
  } catch {
    return [];
  }
}

export function getDraft(id: string): Play | null {
  const drafts = getDrafts();
  return drafts.find((d) => d.id === id) || null;
}

export function saveDraft(play: Partial<Play> & { title: string }): Play {
  const drafts = getDrafts();

  // If the play has an existing draft ID, update it
  const existingIndex = play.id ? drafts.findIndex((d) => d.id === play.id) : -1;

  const savedPlay: Play = {
    id: play.id || generateDraftId(),
    title: play.title,
    description: play.description || '',
    author: play.author,
    cast: play.cast || [],
    lines: play.lines || [],
  };

  if (existingIndex >= 0) {
    drafts[existingIndex] = savedPlay;
  } else {
    drafts.push(savedPlay);
  }

  if (typeof window !== 'undefined') {
    localStorage.setItem(DRAFTS_STORAGE_KEY, JSON.stringify(drafts));
  }

  return savedPlay;
}

export function deleteDraft(id: string): boolean {
  const drafts = getDrafts();
  const newDrafts = drafts.filter((d) => d.id !== id);

  if (newDrafts.length === drafts.length) return false;

  if (typeof window !== 'undefined') {
    localStorage.setItem(DRAFTS_STORAGE_KEY, JSON.stringify(newDrafts));
  }

  return true;
}

export function isDraft(id: string): boolean {
  return id.startsWith('draft-');
}
