import { Link, Stack, useLocalSearchParams } from 'expo-router';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, Text, TextInput, View } from 'react-native';
import { usePlayState } from '../../hooks/usePlayState';
import { ChatPanel } from '../../components/ChatPanel';
import { Lobby } from '../../components/Lobby';
import { Countdown } from '../../components/Countdown';
import { Stage } from '../../components/Stage';
import { Discussion } from '../../components/Discussion';
import { HostControls } from '../../components/HostControls';
import { RoleOfferPrompt } from '../../components/RoleOfferPrompt';
import { VacancyBanner } from '../../components/VacancyBanner';
import { VideoRoom } from '../../components/VideoRoom';
import { PLAYER_NAME_KEY } from '../../lib/player-name';

// Server-side display-name policy errors (§12.4) send the user back to the
// name prompt instead of looping reconnect attempts with a rejected name.
const NAME_ERRORS = [
  'Please enter a name.',
  'Please choose a different name.',
  'Name is too long',
];

export default function StageScreen() {
  const { stageId } = useLocalSearchParams<{ stageId: string }>();
  const [playerName, setPlayerName] = useState<string | null>(null);
  const [nameInput, setNameInput] = useState('');
  const [nameError, setNameError] = useState<string | null>(null);
  const [nameChecked, setNameChecked] = useState(false);

  useEffect(() => {
    AsyncStorage.getItem(PLAYER_NAME_KEY).then((v) => {
      if (v) setPlayerName(v);
      setNameChecked(true);
    });
  }, []);

  const {
    stageState,
    participantId,
    availablePlays,
    chatMessages,
    systemMessages,
    roleOffers,
    roleVacancies,
    isConnected,
    error,
    wasReplaced,
    reconnect,
    toggleCastMember,
    setAudience,
    setReady,
    selectPlay,
    sendChat,
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
    audienceClaimRole,
  } = usePlayState({
    stageIdOrCode: String(stageId),
    participantName: playerName ?? '',
    enabled: !!playerName,
  });

  const [chatOpen, setChatOpen] = useState(false);

  // A name rejected by the server's policy sends the user back to the prompt.
  useEffect(() => {
    if (playerName && error && NAME_ERRORS.some((e) => error.startsWith(e))) {
      setNameInput(playerName);
      setPlayerName(null);
      setNameError(error);
    }
  }, [error, playerName]);

  async function onJoin() {
    const trimmed = nameInput.trim();
    if (!trimmed) {
      setNameError('Enter your name to join');
      return;
    }
    setNameError(null);
    await AsyncStorage.setItem(PLAYER_NAME_KEY, trimmed);
    setPlayerName(trimmed);
  }

  if (!nameChecked) return null;

  if (!playerName) {
    return (
      <View className="flex-1 items-center justify-center bg-stage-bg px-6">
        <Stack.Screen options={{ title: 'Join Stage' }} />
        <Text className="mb-6 text-2xl font-serif text-stage-gold">Enter your name to join</Text>
        <View className="w-full max-w-sm gap-3">
          <TextInput
            value={nameInput}
            onChangeText={setNameInput}
            placeholder="Your name"
            placeholderTextColor="#8a8374"
            autoFocus
            maxLength={32}
            onSubmitEditing={onJoin}
            className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
          />
          {nameError ? <Text className="text-red-400 text-sm">{nameError}</Text> : null}
          <Pressable
            onPress={onJoin}
            className="rounded-md bg-stage-gold px-4 py-3"
          >
            <Text className="text-center font-semibold text-stage-bg">Join</Text>
          </Pressable>
          {/* Optional — guests can join without an account (§5.3). */}
          <Link href="/login" asChild>
            <Pressable className="self-center py-1 active:opacity-70">
              <Text className="text-xs text-stage-muted underline">
                Sign in (optional — needed to host or rate)
              </Text>
            </Pressable>
          </Link>
        </View>
      </View>
    );
  }

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

  const meParticipant = stageState.participants.find((p) => p.id === participantId);

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
      <VacancyBanner
        vacancies={roleVacancies}
        isAudience={!!meParticipant?.isAudience}
        onClaim={audienceClaimRole}
      />
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

      {/* Chat: floating toggle + docked panel (§6). */}
      {chatOpen ? (
        <View className="absolute bottom-16 right-3 z-20 h-80 w-80 max-w-full overflow-hidden rounded-md border border-stage-border bg-stage-bg shadow-lg">
          <View className="flex-row items-center justify-between border-b border-stage-border bg-stage-surface px-3 py-2">
            <Text className="text-xs uppercase tracking-widest text-stage-gold">Chat</Text>
            <Pressable onPress={() => setChatOpen(false)} accessibilityRole="button">
              <Text className="text-xs text-stage-muted">Close</Text>
            </Pressable>
          </View>
          <ChatPanel
            messages={chatMessages}
            systemMessages={systemMessages}
            isAudience={!!meParticipant?.isAudience}
            onSend={sendChat}
          />
        </View>
      ) : null}
      <Pressable
        onPress={() => setChatOpen((v) => !v)}
        accessibilityRole="button"
        accessibilityLabel={chatOpen ? 'Close chat' : 'Open chat'}
        className="absolute bottom-3 right-3 z-20 rounded-full border border-stage-border bg-stage-surface px-4 py-2 shadow-lg active:opacity-80"
      >
        <Text className="text-xs font-semibold text-stage-text">
          {chatOpen ? '💬 Hide chat' : '💬 Chat'}
        </Text>
      </Pressable>
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
