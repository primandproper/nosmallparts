import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import '../global.css';

export default function RootLayout() {
  return (
    <SafeAreaProvider>
      <StatusBar style="light" />
      <Stack
        screenOptions={{
          headerStyle: { backgroundColor: '#0a0a0a' },
          headerTintColor: '#d4af37',
          contentStyle: { backgroundColor: '#0a0a0a' },
          headerTitleStyle: { fontWeight: '600' },
        }}
      >
        <Stack.Screen name="index" options={{ title: 'No Small Parts' }} />
        <Stack.Screen name="login" options={{ title: 'Sign in' }} />
        <Stack.Screen name="register" options={{ title: 'Create account' }} />
        <Stack.Screen name="scripts/index" options={{ title: 'Scripts' }} />
        <Stack.Screen name="scripts/new" options={{ title: 'Script editor' }} />
        <Stack.Screen name="scripts/[id]" options={{ title: 'Script' }} />
        <Stack.Screen name="stage/[stageId]" options={{ title: 'Stage' }} />
      </Stack>
    </SafeAreaProvider>
  );
}
