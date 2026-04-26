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
} from '@nsp/shared';
import type { StageState } from '@nsp/shared';

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

  return (
    <View className="flex-1 bg-stage-bg">
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
            className={`rounded border px-2 py-1 ${
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

      {/* Video + side script panel */}
      <View className="flex-1 min-h-0 flex-row">
        <View className="flex-1 min-h-0">{videoSlot}</View>
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
                {line.direction ? (
                  <Text className="italic text-stage-muted">({line.direction})</Text>
                ) : null}
                <Text className="text-2xl font-serif text-stage-gold leading-snug">
                  {line.text}
                </Text>
              </View>
            ))}
          </View>
        ) : (
          <View>
            <Text className="text-stage-muted text-xs uppercase tracking-wider mb-2">
              Now speaking
            </Text>
            <Text className="mb-2 text-stage-gold text-base">
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
                  <Text className="italic text-stage-muted">({line.direction})</Text>
                ) : null}
                <Text className="text-base text-stage-text">{line.text}</Text>
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
