import { useState, type ReactNode } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import {
  getCurrentLines,
  getCurrentLinesForParticipant,
  getCurrentSpeakerCastMemberIds,
  getNextLinesForParticipant,
  isParticipantSpeaking,
  getCastMemberNameForLine,
  getLineIndices,
  getLinesAtIndex,
} from '@nsp/shared';
import type { PlayLine, StageState } from '@nsp/shared';
import { TeleprompterLine } from './TeleprompterLine';

const DEFAULT_WPM = 150;

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
  const inCast = !!me && !me.isAudience && (me.castMemberIds?.length ?? 0) > 0;
  const turnStatus: 'speaking' | 'on-deck' | 'waiting' | 'done' | 'idle' =
    !inCast || state.phase !== 'performing'
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
    <View className={`flex-1 bg-stage-bg ${frameClass}`}>
      <TurnStatusBanner status={turnStatus} turnsAway={next?.turnsAway ?? null} />
      {/* Top bar: turn counter, code, script toggle */}
      <View className="flex-row items-center justify-between border-b border-stage-border bg-stage-surface px-4 py-2">
        <Text className="text-stage-muted text-xs">
          Turn {state.currentLineIndex + 1} of {totalIndices}
          {state.pausedByHost ? '  ·  ⏸ Paused by host' : ''}
        </Text>
        <View className="flex-row items-center gap-3">
          <Text className="text-stage-muted text-xs">Code: {state.code}</Text>
          <Pressable
            onPress={() => setScriptOpen((v) => !v)}
            accessibilityRole="button"
            accessibilityLabel={scriptOpen ? 'Hide script' : 'Show script'}
            className={`rounded border px-2 py-1 md:hidden ${
              scriptOpen ? 'border-stage-gold bg-stage-goldDim' : 'border-stage-border'
            }`}
          >
            <Text
              className={`text-xs font-semibold ${
                scriptOpen ? 'text-stage-bg' : 'text-stage-text'
              }`}
            >
              ▤ Script
            </Text>
          </Pressable>
        </View>
      </View>

      {me?.forceMuted ? (
        <View className="border-b border-red-500 bg-stage-surface px-4 py-2">
          <Text className="text-xs text-red-400">
            🔇 You are muted by the host. You can still see and read the play.
          </Text>
        </View>
      ) : null}

      {/* Video + line queue (desktop) + script overlay (mobile toggle) */}
      <View className="flex-1 min-h-0 flex-row">
        <View className="flex-1 min-h-0">{videoSlot}</View>
        <LineQueue state={state} participantId={participantId} />
        {scriptOpen ? (
          <ScriptPanel
            state={state}
            participantId={participantId}
            onClose={() => setScriptOpen(false)}
          />
        ) : null}
      </View>

      {/* Bottom bar: director badge + Next line */}
      <View className="border-t border-stage-border bg-stage-surface px-3 py-3">
        {state.directorMode ? (
          <View className="mb-2 flex-row items-center justify-center gap-2">
            <View className="rounded border border-stage-gold px-1.5 py-0.5">
              <Text className="text-[10px] uppercase tracking-wider text-stage-gold font-semibold">
                Director mode
              </Text>
            </View>
            <Text className="text-xs text-stage-muted">
              {isDirector ? 'You advance the lines' : 'The director advances the lines'}
            </Text>
          </View>
        ) : null}
        <Pressable
          onPress={onNextLine}
          disabled={!canAdvance}
          className={`mx-auto w-full max-w-md rounded-md px-4 py-3 ${
            canAdvance ? 'bg-stage-gold' : 'bg-stage-border'
          }`}
        >
          <Text
            className={`text-center font-semibold ${
              canAdvance ? 'text-stage-bg' : 'text-stage-muted'
            }`}
          >
            {advanceLabel}
          </Text>
        </Pressable>
      </View>
    </View>
  );
}

function TurnStatusBanner({
  status,
  turnsAway,
}: {
  status: 'speaking' | 'on-deck' | 'waiting' | 'done' | 'idle';
  turnsAway: number | null;
}) {
  if (status === 'idle') return null;
  if (status === 'speaking') {
    return (
      <View className="bg-red-500 px-4 py-2 items-center">
        <Text className="text-white font-bold text-base uppercase tracking-widest">
          ● Live — your line, speak now
        </Text>
      </View>
    );
  }
  if (status === 'on-deck') {
    return (
      <View className="bg-amber-400 px-4 py-2 items-center animate-pulse">
        <Text className="text-stage-bg font-bold text-base uppercase tracking-widest">
          ▸ On deck — your line is next
        </Text>
      </View>
    );
  }
  if (status === 'done') {
    return (
      <View className="bg-stage-surface border-b border-stage-border px-4 py-1.5 items-center">
        <Text className="text-stage-muted text-xs uppercase tracking-widest">
          Done — no more lines for you
        </Text>
      </View>
    );
  }
  // waiting
  return (
    <View className="bg-stage-surface border-b border-stage-border px-4 py-1.5 items-center">
      <Text className="text-stage-muted text-xs uppercase tracking-widest">
        Your line in {turnsAway} turn{turnsAway === 1 ? '' : 's'}
      </Text>
    </View>
  );
}

function ScriptPanel({
  state,
  participantId,
  onClose,
}: {
  state: StageState;
  participantId: string | null;
  onClose: () => void;
}) {
  const play = state.selectedPlay!;
  const me = state.participants.find((p) => p.id === participantId);
  const myLines = me ? getCurrentLinesForParticipant(state, me.id) : [];
  const currentLines = getCurrentLines(state);
  const speakerIds = new Set(getCurrentSpeakerCastMemberIds(state));
  const speakers = play.cast.filter((c) => speakerIds.has(c.id));
  const isSpeaking = me ? isParticipantSpeaking(state, me.id) : false;
  const next = me ? getNextLinesForParticipant(state, me.id) : null;

  // Mobile: full-bleed overlay. md+: docked right rail.
  return (
    <View className="absolute inset-0 z-10 border-l border-stage-border bg-stage-bg md:relative md:inset-auto md:z-auto md:w-[360px] lg:w-[400px]">
      <View className="flex-row items-center justify-between border-b border-stage-border bg-stage-surface px-3 py-2">
        <Text className="text-xs uppercase tracking-widest text-stage-gold">Script</Text>
        <Pressable
          onPress={onClose}
          accessibilityRole="button"
          accessibilityLabel="Hide script"
          className="rounded border border-stage-border px-2 py-1"
        >
          <Text className="text-xs text-stage-muted">Close</Text>
        </Pressable>
      </View>
      <ScrollView contentContainerStyle={{ padding: 16, gap: 14 }} className="flex-1">
        {isSpeaking && myLines.length > 0 ? (
          <View>
            <Text className="text-stage-muted text-xs uppercase tracking-wider mb-2">
              Your line
            </Text>
            {myLines.map((line, i) => (
              <View key={i} className="mb-3">
                <TeleprompterLine
                  text={line.text}
                  direction={line.direction}
                  wpm={line.autoScrollWpm ?? DEFAULT_WPM}
                  autoScroll={(line.scrollMode ?? 'auto') === 'auto'}
                />
              </View>
            ))}
          </View>
        ) : (
          <View>
            <Text className="text-stage-muted text-xs uppercase tracking-wider mb-2">
              Now speaking
            </Text>
            <Text className="mb-2 text-stage-gold text-sm xl:text-base">
              {speakers.map((s) => s.name).join(', ') || '—'}
            </Text>
            {currentLines.map((line, i) => (
              <View
                key={i}
                className="mb-3 rounded-md border border-stage-border bg-stage-surface p-3"
              >
                <Text className="text-xs text-stage-muted mb-1">
                  {getCastMemberNameForLine(play, line)}
                </Text>
                {line.direction ? (
                  <Text className="italic text-stage-muted text-xs">({line.direction})</Text>
                ) : null}
                <Text className="text-sm xl:text-base text-stage-text leading-snug">
                  {line.text}
                </Text>
              </View>
            ))}
          </View>
        )}

        {next ? (
          <View className="rounded-md border border-stage-border p-3">
            <Text className="text-xs text-stage-muted uppercase tracking-wider">
              Your next line ({next.turnsAway} turn{next.turnsAway === 1 ? '' : 's'} away)
            </Text>
            {next.lines.map((line, i) => (
              <Text key={i} className="mt-1 text-stage-muted">
                {line.text}
              </Text>
            ))}
          </View>
        ) : null}
      </ScrollView>
    </View>
  );
}

function LineQueue({
  state,
  participantId,
}: {
  state: StageState;
  participantId: string | null;
}) {
  const play = state.selectedPlay;
  if (!play) return null;
  const me = state.participants.find((p) => p.id === participantId);
  const myCastIds = new Set(me?.castMemberIds || []);
  const indices = getLineIndices(play);
  const cur = state.currentLineIndex;
  const upcomingIndices = indices.slice(cur, cur + 50);
  const totalLines = indices.length;

  // Find the nearest upcoming "your line" (turnsAway >= 0). 0 means it's now.
  const myUpcomingTurnsAway = (() => {
    for (let i = 0; i < upcomingIndices.length; i++) {
      const idx = upcomingIndices[i];
      if (idx === undefined) continue;
      const lines = getLinesAtIndex(play, idx);
      if (lines.some((l) => (l.castMemberIds || []).some((id) => myCastIds.has(id)))) return i;
    }
    return null;
  })();

  const inCast = !!me && !me.isAudience && myCastIds.size > 0;

  return (
    <View className="hidden md:flex w-[360px] lg:w-[420px] border-l border-stage-border bg-stage-bg">
      <View className="flex-row items-center justify-between border-b border-stage-border bg-stage-surface px-3 py-2">
        <Text className="text-xs uppercase tracking-widest text-stage-gold">Line queue</Text>
        <Text className="text-xs text-stage-muted">
          {Math.min(cur + 1, totalLines)} / {totalLines}
        </Text>
      </View>

      {inCast ? (
        <View className="border-b border-stage-border bg-stage-surface px-3 py-2">
          {myUpcomingTurnsAway === null ? (
            <Text className="text-xs text-stage-muted uppercase tracking-wider">
              Done — no more lines for you
            </Text>
          ) : myUpcomingTurnsAway === 0 ? (
            <Text className="text-xs text-red-400 uppercase tracking-widest font-bold">
              ● Your line is up — speak now
            </Text>
          ) : (
            <Text className="text-xs text-amber-400 uppercase tracking-wider font-semibold">
              Your line in {myUpcomingTurnsAway} line{myUpcomingTurnsAway === 1 ? '' : 's'}
            </Text>
          )}
        </View>
      ) : null}

      <ScrollView contentContainerStyle={{ paddingVertical: 8 }} className="flex-1">
        {upcomingIndices.length === 0 ? (
          <View className="items-center justify-center px-4 py-8">
            <Text className="text-stage-muted text-xs uppercase tracking-widest">
              End of script
            </Text>
          </View>
        ) : null}
        {upcomingIndices.map((idx, i) => {
          const lines = getLinesAtIndex(play, idx);
          const isCurrent = i === 0;
          const isMine = lines.some((l) =>
            (l.castMemberIds || []).some((id) => myCastIds.has(id))
          );
          return (
            <QueueEntry
              key={`${idx}-${i}`}
              lines={lines}
              isCurrent={isCurrent}
              isMine={isMine}
              distance={i}
              play={play}
            />
          );
        })}
      </ScrollView>
    </View>
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
  play: NonNullable<StageState['selectedPlay']>;
}) {
  const speakerNames = Array.from(
    new Set(lines.map((l) => getCastMemberNameForLine(play, l))),
  ).join(', ');

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
    ? 'text-stage-text text-sm xl:text-base leading-snug'
    : isMine
      ? 'text-stage-text text-xs xl:text-sm leading-snug'
      : 'text-stage-muted text-xs xl:text-sm leading-snug';

  return (
    <View className={containerClass}>
      <View className="flex-row items-baseline justify-between mb-0.5">
        <Text className={speakerClass}>
          {isMine ? `You · ${speakerNames}` : speakerNames}
        </Text>
        <Text className="text-[10px] text-stage-muted">
          {isCurrent ? 'NOW' : `+${distance}`}
        </Text>
      </View>
      <View
        className="overflow-hidden"
        style={{ maxHeight: isCurrent ? 160 : 72 }}
      >
        {lines.map((line, i) => (
          <View key={i} className={i > 0 ? 'mt-1' : undefined}>
            {line.direction ? (
              <Text className="italic text-stage-muted text-xs">({line.direction})</Text>
            ) : null}
            <Text className={textClass} numberOfLines={isCurrent ? 6 : 3}>
              {line.text}
            </Text>
          </View>
        ))}
      </View>
    </View>
  );
}
