import {
  groupLinesByIndex,
  getSortedLineIndices,
  getUsedCastMemberIds,
  getAvailableCastMembers,
  getFirstAvailableCastMemberId,
  getNextLineIndex,
  hasSimultaneousSpeakers,
  getSpeakerCountAtIndex,
  swapLineIndices,
  validatePlay,
} from '../lib/scriptUtils';
import { PlayLine, CastMember } from '../lib/types';

describe('scriptUtils', () => {
  // Sample test data
  const cast: CastMember[] = [
    { id: 'alice', name: 'Alice' },
    { id: 'bob', name: 'Bob' },
    { id: 'charlie', name: 'Charlie' },
  ];

  const singleSpeakerLines: PlayLine[] = [
    { castMemberId: 'alice', text: 'Hello!', index: 0 },
    { castMemberId: 'bob', text: 'Hi there!', index: 1 },
    { castMemberId: 'alice', text: 'How are you?', index: 2 },
  ];

  const simultaneousLines: PlayLine[] = [
    { castMemberId: 'alice', text: 'Hello!', index: 0 },
    { castMemberId: 'bob', text: 'Hello!', index: 0 }, // Same index = simultaneous
    { castMemberId: 'charlie', text: 'Hi everyone!', index: 1 },
  ];

  describe('groupLinesByIndex', () => {
    it('groups single speaker lines correctly', () => {
      const groups = groupLinesByIndex(singleSpeakerLines);

      expect(groups.size).toBe(3);
      expect(groups.get(0)?.lines).toHaveLength(1);
      expect(groups.get(1)?.lines).toHaveLength(1);
      expect(groups.get(2)?.lines).toHaveLength(1);
    });

    it('groups simultaneous speakers together', () => {
      const groups = groupLinesByIndex(simultaneousLines);

      expect(groups.size).toBe(2); // Only 2 unique indices
      expect(groups.get(0)?.lines).toHaveLength(2); // Alice and Bob speak simultaneously
      expect(groups.get(1)?.lines).toHaveLength(1);
    });

    it('tracks array indices correctly', () => {
      const groups = groupLinesByIndex(simultaneousLines);

      // Index 0 has two speakers at array positions 0 and 1
      expect(groups.get(0)?.arrayIndices).toEqual([0, 1]);
      // Index 1 has one speaker at array position 2
      expect(groups.get(1)?.arrayIndices).toEqual([2]);
    });

    it('handles empty lines array', () => {
      const groups = groupLinesByIndex([]);
      expect(groups.size).toBe(0);
    });

    it('handles three simultaneous speakers', () => {
      const threeSimultaneous: PlayLine[] = [
        { castMemberId: 'alice', text: 'Together!', index: 0 },
        { castMemberId: 'bob', text: 'Together!', index: 0 },
        { castMemberId: 'charlie', text: 'Together!', index: 0 },
      ];

      const groups = groupLinesByIndex(threeSimultaneous);
      expect(groups.size).toBe(1);
      expect(groups.get(0)?.lines).toHaveLength(3);
    });
  });

  describe('getSortedLineIndices', () => {
    it('returns unique sorted indices', () => {
      const indices = getSortedLineIndices(singleSpeakerLines);
      expect(indices).toEqual([0, 1, 2]);
    });

    it('returns unique indices for simultaneous speakers', () => {
      const indices = getSortedLineIndices(simultaneousLines);
      expect(indices).toEqual([0, 1]);
    });

    it('handles empty lines array', () => {
      const indices = getSortedLineIndices([]);
      expect(indices).toEqual([]);
    });

    it('sorts indices correctly when out of order', () => {
      const outOfOrder: PlayLine[] = [
        { castMemberId: 'alice', text: 'Third', index: 2 },
        { castMemberId: 'bob', text: 'First', index: 0 },
        { castMemberId: 'charlie', text: 'Second', index: 1 },
      ];
      expect(getSortedLineIndices(outOfOrder)).toEqual([0, 1, 2]);
    });

    it('handles gaps in indices', () => {
      const withGaps: PlayLine[] = [
        { castMemberId: 'alice', text: 'First', index: 0 },
        { castMemberId: 'bob', text: 'Third', index: 5 },
        { castMemberId: 'charlie', text: 'Second', index: 2 },
      ];
      expect(getSortedLineIndices(withGaps)).toEqual([0, 2, 5]);
    });
  });

  describe('getUsedCastMemberIds', () => {
    it('returns cast member ids used at a specific index', () => {
      const used = getUsedCastMemberIds(simultaneousLines, 0);
      expect(used).toContain('alice');
      expect(used).toContain('bob');
      expect(used).not.toContain('charlie');
    });

    it('returns single id for single speaker line', () => {
      const used = getUsedCastMemberIds(simultaneousLines, 1);
      expect(used).toEqual(['charlie']);
    });

    it('returns empty array for non-existent index', () => {
      const used = getUsedCastMemberIds(simultaneousLines, 999);
      expect(used).toEqual([]);
    });

    it('filters out empty cast member ids', () => {
      const linesWithEmpty: PlayLine[] = [
        { castMemberId: 'alice', text: 'Hello', index: 0 },
        { castMemberId: '', text: 'Empty', index: 0 },
      ];
      const used = getUsedCastMemberIds(linesWithEmpty, 0);
      expect(used).toEqual(['alice']);
    });
  });

  describe('getAvailableCastMembers', () => {
    it('returns cast members not used in the line', () => {
      const available = getAvailableCastMembers(cast, simultaneousLines, 0);

      expect(available).toHaveLength(1);
      expect(available[0].id).toBe('charlie');
    });

    it('returns all cast members for new line index', () => {
      const available = getAvailableCastMembers(cast, simultaneousLines, 999);

      expect(available).toHaveLength(3);
    });

    it('returns empty array when all cast members are used', () => {
      const allUsed: PlayLine[] = [
        { castMemberId: 'alice', text: 'Hi', index: 0 },
        { castMemberId: 'bob', text: 'Hi', index: 0 },
        { castMemberId: 'charlie', text: 'Hi', index: 0 },
      ];

      const available = getAvailableCastMembers(cast, allUsed, 0);
      expect(available).toHaveLength(0);
    });

    it('works with empty lines array', () => {
      const available = getAvailableCastMembers(cast, [], 0);
      expect(available).toHaveLength(3);
    });
  });

  describe('getFirstAvailableCastMemberId', () => {
    it('returns first available cast member id', () => {
      const firstAvailable = getFirstAvailableCastMemberId(cast, simultaneousLines, 0);
      expect(firstAvailable).toBe('charlie');
    });

    it('returns first cast member for new line', () => {
      const firstAvailable = getFirstAvailableCastMemberId(cast, [], 0);
      expect(firstAvailable).toBe('alice');
    });

    it('returns empty string when no cast members available', () => {
      const allUsed: PlayLine[] = [
        { castMemberId: 'alice', text: 'Hi', index: 0 },
        { castMemberId: 'bob', text: 'Hi', index: 0 },
        { castMemberId: 'charlie', text: 'Hi', index: 0 },
      ];

      const firstAvailable = getFirstAvailableCastMemberId(cast, allUsed, 0);
      expect(firstAvailable).toBe('');
    });

    it('returns empty string for empty cast', () => {
      const firstAvailable = getFirstAvailableCastMemberId([], [], 0);
      expect(firstAvailable).toBe('');
    });
  });

  describe('getNextLineIndex', () => {
    it('returns 0 for empty lines', () => {
      expect(getNextLineIndex([])).toBe(0);
    });

    it('returns max index + 1', () => {
      expect(getNextLineIndex(singleSpeakerLines)).toBe(3);
    });

    it('handles simultaneous lines correctly', () => {
      // simultaneousLines has indices 0 and 1
      expect(getNextLineIndex(simultaneousLines)).toBe(2);
    });

    it('handles gaps in indices', () => {
      const withGaps: PlayLine[] = [
        { castMemberId: 'alice', text: 'First', index: 0 },
        { castMemberId: 'bob', text: 'Last', index: 10 },
      ];
      expect(getNextLineIndex(withGaps)).toBe(11);
    });
  });

  describe('hasSimultaneousSpeakers', () => {
    it('returns true for simultaneous speakers', () => {
      expect(hasSimultaneousSpeakers(simultaneousLines, 0)).toBe(true);
    });

    it('returns false for single speaker', () => {
      expect(hasSimultaneousSpeakers(simultaneousLines, 1)).toBe(false);
    });

    it('returns false for non-existent index', () => {
      expect(hasSimultaneousSpeakers(simultaneousLines, 999)).toBe(false);
    });

    it('returns false for empty lines', () => {
      expect(hasSimultaneousSpeakers([], 0)).toBe(false);
    });
  });

  describe('getSpeakerCountAtIndex', () => {
    it('returns correct count for simultaneous speakers', () => {
      expect(getSpeakerCountAtIndex(simultaneousLines, 0)).toBe(2);
    });

    it('returns 1 for single speaker', () => {
      expect(getSpeakerCountAtIndex(simultaneousLines, 1)).toBe(1);
    });

    it('returns 0 for non-existent index', () => {
      expect(getSpeakerCountAtIndex(simultaneousLines, 999)).toBe(0);
    });

    it('counts three simultaneous speakers', () => {
      const threeSpeakers: PlayLine[] = [
        { castMemberId: 'alice', text: 'Hi', index: 0 },
        { castMemberId: 'bob', text: 'Hi', index: 0 },
        { castMemberId: 'charlie', text: 'Hi', index: 0 },
      ];
      expect(getSpeakerCountAtIndex(threeSpeakers, 0)).toBe(3);
    });
  });

  describe('swapLineIndices', () => {
    it('swaps indices between two lines', () => {
      const swapped = swapLineIndices(singleSpeakerLines, 0, 1);

      const aliceLine = swapped.find((l) => l.castMemberId === 'alice' && l.text === 'Hello!');
      const bobLine = swapped.find((l) => l.castMemberId === 'bob');

      expect(aliceLine?.index).toBe(1);
      expect(bobLine?.index).toBe(0);
    });

    it('swaps all simultaneous speakers together', () => {
      const swapped = swapLineIndices(simultaneousLines, 0, 1);

      // Both Alice and Bob should now have index 1
      const indexZeroSpeakers = swapped.filter((l) => l.index === 0);
      const indexOneSpeakers = swapped.filter((l) => l.index === 1);

      expect(indexZeroSpeakers).toHaveLength(1);
      expect(indexZeroSpeakers[0].castMemberId).toBe('charlie');

      expect(indexOneSpeakers).toHaveLength(2);
      expect(indexOneSpeakers.map((l) => l.castMemberId).sort()).toEqual(['alice', 'bob']);
    });

    it('does not modify lines with other indices', () => {
      const lines: PlayLine[] = [
        { castMemberId: 'alice', text: 'First', index: 0 },
        { castMemberId: 'bob', text: 'Second', index: 1 },
        { castMemberId: 'charlie', text: 'Third', index: 2 },
      ];

      const swapped = swapLineIndices(lines, 0, 1);
      const charlieLine = swapped.find((l) => l.castMemberId === 'charlie');

      expect(charlieLine?.index).toBe(2); // Unchanged
    });

    it('returns unchanged lines if indices do not exist', () => {
      const swapped = swapLineIndices(singleSpeakerLines, 99, 100);
      expect(swapped).toEqual(singleSpeakerLines);
    });
  });

  describe('validatePlay', () => {
    it('returns valid for complete play', () => {
      const result = validatePlay('My Play', cast, singleSpeakerLines);
      expect(result.valid).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('returns error for empty title', () => {
      const result = validatePlay('', cast, singleSpeakerLines);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Please enter a title');
    });

    it('returns error for whitespace-only title', () => {
      const result = validatePlay('   ', cast, singleSpeakerLines);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Please enter a title');
    });

    it('returns error for empty cast', () => {
      const result = validatePlay('My Play', [], singleSpeakerLines);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Please add at least one cast member');
    });

    it('returns error for empty lines', () => {
      const result = validatePlay('My Play', cast, []);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Please add at least one line');
    });

    it('returns error for line without cast member', () => {
      const invalidLines: PlayLine[] = [{ castMemberId: '', text: 'Hello', index: 0 }];
      const result = validatePlay('My Play', cast, invalidLines);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Line 1 needs a character and dialogue');
      expect(result.invalidLineIndices).toEqual(new Set([0]));
    });

    it('returns error for line without text', () => {
      const invalidLines: PlayLine[] = [{ castMemberId: 'alice', text: '', index: 0 }];
      const result = validatePlay('My Play', cast, invalidLines);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Line 1 needs a character and dialogue');
      expect(result.invalidLineIndices).toEqual(new Set([0]));
    });

    it('returns error for line with whitespace-only text', () => {
      const invalidLines: PlayLine[] = [{ castMemberId: 'alice', text: '   ', index: 0 }];
      const result = validatePlay('My Play', cast, invalidLines);
      expect(result.valid).toBe(false);
      expect(result.error).toBe('Line 1 needs a character and dialogue');
      expect(result.invalidLineIndices).toEqual(new Set([0]));
    });

    it('validates simultaneous lines correctly', () => {
      const result = validatePlay('My Play', cast, simultaneousLines);
      expect(result.valid).toBe(true);
    });
  });
});
