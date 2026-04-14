import { Pressable, ScrollView, Text, View } from 'react-native';
import {
  areAllCastMembersAssigned,
  areAllParticipantsReady,
  canStartPlay,
} from '@nsp/shared';
import type { Play, StageState } from '@nsp/shared';

interface Props {
  state: StageState;
  availablePlays: Play[];
  participantId: string | null;
  onSelectPlay: (playId: string) => void;
  onToggleRole: (castMemberId: string) => void;
  onToggleDirector: (v: boolean) => void;
  onToggleAudience: (v: boolean) => void;
  onToggleReady: (v: boolean) => void;
  onStart: () => void;
}

export function Lobby({
  state,
  availablePlays,
  participantId,
  onSelectPlay,
  onToggleRole,
  onToggleDirector,
  onToggleAudience,
  onToggleReady,
  onStart,
}: Props) {
  const me = state.participants.find((p) => p.id === participantId);
  const play = state.selectedPlay;
  const canStart = canStartPlay(state);
  const hasDirector = state.participants.some((p) => p.isDirector);

  return (
    <ScrollView className="flex-1 bg-stage-bg" contentContainerStyle={{ padding: 16, gap: 16 }}>
      <View className="items-center gap-1">
        <Text className="text-xs text-stage-muted">Stage code</Text>
        <Text className="text-3xl font-serif tracking-widest text-stage-gold">{state.code}</Text>
      </View>

      <View>
        <Text className="mb-2 text-lg font-serif text-stage-gold">Play</Text>
        {play ? (
          <View className="rounded-md border border-stage-gold bg-stage-surface p-3">
            <Text className="font-semibold text-stage-text">{play.title}</Text>
            {play.author ? (
              <Text className="text-xs text-stage-muted">by {play.author}</Text>
            ) : null}
            <Text className="mt-1 text-sm text-stage-muted">{play.description}</Text>
            <Pressable onPress={() => onSelectPlay('')} className="mt-2">
              <Text className="text-xs text-stage-muted underline">Change</Text>
            </Pressable>
          </View>
        ) : (
          <View className="gap-2">
            {availablePlays.length === 0 ? (
              <Text className="text-stage-muted">No plays available.</Text>
            ) : (
              availablePlays.map((p) => (
                <Pressable
                  key={p.id}
                  onPress={() => onSelectPlay(p.id)}
                  className="rounded-md border border-stage-border bg-stage-surface p-3"
                >
                  <Text className="font-semibold text-stage-text">{p.title}</Text>
                  <Text className="text-xs text-stage-muted">
                    {p.cast.length} cast · {p.lines.length} lines
                  </Text>
                </Pressable>
              ))
            )}
          </View>
        )}
      </View>

      {play && me ? (
        <View>
          <Text className="mb-2 text-lg font-serif text-stage-gold">Your role</Text>
          <View className="gap-2 rounded-md border border-stage-border bg-stage-surface p-3">
            <View className="flex-row flex-wrap gap-2">
              <Toggle
                label="Director"
                active={me.isDirector}
                disabled={me.isAudience}
                onPress={() => onToggleDirector(!me.isDirector)}
              />
              <Toggle
                label="Audience"
                active={me.isAudience}
                onPress={() => onToggleAudience(!me.isAudience)}
              />
              <Toggle
                label={me.isReady ? 'Ready ✓' : 'Not ready'}
                active={me.isReady}
                disabled={me.isAudience}
                onPress={() => onToggleReady(!me.isReady)}
              />
            </View>
            {!me.isAudience ? (
              <View>
                <Text className="mt-2 mb-1 text-xs text-stage-muted">Characters</Text>
                <View className="flex-row flex-wrap gap-2">
                  {play.cast.map((c) => {
                    const assigned = me.castMemberIds.includes(c.id);
                    const takenByOther = state.participants.some(
                      (p) => p.id !== me.id && p.castMemberIds.includes(c.id)
                    );
                    return (
                      <Pressable
                        key={c.id}
                        onPress={() => onToggleRole(c.id)}
                        disabled={takenByOther && !assigned}
                        className={`rounded-md border px-3 py-1 ${
                          assigned
                            ? 'border-stage-gold bg-stage-goldDim'
                            : takenByOther
                              ? 'border-stage-border opacity-40'
                              : 'border-stage-border'
                        }`}
                      >
                        <Text
                          className={assigned ? 'text-stage-bg font-semibold' : 'text-stage-text'}
                        >
                          {c.name}
                        </Text>
                      </Pressable>
                    );
                  })}
                </View>
              </View>
            ) : null}
          </View>
        </View>
      ) : null}

      <View>
        <Text className="mb-2 text-lg font-serif text-stage-gold">
          Participants ({state.participants.length})
        </Text>
        {state.participants.map((p) => (
          <View key={p.id} className="flex-row items-center justify-between py-1">
            <Text className="text-stage-text">
              {p.name}
              {p.id === participantId ? ' (you)' : ''}
            </Text>
            <Text className="text-xs text-stage-muted">
              {p.isAudience
                ? 'audience'
                : [
                    p.isDirector && 'director',
                    p.castMemberIds.length > 0 &&
                      `${p.castMemberIds.length} role${p.castMemberIds.length === 1 ? '' : 's'}`,
                    p.isReady ? 'ready' : 'waiting',
                  ]
                    .filter(Boolean)
                    .join(' · ')}
            </Text>
          </View>
        ))}
      </View>

      <Pressable
        onPress={onStart}
        disabled={!canStart}
        className={`rounded-md px-4 py-3 ${canStart ? 'bg-stage-gold' : 'bg-stage-border'}`}
      >
        <Text
          className={`text-center font-semibold ${canStart ? 'text-stage-bg' : 'text-stage-muted'}`}
        >
          {!play
            ? 'Select a play to continue'
            : !areAllCastMembersAssigned(state)
              ? 'Assign all characters to start'
              : !areAllParticipantsReady(state)
                ? 'Waiting for performers to be ready'
                : hasDirector
                  ? 'Begin performance'
                  : 'Start'}
        </Text>
      </Pressable>
    </ScrollView>
  );
}

function Toggle({
  label,
  active,
  disabled,
  onPress,
}: {
  label: string;
  active: boolean;
  disabled?: boolean;
  onPress: () => void;
}) {
  return (
    <Pressable
      disabled={disabled}
      onPress={onPress}
      className={`rounded-md border px-3 py-1 ${
        active ? 'border-stage-gold bg-stage-goldDim' : 'border-stage-border'
      } ${disabled ? 'opacity-40' : ''}`}
    >
      <Text className={active ? 'text-stage-bg font-semibold' : 'text-stage-text'}>{label}</Text>
    </Pressable>
  );
}
