'use client';

import {
  StageState,
  getCastMemberById,
  getCastMemberNameForLine,
  getCurrentLinesForParticipant,
  getNextLinesForParticipant,
  isParticipantSpeaking,
  getTotalLineIndices,
} from '@/lib/types';
import { LineDisplay, DirectionBadge } from './LineDisplay';

interface StageProps {
  stageState: StageState;
  participantId: string | null;
  onNextLine: () => void;
}

export function Stage({ stageState, participantId, onNextLine }: StageProps) {
  if (!stageState.selectedPlay || !participantId) {
    return null;
  }

  const currentParticipant = (stageState.participants || []).find((p) => p.id === participantId);
  const isMyTurn = isParticipantSpeaking(stageState, participantId);
  const myCurrentLines = getCurrentLinesForParticipant(stageState, participantId);
  const myNextLines = getNextLinesForParticipant(stageState, participantId);

  const isDirector = currentParticipant?.isDirector || false;
  const hasDirector = (stageState.participants || []).some((p) => p.isDirector);

  // If there's a director, only they can advance. Otherwise, current speaker advances.
  const canAdvance = hasDirector ? isDirector : isMyTurn;

  const totalTurns = getTotalLineIndices(stageState.selectedPlay);

  return (
    <div className="card-elegant rounded-2xl p-6 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-3">
          <span className="text-2xl">🎭</span>
          <h2 className="font-[Cormorant_Garamond] text-xl font-semibold text-gold-gradient">
            {stageState.selectedPlay.title}
          </h2>
        </div>
        <div className="bg-[var(--charcoal)] px-3 py-1 rounded-full">
          <span className="text-[var(--gold)] text-sm font-medium">
            {stageState.currentLineIndex + 1}
          </span>
          <span className="text-[var(--cream)] opacity-40 text-sm"> / {totalTurns}</span>
        </div>
      </div>

      {/* Current speaker's lines */}
      {isMyTurn && myCurrentLines.length > 0 && (
        <div className="relative p-6 rounded-xl bg-gradient-to-br from-[var(--gold)]/20 to-[var(--velvet)]/40 border border-[var(--gold)]/30">
          {/* Spotlight effect */}
          <div className="absolute inset-0 rounded-xl overflow-hidden">
            <div className="absolute top-0 left-1/2 -translate-x-1/2 w-48 h-32 bg-[var(--gold)]/10 blur-3xl" />
          </div>

          <div className="relative">
            <p className="text-[var(--gold)] text-xs font-medium mb-3 tracking-wider uppercase">
              ✨ Your Line{myCurrentLines.length > 1 ? 's' : ''}
            </p>
            {myCurrentLines.map((line, index) => (
              <div key={index} className="mb-4 last:mb-0">
                <span className="text-[var(--gold)] opacity-60 text-sm">
                  {getCastMemberNameForLine(stageState.selectedPlay!, line)}:
                </span>
                <div className="mt-2">
                  <LineDisplay line={line} play={stageState.selectedPlay!} isCurrentLine={true} />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Waiting for turn */}
      {!isMyTurn && currentParticipant && (currentParticipant.castMemberIds || []).length > 0 && (
        <div className="p-4 rounded-xl bg-[var(--charcoal)]/50 border border-[var(--cream)]/10">
          {myNextLines ? (
            <>
              <p className="text-[var(--cream)] opacity-40 text-xs font-medium mb-2 tracking-wider uppercase">
                Coming up ({myNextLines.turnsAway} turn{myNextLines.turnsAway > 1 ? 's' : ''})
              </p>
              {myNextLines.lines.map((line, index) => (
                <div key={index} className="mb-3 last:mb-0">
                  <span className="text-[var(--gold)] opacity-40 text-xs">
                    {getCastMemberNameForLine(stageState.selectedPlay!, line)}:
                  </span>
                  <div className="mt-1">
                    <LineDisplay
                      line={line}
                      play={stageState.selectedPlay!}
                      isCurrentLine={false}
                      isPreview={true}
                    />
                  </div>
                </div>
              ))}
            </>
          ) : (
            <p className="text-[var(--cream)] opacity-40 text-sm text-center py-2">
              No more lines for your character
              {(currentParticipant.castMemberIds || []).length > 1 ? 's' : ''}
            </p>
          )}
        </div>
      )}

      {/* Director view */}
      {currentParticipant &&
        (currentParticipant.castMemberIds || []).length === 0 &&
        isDirector && (
          <div className="p-4 rounded-xl bg-[var(--gold)]/10 border border-[var(--gold)]/20 text-center">
            <span className="text-2xl">🎬</span>
            <p className="text-[var(--gold)] text-sm mt-2">
              Directing • Advance when the speaker is ready
            </p>
          </div>
        )}

      {/* Next button */}
      {canAdvance && (
        <button onClick={onNextLine} className="btn-primary w-full py-4 rounded-xl text-lg">
          <span>{isDirector ? 'Next Line →' : "I've Finished My Line"}</span>
        </button>
      )}

      {/* Waiting message */}
      {!canAdvance && (
        <div className="text-center py-4">
          <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-[var(--charcoal)]/50">
            <div className="w-2 h-2 rounded-full bg-[var(--gold)] animate-pulse" />
            <span className="text-[var(--cream)] opacity-60 text-sm">
              {hasDirector ? 'Waiting for the director...' : 'Waiting for current speaker...'}
            </span>
          </div>
        </div>
      )}
    </div>
  );
}
