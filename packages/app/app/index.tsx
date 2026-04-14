import { Link, useRouter } from 'expo-router';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { useEffect, useState } from 'react';
import { Pressable, Text, TextInput, View } from 'react-native';
import { useSession, signOut } from '../lib/auth-client';
import { api } from '../lib/api';

const PLAYER_NAME_KEY = 'nsp:playerName';

export default function Home() {
  const router = useRouter();
  const { data, isPending } = useSession();
  const user = data?.user;

  const [name, setName] = useState('');
  const [code, setCode] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    AsyncStorage.getItem(PLAYER_NAME_KEY).then((v) => {
      if (v) setName(v);
    });
  }, []);

  useEffect(() => {
    if (user?.name && !name) setName(user.name);
  }, [user, name]);

  async function onEnter() {
    if (!name.trim()) {
      setError('Enter your name');
      return;
    }
    await AsyncStorage.setItem(PLAYER_NAME_KEY, name.trim());
    setBusy(true);
    setError(null);
    try {
      const joinCode = code.trim().toUpperCase();
      if (joinCode) {
        const res = await api<{ exists: boolean; code?: string }>(
          `/api/stages/${encodeURIComponent(joinCode)}/exists`
        );
        if (!res.exists) {
          setError("Stage not found. Leave code blank to create one.");
          return;
        }
        router.push(`/stage/${res.code}`);
      } else {
        const res = await api<{ id: string; code: string }>('/api/stages', {
          method: 'POST',
          body: JSON.stringify({}),
        });
        router.push(`/stage/${res.code}`);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed');
    } finally {
      setBusy(false);
    }
  }

  return (
    <View className="flex-1 items-center bg-stage-bg px-6 pt-16">
      <Text className="mb-2 text-4xl font-serif text-stage-gold">No Small Parts</Text>
      <Text className="mb-10 text-center text-stage-muted">
        Perform short scripts together in real time.
      </Text>

      <View className="w-full max-w-sm gap-3">
        <TextInput
          value={name}
          onChangeText={setName}
          placeholder="Your name"
          placeholderTextColor="#8a8374"
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
        />
        <TextInput
          value={code}
          onChangeText={(v) => setCode(v.toUpperCase())}
          placeholder="Stage code (leave blank to create)"
          placeholderTextColor="#8a8374"
          autoCapitalize="characters"
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
        />
        {error ? <Text className="text-red-400 text-sm">{error}</Text> : null}
        <Pressable
          onPress={onEnter}
          disabled={busy}
          className={`rounded-md px-4 py-3 ${busy ? 'bg-stage-goldDim' : 'bg-stage-gold'}`}
        >
          <Text className="text-center font-semibold text-stage-bg">
            {busy ? '…' : code ? 'Join stage' : 'Create stage'}
          </Text>
        </Pressable>

        <View className="mt-6 gap-2">
          <Link href="/scripts" asChild>
            <Pressable className="rounded-md border border-stage-border px-4 py-3">
              <Text className="text-center text-stage-text">Browse scripts</Text>
            </Pressable>
          </Link>
          {isPending ? null : user ? (
            <Pressable
              className="rounded-md border border-stage-border px-4 py-2"
              onPress={() => signOut()}
            >
              <Text className="text-center text-stage-muted text-sm">
                Signed in as {user.email} · Sign out
              </Text>
            </Pressable>
          ) : (
            <View className="flex-row gap-2">
              <Link href="/login" asChild>
                <Pressable className="flex-1 rounded-md border border-stage-border px-3 py-2">
                  <Text className="text-center text-stage-text">Sign in</Text>
                </Pressable>
              </Link>
              <Link href="/register" asChild>
                <Pressable className="flex-1 rounded-md border border-stage-border px-3 py-2">
                  <Text className="text-center text-stage-text">Register</Text>
                </Pressable>
              </Link>
            </View>
          )}
        </View>
      </View>
    </View>
  );
}
