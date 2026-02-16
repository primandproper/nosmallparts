'use client';

import { PlayLine, CastMember } from '@/lib/types';

interface SingleLineProps {
  line: PlayLine;
  cast: CastMember[];
  availableCast: CastMember[]; // Only cast members not already used in this line group
  onUpdate: (line: PlayLine) => void;
  onDelete?: () => void; // Optional - only shown if there's more than one speaker
  showDeleteButton: boolean;
}

function SingleLine({
  line,
  cast,
  availableCast,
  onUpdate,
  onDelete,
  showDeleteButton,
}: SingleLineProps) {
  // Current cast member should always be shown in dropdown even if "taken"
  const currentMember = cast.find((c) => c.id === line.castMemberId);
  const dropdownOptions = currentMember
    ? [currentMember, ...availableCast.filter((c) => c.id !== currentMember.id)]
    : availableCast;
  return (
    <div className="space-y-2 p-3 rounded-lg bg-[var(--black)]/30">
      {/* Top row: Character, Delete - direction moved below dialogue for tab flow */}
      <div className="flex gap-3 items-center">
        {/* Character selector - first in tab order */}
        <div className="w-40 shrink-0">
          <select
            value={line.castMemberId}
            onChange={(e) => onUpdate({ ...line, castMemberId: e.target.value })}
            className="input-elegant w-full px-3 py-2 rounded-lg text-[var(--cream)] text-sm"
          >
            <option value="">Character...</option>
            {dropdownOptions.map((member) => (
              <option key={member.id} value={member.id}>
                {member.name}
              </option>
            ))}
          </select>
        </div>

        {/* Delete button - not in main tab flow for transcription */}
        {showDeleteButton && onDelete && (
          <button
            type="button"
            onClick={onDelete}
            tabIndex={-1}
            className="p-2 rounded hover:bg-red-500/20 text-red-400 opacity-60 hover:opacity-100 transition-opacity shrink-0"
            title="Remove this line"
          >
            ✕
          </button>
        )}
      </div>

      {/* Dialogue textarea - second in tab order (main typing area) */}
      <textarea
        value={line.text}
        onChange={(e) => onUpdate({ ...line, text: e.target.value })}
        placeholder="Enter dialogue..."
        className="input-elegant w-full px-3 py-3 rounded-lg text-[var(--cream)] text-base min-h-[80px] resize-y"
        rows={2}
      />

      {/* Direction - last, tab past when not needed */}
      <input
        type="text"
        value={line.direction || ''}
        onChange={(e) => onUpdate({ ...line, direction: e.target.value || undefined })}
        placeholder="Stage direction (optional, e.g., angrily, whispering)"
        className="input-elegant w-full px-3 py-2 rounded-lg text-[var(--cream)] text-sm italic opacity-80"
      />
    </div>
  );
}

interface LineGroupEditorProps {
  lineIndex: number;
  lineIdx: number; // Actual line index (for validation highlighting)
  lines: PlayLine[]; // All lines spoken simultaneously
  cast: CastMember[];
  invalidLineIndices: Set<number> | null;
  onUpdateLine: (lineArrayIndex: number, line: PlayLine) => void;
  onDeleteLine: (lineArrayIndex: number) => void;
  onAddSimultaneous: () => void;
  onDeleteLineGroup: () => void;
  onMoveLineGroup: (direction: 'up' | 'down') => void;
  canMoveUp: boolean;
  canMoveDown: boolean;
  lineArrayIndices: number[]; // Maps local index to the array index in the parent
}

export function TurnEditor({
  lineIndex,
  lineIdx,
  lines,
  cast,
  invalidLineIndices,
  onUpdateLine,
  onDeleteLine,
  onAddSimultaneous,
  onDeleteLineGroup,
  onMoveLineGroup,
  canMoveUp,
  canMoveDown,
  lineArrayIndices,
}: LineGroupEditorProps) {
  const isInvalid = invalidLineIndices?.has(lineIdx) ?? false;

  return (
    <div
      className={`p-4 rounded-xl border space-y-3 transition-colors ${
        isInvalid
          ? 'bg-red-500/10 border-red-500/50'
          : 'bg-[var(--charcoal)]/30 border-[var(--cream)]/10'
      }`}
    >
      {/* Line header */}
      <div className="flex items-center justify-between">
        <span className="text-[var(--gold)] text-xs font-medium">
          Line {lineIndex + 1}
          {isInvalid && (
            <span className="ml-2 text-red-400 text-xs font-normal">
              — needs character and dialogue
            </span>
          )}
          {lines.length > 1 && !isInvalid && (
            <span className="ml-2 px-2 py-0.5 rounded-full text-xs bg-purple-500/20 text-purple-400 border border-purple-500/30">
              {lines.length} speaking
            </span>
          )}
        </span>
        <div className="flex items-center gap-1">
          {canMoveUp && (
            <button
              type="button"
              tabIndex={-1}
              onClick={() => onMoveLineGroup('up')}
              className="p-1.5 rounded hover:bg-[var(--cream)]/10 text-[var(--cream)] opacity-40 hover:opacity-100 transition-opacity"
              title="Move line up"
            >
              ↑
            </button>
          )}
          {canMoveDown && (
            <button
              type="button"
              tabIndex={-1}
              onClick={() => onMoveLineGroup('down')}
              className="p-1.5 rounded hover:bg-[var(--cream)]/10 text-[var(--cream)] opacity-40 hover:opacity-100 transition-opacity"
              title="Move line down"
            >
              ↓
            </button>
          )}
          <button
            type="button"
            tabIndex={-1}
            onClick={onDeleteLineGroup}
            className="p-1.5 rounded hover:bg-red-500/20 text-red-400 opacity-60 hover:opacity-100 transition-opacity ml-2"
            title="Delete this line"
          >
            ✕
          </button>
        </div>
      </div>

      {/* Speakers in this line */}
      <div className="space-y-2">
        {lines.map((line, localIndex) => {
          // Get cast members already used by OTHER speakers in this line
          const usedCastIds = lines
            .filter((_, i) => i !== localIndex)
            .map((l) => l.castMemberId)
            .filter(Boolean);
          const availableCast = cast.filter((c) => !usedCastIds.includes(c.id));

          return (
            <SingleLine
              key={lineArrayIndices[localIndex]}
              line={line}
              cast={cast}
              availableCast={availableCast}
              onUpdate={(updated) => onUpdateLine(lineArrayIndices[localIndex], updated)}
              onDelete={() => onDeleteLine(lineArrayIndices[localIndex])}
              showDeleteButton={lines.length > 1}
            />
          );
        })}
      </div>

      {/* Add simultaneous speaker - subtle link, skipped when tabbing */}
      {(() => {
        const usedCastIds = lines.map((l) => l.castMemberId).filter(Boolean);
        const availableCastCount = cast.filter((c) => !usedCastIds.includes(c.id)).length;
        const canAddMore = availableCastCount > 0;

        return (
          <button
            type="button"
            onClick={onAddSimultaneous}
            disabled={!canAddMore}
            tabIndex={-1}
            className="text-xs text-[var(--cream)]/40 hover:text-[var(--cream)]/70 transition-colors disabled:opacity-30 disabled:cursor-not-allowed disabled:hover:text-[var(--cream)]/40"
            title={
              canAddMore
                ? 'Add another character speaking at the same time'
                : 'All cast members are already speaking'
            }
          >
            + add simultaneous speaker
          </button>
        );
      })()}
    </div>
  );
}
