import { useState } from 'react';
import { Pressable, Text, View } from 'react-native';
import type { StageState } from '@nsp/shared';
import { useSession } from '../lib/auth-client';
import { ratePlayScript } from '../lib/scripts-api';

interface Props {
  state: StageState;
  onReset: () => void;
}

export function Discussion({ state, onReset }: Props) {
  const { data: session } = useSession();
  const [rating, setRating] = useState<number | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const play = state.selectedPlay;
  async function submit(r: number) {
    if (!play) return;
    setRating(r);
    setSaving(true);
    try {
      await ratePlayScript(play.id, r);
      setSaved(true);
    } catch {
      // ignore
    } finally {
      setSaving(false);
    }
  }

  return (
    <View className="flex-1 items-center justify-center bg-stage-bg px-6 gap-5">
      <Text className="text-5xl font-serif text-stage-gold">Bravo!</Text>
      {play ? <Text className="text-stage-muted">That was "{play.title}"</Text> : null}

      <View className="mt-4 gap-2 items-center">
        <Text className="text-stage-muted">Cast</Text>
        {state.participants
          .filter((p) => !p.isAudience && p.castMemberIds.length > 0)
          .map((p) => (
            <Text key={p.id} className="text-stage-text">
              {p.name}
            </Text>
          ))}
      </View>

      {session?.user && play ? (
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
        </View>
      ) : null}

      <Pressable onPress={onReset} className="mt-6 rounded-md bg-stage-gold px-5 py-3">
        <Text className="font-semibold text-stage-bg">Perform again</Text>
      </Pressable>
    </View>
  );
}
