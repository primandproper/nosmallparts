import { createAuthClient } from 'better-auth/react';
import { expoClient } from '@better-auth/expo/client';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';
import { getApiBaseUrl } from './config';

/**
 * Better Auth client. On native, uses expo-secure-store via the Expo plugin
 * to persist the session token. On web, falls back to cookies.
 */
export const authClient = createAuthClient({
  baseURL: getApiBaseUrl(),
  plugins:
    Platform.OS === 'web'
      ? []
      : [
          expoClient({
            scheme: 'nosmallparts',
            storagePrefix: 'nsp',
            storage: SecureStore,
          }),
        ],
});

export const { useSession, signIn, signUp, signOut } = authClient;
