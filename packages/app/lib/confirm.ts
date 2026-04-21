import { Alert, Platform } from 'react-native';

/**
 * Cross-platform confirm dialog. Resolves true if the user confirms.
 *
 * - Web: uses window.confirm (synchronous; resolved immediately).
 * - Native: uses Alert.alert with a destructive button.
 */
export function confirm(opts: {
  title: string;
  message?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  destructive?: boolean;
}): Promise<boolean> {
  const { title, message, confirmLabel = 'Confirm', cancelLabel = 'Cancel', destructive } = opts;

  if (Platform.OS === 'web') {
    const text = message ? `${title}\n\n${message}` : title;
    return Promise.resolve(typeof window !== 'undefined' && window.confirm(text));
  }

  return new Promise((resolve) => {
    Alert.alert(title, message, [
      { text: cancelLabel, style: 'cancel', onPress: () => resolve(false) },
      {
        text: confirmLabel,
        style: destructive ? 'destructive' : 'default',
        onPress: () => resolve(true),
      },
    ]);
  });
}
