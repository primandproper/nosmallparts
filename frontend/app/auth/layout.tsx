'use client';

import { useState, useEffect } from 'react';
import Link from 'next/link';

function Star({ style }: { style: React.CSSProperties }) {
  return <div className="star" style={style} />;
}

function Stars() {
  const [stars, setStars] = useState<Array<{ id: number; style: React.CSSProperties }>>([]);

  useEffect(() => {
    const newStars = Array.from({ length: 50 }, (_, i) => ({
      id: i,
      style: {
        left: `${Math.random() * 100}%`,
        top: `${Math.random() * 100}%`,
        animationDelay: `${Math.random() * 3}s`,
        animationDuration: `${2 + Math.random() * 3}s`,
      },
    }));
    setStars(newStars);
  }, []);

  return (
    <>
      {stars.map((star) => (
        <Star key={star.id} style={star.style} />
      ))}
    </>
  );
}

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="min-h-screen stage-bg relative overflow-hidden">
      {/* Decorative elements */}
      <div className="curtain-left" />
      <div className="curtain-right" />
      <Stars />

      {/* Main content */}
      <div className="relative z-20 min-h-screen flex flex-col items-center justify-center px-6 py-12">
        <div className="spotlight w-full max-w-md">
          {/* Logo/Title */}
          <div className="text-center mb-8 animate-fade-up">
            <Link href="/" className="inline-block">
              <div className="mb-4">
                <span className="text-5xl">🎭</span>
              </div>
              <h1 className="font-[Cormorant_Garamond] text-4xl md:text-5xl font-bold text-gold-gradient mb-2 tracking-wide">
                No Small Parts
              </h1>
            </Link>
            <div className="decorative-line w-32 mx-auto" />
          </div>

          {children}
        </div>
      </div>
    </main>
  );
}
