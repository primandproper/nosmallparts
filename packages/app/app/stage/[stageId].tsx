import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useEffect, useState } from 'react';
import { ActivityIndicator, Text, View } from 'react-native';
import { usePlayState } from '../../hooks/usePlayState';
import { Lobby } from '../../components/Lobby';
import { Countdown } from '../../components/Countdown';
import { Stage } from '../../components/Stage';
import { Discussion } from '../../components/Discussion';
import { VideoRoom } from '../../components/VideoRoom';

const PLAYER_NAME_KEY = 'nsp:playerName';

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
    isConnected,
    error,
    toggleCastMember,
    setDirector,
    setAudience,
    setReady,
    selectPlay,
    startPlay,
    nextLine,
    resetPlay,
  } = usePlayState({
    stageIdOrCode: String(stageId),
    participantName: playerName ?? '',
    enabled: !!playerName,
  });

  if (!playerName) return null;

  if (!stageState) {
    return (
      <View className="flex-1 items-center justify-center bg-stage-bg">
        <Stack.Screen options={{ title: 'Stage' }} />
        <ActivityIndicator color="#d4af37" />
        <Text className="mt-3 text-stage-muted">
          {isConnected ? 'Joining stage…' : error ? error : 'Connecting…'}
        </Text>
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
      {stageState.phase === 'lobby' && (
        <Lobby
          state={stageState}
          availablePlays={availablePlays}
          participantId={participantId}
          onSelectPlay={selectPlay}
          onToggleRole={toggleCastMember}
          onToggleDirector={setDirector}
          onToggleAudience={setAudience}
          onToggleReady={setReady}
          onStart={startPlay}
        />
      )}
      {stageState.phase === 'countdown' && <Countdown />}
      {stageState.phase === 'performing' && (
        <View className="flex-1">
          <VideoRoom
            stageId={stageState.id}
            participantId={participantId}
            state={stageState}
          />
          <View className="flex-1">
            <Stage state={stageState} participantId={participantId} onNextLine={nextLine} />
          </View>
        </View>
      )}
      {stageState.phase === 'discussion' && (
        <Discussion state={stageState} onReset={resetPlay} />
      )}
    </View>
  );
}
