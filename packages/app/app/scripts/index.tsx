import { Link, Stack, useFocusEffect } from 'expo-router';
import { useCallback, useState } from 'react';
import { ActivityIndicator, FlatList, Pressable, Text, TextInput, View } from 'react-native';
import type { Play } from '@nsp/shared';
import { listPlayScripts } from '../../lib/scripts-api';
import { getDrafts, isDraftId } from '../../lib/drafts';

export default function ScriptsLibrary() {
  const [search, setSearch] = useState('');
  const [plays, setPlays] = useState<Play[]>([]);
  const [drafts, setDrafts] = useState<Play[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (q: string) => {
    setLoading(true);
    setError(null);
    try {
      const [remote, localDrafts] = await Promise.all([listPlayScripts(q), getDrafts()]);
      setPlays(remote);
      setDrafts(localDrafts);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load');
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      load(search);
    }, [load, search])
  );

  const combined = [...drafts, ...plays];

  return (
    <View className="flex-1 bg-stage-bg">
      <Stack.Screen
        options={{
          title: 'Scripts',
          headerRight: () => (
            <Link href="/scripts/new" asChild>
              <Pressable hitSlop={10}>
                <Text className="text-stage-gold text-base font-semibold">+ New</Text>
              </Pressable>
            </Link>
          ),
        }}
      />
      <View className="mx-auto w-full max-w-3xl px-4 pt-3">
        <TextInput
          placeholder="Search title, author, cast, tag…"
          placeholderTextColor="#8a8374"
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-2 text-stage-text"
          value={search}
          onChangeText={(v) => {
            setSearch(v);
            load(v);
          }}
        />
      </View>
      {loading ? (
        <View className="flex-1 items-center justify-center">
          <ActivityIndicator color="#d4af37" />
        </View>
      ) : error ? (
        <View className="flex-1 items-center justify-center px-6">
          <Text className="text-red-400">{error}</Text>
        </View>
      ) : (
        <FlatList
          data={combined}
          keyExtractor={(p) => p.id}
          contentContainerStyle={{
            padding: 16,
            gap: 10,
            width: '100%',
            maxWidth: 768,
            alignSelf: 'center',
          }}
          ListEmptyComponent={
            <Text className="text-center text-stage-muted mt-12">No scripts yet.</Text>
          }
          renderItem={({ item }) => <ScriptCard play={item} isDraft={isDraftId(item.id)} />}
        />
      )}
    </View>
  );
}

function ScriptCard({ play, isDraft }: { play: Play; isDraft: boolean }) {
  return (
    <Link
      href={isDraft ? `/scripts/new?draftId=${play.id}` : `/scripts/${play.id}`}
      asChild
    >
      <Pressable className="rounded-lg border border-stage-border bg-stage-surface p-4">
        <View className="flex-row items-center justify-between gap-2">
          <Text className="flex-1 text-lg font-semibold text-stage-text" numberOfLines={1}>
            {play.title || 'Untitled'}
          </Text>
          {play.enabled === false ? (
            <Text className="rounded bg-red-900 px-2 py-0.5 text-xs text-red-200">disabled</Text>
          ) : null}
          {isDraft ? (
            <Text className="rounded bg-stage-border px-2 py-0.5 text-xs text-stage-muted">
              draft
            </Text>
          ) : play.isPublic ? (
            <Text className="rounded bg-stage-goldDim px-2 py-0.5 text-xs text-stage-bg">
              public
            </Text>
          ) : null}
        </View>
        {play.author ? (
          <Text className="mt-1 text-sm text-stage-muted">by {play.author}</Text>
        ) : null}
        {play.description ? (
          <Text className="mt-2 text-sm text-stage-text" numberOfLines={2}>
            {play.description}
          </Text>
        ) : null}
        <View className="mt-3 flex-row gap-3">
          <Text className="text-xs text-stage-muted">{play.cast.length} cast</Text>
          <Text className="text-xs text-stage-muted">{play.lines.length} lines</Text>
          {play.ratingStats ? (
            <Text className="text-xs text-stage-muted">
              ★ {play.ratingStats.averageRating.toFixed(1)} ({play.ratingStats.ratingCount})
            </Text>
          ) : null}
        </View>
      </Pressable>
    </Link>
  );
}
