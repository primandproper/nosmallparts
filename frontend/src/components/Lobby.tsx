import { useEffect, useRef, useState } from 'react';

import type { Play } from '../generated/nosmallparts/scripts/v1/scripts';
import {
  type Participant,
  StagePhase,
  type StageState,
  type SystemMessage,
} from '../generated/nosmallparts/stage/v1/stage';
import { areAllCastMembersAssigned, areAllParticipantsReady } from '../lib/stage-logic';
import { HostControls } from './HostControls';

interface Props {
  state: StageState;
  availablePlays: Play[];
  participantId: string | null;
  systemMessages?: SystemMessage[];
  onSelectPlay: (playId: string) => void;
  onToggleRole: (castMemberId: string) => void;
  onToggleAudience: (v: boolean) => void;
  onToggleReady: (v: boolean) => void;
  onStart: () => void;
  // Host actions
  onHostPause: () => void;
  onHostResume: () => void;
  onHostBackLine: () => void;
  onHostSkipLine: () => void;
  onHostEndPlay: () => void;
  onHostTransferHost: (targetParticipantId: string) => void;
  onHostKick: (targetParticipantId: string) => void;
  onHostBan: (targetParticipantId: string) => void;
  onHostMute: (targetParticipantId: string, muted: boolean) => void;
  onHostToggleDirectorMode: (enabled: boolean) => void;
  onHostReassignRole: (targetParticipantId: string, castMemberId: string) => void;
  onHostUnassignRole: (targetParticipantId: string, castMemberId: string) => void;
}

export function Lobby({
  state,
  availablePlays,
  participantId,
  systemMessages,
  onSelectPlay,
  onToggleRole,
  onToggleAudience,
  onToggleReady,
  onStart,
  onHostPause,
  onHostResume,
  onHostBackLine,
  onHostSkipLine,
  onHostEndPlay,
  onHostTransferHost,
  onHostKick,
  onHostBan,
  onHostMute,
  onHostToggleDirectorMode,
  onHostReassignRole,
  onHostUnassignRole,
}: Props) {
  const me = state.participants.find((p) => p.id === participantId);
  const iAmHost = !!me?.isHost;
  const directorMode = state.directorMode;
  const meIsDirector = iAmHost && directorMode;
  const play = state.selectedPlay;
  const allAssigned = areAllCastMembersAssigned(state);
  const allReady = areAllParticipantsReady(state);
  const performers = state.participants.filter((p) => !p.isAudience);
  const unreadyPerformers = performers.filter((p) => !p.isReady);
  const firstPerformerId = performers[0]?.id;
  // Only signed-in users can host; a stage of guests waits for one (§1.4).
  const hasHost = state.participants.some((p) => p.isHost);
  const assignedIds = new Set(performers.flatMap((p) => p.castMemberIds));
  const unfilledRoles = play ? play.cast.filter((c) => !assignedIds.has(c.id)).length : 0;
  const [copied, setCopied] = useState(false);
  // Copying can fail (an insecure context, a refused permission), so the link
  // is then shown to copy by hand.
  const [copyFallbackUrl, setCopyFallbackUrl] = useState<string | null>(null);
  const copiedTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (copiedTimer.current) clearTimeout(copiedTimer.current);
    };
  }, []);

  const copyInviteLink = async () => {
    const url = `${window.location.origin}/stage/${state.code}`;
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setCopyFallbackUrl(null);
      if (copiedTimer.current) clearTimeout(copiedTimer.current);
      copiedTimer.current = setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopyFallbackUrl(url);
    }
  };

  // Start on its own once everyone is ready, unless in director mode, where
  // the director starts it. One client (the first performer, by the server's
  // stable order) sends it, so the server is not sent a burst of duplicates.
  const shouldAutoStart =
    state.phase === StagePhase.STAGE_PHASE_LOBBY &&
    !!play &&
    hasHost &&
    allAssigned &&
    allReady &&
    !directorMode &&
    !!participantId &&
    firstPerformerId === participantId;
  useEffect(() => {
    if (shouldAutoStart) onStart();
  }, [shouldAutoStart, onStart]);

  const status = getStatus({
    play,
    me,
    hasHost,
    allAssigned,
    allReady,
    directorMode,
    iAmHost,
    unfilledRoles,
    unreadyCount: unreadyPerformers.length,
    onStart,
    onToggleReady,
  });

  const directorName = state.participants.find((p) => p.isHost)?.name ?? 'The host';

  return (
    <div className="flex-1 overflow-y-auto bg-stage-bg p-4">
      <div className="mx-auto flex w-full max-w-[1400px] flex-col gap-4">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:gap-8">
          <div className="flex flex-1 flex-col gap-4">
            <HostControls
              state={state}
              participantId={participantId}
              onPause={onHostPause}
              onResume={onHostResume}
              onBackLine={onHostBackLine}
              onSkipLine={onHostSkipLine}
              onEndPlay={onHostEndPlay}
              onToggleDirectorMode={onHostToggleDirectorMode}
              onReassignRole={onHostReassignRole}
              onUnassignRole={onHostUnassignRole}
            />

            {!hasHost ? (
              <div className="rounded-md border border-amber-500 bg-stage-surface px-3 py-2">
                <p className="text-xs uppercase tracking-widest text-amber-400">Waiting for a signed-in user to host</p>
                <p className="mt-1 text-xs text-stage-muted">
                  Guests can join and claim roles, but the play can't start until someone signs in to host this stage.
                </p>
              </div>
            ) : null}

            {directorMode ? (
              <div className="rounded-md border border-stage-gold bg-stage-surface px-3 py-2">
                <p className="text-xs uppercase tracking-widest text-stage-gold">
                  Director mode {iAmHost ? '· You are the director' : `· ${directorName} is the director`}
                </p>
                <p className="mt-1 text-xs text-stage-muted">The host advances lines. Actors just read.</p>
              </div>
            ) : null}

            <div className="flex flex-col items-center gap-2">
              <p className="text-xs uppercase tracking-widest text-stage-muted">Stage code</p>
              <div className="rounded-md border border-stage-border bg-stage-surface px-5 py-2">
                <p className="font-mono text-4xl font-semibold tracking-[0.3em] text-stage-gold">{state.code}</p>
              </div>
              <button
                type="button"
                onClick={copyInviteLink}
                aria-label="Copy invite link"
                className="rounded-md border border-stage-border px-3 py-1.5 text-xs text-stage-text active:opacity-70"
              >
                {copied ? 'Invite link copied' : 'Copy invite link'}
              </button>
              {copyFallbackUrl ? (
                <div className="rounded-md border border-stage-border bg-stage-surface px-3 py-1.5">
                  <p className="text-[10px] text-stage-muted">Copy this link manually:</p>
                  <p className="select-all text-xs text-stage-text">{copyFallbackUrl}</p>
                </div>
              ) : null}
            </div>

            <section>
              <h2 className="mb-2 font-serif text-lg text-stage-gold">Play</h2>
              {play ? (
                <div className="rounded-md border border-stage-gold bg-stage-surface p-3">
                  <p className="font-semibold text-stage-text">{play.title}</p>
                  {play.author ? <p className="text-xs text-stage-muted">by {play.author}</p> : null}
                  <p className="mt-1 text-sm text-stage-muted">{play.description}</p>
                  <button
                    type="button"
                    onClick={() => onSelectPlay('')}
                    className="mt-2 text-xs text-stage-muted underline"
                  >
                    Change
                  </button>
                </div>
              ) : (
                <div className="flex flex-col gap-2">
                  {availablePlays.length === 0 ? (
                    <p className="text-stage-muted">No plays available.</p>
                  ) : (
                    availablePlays.map((p) => (
                      <button
                        type="button"
                        key={p.id}
                        onClick={() => onSelectPlay(p.id)}
                        className="rounded-md border border-stage-border bg-stage-surface p-3 text-left"
                      >
                        <span className="block font-semibold text-stage-text">{p.title}</span>
                        <span className="block text-xs text-stage-muted">
                          {p.cast.length} cast · {p.lines.length} lines
                        </span>
                      </button>
                    ))
                  )}
                </div>
              )}
            </section>

            {play && me && !meIsDirector ? (
              <section>
                <h2 className="mb-2 font-serif text-lg text-stage-gold">Your role</h2>
                <div className="flex flex-col gap-2 rounded-md border border-stage-border bg-stage-surface p-3">
                  <div className="flex flex-wrap gap-2">
                    <Toggle label="Audience" active={me.isAudience} onPress={() => onToggleAudience(!me.isAudience)} />
                  </div>
                  {!me.isAudience ? (
                    <div>
                      <p className="mb-1 mt-2 text-xs text-stage-muted">Characters</p>
                      <div className="flex flex-wrap gap-2">
                        {play.cast.map((c) => {
                          const assigned = me.castMemberIds.includes(c.id);
                          const takenByOther = state.participants.some(
                            (p) => p.id !== me.id && p.castMemberIds.includes(c.id),
                          );
                          return (
                            <button
                              type="button"
                              key={c.id}
                              onClick={() => onToggleRole(c.id)}
                              disabled={takenByOther && !assigned}
                              aria-pressed={assigned}
                              className={`rounded-md border px-3 py-1 ${
                                assigned
                                  ? 'border-stage-gold bg-stage-goldDim font-semibold text-stage-bg'
                                  : takenByOther
                                    ? 'border-stage-border text-stage-text opacity-40'
                                    : 'border-stage-border text-stage-text'
                              }`}
                            >
                              {c.name}
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  ) : null}
                </div>
              </section>
            ) : null}
          </div>

          <div className="flex flex-1 flex-col gap-4">
            <section>
              <h2 className="mb-2 font-serif text-lg text-stage-gold">Participants ({state.participants.length})</h2>
              {state.participants.map((p) => (
                <div key={p.id} className="py-1">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className="text-stage-text">
                        {p.isHost ? '★ ' : ''}
                        {p.name}
                        {p.id === participantId ? ' (you)' : ''}
                      </span>
                      {p.isHost && directorMode ? (
                        <span className="rounded border border-stage-gold px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-stage-gold">
                          Director
                        </span>
                      ) : null}
                    </div>
                    <span className="text-xs text-stage-muted">{describeParticipant(p, directorMode)}</span>
                  </div>
                  {iAmHost && p.id !== participantId ? (
                    <div className="mt-1 flex flex-wrap gap-1">
                      <RowAction
                        label={p.forceMuted ? 'Unmute' : 'Mute'}
                        onPress={() => onHostMute(p.id, !p.forceMuted)}
                      />
                      <RowAction label="Make host" onPress={() => onHostTransferHost(p.id)} />
                      <RowAction
                        label="Kick"
                        onPress={() => {
                          if (
                            window.confirm(
                              `Remove ${p.name}?\n\nThey will be disconnected. They can rejoin if they want.`,
                            )
                          ) {
                            onHostKick(p.id);
                          }
                        }}
                      />
                      <RowAction
                        label="Ban"
                        danger
                        onPress={() => {
                          if (window.confirm(`Ban ${p.name}?\n\nThey cannot rejoin this stage.`)) onHostBan(p.id);
                        }}
                      />
                    </div>
                  ) : null}
                  {p.id === participantId && p.forceMuted ? (
                    <p className="mt-1 text-xs text-red-400">You are muted by the host.</p>
                  ) : null}
                </div>
              ))}
            </section>

            {systemMessages && systemMessages.length > 0 ? (
              <div className="rounded-md border border-stage-border bg-stage-surface p-3">
                <p className="mb-1 text-xs uppercase tracking-widest text-stage-muted">Notices</p>
                {systemMessages.slice(-3).map((m) => (
                  <p key={m.id} className="text-xs text-stage-muted">
                    · {m.text}
                  </p>
                ))}
              </div>
            ) : null}
          </div>
        </div>

        <div className="lg:mx-auto lg:w-full lg:max-w-md">
          <button
            type="button"
            onClick={status.onPress ?? undefined}
            disabled={!status.onPress}
            className={`w-full rounded-md px-4 py-3 text-center font-semibold ${
              status.primary
                ? 'bg-stage-gold text-stage-bg'
                : status.onPress
                  ? 'bg-stage-goldDim text-stage-bg'
                  : 'bg-stage-border text-stage-muted'
            }`}
          >
            {status.label}
          </button>
        </div>
      </div>
    </div>
  );
}

function describeParticipant(p: Participant, directorMode: boolean): string {
  if (p.isAudience) return 'audience';

  return [
    p.isHost && (directorMode ? 'host · director' : 'host'),
    p.forceMuted && 'muted',
    p.castMemberIds.length > 0 && `${p.castMemberIds.length} role${p.castMemberIds.length === 1 ? '' : 's'}`,
    // The host directing is not a performer, so is neither ready nor waiting.
    !(p.isHost && directorMode) && (p.isReady ? 'ready' : 'waiting'),
  ]
    .filter(Boolean)
    .join(' · ');
}

interface StatusInput {
  play: Play | undefined;
  me: Participant | undefined;
  hasHost: boolean;
  allAssigned: boolean;
  allReady: boolean;
  directorMode: boolean;
  iAmHost: boolean;
  unfilledRoles: number;
  unreadyCount: number;
  onStart: () => void;
  onToggleReady: (v: boolean) => void;
}

interface StatusOutput {
  label: string;
  onPress: (() => void) | null;
  primary: boolean;
}

function getStatus(i: StatusInput): StatusOutput {
  const {
    play,
    me,
    hasHost,
    allAssigned,
    allReady,
    directorMode,
    iAmHost,
    unfilledRoles,
    unreadyCount,
    onStart,
    onToggleReady,
  } = i;

  if (!hasHost) return { label: 'Waiting for a signed-in user to host', onPress: null, primary: false };
  if (!play) return { label: 'Select a play to continue', onPress: null, primary: false };
  if (!allAssigned) {
    // Spell out what is missing (§3.4).
    const parts = [`${unfilledRoles} unfilled role${unfilledRoles === 1 ? '' : 's'}`];
    if (unreadyCount > 0) parts.push(`${unreadyCount} not ready`);
    return { label: `Waiting: ${parts.join(', ')}`, onPress: null, primary: false };
  }
  if (!me) return { label: 'Joining…', onPress: null, primary: false };

  // The host directing starts the play once everyone is ready.
  if (directorMode && iAmHost) {
    if (!allReady) return { label: `Waiting: ${unreadyCount} not ready`, onPress: null, primary: false };
    return { label: 'Begin performance', onPress: onStart, primary: true };
  }

  if (me.isAudience) {
    if (!allReady) return { label: 'Waiting for performers to ready up', onPress: null, primary: false };
    return { label: directorMode ? 'Waiting for director to begin' : 'Starting…', onPress: null, primary: false };
  }

  // A performer: in the cast, not the audience, not the host directing.
  if (!me.isReady) return { label: 'Ready up', onPress: () => onToggleReady(true), primary: true };
  if (!allReady) {
    const waiting = unreadyCount === 1 ? '1 other' : `${unreadyCount} others`;
    return {
      label: `Ready ✓ · Waiting on ${waiting} (tap to unready)`,
      onPress: () => onToggleReady(false),
      primary: false,
    };
  }
  return { label: directorMode ? 'Waiting for director to begin' : 'Starting…', onPress: null, primary: false };
}

function Toggle({ label, active, onPress }: { label: string; active: boolean; onPress: () => void }) {
  return (
    <button
      type="button"
      onClick={onPress}
      aria-pressed={active}
      className={`rounded-md border px-3 py-1 ${
        active
          ? 'border-stage-gold bg-stage-goldDim font-semibold text-stage-bg'
          : 'border-stage-border text-stage-text'
      }`}
    >
      {label}
    </button>
  );
}

function RowAction({ label, onPress, danger }: { label: string; onPress: () => void; danger?: boolean }) {
  return (
    <button
      type="button"
      onClick={onPress}
      className={`rounded border px-2 py-0.5 text-xs ${danger ? 'border-red-500 text-red-400' : 'border-stage-border text-stage-muted'}`}
    >
      {label}
    </button>
  );
}
