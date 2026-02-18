'use client';

import { useState, useEffect } from 'react';

interface CountdownProps {
  onComplete: () => void;
}

export function Countdown({ onComplete }: CountdownProps) {
  const [count, setCount] = useState(3);

  useEffect(() => {
    if (count === 0) {
      onComplete();
      return;
    }

    const timer = setTimeout(() => {
      setCount(count - 1);
    }, 1000);

    return () => clearTimeout(timer);
  }, [count, onComplete]);

  return (
    <div className="card-elegant rounded-2xl p-8 text-center">
      <h2 className="font-[Cormorant_Garamond] text-2xl text-gold-gradient mb-6">
        The Performance Begins...
      </h2>

      <div className="relative">
        {/* Glow effect */}
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="w-32 h-32 rounded-full bg-[var(--gold)]/20 blur-2xl animate-pulse" />
        </div>

        {/* Number */}
        <div className="relative countdown-number text-8xl font-[Cormorant_Garamond] font-bold text-gold-gradient">
          {count > 0 ? count : '🎭'}
        </div>
      </div>

      <p className="text-[var(--cream)] opacity-40 mt-6 text-lg">
        {count > 0 ? 'Places, everyone!' : 'Action!'}
      </p>
    </div>
  );
}
