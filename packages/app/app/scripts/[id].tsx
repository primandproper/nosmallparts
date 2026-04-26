import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { useEffect, useState } from 'react';
import { ActivityIndicator, Pressable, ScrollView, Text, View } from 'react-native';
import type { Play } from '@nsp/shared';
import { getPlayScriptById, ratePlayScript } from '../../lib/scripts-api';
import { useSession } from '../../lib/auth-client';
import { api } from '../../lib/api';
import { getStoredPlayerName } from '../../lib/player-name';

export default function ScriptDetail() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { data: session } = useSession();
  const [play, setPlay] = useState<Play | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [rating, setRating] = useState<number | null>(null);
  const [savingRating, setSavingRating] = useState(false);
  const [startingStage, setStartingStage] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        setLoading(true);
        const p = await getPlayScriptById(String(id));
        if (!cancelled) {
          setPlay(p);
          setRating(p.userRating ?? null);
        }
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Failed');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id]);

  async function submitRating(r: number) {
    if (!play) return;
    setSavingRating(true);
    try {
      const stats = await ratePlayScript(play.id, r);
      setRating(stats.userRating);
      setPlay({
        ...play,
        userRating: stats.userRating,
        ratingStats: {
          averageRating: stats.averageRating,
          ratingCount: stats.ratingCount,
        },
      });
    } finally {
      setSavingRating(false);
    }
  }

  if (loading) {
    return (
      <View className="flex-1 items-center justify-center bg-stage-bg">
        <ActivityIndicator color="#d4af37" />
      </View>
    );
  }
  if (error || !play) {
    return (
      <View className="flex-1 items-center justify-center bg-stage-bg px-6">
        <Text className="text-red-400">{error || 'Not found'}</Text>
      </View>
    );
  }

  const castById = new Map(play.cast.map((c) => [c.id, c]));
  const sortedLines = [...play.lines].sort((a, b) => a.index - b.index);

  return (
    <ScrollView
      className="flex-1 bg-stage-bg"
      contentContainerStyle={{
        padding: 16,
        width: '100%',
        maxWidth: 768,
        alignSelf: 'center',
      }}
    >
      <Stack.Screen options={{ title: play.title }} />

      <Text className="text-3xl font-serif text-stage-gold">{play.title}</Text>
      {play.author ? (
        <Text className="mt-1 text-stage-muted">by {play.author}</Text>
      ) : null}
      {play.description ? (
        <Text className="mt-3 text-stage-text">{play.description}</Text>
      ) : null}

      <View className="mt-4 flex-row flex-wrap gap-2">
        {play.cast.map((c) => (
          <Text
            key={c.id}
            className="rounded-full border border-stage-border bg-stage-surface px-3 py-1 text-xs text-stage-text"
          >
            {c.name}
          </Text>
        ))}
      </View>

      {session?.user ? (
        <View className="mt-6">
          <Text className="mb-2 text-stage-muted">Your rating:</Text>
          <View className="flex-row gap-2">
            {[1, 2, 3, 4, 5].map((n) => (
              <Pressable
                key={n}
                disabled={savingRating}
                onPress={() => submitRating(n)}
                className="rounded-md border border-stage-border px-3 py-1"
              >
                <Text
                  className={
                    rating && n <= rating ? 'text-stage-gold' : 'text-stage-muted'
                  }
                >
                  ★ {n}
                </Text>
              </Pressable>
            ))}
          </View>
          {play.ratingStats ? (
            <Text className="mt-2 text-xs text-stage-muted">
              Avg {play.ratingStats.averageRating.toFixed(1)} across{' '}
              {play.ratingStats.ratingCount} rating
              {play.ratingStats.ratingCount === 1 ? '' : 's'}
            </Text>
          ) : null}
        </View>
      ) : null}

      <Pressable
        onPress={async () => {
          if (startingStage) return;
          setStartingStage(true);
          try {
            const name = await getStoredPlayerName();
            if (!name) {
              router.push({ pathname: '/', params: { s: play.id } });
              return;
            }
            const res = await api<{ id: string; code: string }>('/api/stages', {
              method: 'POST',
              body: JSON.stringify({ playScriptId: play.id }),
            });
            router.replace(`/stage/${res.code}`);
          } finally {
            setStartingStage(false);
          }
        }}
        disabled={startingStage}
        className={`mt-6 rounded-md px-4 py-3 ${startingStage ? 'bg-stage-goldDim' : 'bg-stage-gold'}`}
      >
        <Text className="text-center font-semibold text-stage-bg">
          {startingStage ? '…' : 'Start a stage with this script'}
        </Text>
      </Pressable>

      <Text className="mt-8 mb-2 text-xl font-serif text-stage-gold">Script</Text>
      {sortedLines.map((l, i) => {
        const c = castById.get(l.castMemberId);
        return (
          <View key={`${l.castMemberId}-${i}`} className="mb-3">
            <Text className="text-stage-gold">{c?.name || 'Unknown'}</Text>
            {l.direction ? (
              <Text className="italic text-stage-muted">({l.direction})</Text>
            ) : null}
            <Text className="text-stage-text">{l.text}</Text>
          </View>
        );
      })}
    </ScrollView>
  );
}
