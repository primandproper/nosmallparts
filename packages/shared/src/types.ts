// Core domain types shared by server and app.

export interface CastMember {
  id: string;
  name: string;
  description?: string;
}

export interface PlayLine {
  castMemberId: string;
  text: string;
  index: number;
  direction?: string; // Stage direction/emotion, e.g., "skeptically"
  scrollMode?: 'auto' | 'manual';
  autoScrollWpm?: number;
}

export interface PlayRatingStats {
  averageRating: number;
  ratingCount: number;
}

export interface ScriptTag {
  id: string;
  name: string;
  description?: string;
}

export interface Play {
  id: string;
  title: string;
  description: string;
  author?: string;
  isPublic?: boolean;
  cast: CastMember[];
  lines: PlayLine[];
  tags?: ScriptTag[];
  ratingStats?: PlayRatingStats;
  userRating?: number;
}

export interface Participant {
  id: string;
  name: string;
  castMemberIds: string[];
  isReady: boolean;
  isAudience: boolean;
  isHost: boolean;
  forceMuted: boolean;
}

export type StagePhase = 'lobby' | 'countdown' | 'performing' | 'discussion';

export interface StageState {
  id: string;
  code: string;
  participants: Participant[];
  selectedPlay: Play | null;
  currentLineIndex: number;
  phase: StagePhase;
  pausedByHost: boolean;
  directorMode: boolean;
}

export interface ChatMessage {
  id: string;
  senderId: string;
  senderName: string;
  text: string;
  castOnly: boolean;
  timestamp: number;
}

export interface SystemMessage {
  id: string;
  text: string;
  timestamp: number;
}

export interface RoleOffer {
  offerId: string;
  castMemberId: string;
  castMemberName: string;
}

// Client -> Server WebSocket messages
export type ClientMessage =
  | { type: 'join'; stageId: string; participantName: string }
  | { type: 'heartbeat' }
  | { type: 'toggleCastMember'; participantId: string; castMemberId: string }
  | { type: 'setAudience'; participantId: string; isAudience: boolean }
  | { type: 'selectPlay'; playId: string }
  | { type: 'setReady'; participantId: string; isReady: boolean }
  | { type: 'sendChat'; participantId: string; text: string; castOnly: boolean }
  | { type: 'startPlay' }
  | { type: 'nextLine' }
  | { type: 'resetPlay' }
  // Host actions (server gates on requireHost)
  | { type: 'hostPause' }
  | { type: 'hostResume' }
  | { type: 'hostBackLine' }
  | { type: 'hostSkipLine' }
  | { type: 'hostEndPlay' }
  | { type: 'hostReassignRole'; targetParticipantId: string; castMemberId: string }
  | { type: 'hostUnassignRole'; targetParticipantId: string; castMemberId: string }
  | { type: 'hostTransferHost'; targetParticipantId: string }
  | { type: 'hostKick'; targetParticipantId: string }
  | { type: 'hostBan'; targetParticipantId: string }
  | { type: 'hostMute'; targetParticipantId: string; muted: boolean }
  | { type: 'hostToggleDirectorMode'; enabled: boolean }
  | { type: 'respondRoleOffer'; offerId: string; accept: boolean };

// Server -> Client WebSocket messages
export type ServerMessage =
  | { type: 'stageState'; state: StageState }
  | { type: 'joined'; participantId: string; stageState: StageState }
  | { type: 'error'; message: string }
  | { type: 'availablePlays'; plays: Play[] }
  | { type: 'countdownStart'; duration: number }
  | { type: 'castMembersUpdated'; participants: Participant[] }
  | { type: 'chatMessage'; message: ChatMessage }
  | { type: 'systemMessage'; message: SystemMessage }
  | { type: 'roleOffer'; offer: RoleOffer }
  | { type: 'roleOfferCleared'; offerId: string }
  | { type: 'playStarted' }
  | { type: 'lineAdvanced'; currentLineIndex: number }
  | { type: 'playEnded' }
  | { type: 'sessionReplaced' };
