import { useState } from 'react';

import { StagePhase, type StageState } from '../generated/nosmallparts/stage/v1/stage';

interface Props {
  state: StageState;
  participantId: string | null;
  onPause: () => void;
  onResume: () => void;
  onBackLine: () => void;
  onSkipLine: () => void;
  onEndPlay: () => void;
  onToggleDirectorMode?: (enabled: boolean) => void;
  onReassignRole?: (targetParticipantId: string, castMemberId: string) => void;
  onUnassignRole?: (targetParticipantId: string, castMemberId: string) => void;
  onMute?: (targetParticipantId: string, muted: boolean) => void;
}

export function HostControls({
  state,
  participantId,
  onPause,
  onResume,
  onBackLine,
  onSkipLine,
  onEndPlay,
  onToggleDirectorMode,
  onReassignRole,
  onUnassignRole,
  onMute,
}: Props) {
  const me = state.participants.find((p) => p.id === participantId);
  const [open, setOpen] = useState(true);
  const [castOpen, setCastOpen] = useState(false);
  const [pickerForCastId, setPickerForCastId] = useState<string | null>(null);
  if (!me?.isHost) return null;

  const performing = state.phase === StagePhase.STAGE_PHASE_PERFORMING;
  const canEnd = performing || state.phase === StagePhase.STAGE_PHASE_DISCUSSION;
  const paused = state.pausedByHost;
  const inLobby = state.phase === StagePhase.STAGE_PHASE_LOBBY;
  const play = state.selectedPlay;
  // Eligible assignees: performers, and not the host while directing.
  const candidates = state.participants.filter((p) => !p.isAudience && !(p.isHost && state.directorMode));

  return (
    <div className="rounded-md border border-stage-gold bg-stage-surface">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-label={open ? 'Hide host controls' : 'Show host controls'}
        aria-expanded={open}
        className="flex w-full items-center justify-between px-3 py-2"
      >
        <span className="text-xs uppercase tracking-widest text-stage-gold">★ Host controls</span>
        <span className="text-xs text-stage-muted">{open ? '▾' : '▸'}</span>
      </button>
      <div className={`flex-wrap gap-2 border-t border-stage-border px-3 py-3 ${open ? 'flex' : 'hidden'}`}>
        <ActionButton
          label={paused ? 'Resume' : 'Pause'}
          disabled={!performing}
          onPress={paused ? onResume : onPause}
        />
        <ActionButton label="Back" disabled={!performing} onPress={onBackLine} />
        <ActionButton label="Skip" disabled={!performing} onPress={onSkipLine} />
        <ActionButton
          label="End play"
          disabled={!canEnd}
          onPress={() => {
            // Only confirm mid-scene; in the discussion the play already ended.
            if (performing && !window.confirm('End play?\n\nProgress will be lost.')) return;
            onEndPlay();
          }}
        />
        <ActionButton
          label={state.directorMode ? 'Exit director mode' : 'Director mode'}
          disabled={!inLobby || !onToggleDirectorMode}
          onPress={onToggleDirectorMode ? () => onToggleDirectorMode(!state.directorMode) : undefined}
        />
        {play && onReassignRole ? (
          <ActionButton label={castOpen ? 'Hide cast' : 'Manage cast'} onPress={() => setCastOpen((v) => !v)} />
        ) : null}
      </div>

      {open && castOpen && play && onReassignRole ? (
        <div className="flex flex-col gap-2 border-t border-stage-border px-3 py-2">
          {play.cast.map((cm) => {
            const assignee = state.participants.find((p) => p.castMemberIds.includes(cm.id));
            const showPicker = pickerForCastId === cm.id;
            return (
              <div key={cm.id} className="rounded-md border border-stage-border p-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-stage-text">{cm.name}</span>
                  <span className="text-xs text-stage-muted">{assignee ? assignee.name : 'unassigned'}</span>
                </div>
                <div className="mt-1 flex flex-wrap gap-1">
                  <TinyAction
                    label={showPicker ? 'Cancel' : 'Reassign'}
                    onPress={() => setPickerForCastId(showPicker ? null : cm.id)}
                  />
                  {assignee && onUnassignRole ? (
                    <TinyAction label="Unassign" onPress={() => onUnassignRole(assignee.id, cm.id)} />
                  ) : null}
                </div>
                {showPicker ? (
                  <div className="mt-2 flex flex-wrap gap-1">
                    {candidates.length === 0 ? (
                      <span className="text-xs text-stage-muted">No eligible participants.</span>
                    ) : (
                      candidates.map((p) => (
                        <TinyAction
                          key={p.id}
                          label={p.name}
                          onPress={() => {
                            onReassignRole(p.id, cm.id);
                            setPickerForCastId(null);
                          }}
                        />
                      ))
                    )}
                  </div>
                ) : null}
              </div>
            );
          })}
          {onMute ? (
            <div className="rounded-md border border-stage-border p-2">
              <p className="mb-1 text-xs uppercase tracking-wider text-stage-muted">Participants</p>
              {state.participants
                .filter((p) => p.id !== participantId)
                .map((p) => (
                  <div key={p.id} className="flex items-center justify-between py-0.5">
                    <span className="text-xs text-stage-text">
                      {p.name}
                      {p.forceMuted ? ' · 🔇 muted' : ''}
                    </span>
                    <TinyAction label={p.forceMuted ? 'Unmute' : 'Mute'} onPress={() => onMute(p.id, !p.forceMuted)} />
                  </div>
                ))}
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function TinyAction({ label, onPress }: { label: string; onPress: () => void }) {
  return (
    <button
      type="button"
      onClick={onPress}
      className="rounded border border-stage-border px-2 py-0.5 text-xs text-stage-muted"
    >
      {label}
    </button>
  );
}

function ActionButton({ label, onPress, disabled }: { label: string; onPress?: () => void; disabled?: boolean }) {
  const off = disabled || !onPress;
  return (
    <button
      type="button"
      onClick={onPress}
      disabled={off}
      className={`rounded-md border px-3 py-1 text-xs ${
        off ? 'border-stage-border text-stage-muted opacity-40' : 'border-stage-gold text-stage-text'
      }`}
    >
      {label}
    </button>
  );
}
