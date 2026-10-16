import { readItem, writeItem } from './storage';

const PLAYER_NAME_KEY = 'nsp:playerName';

/** getStoredPlayerName is the name this browser last joined a stage as. */
export function getStoredPlayerName(): string | null {
  return readItem(PLAYER_NAME_KEY);
}

export function storePlayerName(name: string): void {
  writeItem(PLAYER_NAME_KEY, name);
}
