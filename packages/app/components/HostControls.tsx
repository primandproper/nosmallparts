import { useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import type { StageState } from '@nsp/shared';
import { confirm } from '../lib/confirm';

interface Props {
  state: StageState;
  participantId: string | null;
  onPause: () => void;
  onResume: () => void;
  onBackLine: () => void;
  onSkipLine: () => void;
  onEndPlay: () => void;
  onToggleDirectorMode?: (enabled: boolean) => void;
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
}: Props) {
  const me = state.participants.find((p) => p.id === participantId);
  const [open, setOpen] = useState(false);
  if (!me?.isHost) return null;

  const performing = state.phase === 'performing';
  const canEnd = state.phase === 'performing' || state.phase === 'discussion';
  const paused = state.pausedByHost;
  const inLobby = state.phase === 'lobby';

  return (
    <View className="rounded-md border border-stage-gold bg-stage-surface">
      <Pressable
        onPress={() => setOpen((v) => !v)}
        className="flex-row items-center justify-between px-3 py-2"
      >
        <Text className="text-xs uppercase tracking-widest text-stage-gold">
          ★ Host controls
        </Text>
        <Text className="text-xs text-stage-muted">{open ? '▾' : '▸'}</Text>
      </Pressable>
      {open ? (
        <View className="flex-row flex-wrap gap-2 border-t border-stage-border px-3 py-3">
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
            onPress={async () => {
              // Only confirm mid-scene; in discussion the play already ended naturally.
              if (performing) {
                const ok = await confirm({
                  title: 'End play?',
                  message: 'Progress will be lost.',
                  confirmLabel: 'End play',
                  destructive: true,
                });
                if (!ok) return;
              }
              onEndPlay();
            }}
          />
          <ActionButton
            label={state.directorMode ? 'Exit director mode' : 'Director mode'}
            disabled={!inLobby || !onToggleDirectorMode}
            onPress={
              onToggleDirectorMode ? () => onToggleDirectorMode(!state.directorMode) : undefined
            }
          />
        </View>
      ) : null}
    </View>
  );
}

function ActionButton({
  label,
  onPress,
  disabled,
}: {
  label: string;
  onPress?: () => void;
  disabled?: boolean;
}) {
  return (
    <Pressable
      onPress={onPress}
      disabled={disabled || !onPress}
      className={`rounded-md border px-3 py-1 ${
        disabled ? 'border-stage-border opacity-40' : 'border-stage-gold'
      }`}
    >
      <Text className={disabled ? 'text-stage-muted text-xs' : 'text-stage-text text-xs'}>
        {label}
      </Text>
    </Pressable>
  );
}
