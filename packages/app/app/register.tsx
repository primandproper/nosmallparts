import { useRouter } from 'expo-router';
import { useState } from 'react';
import { Pressable, Text, TextInput, View } from 'react-native';
import { signUp } from '../lib/auth-client';

export default function Register() {
  const router = useRouter();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit() {
    setError(null);
    setBusy(true);
    try {
      const res = await signUp.email({ email: email.trim(), password, name: name.trim() });
      if (res.error) {
        setError(res.error.message || 'Sign-up failed');
      } else {
        router.replace('/');
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Sign-up failed');
    } finally {
      setBusy(false);
    }
  }

  return (
    <View className="flex-1 items-center justify-center bg-stage-bg px-6">
      <View className="w-full max-w-sm gap-3">
        <Text className="mb-4 text-center text-2xl font-serif text-stage-gold">
          Create account
        </Text>
        <TextInput
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
          placeholder="Name"
          placeholderTextColor="#8a8374"
          value={name}
          onChangeText={setName}
        />
        <TextInput
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
          placeholder="Email"
          placeholderTextColor="#8a8374"
          autoCapitalize="none"
          autoComplete="email"
          keyboardType="email-address"
          value={email}
          onChangeText={setEmail}
        />
        <TextInput
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
          placeholder="Password (min 8 chars)"
          placeholderTextColor="#8a8374"
          secureTextEntry
          value={password}
          onChangeText={setPassword}
        />
        {error ? <Text className="text-red-400">{error}</Text> : null}
        <Pressable
          className="rounded-md bg-stage-gold px-4 py-3 disabled:opacity-50"
          onPress={onSubmit}
          disabled={busy || !name || !email || password.length < 8}
        >
          <Text className="text-center font-semibold text-stage-bg">
            {busy ? 'Creating…' : 'Create account'}
          </Text>
        </Pressable>
      </View>
    </View>
  );
}
