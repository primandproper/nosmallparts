import type { Toast } from '../hooks/usePlayState';

interface Props {
  toasts: Toast[];
  onDismiss: (id: number) => void;
}

/**
 * Top-center stack of transient notices. Each toast auto-dismisses (see
 * usePlayState) and can also be clicked to dismiss at once; the stack is capped
 * so notices never pile up (§11).
 */
export function ToastHost({ toasts, onDismiss }: Props) {
  if (toasts.length === 0) return null;
  return (
    <div className="pointer-events-none fixed inset-x-0 top-3 z-30 flex flex-col items-center gap-2 px-4">
      {toasts.map((t) => (
        <button
          type="button"
          key={t.id}
          onClick={() => onDismiss(t.id)}
          aria-label={`Dismiss notice: ${t.message}`}
          className="pointer-events-auto flex w-full max-w-md items-start justify-between gap-3 rounded-md border border-stage-border bg-stage-surface px-4 py-3 text-left shadow-lg active:opacity-80"
        >
          <span className="flex-1 text-sm text-stage-text">{t.message}</span>
          <span className="text-base leading-none text-stage-muted" aria-hidden>
            ✕
          </span>
        </button>
      ))}
    </div>
  );
}
