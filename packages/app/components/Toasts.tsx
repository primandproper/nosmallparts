import { Pressable, Text, View } from 'react-native';
import type { Toast } from '../hooks/usePlayState';

interface Props {
  toasts: Toast[];
  onDismiss: (id: number) => void;
}

/**
 * Top-center stack of transient notices. Each toast auto-dismisses (see
 * usePlayState) and can also be tapped to dismiss immediately; the stack is
 * capped so notices never pile up indefinitely (§11).
 */
export function ToastHost({ toasts, onDismiss }: Props) {
  if (toasts.length === 0) return null;
  return (
    <View
      pointerEvents="box-none"
      className="absolute left-0 right-0 top-3 z-30 items-center gap-2 px-4"
    >
      {toasts.map((t) => (
        <Pressable
          key={t.id}
          onPress={() => onDismiss(t.id)}
          accessibilityRole="button"
          accessibilityLabel={`Dismiss notice: ${t.message}`}
          className="w-full max-w-md flex-row items-start justify-between gap-3 rounded-md border border-stage-border bg-stage-surface px-4 py-3 shadow-lg active:opacity-80"
        >
          <Text className="flex-1 text-sm text-stage-text">{t.message}</Text>
          <Text className="text-base leading-none text-stage-muted" aria-hidden>
            ✕
          </Text>
        </Pressable>
      ))}
    </View>
  );
}
