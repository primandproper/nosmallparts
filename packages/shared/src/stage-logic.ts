// Pure logic shared by server and app: resolving speakers, lines, readiness, etc.
// No platform dependencies.

import type { CastMember, Participant, Play, PlayLine, StageState } from './types';

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
  const performers = getPerformers(state);
  const assignedCastMemberIds = new Set(performers.flatMap((p) => p.castMemberIds));
  return (state.selectedPlay.cast || []).every((cm) => assignedCastMemberIds.has(cm.id));
}

export function areAllParticipantsReady(state: StageState): boolean {
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
  if (currentIndex === undefined) return [];
  return getLinesAtIndex(state.selectedPlay, currentIndex);
}

export function getCurrentSpeakerCastMemberIds(state: StageState): string[] {
  return getCurrentLines(state).map((l) => l.castMemberId);
}

export function getCurrentSpeakers(state: StageState): Participant[] {
  const speakerIds = getCurrentSpeakerCastMemberIds(state);
  return (state.participants || []).filter((p) =>
    (p.castMemberIds || []).some((id) => speakerIds.includes(id))
  );
}

export function isParticipantSpeaking(state: StageState, participantId: string): boolean {
  return getCurrentSpeakers(state).some((s) => s.id === participantId);
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
  for (let i = state.currentLineIndex + 1; i < indices.length; i++) {
    const idx = indices[i];
    if (idx === undefined) continue;
    const linesAtIndex = getLinesAtIndex(state.selectedPlay, idx);
    const participantLines = linesAtIndex.filter((l) =>
      (participant.castMemberIds || []).includes(l.castMemberId)
    );
    if (participantLines.length > 0) {
      return { lines: participantLines, turnsAway: i - state.currentLineIndex };
    }
  }
  return null;
}

export function getTotalLineIndices(play: Play): number {
  return getLineIndices(play).length;
}

export function getCurrentSpeaker(state: StageState): Participant | null {
  return getCurrentSpeakers(state)[0] || null;
}
