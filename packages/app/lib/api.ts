import { Platform } from 'react-native';
import * as SecureStore from 'expo-secure-store';
import { getApiBaseUrl } from './config';

/** Read the Better Auth session token (native) to attach as a bearer. */
async function getBearerToken(): Promise<string | null> {
  if (Platform.OS === 'web') return null;
  try {
    return (await SecureStore.getItemAsync('nsp.session_token')) || null;
  } catch {
    return null;
  }
}

export async function api<T = unknown>(
  path: string,
  init: RequestInit = {}
): Promise<T> {
  const url = getApiBaseUrl() + path;
  const headers = new Headers(init.headers);
  if (!headers.has('Content-Type') && init.body) {
    headers.set('Content-Type', 'application/json');
  }
  const token = await getBearerToken();
  if (token) headers.set('Authorization', `Bearer ${token}`);

  const res = await fetch(url, {
    ...init,
    headers,
    credentials: Platform.OS === 'web' ? 'include' : undefined,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${res.status}: ${text || res.statusText}`);
  }
  const ct = res.headers.get('Content-Type') || '';
  if (ct.includes('application/json')) {
    return (await res.json()) as T;
  }
  return (await res.text()) as unknown as T;
}
