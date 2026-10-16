import { useEffect, useState } from 'react';

export function Countdown({ seconds = 3 }: { seconds?: number }) {
  const [remaining, setRemaining] = useState(seconds);
  useEffect(() => {
    if (remaining <= 0) return;
    const t = setTimeout(() => setRemaining((r) => r - 1), 1000);
    return () => clearTimeout(t);
  }, [remaining]);
  return (
    <div className="flex flex-1 flex-col items-center justify-center bg-stage-bg">
      <p className="text-stage-muted">Get ready…</p>
      <p className="mt-3 font-serif text-9xl text-stage-gold">{remaining > 0 ? remaining : '🎭'}</p>
    </div>
  );
}
