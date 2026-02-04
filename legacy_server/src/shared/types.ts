// Shared types for the play performance app
// This is a local copy for the server - the frontend has its own copy

export interface CastMember {
  id: string;
  name: string;
}

export interface PlayLine {
  castMemberId: string;
  text: string;
  index: number;
  direction?: string;
  scrollMode?: 'auto' | 'manual';
  autoScrollWpm?: number;
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

export interface StageState {
  id: string;
  participants: Participant[];
  selectedPlay: Play | null;
  currentLineIndex: number;
  phase: StagePhase;
}

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
export function getLineIndices(play: Play): number[] {
  const indices = new Set(play.lines.map((l) => l.index));
  return Array.from(indices).sort((a, b) => a - b);
}
