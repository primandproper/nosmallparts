import Constants from 'expo-constants';
import { Platform } from 'react-native';

/**
 * Base URL of the Hono server.
 *
 * - Web: proxy via same origin when possible; otherwise env override.
 * - Native: must be the LAN IP of the dev machine so phones can reach it.
 *
 * Override by setting `EXPO_PUBLIC_API_URL` in the environment or app.json.
 */
export function getApiBaseUrl(): string {
  const override = process.env.EXPO_PUBLIC_API_URL;
  if (override) return override.replace(/\/$/, '');

  // Try to infer the host Metro is served from (native dev only).
  const hostUri =
    (Constants.expoConfig as any)?.hostUri || (Constants as any).manifest2?.extra?.expoGo?.debuggerHost;
  if (Platform.OS !== 'web' && hostUri) {
    const host = String(hostUri).split(':')[0];
    return `http://${host}:3001`;
  }

  return 'http://localhost:3001';
}

export function getWsUrl(path = '/ws/stage'): string {
  const base = getApiBaseUrl();
  return base.replace(/^http/, 'ws') + path;
}
