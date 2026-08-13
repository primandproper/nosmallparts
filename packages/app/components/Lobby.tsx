import { useEffect, useRef, useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import * as Clipboard from 'expo-clipboard';
import * as Linking from 'expo-linking';
import {
  areAllCastMembersAssigned,
  areAllParticipantsReady,
  canStartPlay,
} from '@nsp/shared';
import type { Play, StageState, SystemMessage } from '@nsp/shared';
import { HostControls } from './HostControls';
import { confirm } from '../lib/confirm';

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
  // Only signed-in users can host; a guest-only stage waits for one (§1.4).
  const hasHost = state.participants.some((p) => p.isHost);
  const assignedIds = new Set(performers.flatMap((p) => p.castMemberIds));
  const unfilledRoles = play ? play.cast.filter((c) => !assignedIds.has(c.id)).length : 0;
  const [copied, setCopied] = useState(false);
  // Clipboard can fail (e.g. insecure context) — fall back to showing the URL.
  const [copyFallbackUrl, setCopyFallbackUrl] = useState<string | null>(null);
  const copiedTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (copiedTimer.current) clearTimeout(copiedTimer.current);
    };
  }, []);

  const copyInviteLink = async () => {
    const url = Linking.createURL(`/stage/${state.code}`);
    try {
      await Clipboard.setStringAsync(url);
      setCopied(true);
      setCopyFallbackUrl(null);
      if (copiedTimer.current) clearTimeout(copiedTimer.current);
      copiedTimer.current = setTimeout(() => setCopied(false), 1500);
    } catch {
      setCopyFallbackUrl(url);
    }
  };

  // Auto-start when everyone is ready and director mode is OFF. With director
  // mode on, the host (director) starts the play explicitly. One client (the
  // first performer, by stable server-broadcast order) issues the auto-start
  // so we don't stampede the server with duplicate startPlay messages.
  const shouldAutoStart =
    state.phase === 'lobby' &&
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

  return (
    <ScrollView className="flex-1 bg-stage-bg" contentContainerStyle={{ padding: 16 }}>
      <View className="mx-auto w-full max-w-[1400px] gap-4">
        <View className="gap-4 lg:flex-row lg:items-start lg:gap-8">
          <View className="flex-1 gap-4">
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
        <View className="rounded-md border border-amber-500 bg-stage-surface px-3 py-2">
          <Text className="text-xs uppercase tracking-widest text-amber-400">
            Waiting for a signed-in user to host
          </Text>
          <Text className="mt-1 text-xs text-stage-muted">
            Guests can join and claim roles, but the play can't start until someone signs in to
            host this stage.
          </Text>
        </View>
      ) : null}

      {directorMode ? (
        <View className="rounded-md border border-stage-gold bg-stage-surface px-3 py-2">
          <Text className="text-xs uppercase tracking-widest text-stage-gold">
            Director mode {iAmHost ? '· You are the director' : `· ${state.participants.find((p) => p.isHost)?.name ?? 'The host'} is the director`}
          </Text>
          <Text className="text-xs text-stage-muted mt-1">
            The host advances lines. Actors just read.
          </Text>
        </View>
      ) : null}

      <View className="items-center gap-2">
        <Text className="text-xs uppercase tracking-widest text-stage-muted">Stage code</Text>
        <View className="rounded-md border border-stage-border bg-stage-surface px-5 py-2">
          <Text className="text-4xl font-mono font-semibold tracking-[0.3em] text-stage-gold">
            {state.code}
          </Text>
        </View>
        <Pressable
          onPress={copyInviteLink}
          accessibilityRole="button"
          accessibilityLabel="Copy invite link"
          className="rounded-md border border-stage-border px-3 py-1.5 active:opacity-70"
        >
          <Text className="text-xs text-stage-text">
            {copied ? 'Invite link copied' : 'Copy invite link'}
          </Text>
        </Pressable>
        {copyFallbackUrl ? (
          <View className="rounded-md border border-stage-border bg-stage-surface px-3 py-1.5">
            <Text className="text-[10px] text-stage-muted">Copy this link manually:</Text>
            <Text selectable className="text-xs text-stage-text">
              {copyFallbackUrl}
            </Text>
          </View>
        ) : null}
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

      {play && me && !meIsDirector ? (
        <View>
          <Text className="mb-2 text-lg font-serif text-stage-gold">Your role</Text>
          <View className="gap-2 rounded-md border border-stage-border bg-stage-surface p-3">
            <View className="flex-row flex-wrap gap-2">
              <Toggle
                label="Audience"
                active={me.isAudience}
                onPress={() => onToggleAudience(!me.isAudience)}
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
          </View>

          <View className="flex-1 gap-4">
      <View>
        <Text className="mb-2 text-lg font-serif text-stage-gold">
          Participants ({state.participants.length})
        </Text>
        {state.participants.map((p) => (
          <View key={p.id} className="py-1">
            <View className="flex-row items-center justify-between">
              <View className="flex-row items-center gap-2">
                <Text className="text-stage-text">
                  {p.isHost ? '★ ' : ''}
                  {p.name}
                  {p.id === participantId ? ' (you)' : ''}
                </Text>
                {p.isHost && directorMode ? (
                  <View className="rounded border border-stage-gold px-1.5 py-0.5">
                    <Text className="text-[10px] uppercase tracking-wider text-stage-gold font-semibold">
                      Director
                    </Text>
                  </View>
                ) : null}
              </View>
              <Text className="text-xs text-stage-muted">
                {p.isAudience
                  ? 'audience'
                  : [
                      p.isHost && (directorMode ? 'host · director' : 'host'),
                      p.forceMuted && 'muted',
                      p.castMemberIds.length > 0 &&
                        `${p.castMemberIds.length} role${p.castMemberIds.length === 1 ? '' : 's'}`,
                      // The host-as-director isn't a performer and shouldn't be tagged ready/waiting.
                      !(p.isHost && directorMode) && (p.isReady ? 'ready' : 'waiting'),
                    ]
                      .filter(Boolean)
                      .join(' · ')}
              </Text>
            </View>
            {iAmHost && p.id !== participantId ? (
              <View className="mt-1 flex-row flex-wrap gap-1">
                <RowAction label={p.forceMuted ? 'Unmute' : 'Mute'} onPress={() => onHostMute(p.id, !p.forceMuted)} />
                <RowAction label="Make host" onPress={() => onHostTransferHost(p.id)} />
                <RowAction
                  label="Kick"
                  onPress={async () => {
                    const ok = await confirm({
                      title: `Remove ${p.name}?`,
                      message: 'They will be disconnected. They can rejoin if they want.',
                      confirmLabel: 'Remove',
                      destructive: true,
                    });
                    if (ok) onHostKick(p.id);
                  }}
                />
                <RowAction
                  label="Ban"
                  danger
                  onPress={async () => {
                    const ok = await confirm({
                      title: `Ban ${p.name}?`,
                      message: 'They cannot rejoin this stage.',
                      confirmLabel: 'Ban',
                      destructive: true,
                    });
                    if (ok) onHostBan(p.id);
                  }}
                />
              </View>
            ) : null}
            {p.id === participantId && p.forceMuted ? (
              <Text className="mt-1 text-xs text-red-400">You are muted by the host.</Text>
            ) : null}
          </View>
        ))}
      </View>

      {systemMessages && systemMessages.length > 0 ? (
        <View className="rounded-md border border-stage-border bg-stage-surface p-3">
          <Text className="mb-1 text-xs uppercase tracking-widest text-stage-muted">Notices</Text>
          {systemMessages.slice(-3).map((m) => (
            <Text key={m.id} className="text-xs text-stage-muted">
              · {m.text}
            </Text>
          ))}
        </View>
      ) : null}
          </View>
        </View>

        <View className="lg:mx-auto lg:w-full lg:max-w-md">
      <Pressable
        onPress={status.onPress ?? undefined}
        disabled={!status.onPress}
        className={`rounded-md px-4 py-3 ${status.primary ? 'bg-stage-gold' : status.onPress ? 'bg-stage-goldDim' : 'bg-stage-border'}`}
      >
        <Text
          className={`text-center font-semibold ${status.primary ? 'text-stage-bg' : status.onPress ? 'text-stage-bg' : 'text-stage-muted'}`}
        >
          {status.label}
        </Text>
      </Pressable>
        </View>
      </View>
    </ScrollView>
  );
}

interface StatusInput {
  play: Play | null | undefined;
  me: StageState['participants'][number] | undefined;
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
  const { play, me, hasHost, allAssigned, allReady, directorMode, iAmHost, unfilledRoles, unreadyCount, onStart, onToggleReady } = i;

  if (!hasHost) {
    return { label: 'Waiting for a signed-in user to host', onPress: null, primary: false };
  }
  if (!play) return { label: 'Select a play to continue', onPress: null, primary: false };
  if (!allAssigned) {
    // Spell out what's missing (§3.4).
    const parts = [`${unfilledRoles} unfilled role${unfilledRoles === 1 ? '' : 's'}`];
    if (unreadyCount > 0) parts.push(`${unreadyCount} not ready`);
    return { label: `Waiting: ${parts.join(', ')}`, onPress: null, primary: false };
  }
  if (!me) return { label: 'Joining…', onPress: null, primary: false };

  // Host-as-director: they start the play once everyone's ready.
  if (directorMode && iAmHost) {
    if (!allReady) {
      return {
        label: `Waiting: ${unreadyCount} not ready`,
        onPress: null,
        primary: false,
      };
    }
    return { label: 'Begin performance', onPress: onStart, primary: true };
  }

  if (me.isAudience) {
    if (!allReady) return { label: 'Waiting for performers to ready up', onPress: null, primary: false };
    return {
      label: directorMode ? 'Waiting for director to begin' : 'Starting…',
      onPress: null,
      primary: false,
    };
  }

  // Performer (cast member, not audience, not host-as-director)
  if (!me.isReady) return { label: 'Ready up', onPress: () => onToggleReady(true), primary: true };
  if (!allReady) {
    const waiting =
      unreadyCount === 1 ? '1 other' : `${unreadyCount} others`;
    return {
      label: `Ready ✓ · Waiting on ${waiting} (tap to unready)`,
      onPress: () => onToggleReady(false),
      primary: false,
    };
  }
  return {
    label: directorMode ? 'Waiting for director to begin' : 'Starting…',
    onPress: null,
    primary: false,
  };
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

function RowAction({
  label,
  onPress,
  danger,
}: {
  label: string;
  onPress: () => void;
  danger?: boolean;
}) {
  return (
    <Pressable
      onPress={onPress}
      className={`rounded border px-2 py-0.5 ${danger ? 'border-red-500' : 'border-stage-border'}`}
    >
      <Text className={`text-xs ${danger ? 'text-red-400' : 'text-stage-muted'}`}>{label}</Text>
    </Pressable>
  );
}
