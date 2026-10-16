import { type ReactNode, useState } from 'react';

import { type Play, type PlayLine, ScrollMode } from '../generated/nosmallparts/scripts/v1/scripts';
import { StagePhase, type StageState } from '../generated/nosmallparts/stage/v1/stage';
import {
  getCastMemberNameForLine,
  getCurrentLines,
  getCurrentLinesForParticipant,
  getCurrentSpeakerCastMemberIds,
  getLineIndices,
  getLinesAtIndex,
  getNextLinesForParticipant,
  isParticipantSpeaking,
} from '../lib/stage-logic';
import { TeleprompterLine } from './TeleprompterLine';

const DEFAULT_WPM = 150;

type TurnStatus = 'speaking' | 'on-deck' | 'waiting' | 'done' | 'idle';

interface Props {
  state: StageState;
  participantId: string | null;
  onNextLine: () => void;
  videoSlot: ReactNode;
}

export function Stage({ state, participantId, onNextLine, videoSlot }: Props) {
  const play = state.selectedPlay;
  const [scriptOpen, setScriptOpen] = useState(false);

  if (!play) return null;

  const me = state.participants.find((p) => p.id === participantId);
  const isSpeaking = me ? isParticipantSpeaking(state, me.id) : false;
  const isDirector = state.directorMode && !!me?.isHost;
  const totalIndices = getLineIndices(play).length;
  const canAdvance = !state.pausedByHost && (state.directorMode ? isDirector : isSpeaking);
  const advanceLabel = state.directorMode && isDirector ? 'Advance' : 'Next line';

  const next = me ? getNextLinesForParticipant(state, me.id) : null;
  const inCast = !!me && !me.isAudience && me.castMemberIds.length > 0;
  const turnStatus: TurnStatus =
    !inCast || state.phase !== StagePhase.STAGE_PHASE_PERFORMING
      ? 'idle'
      : isSpeaking
        ? 'speaking'
        : next && next.turnsAway === 1
          ? 'on-deck'
          : next && next.turnsAway <= 3
            ? 'waiting'
            : next === null
              ? 'done'
              : 'idle';

  const frameClass =
    turnStatus === 'speaking'
      ? 'border-4 border-red-500'
      : turnStatus === 'on-deck'
        ? 'border-4 border-amber-400'
        : 'border-0 border-transparent';

  return (
    <div className={`flex min-h-0 flex-1 flex-col bg-stage-bg ${frameClass}`}>
      <TurnStatusBanner status={turnStatus} turnsAway={next?.turnsAway ?? null} />
      {/* Top bar: turn counter, code, script toggle */}
      <div className="flex items-center justify-between border-b border-stage-border bg-stage-surface px-4 py-2">
        <span className="text-xs text-stage-muted">
          Turn {state.currentLineIndex + 1} of {totalIndices}
          {state.pausedByHost ? '  ·  ⏸ Paused by host' : ''}
        </span>
        <div className="flex items-center gap-3">
          <span className="text-xs text-stage-muted">Code: {state.code}</span>
          <button
            type="button"
            onClick={() => setScriptOpen((v) => !v)}
            aria-label={scriptOpen ? 'Hide script' : 'Show script'}
            className={`rounded border px-2 py-1 text-xs font-semibold md:hidden ${
              scriptOpen ? 'border-stage-gold bg-stage-goldDim text-stage-bg' : 'border-stage-border text-stage-text'
            }`}
          >
            ▤ Script
          </button>
        </div>
      </div>

      {me?.forceMuted ? (
        <div className="border-b border-red-500 bg-stage-surface px-4 py-2">
          <p className="text-xs text-red-400">🔇 You are muted by the host. You can still see and read the play.</p>
        </div>
      ) : null}

      {/* Video, the line queue (desktop) and the script overlay (mobile toggle) */}
      <div className="relative flex min-h-0 flex-1">
        <div className="flex min-h-0 flex-1 flex-col">{videoSlot}</div>
        <LineQueue state={state} play={play} participantId={participantId} />
        {scriptOpen ? (
          <ScriptPanel state={state} play={play} participantId={participantId} onClose={() => setScriptOpen(false)} />
        ) : null}
      </div>

      {/* Bottom bar: director badge and Next line */}
      <div className="border-t border-stage-border bg-stage-surface px-3 py-3">
        {state.directorMode ? (
          <div className="mb-2 flex items-center justify-center gap-2">
            <span className="rounded border border-stage-gold px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-stage-gold">
              Director mode
            </span>
            <span className="text-xs text-stage-muted">
              {isDirector ? 'You advance the lines' : 'The director advances the lines'}
            </span>
          </div>
        ) : null}
        <button
          type="button"
          onClick={onNextLine}
          disabled={!canAdvance}
          className={`mx-auto block w-full max-w-md rounded-md px-4 py-3 text-center font-semibold ${
            canAdvance ? 'bg-stage-gold text-stage-bg' : 'bg-stage-border text-stage-muted'
          }`}
        >
          {advanceLabel}
        </button>
      </div>
    </div>
  );
}

function TurnStatusBanner({ status, turnsAway }: { status: TurnStatus; turnsAway: number | null }) {
  if (status === 'idle') return null;
  if (status === 'speaking') {
    return (
      <div className="flex justify-center bg-red-500 px-4 py-2">
        <p className="text-base font-bold uppercase tracking-widest text-white">● Live: your line, speak now</p>
      </div>
    );
  }
  if (status === 'on-deck') {
    return (
      <div className="flex animate-pulse justify-center bg-amber-400 px-4 py-2">
        <p className="text-base font-bold uppercase tracking-widest text-stage-bg">▸ On deck: your line is next</p>
      </div>
    );
  }
  if (status === 'done') {
    return (
      <div className="flex justify-center border-b border-stage-border bg-stage-surface px-4 py-1.5">
        <p className="text-xs uppercase tracking-widest text-stage-muted">Done: no more lines for you</p>
      </div>
    );
  }
  return (
    <div className="flex justify-center border-b border-stage-border bg-stage-surface px-4 py-1.5">
      <p className="text-xs uppercase tracking-widest text-stage-muted">
        Your line in {turnsAway} turn{turnsAway === 1 ? '' : 's'}
      </p>
    </div>
  );
}

function ScriptPanel({
  state,
  play,
  participantId,
  onClose,
}: {
  state: StageState;
  play: Play;
  participantId: string | null;
  onClose: () => void;
}) {
  const me = state.participants.find((p) => p.id === participantId);
  const myLines = me ? getCurrentLinesForParticipant(state, me.id) : [];
  const currentLines = getCurrentLines(state);
  const speakerIds = new Set(getCurrentSpeakerCastMemberIds(state));
  const speakers = play.cast.filter((c) => speakerIds.has(c.id));
  const isSpeaking = me ? isParticipantSpeaking(state, me.id) : false;
  const next = me ? getNextLinesForParticipant(state, me.id) : null;

  // Mobile: a full-bleed overlay. md and up: a docked right rail.
  return (
    <aside className="absolute inset-0 z-10 flex flex-col border-l border-stage-border bg-stage-bg md:relative md:inset-auto md:z-auto md:w-[360px] lg:w-[400px]">
      <div className="flex items-center justify-between border-b border-stage-border bg-stage-surface px-3 py-2">
        <span className="text-xs uppercase tracking-widest text-stage-gold">Script</span>
        <button
          type="button"
          onClick={onClose}
          aria-label="Hide script"
          className="rounded border border-stage-border px-2 py-1 text-xs text-stage-muted"
        >
          Close
        </button>
      </div>
      <div className="flex flex-1 flex-col gap-3.5 overflow-y-auto p-4">
        {isSpeaking && myLines.length > 0 ? (
          <div>
            <p className="mb-2 text-xs uppercase tracking-wider text-stage-muted">Your line</p>
            {myLines.map((line, i) => (
              <div key={i} className="mb-3">
                <TeleprompterLine
                  text={line.text}
                  direction={line.direction}
                  wpm={line.autoScrollWpm ?? DEFAULT_WPM}
                  autoScroll={line.scrollMode !== ScrollMode.SCROLL_MODE_MANUAL}
                />
              </div>
            ))}
          </div>
        ) : (
          <div>
            <p className="mb-2 text-xs uppercase tracking-wider text-stage-muted">Now speaking</p>
            <p className="mb-2 text-sm text-stage-gold xl:text-base">{speakers.map((s) => s.name).join(', ') || '—'}</p>
            {currentLines.map((line, i) => (
              <div key={i} className="mb-3 rounded-md border border-stage-border bg-stage-surface p-3">
                <p className="mb-1 text-xs text-stage-muted">{getCastMemberNameForLine(play, line)}</p>
                {line.direction ? <p className="text-xs italic text-stage-muted">({line.direction})</p> : null}
                <p className="text-sm leading-snug text-stage-text xl:text-base">{line.text}</p>
              </div>
            ))}
          </div>
        )}

        {next ? (
          <div className="rounded-md border border-stage-border p-3">
            <p className="text-xs uppercase tracking-wider text-stage-muted">
              Your next line ({next.turnsAway} turn{next.turnsAway === 1 ? '' : 's'} away)
            </p>
            {next.lines.map((line, i) => (
              <p key={i} className="mt-1 text-stage-muted">
                {line.text}
              </p>
            ))}
          </div>
        ) : null}
      </div>
    </aside>
  );
}

function LineQueue({ state, play, participantId }: { state: StageState; play: Play; participantId: string | null }) {
  const me = state.participants.find((p) => p.id === participantId);
  const myCastIds = new Set(me?.castMemberIds ?? []);
  const indices = getLineIndices(play);
  const cur = state.currentLineIndex;
  const upcomingIndices = indices.slice(cur, cur + 50);
  const totalLines = indices.length;

  // The nearest upcoming line of mine, in lines away; 0 is now.
  const myUpcomingTurnsAway = (() => {
    for (let i = 0; i < upcomingIndices.length; i++) {
      const idx = upcomingIndices[i];
      if (idx === undefined) continue;
      if (getLinesAtIndex(play, idx).some((l) => l.castMemberIds.some((id) => myCastIds.has(id)))) return i;
    }
    return null;
  })();

  const inCast = !!me && !me.isAudience && myCastIds.size > 0;

  return (
    <aside className="hidden w-[360px] flex-col border-l border-stage-border bg-stage-bg md:flex lg:w-[420px]">
      <div className="flex items-center justify-between border-b border-stage-border bg-stage-surface px-3 py-2">
        <span className="text-xs uppercase tracking-widest text-stage-gold">Line queue</span>
        <span className="text-xs text-stage-muted">
          {Math.min(cur + 1, totalLines)} / {totalLines}
        </span>
      </div>

      {inCast ? (
        <div className="border-b border-stage-border bg-stage-surface px-3 py-2">
          {myUpcomingTurnsAway === null ? (
            <p className="text-xs uppercase tracking-wider text-stage-muted">Done: no more lines for you</p>
          ) : myUpcomingTurnsAway === 0 ? (
            <p className="text-xs font-bold uppercase tracking-widest text-red-400">● Your line is up, speak now</p>
          ) : (
            <p className="text-xs font-semibold uppercase tracking-wider text-amber-400">
              Your line in {myUpcomingTurnsAway} line{myUpcomingTurnsAway === 1 ? '' : 's'}
            </p>
          )}
        </div>
      ) : null}

      <div className="flex-1 overflow-y-auto py-2">
        {upcomingIndices.length === 0 ? (
          <div className="flex justify-center px-4 py-8">
            <p className="text-xs uppercase tracking-widest text-stage-muted">End of script</p>
          </div>
        ) : null}
        {upcomingIndices.map((idx, i) => {
          const lines = getLinesAtIndex(play, idx);
          return (
            <QueueEntry
              key={`${idx}-${i}`}
              lines={lines}
              isCurrent={i === 0}
              isMine={lines.some((l) => l.castMemberIds.some((id) => myCastIds.has(id)))}
              distance={i}
              play={play}
            />
          );
        })}
      </div>
    </aside>
  );
}

function QueueEntry({
  lines,
  isCurrent,
  isMine,
  distance,
  play,
}: {
  lines: PlayLine[];
  isCurrent: boolean;
  isMine: boolean;
  distance: number;
  play: Play;
}) {
  const speakerNames = Array.from(new Set(lines.map((l) => getCastMemberNameForLine(play, l)))).join(', ');

  const containerClass = isCurrent
    ? isMine
      ? 'border-l-4 border-red-500 bg-red-500/10 px-3 py-2'
      : 'border-l-4 border-stage-gold bg-stage-goldDim/20 px-3 py-2'
    : isMine
      ? 'border-l-4 border-amber-400 bg-amber-400/5 px-3 py-2'
      : 'border-l-4 border-transparent px-3 py-2';

  const speakerClass = isMine
    ? 'text-amber-300 text-xs uppercase tracking-widest font-bold'
    : 'text-stage-muted text-xs uppercase tracking-wider';

  const textClass = isCurrent
    ? 'text-stage-text text-sm xl:text-base leading-snug line-clamp-6'
    : isMine
      ? 'text-stage-text text-xs xl:text-sm leading-snug line-clamp-3'
      : 'text-stage-muted text-xs xl:text-sm leading-snug line-clamp-3';

  return (
    <div className={containerClass}>
      <div className="mb-0.5 flex items-baseline justify-between">
        <span className={speakerClass}>{isMine ? `You · ${speakerNames}` : speakerNames}</span>
        <span className="text-[10px] text-stage-muted">{isCurrent ? 'NOW' : `+${distance}`}</span>
      </div>
      <div className="overflow-hidden" style={{ maxHeight: isCurrent ? 160 : 72 }}>
        {lines.map((line, i) => (
          <div key={i} className={i > 0 ? 'mt-1' : undefined}>
            {line.direction ? <p className="text-xs italic text-stage-muted">({line.direction})</p> : null}
            <p className={textClass}>{line.text}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
