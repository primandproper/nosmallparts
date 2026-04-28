import { Stack, useLocalSearchParams, useRouter } from 'expo-router';
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  ScrollView,
  Switch,
  Text,
  TextInput,
  View,
} from 'react-native';
import type { CastMember, Play, PlayLine } from '@nsp/shared';
import { validatePlay } from '@nsp/shared';
import { useSession } from '../../lib/auth-client';
import { createPlayScript, getPlayScriptById, updatePlayScript } from '../../lib/scripts-api';
import { deleteDraft, getDraft, saveDraft, isDraftId } from '../../lib/drafts';

interface CastRow extends CastMember {
  key: string;
}

interface LineRow extends PlayLine {
  key: string;
}

export default function ScriptEditor() {
  const router = useRouter();
  const { draftId, id } = useLocalSearchParams<{ draftId?: string; id?: string }>();
  const editingRemoteId = id && typeof id === 'string' && !isDraftId(id) ? id : null;
  const editingDraftId = (draftId ?? (id && isDraftId(String(id)) ? String(id) : null)) || null;

  const { data: session } = useSession();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [author, setAuthor] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [cast, setCast] = useState<CastRow[]>([]);
  const [lines, setLines] = useState<LineRow[]>([]);
  const [currentDraftId, setCurrentDraftId] = useState<string | null>(editingDraftId);
  const [loading, setLoading] = useState(true);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const autosaveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        if (editingRemoteId) {
          const p = await getPlayScriptById(editingRemoteId);
          if (cancelled) return;
          hydrateFromPlay(p);
        } else if (editingDraftId) {
          const d = await getDraft(editingDraftId);
          if (cancelled) return;
          if (d) hydrateFromPlay(d);
        }
      } catch (e) {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Failed to load');
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [editingRemoteId, editingDraftId]);

  function hydrateFromPlay(p: Play) {
    setTitle(p.title);
    setDescription(p.description ?? '');
    setAuthor(p.author ?? '');
    setIsPublic(p.isPublic ?? false);
    setCast(p.cast.map((c, i) => ({ ...c, key: `c-${i}-${c.id}` })));
    setLines(
      p.lines.map((l, i) => ({
        ...l,
        key: `l-${i}-${l.index}`,
      }))
    );
  }

  // Autosave drafts every 3s while editing.
  useEffect(() => {
    if (loading) return;
    if (editingRemoteId) return; // Editing a published script — don't silently save as draft.
    if (autosaveTimer.current) clearTimeout(autosaveTimer.current);
    autosaveTimer.current = setTimeout(async () => {
      if (!title.trim() && cast.length === 0 && lines.length === 0) return;
      const saved = await saveDraft({
        id: currentDraftId ?? undefined,
        title: title || 'Untitled',
        description,
        author: author || undefined,
        cast: cast.map(({ key, ...c }) => c),
        lines: lines.map(({ key, ...l }) => l),
      });
      if (!currentDraftId) setCurrentDraftId(saved.id);
    }, 3000);
    return () => {
      if (autosaveTimer.current) clearTimeout(autosaveTimer.current);
    };
  }, [title, description, author, cast, lines, editingRemoteId, currentDraftId, loading]);

  function addCastMember() {
    const n = cast.length + 1;
    setCast((prev) => [
      ...prev,
      { id: `new-${Date.now()}-${n}`, name: `Character ${n}`, key: `c-${Date.now()}` },
    ]);
  }
  function updateCast(key: string, patch: Partial<CastMember>) {
    setCast((prev) => prev.map((c) => (c.key === key ? { ...c, ...patch } : c)));
  }
  function removeCast(key: string) {
    const removed = cast.find((c) => c.key === key);
    setCast((prev) => prev.filter((c) => c.key !== key));
    if (removed) {
      setLines((prev) =>
        prev
          .map((l) => ({
            ...l,
            castMemberIds: (l.castMemberIds || []).filter((id) => id !== removed.id),
          }))
          .filter((l) => l.castMemberIds.length > 0)
      );
    }
  }

  function addLine() {
    const nextIndex = lines.length === 0 ? 0 : Math.max(...lines.map((l) => l.index)) + 1;
    const firstCast = cast[0];
    setLines((prev) => [
      ...prev,
      {
        key: `l-${Date.now()}`,
        castMemberIds: firstCast ? [firstCast.id] : [],
        text: '',
        index: nextIndex,
      },
    ]);
  }
  function addOverlappingLine(atIndex: number) {
    const used = new Set(
      lines.filter((l) => l.index === atIndex).flatMap((l) => l.castMemberIds || [])
    );
    const remaining = cast.find((c) => !used.has(c.id)) ?? cast[0];
    setLines((prev) => [
      ...prev,
      {
        key: `l-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
        castMemberIds: remaining ? [remaining.id] : [],
        text: '',
        index: atIndex,
      },
    ]);
  }
  function updateLine(key: string, patch: Partial<PlayLine>) {
    setLines((prev) => prev.map((l) => (l.key === key ? { ...l, ...patch } : l)));
  }
  function toggleSpeaker(key: string, castMemberId: string) {
    setLines((prev) =>
      prev.map((l) => {
        if (l.key !== key) return l;
        const has = (l.castMemberIds || []).includes(castMemberId);
        const next = has
          ? l.castMemberIds.filter((id) => id !== castMemberId)
          : [...(l.castMemberIds || []), castMemberId];
        // Don't allow removing the last speaker.
        return { ...l, castMemberIds: next.length === 0 ? l.castMemberIds : next };
      })
    );
  }
  function removeLine(key: string) {
    setLines((prev) => prev.filter((l) => l.key !== key));
  }

  const validation = useMemo(
    () =>
      validatePlay(
        title,
        cast.map(({ key, ...c }) => c),
        lines.map(({ key, ...l }) => l)
      ),
    [title, cast, lines]
  );

  async function publish() {
    if (!validation.valid) {
      setError(validation.error ?? 'Invalid script');
      return;
    }
    if (!session?.user) {
      Alert.alert('Sign in required', 'Please sign in to publish scripts.');
      return;
    }
    setPublishing(true);
    setError(null);
    try {
      const payload = {
        title,
        description,
        author: author.trim() || null,
        isPublic,
        cast: cast.map(({ key, ...c }) => c),
        lines: lines.map(({ key, ...l }) => l),
      };
      const saved = editingRemoteId
        ? await updatePlayScript(editingRemoteId, payload)
        : await createPlayScript(payload);
      if (currentDraftId) await deleteDraft(currentDraftId);
      router.replace(`/scripts/${saved.id}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Publish failed');
    } finally {
      setPublishing(false);
    }
  }

  async function discardDraft() {
    if (!currentDraftId) {
      router.back();
      return;
    }
    await deleteDraft(currentDraftId);
    router.back();
  }

  if (loading) {
    return (
      <View className="flex-1 items-center justify-center bg-stage-bg">
        <ActivityIndicator color="#d4af37" />
      </View>
    );
  }

  return (
    <KeyboardAvoidingView
      className="flex-1 bg-stage-bg"
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <Stack.Screen options={{ title: editingRemoteId ? 'Edit script' : 'New script' }} />
      <ScrollView
        contentContainerStyle={{
          padding: 16,
          gap: 14,
          width: '100%',
          maxWidth: 768,
          alignSelf: 'center',
        }}
      >
        <TextInput
          value={title}
          onChangeText={setTitle}
          placeholder="Title"
          placeholderTextColor="#8a8374"
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-lg text-stage-text"
        />
        <TextInput
          value={author}
          onChangeText={setAuthor}
          placeholder="Author (optional)"
          placeholderTextColor="#8a8374"
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
        />
        <TextInput
          value={description}
          onChangeText={setDescription}
          placeholder="Description"
          placeholderTextColor="#8a8374"
          multiline
          numberOfLines={3}
          className="rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text"
        />
        <View className="flex-row items-center gap-3">
          <Switch
            value={isPublic}
            onValueChange={setIsPublic}
            trackColor={{ false: '#2a2a2a', true: '#9a7e27' }}
            thumbColor={isPublic ? '#d4af37' : '#8a8374'}
          />
          <Text className="text-stage-text">Public</Text>
        </View>

        <View className="mt-2">
          <View className="flex-row items-center justify-between">
            <Text className="text-xl font-serif text-stage-gold">Cast</Text>
            <Pressable onPress={addCastMember} className="px-2 py-1">
              <Text className="text-stage-gold">+ Add character</Text>
            </Pressable>
          </View>
          {cast.length === 0 ? (
            <Text className="mt-2 text-sm text-stage-muted">No characters yet.</Text>
          ) : (
            <View className="mt-2 gap-2">
              {cast.map((c) => (
                <View
                  key={c.key}
                  className="rounded-md border border-stage-border bg-stage-surface p-3"
                >
                  <TextInput
                    value={c.name}
                    onChangeText={(v) => updateCast(c.key, { name: v })}
                    placeholder="Character name"
                    placeholderTextColor="#8a8374"
                    className="text-stage-text"
                  />
                  <TextInput
                    value={c.description ?? ''}
                    onChangeText={(v) => updateCast(c.key, { description: v })}
                    placeholder="Description (optional)"
                    placeholderTextColor="#8a8374"
                    className="mt-1 text-sm text-stage-muted"
                  />
                  <Pressable onPress={() => removeCast(c.key)} className="mt-2 self-end">
                    <Text className="text-red-400 text-sm">Remove</Text>
                  </Pressable>
                </View>
              ))}
            </View>
          )}
        </View>

        <View className="mt-4">
          <View className="flex-row items-center justify-between">
            <Text className="text-xl font-serif text-stage-gold">Lines</Text>
            <Pressable
              onPress={addLine}
              disabled={cast.length === 0}
              className="px-2 py-1"
            >
              <Text className={cast.length === 0 ? 'text-stage-muted' : 'text-stage-gold'}>
                + Add line
              </Text>
            </Pressable>
          </View>
          {lines.length === 0 ? (
            <Text className="mt-2 text-sm text-stage-muted">
              {cast.length === 0 ? 'Add a character first.' : 'No lines yet.'}
            </Text>
          ) : (
            <View className="mt-2 gap-3">
              {Array.from(new Set(lines.map((l) => l.index)))
                .sort((a, b) => a - b)
                .map((turnIndex) => {
                  const groupLines = lines.filter((l) => l.index === turnIndex);
                  return (
                    <View key={`turn-${turnIndex}`} className="gap-1.5">
                      <View className="flex-row items-center justify-between">
                        <Text className="text-xs text-stage-muted">
                          Turn {turnIndex + 1}
                          {groupLines.length > 1 ? ' · overlapping' : ''}
                        </Text>
                        <Pressable
                          onPress={() => addOverlappingLine(turnIndex)}
                          disabled={cast.length === 0}
                          className="px-2 py-0.5"
                        >
                          <Text
                            className={`text-xs ${
                              cast.length === 0 ? 'text-stage-muted' : 'text-stage-gold'
                            }`}
                          >
                            + Overlap
                          </Text>
                        </Pressable>
                      </View>
                      {groupLines.map((l) => {
                        const selectedIds = new Set(l.castMemberIds || []);
                        const isUnison = selectedIds.size > 1;
                        return (
                          <View
                            key={l.key}
                            className="rounded-md border border-stage-border bg-stage-surface p-3"
                          >
                            <View className="flex-row flex-wrap items-center gap-1">
                              {isUnison ? (
                                <Text className="mr-1 text-[10px] uppercase tracking-wider text-stage-gold font-semibold">
                                  Unison
                                </Text>
                              ) : null}
                              {cast.map((c) => {
                                const selected = selectedIds.has(c.id);
                                return (
                                  <Pressable
                                    key={c.id}
                                    onPress={() => toggleSpeaker(l.key, c.id)}
                                    className={`rounded-full border px-2 py-0.5 ${
                                      selected
                                        ? 'border-stage-gold bg-stage-goldDim'
                                        : 'border-stage-border'
                                    }`}
                                  >
                                    <Text
                                      className={`text-xs ${
                                        selected ? 'text-stage-bg' : 'text-stage-muted'
                                      }`}
                                    >
                                      {c.name}
                                    </Text>
                                  </Pressable>
                                );
                              })}
                            </View>
                            <TextInput
                              value={l.direction ?? ''}
                              onChangeText={(v) =>
                                updateLine(l.key, { direction: v || undefined })
                              }
                              placeholder="Stage direction (optional)"
                              placeholderTextColor="#8a8374"
                              className="mt-2 italic text-stage-muted"
                            />
                            <TextInput
                              value={l.text}
                              onChangeText={(v) => updateLine(l.key, { text: v })}
                              placeholder="Line text"
                              placeholderTextColor="#8a8374"
                              multiline
                              className="mt-1 text-stage-text"
                            />
                            <Pressable onPress={() => removeLine(l.key)} className="mt-2 self-end">
                              <Text className="text-red-400 text-sm">Remove</Text>
                            </Pressable>
                          </View>
                        );
                      })}
                    </View>
                  );
                })}
            </View>
          )}
        </View>

        {error ? <Text className="text-red-400">{error}</Text> : null}
        {!validation.valid && !error ? (
          <Text className="text-stage-muted text-sm">{validation.error}</Text>
        ) : null}

        <Pressable
          onPress={publish}
          disabled={publishing}
          className={`mt-4 rounded-md px-4 py-3 ${publishing ? 'bg-stage-goldDim' : 'bg-stage-gold'}`}
        >
          <Text className="text-center font-semibold text-stage-bg">
            {publishing
              ? 'Publishing…'
              : editingRemoteId
                ? 'Save changes'
                : session?.user
                  ? 'Publish'
                  : 'Sign in to publish'}
          </Text>
        </Pressable>

        {!editingRemoteId ? (
          <Pressable onPress={discardDraft} className="mt-2">
            <Text className="text-center text-stage-muted">Discard draft</Text>
          </Pressable>
        ) : null}
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
