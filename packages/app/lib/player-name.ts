import AsyncStorage from '@react-native-async-storage/async-storage';

export const PLAYER_NAME_KEY = 'nsp:playerName';

export function getStoredPlayerName(): Promise<string | null> {
  return AsyncStorage.getItem(PLAYER_NAME_KEY);
}
