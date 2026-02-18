import { PlayLine, CastMember } from './types';

/**
 * Groups lines by their index (for simultaneous speakers)
 * Lines with the same index are spoken at the same time by different cast members
 */
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

/**
 * Get sorted unique line indices
 */
export function getSortedLineIndices(lines: PlayLine[]): number[] {
  const indices = new Set(lines.map((l) => l.index));
  return Array.from(indices).sort((a, b) => a - b);
}

/**
 * Find cast members that are already assigned to a specific line index
 */
export function getUsedCastMemberIds(lines: PlayLine[], lineIndex: number): string[] {
  return lines
    .filter((l) => l.index === lineIndex)
    .map((l) => l.castMemberId)
    .filter(Boolean);
}

/**
 * Find cast members available for a line (not already speaking in that line)
 */
export function getAvailableCastMembers(
  cast: CastMember[],
  lines: PlayLine[],
  lineIndex: number
): CastMember[] {
  const usedCastIds = getUsedCastMemberIds(lines, lineIndex);
  return cast.filter((c) => !usedCastIds.includes(c.id));
}

/**
 * Get the first available cast member for a simultaneous speaker
 * Returns empty string if no cast members are available
 */
export function getFirstAvailableCastMemberId(
  cast: CastMember[],
  lines: PlayLine[],
  lineIndex: number
): string {
  const available = getAvailableCastMembers(cast, lines, lineIndex);
  return available.length > 0 ? available[0].id : '';
}

/**
 * Calculate the next line index for a new line
 */
export function getNextLineIndex(lines: PlayLine[]): number {
  if (lines.length === 0) return 0;
  return Math.max(...lines.map((l) => l.index)) + 1;
}

/**
 * Check if a line index has simultaneous speakers (multiple lines at same index)
 */
export function hasSimultaneousSpeakers(lines: PlayLine[], lineIndex: number): boolean {
  return lines.filter((l) => l.index === lineIndex).length > 1;
}

/**
 * Get count of speakers at a specific line index
 */
export function getSpeakerCountAtIndex(lines: PlayLine[], lineIndex: number): number {
  return lines.filter((l) => l.index === lineIndex).length;
}

/**
 * Swap line indices between two positions (for reordering)
 */
export function swapLineIndices(lines: PlayLine[], fromIndex: number, toIndex: number): PlayLine[] {
  return lines.map((line) => {
    if (line.index === fromIndex) {
      return { ...line, index: toIndex };
    }
    if (line.index === toIndex) {
      return { ...line, index: fromIndex };
    }
    return line;
  });
}

/**
 * Returns the set of line indices (group keys) that have at least one line
 * missing character or dialogue. Used to highlight invalid lines in the UI.
 */
export function getInvalidLineIndices(lines: PlayLine[]): Set<number> {
  return new Set(lines.filter((l) => !l.castMemberId || !l.text.trim()).map((l) => l.index));
}

/**
 * Returns human-readable line numbers (1-based) for invalid lines.
 * Used for error messages.
 */
export function getInvalidLineNumbers(lines: PlayLine[]): {
  invalidIndices: Set<number>;
  displayNumbers: number[];
} {
  const invalidIndices = getInvalidLineIndices(lines);
  const sortedIndices = getSortedLineIndices(lines);
  const displayNumbers = Array.from(invalidIndices)
    .map((lineIdx) => sortedIndices.indexOf(lineIdx) + 1)
    .filter((n) => n > 0)
    .sort((a, b) => a - b);
  return { invalidIndices, displayNumbers };
}

/**
 * Validate a play before publishing
 */
export function validatePlay(
  title: string,
  cast: CastMember[],
  lines: PlayLine[]
): { valid: boolean; error?: string; invalidLineIndices?: Set<number> } {
  if (!title.trim()) {
    return { valid: false, error: 'Please enter a title' };
  }
  if (cast.length === 0) {
    return { valid: false, error: 'Please add at least one cast member' };
  }
  if (lines.length === 0) {
    return { valid: false, error: 'Please add at least one line' };
  }
  const { invalidIndices, displayNumbers } = getInvalidLineNumbers(lines);
  if (invalidIndices.size > 0) {
    const lineList =
      displayNumbers.length <= 3
        ? displayNumbers.join(', ')
        : `${displayNumbers.slice(0, 2).join(', ')}, and ${displayNumbers.length - 2} more`;
    const verb = displayNumbers.length > 1 ? 'need' : 'needs';
    return {
      valid: false,
      error: `Line${displayNumbers.length > 1 ? 's' : ''} ${lineList} ${verb} a character and dialogue`,
      invalidLineIndices: invalidIndices,
    };
  }
  return { valid: true };
}
