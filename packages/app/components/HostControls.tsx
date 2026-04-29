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
  onReassignRole?: (targetParticipantId: string, castMemberId: string) => void;
  onUnassignRole?: (targetParticipantId: string, castMemberId: string) => void;
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
}: Props) {
  const me = state.participants.find((p) => p.id === participantId);
  const [open, setOpen] = useState(true);
  const [castOpen, setCastOpen] = useState(false);
  const [pickerForCastId, setPickerForCastId] = useState<string | null>(null);
  if (!me?.isHost) return null;

  const performing = state.phase === 'performing';
  const canEnd = state.phase === 'performing' || state.phase === 'discussion';
  const paused = state.pausedByHost;
  const inLobby = state.phase === 'lobby';
  const play = state.selectedPlay;
  // Eligible assignees: non-audience participants. Exclude host while in director mode.
  const candidates = state.participants.filter(
    (p) => !p.isAudience && !(p.isHost && state.directorMode)
  );

  return (
    <View className="rounded-md border border-stage-gold bg-stage-surface">
      <Pressable
        onPress={() => setOpen((v) => !v)}
        accessibilityRole="button"
        accessibilityLabel={open ? 'Hide host controls' : 'Show host controls'}
        className="flex-row items-center justify-between px-3 py-2"
      >
        <Text className="text-xs uppercase tracking-widest text-stage-gold">
          ★ Host controls
        </Text>
        <Text className="text-xs text-stage-muted">{open ? '▾' : '▸'}</Text>
      </Pressable>
      <View
        className={`flex-row flex-wrap gap-2 border-t border-stage-border px-3 py-3 ${open ? 'flex' : 'hidden'}`}
      >
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
          {play && onReassignRole ? (
            <ActionButton
              label={castOpen ? 'Hide cast' : 'Manage cast'}
              onPress={() => setCastOpen((v) => !v)}
            />
          ) : null}
        </View>

      {open && castOpen && play && onReassignRole ? (
        <View className="border-t border-stage-border px-3 py-2 gap-2">
          {play.cast.map((cm) => {
            const assignee = state.participants.find((p) => p.castMemberIds.includes(cm.id));
            const showPicker = pickerForCastId === cm.id;
            return (
              <View key={cm.id} className="rounded-md border border-stage-border p-2">
                <View className="flex-row items-center justify-between">
                  <Text className="text-stage-text text-sm">{cm.name}</Text>
                  <Text className="text-xs text-stage-muted">
                    {assignee ? assignee.name : 'unassigned'}
                  </Text>
                </View>
                <View className="mt-1 flex-row flex-wrap gap-1">
                  <TinyAction
                    label={showPicker ? 'Cancel' : 'Reassign'}
                    onPress={() => setPickerForCastId(showPicker ? null : cm.id)}
                  />
                  {assignee && onUnassignRole ? (
                    <TinyAction
                      label="Unassign"
                      onPress={() => onUnassignRole(assignee.id, cm.id)}
                    />
                  ) : null}
                </View>
                {showPicker ? (
                  <View className="mt-2 flex-row flex-wrap gap-1">
                    {candidates.length === 0 ? (
                      <Text className="text-xs text-stage-muted">No eligible participants.</Text>
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
                  </View>
                ) : null}
              </View>
            );
          })}
        </View>
      ) : null}
    </View>
  );
}

function TinyAction({ label, onPress }: { label: string; onPress: () => void }) {
  return (
    <Pressable onPress={onPress} className="rounded border border-stage-border px-2 py-0.5">
      <Text className="text-xs text-stage-muted">{label}</Text>
    </Pressable>
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
