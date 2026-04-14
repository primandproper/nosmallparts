import { useEffect, useState } from 'react';
import { Text, View } from 'react-native';

export function Countdown({ seconds = 3 }: { seconds?: number }) {
  const [remaining, setRemaining] = useState(seconds);
  useEffect(() => {
    if (remaining <= 0) return;
    const t = setTimeout(() => setRemaining((r) => r - 1), 1000);
    return () => clearTimeout(t);
  }, [remaining]);
  return (
    <View className="flex-1 items-center justify-center bg-stage-bg">
      <Text className="text-stage-muted">Get ready…</Text>
      <Text className="mt-3 text-9xl font-serif text-stage-gold">
        {remaining > 0 ? remaining : '🎭'}
      </Text>
    </View>
  );
}
