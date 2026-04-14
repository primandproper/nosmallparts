// Pure helpers for editing & validating play scripts. No platform deps.

import type { CastMember, PlayLine } from './types';

export function groupLinesByIndex(
  lines: PlayLine[]
): Map<number, { lines: PlayLine[]; arrayIndices: number[] }> {
  const groups = new Map<number, { lines: PlayLine[]; arrayIndices: number[] }>();
  lines.forEach((line, arrayIndex) => {
    const existing = groups.get(line.index);
    if (existing) {
      existing.lines.push(line);
      existing.arrayIndices.push(arrayIndex);
    } else {
      groups.set(line.index, { lines: [line], arrayIndices: [arrayIndex] });
    }
  });
  return groups;
}

export function getSortedLineIndices(lines: PlayLine[]): number[] {
  const indices = new Set(lines.map((l) => l.index));
  return Array.from(indices).sort((a, b) => a - b);
}

export function getUsedCastMemberIds(lines: PlayLine[], lineIndex: number): string[] {
  return lines
    .filter((l) => l.index === lineIndex)
    .map((l) => l.castMemberId)
    .filter(Boolean);
}

export function getAvailableCastMembers(
  cast: CastMember[],
  lines: PlayLine[],
  lineIndex: number
): CastMember[] {
  const used = getUsedCastMemberIds(lines, lineIndex);
  return cast.filter((c) => !used.includes(c.id));
}

export function getFirstAvailableCastMemberId(
  cast: CastMember[],
  lines: PlayLine[],
  lineIndex: number
): string {
  const available = getAvailableCastMembers(cast, lines, lineIndex);
  return available.length > 0 ? available[0]!.id : '';
}

export function getNextLineIndex(lines: PlayLine[]): number {
  if (lines.length === 0) return 0;
  return Math.max(...lines.map((l) => l.index)) + 1;
}

export function hasSimultaneousSpeakers(lines: PlayLine[], lineIndex: number): boolean {
  return lines.filter((l) => l.index === lineIndex).length > 1;
}

export function getInvalidLineIndices(lines: PlayLine[]): Set<number> {
  return new Set(lines.filter((l) => !l.castMemberId || !l.text.trim()).map((l) => l.index));
}

export function validatePlay(
  title: string,
  cast: CastMember[],
  lines: PlayLine[]
): { valid: boolean; error?: string; invalidLineIndices?: Set<number> } {
  if (!title.trim()) return { valid: false, error: 'Please enter a title' };
  if (cast.length === 0) return { valid: false, error: 'Please add at least one cast member' };
  if (lines.length === 0) return { valid: false, error: 'Please add at least one line' };
  const invalid = getInvalidLineIndices(lines);
  if (invalid.size > 0) {
    const sorted = getSortedLineIndices(lines);
    const displayNumbers = Array.from(invalid)
      .map((i) => sorted.indexOf(i) + 1)
      .filter((n) => n > 0)
      .sort((a, b) => a - b);
    const list =
      displayNumbers.length <= 3
        ? displayNumbers.join(', ')
        : `${displayNumbers.slice(0, 2).join(', ')}, and ${displayNumbers.length - 2} more`;
    return {
      valid: false,
      error: `Line${displayNumbers.length > 1 ? 's' : ''} ${list} need a character and dialogue`,
      invalidLineIndices: invalid,
    };
  }
  return { valid: true };
}
