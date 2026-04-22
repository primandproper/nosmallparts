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
}

export function Stage({ state, participantId, onNextLine }: Props) {
  const play = state.selectedPlay;
  if (!play) return null;

  const me = state.participants.find((p) => p.id === participantId);
  const myLines = me ? getCurrentLinesForParticipant(state, me.id) : [];
  const currentLines = getCurrentLines(state);
  const speakerIds = new Set(getCurrentSpeakerCastMemberIds(state));
  const speakers = play.cast.filter((c) => speakerIds.has(c.id));
  const isSpeaking = me ? isParticipantSpeaking(state, me.id) : false;
  const isDirector = state.directorMode && !!me?.isHost;
  const next = me ? getNextLinesForParticipant(state, me.id) : null;

  const totalIndices = getLineIndices(play).length;

  return (
    <View className="flex-1 bg-stage-bg">
      <View className="flex-row items-center justify-between border-b border-stage-border bg-stage-surface px-4 py-3">
        <Text className="text-stage-muted text-xs">
          Turn {state.currentLineIndex + 1} of {totalIndices}
          {state.pausedByHost ? '  ·  ⏸ Paused by host' : ''}
        </Text>
        <Text className="text-stage-muted text-xs">Code: {state.code}</Text>
      </View>

      {me?.forceMuted ? (
        <View className="border-b border-red-500 bg-stage-surface px-4 py-2">
          <Text className="text-xs text-red-400">
            🔇 You are muted by the host. You can still see and read the play.
          </Text>
        </View>
      ) : null}

      <ScrollView contentContainerStyle={{ padding: 20, gap: 16 }} className="flex-1">
        {isSpeaking && myLines.length > 0 ? (
          <View>
            <Text className="text-stage-muted text-sm uppercase tracking-wider mb-2">
              Your line
            </Text>
            {myLines.map((line, i) => (
              <View key={i} className="mb-3">
                {line.direction ? (
                  <Text className="italic text-stage-muted">({line.direction})</Text>
                ) : null}
                <Text className="text-3xl font-serif text-stage-gold leading-relaxed">
                  {line.text}
                </Text>
              </View>
            ))}
          </View>
        ) : (
          <View>
            <Text className="text-stage-muted text-sm uppercase tracking-wider mb-2">
              Now speaking
            </Text>
            <Text className="mb-2 text-stage-gold text-lg">
              {speakers.map((s) => s.name).join(', ') || '—'}
            </Text>
            {currentLines.map((line, i) => (
              <View key={i} className="mb-3 rounded-md border border-stage-border bg-stage-surface p-3">
                <Text className="text-xs text-stage-muted mb-1">
                  {getCastMemberNameForLine(play, line)}
                </Text>
                {line.direction ? (
                  <Text className="italic text-stage-muted">({line.direction})</Text>
                ) : null}
                <Text className="text-lg text-stage-text">{line.text}</Text>
              </View>
            ))}
          </View>
        )}

        {next ? (
          <View className="mt-4 rounded-md border border-stage-border p-3">
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

      <View className="border-t border-stage-border bg-stage-surface p-3">
        <Pressable
          onPress={onNextLine}
          disabled={state.pausedByHost || (state.directorMode ? !isDirector : !isSpeaking)}
          className={`rounded-md px-4 py-3 ${
            !state.pausedByHost && (state.directorMode ? isDirector : isSpeaking)
              ? 'bg-stage-gold'
              : 'bg-stage-border'
          }`}
        >
          <Text
            className={`text-center font-semibold ${
              !state.pausedByHost && (state.directorMode ? isDirector : isSpeaking)
                ? 'text-stage-bg'
                : 'text-stage-muted'
            }`}
          >
            {state.directorMode && isDirector ? 'Advance' : 'Next line'}
          </Text>
        </Pressable>
      </View>
    </View>
  );
}
