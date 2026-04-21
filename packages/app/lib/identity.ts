import { Platform } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as SecureStore from 'expo-secure-store';
import { api } from './api';

export const GUEST_ID_KEY = 'nsp:guestId';

export interface Identity {
  userId: string | null;
  guestId: string | null;
}

async function readLocalGuestId(): Promise<string | null> {
  return AsyncStorage.getItem(GUEST_ID_KEY);
}

async function writeLocalGuestId(id: string): Promise<void> {
  await AsyncStorage.setItem(GUEST_ID_KEY, id);
}

/**
 * Ensures we have a stable identity for the current user/device. Calls
 * `/api/identity` which returns the authed user's ID if signed in, or
 * a server-minted guest ULID otherwise. Guest ID is persisted locally
 * (cookie on web, AsyncStorage on native) so reconnects stay stable.
 */
export async function ensureIdentity(): Promise<Identity> {
  const existing = await readLocalGuestId();
  const qs = existing ? `?guest=${encodeURIComponent(existing)}` : '';
  const res = await api<Identity>(`/api/identity${qs}`);
  if (res.guestId) await writeLocalGuestId(res.guestId);
  return res;
}

/**
 * Read the Better Auth bearer token on native so it can be passed as a
 * WebSocket query param (WebSocket API can't set headers on native).
 */
export async function getSessionToken(): Promise<string | null> {
  if (Platform.OS === 'web') return null;
  try {
    return (await SecureStore.getItemAsync('nsp.session_token')) || null;
  } catch {
    return null;
  }
}
