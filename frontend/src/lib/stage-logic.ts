// Pure helpers over a stage's state: who is performing, which lines are
// current, who speaks them.

import type { CastMember, Play, PlayLine } from '../generated/nosmallparts/scripts/v1/scripts';
import { type Participant, StagePhase, type StageState } from '../generated/nosmallparts/stage/v1/stage';

export function getCastMemberById(play: Play, castMemberId: string): CastMember | undefined {
  return play.cast.find((c) => c.id === castMemberId);
}

export function getCastMemberNamesForLine(play: Play, line: PlayLine): string[] {
  return line.castMemberIds.map((id) => getCastMemberById(play, id)?.name).filter((n): n is string => !!n);
}

export function getCastMemberNameForLine(play: Play, line: PlayLine): string {
  const names = getCastMemberNamesForLine(play, line);
  if (names.length === 0) return 'Unknown';
  return names.join(' & ');
}

export function getPerformers(state: StageState): Participant[] {
  return state.participants.filter((p) => !p.isAudience);
}

export function getAudience(state: StageState): Participant[] {
  return state.participants.filter((p) => p.isAudience);
}

export function areAllCastMembersAssigned(state: StageState): boolean {
  if (!state.selectedPlay) return false;
  const assigned = new Set(getPerformers(state).flatMap((p) => p.castMemberIds));
  return state.selectedPlay.cast.every((cm) => assigned.has(cm.id));
}

export function areAllParticipantsReady(state: StageState): boolean {
  const performers = getPerformers(state);
  return performers.length > 0 && performers.every((p) => p.isReady);
}

export function canStartPlay(state: StageState): boolean {
  return areAllCastMembersAssigned(state) && areAllParticipantsReady(state);
}

/** getLineIndices are the play's distinct line indexes, in order; the playhead counts positions in it. */
export function getLineIndices(play: Play): number[] {
  return Array.from(new Set(play.lines.map((l) => l.index))).sort((a, b) => a - b);
}

export function getLinesAtIndex(play: Play, index: number): PlayLine[] {
  return play.lines.filter((l) => l.index === index);
}

export function getCurrentLines(state: StageState): PlayLine[] {
  if (!state.selectedPlay || state.phase !== StagePhase.STAGE_PHASE_PERFORMING) return [];
  const indices = getLineIndices(state.selectedPlay);
  const currentIndex = indices[state.currentLineIndex];
  if (currentIndex === undefined) return [];
  return getLinesAtIndex(state.selectedPlay, currentIndex);
}

export function getCurrentSpeakerCastMemberIds(state: StageState): string[] {
  return Array.from(new Set(getCurrentLines(state).flatMap((l) => l.castMemberIds)));
}

export function getCurrentSpeakers(state: StageState): Participant[] {
  const speakerIds = getCurrentSpeakerCastMemberIds(state);
  return state.participants.filter((p) => p.castMemberIds.some((id) => speakerIds.includes(id)));
}

export function isParticipantSpeaking(state: StageState, participantId: string): boolean {
  return getCurrentSpeakers(state).some((s) => s.id === participantId);
}

export function getCurrentLinesForParticipant(state: StageState, participantId: string): PlayLine[] {
  const participant = state.participants.find((p) => p.id === participantId);
  if (!participant) return [];
  return getCurrentLines(state).filter((l) => l.castMemberIds.some((id) => participant.castMemberIds.includes(id)));
}

export function getNextLinesForParticipant(
  state: StageState,
  participantId: string,
): { lines: PlayLine[]; turnsAway: number } | null {
  if (!state.selectedPlay || state.phase !== StagePhase.STAGE_PHASE_PERFORMING) return null;
  const participant = state.participants.find((p) => p.id === participantId);
  if (!participant || participant.castMemberIds.length === 0) return null;

  const indices = getLineIndices(state.selectedPlay);
  for (let i = state.currentLineIndex + 1; i < indices.length; i++) {
    const idx = indices[i];
    if (idx === undefined) continue;
    const mine = getLinesAtIndex(state.selectedPlay, idx).filter((l) =>
      l.castMemberIds.some((id) => participant.castMemberIds.includes(id)),
    );
    if (mine.length > 0) return { lines: mine, turnsAway: i - state.currentLineIndex };
  }
  return null;
}

export function getTotalLineIndices(play: Play): number {
  return getLineIndices(play).length;
}

export function getCurrentSpeaker(state: StageState): Participant | null {
  return getCurrentSpeakers(state)[0] ?? null;
}
