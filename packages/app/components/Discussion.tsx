import { useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import { Link } from 'expo-router';
import type { StageState } from '@nsp/shared';
import { useSession } from '../lib/auth-client';
import { ratePlayScript } from '../lib/scripts-api';

interface Props {
  state: StageState;
  participantId?: string | null;
  onReset: () => void;
}

export function Discussion({ state, participantId, onReset }: Props) {
  const me = participantId ? state.participants.find((p) => p.id === participantId) : null;
  const { data: session } = useSession();
  const [rating, setRating] = useState<number | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [rateError, setRateError] = useState<string | null>(null);

  const play = state.selectedPlay;
  const isAuthor =
    !!session?.user && !!play?.createdByUserId && session.user.id === play.createdByUserId;

  async function submit(r: number) {
    if (!play) return;
    setRating(r);
    setSaving(true);
    setRateError(null);
    try {
      await ratePlayScript(play.id, r);
      setSaved(true);
    } catch {
      setRateError('Could not save your rating.');
      setRating(null);
    } finally {
      setSaving(false);
    }
  }

  return (
    <View className="flex-1 items-center justify-center bg-stage-bg px-6 gap-5">
      {me?.forceMuted ? (
        <Text className="text-xs text-red-400">🔇 You are muted by the host.</Text>
      ) : null}
      <Text className="text-5xl font-serif text-stage-gold">Bravo!</Text>
      {play ? <Text className="text-stage-muted">That was "{play.title}"</Text> : null}

      <View className="mt-4 gap-2 items-center">
        <Text className="text-stage-muted">Cast</Text>
        {state.participants
          .filter((p) => !p.isAudience && p.castMemberIds.length > 0)
          .map((p) => (
            <Text key={p.id} className="text-stage-text">
              {p.name}
              {play
                ? ` — ${p.castMemberIds
                    .map((id) => play.cast.find((c) => c.id === id)?.name)
                    .filter(Boolean)
                    .join(', ')}`
                : ''}
            </Text>
          ))}
      </View>

      {play ? (
        session?.user ? (
          isAuthor ? (
            <View className="items-center gap-1">
              <View className="flex-row gap-2 opacity-40">
                {[1, 2, 3, 4, 5].map((n) => (
                  <View key={n} className="rounded-md border border-stage-border px-3 py-1">
                    <Text className="text-stage-muted">★ {n}</Text>
                  </View>
                ))}
              </View>
              <Text className="text-xs text-stage-muted">Authors can't rate their own work.</Text>
            </View>
          ) : (
            <View className="items-center gap-2">
              <Text className="text-stage-muted">Rate this script:</Text>
              <View className="flex-row gap-2">
                {[1, 2, 3, 4, 5].map((n) => (
                  <Pressable
                    key={n}
                    onPress={() => submit(n)}
                    disabled={saving || saved}
                    className="rounded-md border border-stage-border px-3 py-1"
                  >
                    <Text
                      className={rating && n <= rating ? 'text-stage-gold' : 'text-stage-muted'}
                    >
                      ★ {n}
                    </Text>
                  </Pressable>
                ))}
              </View>
              {saved ? <Text className="text-xs text-stage-muted">Thanks!</Text> : null}
              {rateError ? <Text className="text-xs text-red-400">{rateError}</Text> : null}
            </View>
          )
        ) : (
          <View className="items-center gap-1">
            <Link href="/login" asChild>
              <Pressable className="rounded-md border border-stage-gold px-4 py-2 active:opacity-70">
                <Text className="font-semibold text-stage-gold">Sign in to rate</Text>
              </Pressable>
            </Link>
            <Text className="text-xs text-stage-muted">
              Only signed-in users can rate scripts.
            </Text>
          </View>
        )
      ) : null}

      <Pressable onPress={onReset} className="mt-6 rounded-md bg-stage-gold px-5 py-3">
        <Text className="font-semibold text-stage-bg">Perform again</Text>
      </Pressable>
    </View>
  );
}
