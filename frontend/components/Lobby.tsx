'use client';

import { useEffect, useRef, useState } from 'react';
import {
  StageState,
  Play,
  areAllCastMembersAssigned,
  areAllParticipantsReady,
  canStartPlay,
  getCastMemberById,
  getTotalLineIndices,
} from '@/lib/types';

interface LobbyProps {
  stageState: StageState;
  participantId: string | null;
  availablePlays: Play[];
  onToggleCastMember: (castMemberId: string) => void;
  onSetDirector: (isDirector: boolean) => void;
  onSetAudience: (isAudience: boolean) => void;
  onSetReady: (isReady: boolean) => void;
  onSelectPlay: (playId: string) => void;
  onStartPlay: () => void;
}

function PlaySelector({
  plays,
  selectedPlay,
  onSelect,
}: {
  plays: Play[];
  selectedPlay: Play | null;
  onSelect: (playId: string) => void;
}) {
  const [searchQuery, setSearchQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Filter plays based on search
  const filteredPlays = (plays || []).filter((play) => {
    const query = searchQuery.toLowerCase().trim();
    if (!query) return true;
    return (
      play.title.toLowerCase().includes(query) ||
      play.description.toLowerCase().includes(query) ||
      (play.author && play.author.toLowerCase().includes(query)) ||
      (play.cast || []).some((member) => member.name.toLowerCase().includes(query))
    );
  });

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={containerRef} className="relative">
      {/* Search Input */}
      <div
        className="input-elegant w-full px-4 py-3 rounded-xl text-[var(--cream)] cursor-pointer flex items-center justify-between"
        onClick={() => setIsOpen(!isOpen)}
      >
        {selectedPlay ? (
          <div className="flex-1 min-w-0">
            <div className="font-medium truncate">{selectedPlay.title}</div>
            <div className="text-xs text-[var(--cream)] opacity-40 truncate">
              {(selectedPlay.cast || []).length} roles • {getTotalLineIndices(selectedPlay)} turns
            </div>
          </div>
        ) : (
          <span className="text-[var(--cream)] opacity-40">Search for a play...</span>
        )}
        <svg
          className={`w-5 h-5 text-[var(--gold)] transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </div>

      {/* Dropdown */}
      {isOpen && (
        <div className="absolute z-50 w-full mt-2 rounded-xl bg-[var(--black-light)] border border-[var(--gold)]/20 shadow-xl overflow-hidden">
          {/* Search input */}
          <div className="p-3 border-b border-[var(--cream)]/10">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search by title, author, or character..."
              className="w-full px-3 py-2 rounded-lg bg-[var(--black)] border border-[var(--cream)]/10 text-[var(--cream)] text-sm focus:outline-none focus:border-[var(--gold)]/50"
              autoFocus
            />
          </div>

          {/* Play list */}
          <div className="max-h-64 overflow-y-auto">
            {filteredPlays.length === 0 ? (
              <div className="p-4 text-center text-[var(--cream)] opacity-40 text-sm">
                No plays found matching &quot;{searchQuery}&quot;
              </div>
            ) : (
              filteredPlays.map((play) => (
                <div
                  key={play.id}
                  onClick={() => {
                    onSelect(play.id);
                    setIsOpen(false);
                    setSearchQuery('');
                  }}
                  className={`p-3 cursor-pointer transition-colors hover:bg-[var(--gold)]/10 ${
                    selectedPlay?.id === play.id ? 'bg-[var(--gold)]/20' : ''
                  }`}
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-[var(--cream)]">{play.title}</div>
                      {play.author && (
                        <div className="text-xs text-[var(--gold)] opacity-80">
                          by {play.author}
                        </div>
                      )}
                      <div className="text-xs text-[var(--cream)] opacity-50 mt-1 line-clamp-2">
                        {play.description}
                      </div>
                    </div>
                    <div className="text-right shrink-0">
                      <div className="text-xs text-[var(--gold)]">{(play.cast || []).length} roles</div>
                      <div className="text-xs text-[var(--cream)] opacity-40">
                        {getTotalLineIndices(play)} turns
                      </div>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Footer */}
          <div className="p-2 border-t border-[var(--cream)]/10 text-center">
            <span className="text-xs text-[var(--cream)] opacity-30">
              {(plays || []).length} plays available
            </span>
          </div>
        </div>
      )}
    </div>
  );
}

export function Lobby({
  stageState,
  participantId,
  availablePlays,
  onToggleCastMember,
  onSetDirector,
  onSetAudience,
  onSetReady,
  onSelectPlay,
  onStartPlay,
}: LobbyProps) {
  const currentParticipant = (stageState.participants || []).find((p) => p.id === participantId);
  const allCastMembersAssigned = stageState.selectedPlay
    ? areAllCastMembersAssigned(stageState)
    : false;
  const allReady = areAllParticipantsReady(stageState);
  const canStart = canStartPlay(stageState);

  // Check if there's a director
  const hasDirector = (stageState.participants || []).some((p) => p.isDirector);
  const isDirector = currentParticipant?.isDirector || false;
  const isAudience = currentParticipant?.isAudience || false;

  // Track if we've already triggered auto-start to prevent multiple calls
  const hasAutoStarted = useRef(false);

  // Auto-start when no director and all conditions are met
  useEffect(() => {
    if (!hasDirector && canStart && !hasAutoStarted.current) {
      hasAutoStarted.current = true;
      // Small delay to ensure all participants see the "starting" state
      const timer = setTimeout(() => {
        onStartPlay();
      }, 1000);
      return () => clearTimeout(timer);
    }
    // Reset the flag if conditions change
    if (!canStart) {
      hasAutoStarted.current = false;
    }
  }, [hasDirector, canStart, onStartPlay]);

  const getCastMemberAssignment = (castMemberId: string) => {
    return (stageState.participants || []).find((p) => (p.castMemberIds || []).includes(castMemberId));
  };

  return (
    <div className="card-elegant rounded-2xl p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center gap-3">
        <span className="text-2xl">📜</span>
        <h2 className="font-[Cormorant_Garamond] text-2xl font-semibold text-gold-gradient">
          Stage Setup
        </h2>
      </div>

      {/* Play Selection */}
      <div>
        <label className="block text-[var(--gold)] text-xs font-medium mb-2 tracking-wide uppercase">
          Select a Play
        </label>
        <PlaySelector
          plays={availablePlays}
          selectedPlay={stageState.selectedPlay}
          onSelect={onSelectPlay}
        />
      </div>

      {/* Selected Play Info */}
      {stageState.selectedPlay && (
        <div className="p-3 rounded-xl bg-[var(--charcoal)]/50 border border-[var(--cream)]/5">
          <p className="text-[var(--cream)] opacity-60 text-sm">
            {stageState.selectedPlay.description}
          </p>
          {stageState.selectedPlay.author && (
            <p className="text-[var(--gold)] opacity-60 text-xs mt-1">
              — {stageState.selectedPlay.author}
            </p>
          )}
        </div>
      )}

      {/* Cast Assignment - only show for non-audience */}
      {stageState.selectedPlay && !isAudience && (
        <div>
          <label className="block text-[var(--gold)] text-xs font-medium mb-3 tracking-wide uppercase">
            Cast Roles
          </label>
          <div className="space-y-2">
            {(stageState.selectedPlay.cast || []).map((castMember) => {
              const assignedTo = getCastMemberAssignment(castMember.id);
              const isAssignedToMe = (currentParticipant?.castMemberIds || []).includes(castMember.id);
              const lineCount = (stageState.selectedPlay!.lines || []).filter(
                (l) => l.castMemberId === castMember.id
              ).length;

              return (
                <div
                  key={castMember.id}
                  className={`flex items-center justify-between p-3 rounded-xl border transition-all ${
                    isAssignedToMe
                      ? 'border-[var(--gold)] bg-[var(--gold)]/10'
                      : assignedTo
                        ? 'border-[var(--cream)]/10 bg-[var(--cream)]/5'
                        : 'border-[var(--cream)]/5 hover:border-[var(--cream)]/20'
                  }`}
                >
                  <div>
                    <span className="text-[var(--cream)] font-medium">{castMember.name}</span>
                    <span className="text-[var(--cream)] opacity-40 text-sm ml-2">
                      {lineCount} line{lineCount !== 1 ? 's' : ''}
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    {assignedTo && !isAssignedToMe && (
                      <span className="text-[var(--gold)] text-sm opacity-80">
                        {assignedTo.name}
                      </span>
                    )}
                    <button
                      onClick={() => onToggleCastMember(castMember.id)}
                      className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-all ${
                        isAssignedToMe
                          ? 'bg-[var(--black)] text-[var(--cream)] hover:bg-[var(--black-light)]'
                          : 'bg-[var(--gold)]/20 text-[var(--gold)] hover:bg-[var(--gold)]/30'
                      }`}
                    >
                      {isAssignedToMe ? 'Leave' : assignedTo ? 'Take' : 'Claim'}
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Director Toggle - only show after play is selected and not audience */}
      {currentParticipant && stageState.selectedPlay && !isAudience && (
        <div
          className={`p-3 rounded-xl border transition-all ${
            currentParticipant.isDirector
              ? 'border-[var(--gold)]/50 bg-[var(--gold)]/10'
              : 'border-[var(--cream)]/10 bg-[var(--cream)]/5'
          }`}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span>🎬</span>
              <span className="text-[var(--cream)]">Be Director</span>
            </div>
            <button
              onClick={() => onSetDirector(!currentParticipant.isDirector)}
              className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${
                currentParticipant.isDirector
                  ? 'bg-[var(--gold)] text-[var(--charcoal)]'
                  : 'bg-[var(--cream)]/10 text-[var(--cream)] hover:bg-[var(--cream)]/20'
              }`}
            >
              {currentParticipant.isDirector ? '✓ Director' : 'Volunteer'}
            </button>
          </div>
          <p className="text-[var(--cream)] opacity-40 text-xs mt-2">
            {currentParticipant.isDirector
              ? 'You control when the performance starts and progresses. Claim a role to become an actor instead.'
              : currentParticipant.castMemberIds.length > 0
                ? 'Becoming director will release your assigned roles.'
                : "Directors control flow but don't perform. Choose roles above to act instead."}
          </p>
        </div>
      )}

      {/* Audience Mode Toggle - only show after play is selected */}
      {currentParticipant && stageState.selectedPlay && (
        <div
          className={`p-3 rounded-xl border transition-all ${
            isAudience
              ? 'border-purple-500/50 bg-purple-500/10'
              : 'border-[var(--cream)]/10 bg-[var(--cream)]/5'
          }`}
        >
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span>👁️</span>
              <span className="text-[var(--cream)]">Join as Audience</span>
            </div>
            <button
              onClick={() => onSetAudience(!isAudience)}
              className={`px-4 py-1.5 rounded-lg text-sm font-medium transition-all ${
                isAudience
                  ? 'bg-purple-500 text-white'
                  : 'bg-[var(--cream)]/10 text-[var(--cream)] hover:bg-[var(--cream)]/20'
              }`}
            >
              {isAudience ? '✓ Watching' : 'Watch Only'}
            </button>
          </div>
          <p className="text-[var(--cream)] opacity-40 text-xs mt-2">
            {isAudience
              ? "You'll watch the performance without a webcam or role. Click again to participate."
              : "Audience members can watch without performing. Your camera won't be turned on."}
          </p>
        </div>
      )}

      {/* Ready Toggle - only show for non-audience */}
      {currentParticipant && stageState.selectedPlay && !isAudience && (
        <div
          className={`p-4 rounded-xl border-2 transition-all ${
            currentParticipant.isReady
              ? 'border-green-500/50 bg-green-500/10'
              : 'border-[var(--gold)]/30 bg-[var(--gold)]/5'
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="text-[var(--cream)] font-medium">Ready to perform?</span>
            <button
              onClick={() => onSetReady(!currentParticipant.isReady)}
              className={`px-5 py-2 rounded-lg text-sm font-semibold transition-all ${
                currentParticipant.isReady
                  ? 'bg-green-600 text-white hover:bg-green-700'
                  : 'btn-primary'
              }`}
            >
              <span>{currentParticipant.isReady ? '✓ Ready!' : 'Click When Ready'}</span>
            </button>
          </div>
        </div>
      )}

      {/* Participants */}
      <div>
        <label className="block text-[var(--gold)] text-xs font-medium mb-3 tracking-wide uppercase">
          Present ({(stageState.participants || []).length})
        </label>
        <div className="space-y-2">
          {(stageState.participants || []).map((participant) => (
            <div
              key={participant.id}
              className={`flex items-center justify-between text-sm p-3 rounded-xl ${
                participant.isAudience
                  ? 'bg-purple-500/10 border border-purple-500/20'
                  : 'bg-[var(--charcoal)]/50'
              }`}
            >
              <div className="flex items-center gap-2">
                <div
                  className={`w-8 h-8 rounded-full flex items-center justify-center text-xs font-medium ${
                    participant.isAudience
                      ? 'bg-purple-500/30 text-purple-300'
                      : 'bg-[var(--gold)]/20 text-[var(--gold)]'
                  }`}
                >
                  {participant.isAudience ? '👁️' : participant.name.charAt(0).toUpperCase()}
                </div>
                <span className="text-[var(--cream)]">
                  {participant.name}
                  {participant.id === participantId && (
                    <span className="text-[var(--gold)] opacity-60 ml-1">(you)</span>
                  )}
                </span>
                {participant.isAudience && (
                  <span className="text-purple-400 text-xs px-2 py-0.5 rounded-full bg-purple-500/20">
                    audience
                  </span>
                )}
                {participant.isDirector && <span className="text-[var(--gold)] text-xs">🎬</span>}
              </div>
              <div className="flex items-center gap-2">
                {!participant.isAudience &&
                  (participant.castMemberIds || []).length > 0 &&
                  stageState.selectedPlay && (
                    <span className="text-[var(--cream)] opacity-60 text-xs">
                      {(participant.castMemberIds || [])
                        .map((id) => getCastMemberById(stageState.selectedPlay!, id)?.name)
                        .filter(Boolean)
                        .join(', ')}
                    </span>
                  )}
                {stageState.selectedPlay && !participant.isAudience && (
                  <span
                    className={`text-xs px-2 py-0.5 rounded-full ${
                      participant.isReady
                        ? 'bg-green-600/20 text-green-400'
                        : 'bg-[var(--cream)]/10 text-[var(--cream)]/40'
                    }`}
                  >
                    {participant.isReady ? '✓' : '○'}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Status / Start Button */}
      {stageState.selectedPlay && (
        <div className="space-y-3 pt-2">
          {!allCastMembersAssigned && (
            <p className="text-[var(--cream)] opacity-40 text-sm text-center">
              All roles must be assigned before starting
            </p>
          )}
          {allCastMembersAssigned && !allReady && (
            <p className="text-[var(--cream)] opacity-40 text-sm text-center">
              Waiting for all performers to be ready...
            </p>
          )}

          {/* Auto-start message when no director */}
          {!hasDirector && allCastMembersAssigned && allReady && (
            <div className="text-center py-4">
              <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-green-500/20 border border-green-500/30">
                <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                <span className="text-green-400 text-sm font-medium">Starting performance...</span>
              </div>
            </div>
          )}

          {/* Only show button if there's a director */}
          {hasDirector && (
            <>
              {isDirector ? (
                <button
                  onClick={onStartPlay}
                  disabled={!canStart}
                  className="btn-primary w-full py-4 rounded-xl text-lg"
                >
                  <span>🎭 Begin Performance</span>
                </button>
              ) : (
                <div className="text-center py-3">
                  <p className="text-[var(--cream)] opacity-40 text-sm">
                    Waiting for the director to start...
                  </p>
                </div>
              )}
            </>
          )}

          {/* Message when no director and not ready */}
          {!hasDirector && !(allCastMembersAssigned && allReady) && (
            <p className="text-[var(--cream)] opacity-40 text-sm text-center">
              Performance will auto-start when everyone is ready
            </p>
          )}
        </div>
      )}
    </div>
  );
}
