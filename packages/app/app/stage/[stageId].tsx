import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';
import { usePlayState } from '../../hooks/usePlayState';
import { Lobby } from '../../components/Lobby';
import { Countdown } from '../../components/Countdown';
import { Stage } from '../../components/Stage';
import { Discussion } from '../../components/Discussion';
import { HostControls } from '../../components/HostControls';
import { RoleOfferPrompt } from '../../components/RoleOfferPrompt';
import { VideoRoom } from '../../components/VideoRoom';
import { PLAYER_NAME_KEY } from '../../lib/player-name';

export default function StageScreen() {
  const { stageId } = useLocalSearchParams<{ stageId: string }>();
  const router = useRouter();
  const [playerName, setPlayerName] = useState<string | null>(null);

  useEffect(() => {
    AsyncStorage.getItem(PLAYER_NAME_KEY).then((v) => {
      if (v) setPlayerName(v);
      else router.replace('/');
    });
  }, [router]);

  const {
    stageState,
    participantId,
    availablePlays,
    systemMessages,
    roleOffers,
    isConnected,
    error,
    wasReplaced,
    reconnect,
    toggleCastMember,
    setAudience,
    setReady,
    selectPlay,
    startPlay,
    nextLine,
    resetPlay,
    hostPause,
    hostResume,
    hostBackLine,
    hostSkipLine,
    hostEndPlay,
    hostTransferHost,
    hostKick,
    hostBan,
    hostMute,
    hostToggleDirectorMode,
    hostReassignRole,
    hostUnassignRole,
    respondRoleOffer,
  } = usePlayState({
    stageIdOrCode: String(stageId),
    participantName: playerName ?? '',
    enabled: !!playerName,
  });

  if (!playerName) return null;

  if (!stageState) {
    return (
      <View className="flex-1 items-center justify-center bg-stage-bg px-6">
        <Stack.Screen options={{ title: 'Stage' }} />
        {wasReplaced ? null : <ActivityIndicator color="#d4af37" />}
        <Text className="mt-3 text-center text-stage-muted">
          {wasReplaced
            ? 'This stage was opened in another window.'
            : isConnected
              ? 'Joining stage…'
              : error
                ? error
                : 'Connecting…'}
        </Text>
        {wasReplaced ? (
          <Pressable
            onPress={reconnect}
            className="mt-4 rounded-md border border-stage-gold px-4 py-2 active:opacity-70"
          >
            <Text className="font-semibold text-stage-gold">Take over here</Text>
          </Pressable>
        ) : null}
      </View>
    );
  }

  const screenTitle =
    stageState.phase === 'lobby'
      ? `Lobby · ${stageState.code}`
      : stageState.phase === 'performing'
        ? stageState.selectedPlay?.title || 'Performing'
        : stageState.phase === 'countdown'
          ? 'Starting…'
          : 'Discussion';

  return (
    <View className="flex-1 bg-stage-bg">
      <Stack.Screen options={{ title: screenTitle }} />
      <ConnectionBanner
        isConnected={isConnected}
        wasReplaced={wasReplaced}
        error={error}
        onReconnect={reconnect}
      />
      <RoleOfferPrompt offers={roleOffers} onRespond={respondRoleOffer} />
      {stageState.phase === 'lobby' && (
        <Lobby
          state={stageState}
          availablePlays={availablePlays}
          participantId={participantId}
          systemMessages={systemMessages}
          onSelectPlay={selectPlay}
          onToggleRole={toggleCastMember}
          onToggleAudience={setAudience}
          onToggleReady={setReady}
          onStart={startPlay}
          onHostPause={hostPause}
          onHostResume={hostResume}
          onHostBackLine={hostBackLine}
          onHostSkipLine={hostSkipLine}
          onHostEndPlay={hostEndPlay}
          onHostTransferHost={hostTransferHost}
          onHostKick={hostKick}
          onHostBan={hostBan}
          onHostMute={hostMute}
          onHostToggleDirectorMode={hostToggleDirectorMode}
          onHostReassignRole={hostReassignRole}
          onHostUnassignRole={hostUnassignRole}
        />
      )}
      {stageState.phase === 'countdown' && <Countdown />}
      {stageState.phase === 'performing' && (
        <View className="flex-1 min-h-0">
          <View className="px-3 pt-3">
            <HostControls
              state={stageState}
              participantId={participantId}
              onPause={hostPause}
              onResume={hostResume}
              onBackLine={hostBackLine}
              onSkipLine={hostSkipLine}
              onEndPlay={hostEndPlay}
              onToggleDirectorMode={hostToggleDirectorMode}
              onReassignRole={hostReassignRole}
              onUnassignRole={hostUnassignRole}
            />
          </View>
          <Stage
            state={stageState}
            participantId={participantId}
            onNextLine={nextLine}
            videoSlot={
              <VideoRoom
                stageId={stageState.id}
                participantId={participantId}
                state={stageState}
              />
            }
          />
        </View>
      )}
      {stageState.phase === 'discussion' && (
        <View className="flex-1">
          <View className="px-3 pt-3">
            <HostControls
              state={stageState}
              participantId={participantId}
              onPause={hostPause}
              onResume={hostResume}
              onBackLine={hostBackLine}
              onSkipLine={hostSkipLine}
              onEndPlay={hostEndPlay}
              onToggleDirectorMode={hostToggleDirectorMode}
              onReassignRole={hostReassignRole}
              onUnassignRole={hostUnassignRole}
            />
          </View>
          <Discussion state={stageState} participantId={participantId} onReset={resetPlay} />
        </View>
      )}
    </View>
  );
}

function ConnectionBanner({
  isConnected,
  wasReplaced,
  error,
  onReconnect,
}: {
  isConnected: boolean;
  wasReplaced: boolean;
  error: string | null;
  onReconnect: () => void;
}) {
  if (isConnected && !wasReplaced) return null;

  const replacedCopy = 'This stage was opened in another window. Controls here are inactive.';
  const reconnectingCopy = error ?? 'Connection lost. Reconnecting…';

  return (
    <View
      className={`flex-row items-center justify-between gap-3 px-4 py-2 ${
        wasReplaced ? 'bg-red-900' : 'bg-amber-900'
      }`}
    >
      <Text className="flex-1 text-xs text-white">
        {wasReplaced ? replacedCopy : reconnectingCopy}
      </Text>
      {wasReplaced ? (
        <Pressable
          onPress={onReconnect}
          accessibilityRole="button"
          className="rounded border border-white/40 px-2 py-1 active:opacity-70"
        >
          <Text className="text-xs font-semibold text-white">Take over here</Text>
        </Pressable>
      ) : null}
    </View>
  );
}
