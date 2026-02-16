'use client';

import { useState, useEffect, useRef, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { useAuth } from '@/hooks/useAuth';

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

function HomeContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isAuthenticated, isLoading: authLoading, logout } = useAuth();
  const [playerName, setPlayerName] = useState('');
  const [stageCode, setStageCode] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState('');
  const nameInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    const stageFromUrl = searchParams.get('stage') || searchParams.get('code');
    if (stageFromUrl) {
      setStageCode(stageFromUrl.toUpperCase());
    }
    nameInputRef.current?.focus();
  }, [searchParams]);

  const extractStageCode = (input: string): string => {
    try {
      const url = new URL(input);
      const stageParam = url.searchParams.get('stage') || url.searchParams.get('code');
      if (stageParam) {
        return stageParam.toUpperCase();
      }
    } catch {
      // Not a valid URL
    }
    return input.toUpperCase();
  };

  const handleStageCodeChange = (value: string) => {
    const extracted = extractStageCode(value);
    setStageCode(extracted);
    setError('');
  };

  const isValidStageCode = stageCode.length >= 6;
  const hasStageCode = stageCode.length > 0;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!playerName.trim()) {
      setError('Please enter your name');
      return;
    }

    if (hasStageCode && !isValidStageCode) {
      setError('Stage code must be at least 6 characters');
      return;
    }

    setIsCreating(true);
    setError('');

    try {
      sessionStorage.setItem('playerName', playerName);

      if (stageCode) {
        const response = await fetch(`/api/stages/${stageCode}/exists`);
        const data = await response.json();

        if (data.exists) {
          router.push(`/stage/${stageCode}`);
        } else {
          setError('Stage not found. Check the code and try again.');
          setIsCreating(false);
        }
      } else {
        const response = await fetch('/api/stages', {
          method: 'POST',
        });
        const data = await response.json();

        if (!response.ok || !data.code) {
          setError(data.error || 'Failed to create stage. Please try again.');
          setIsCreating(false);
          return;
        }

        router.push(`/stage/${data.code}`);
      }
    } catch (err) {
      setError('Failed to connect to server. Please try again.');
      setIsCreating(false);
    }
  };

  const buttonText = isCreating
    ? hasStageCode
      ? 'Entering...'
      : 'Creating Stage...'
    : hasStageCode
      ? 'Enter Stage'
      : 'Create New Stage';

  const handleLogout = async () => {
    await logout();
    router.refresh();
  };

  return (
    <main className="min-h-screen stage-bg relative overflow-hidden">
      {/* Decorative elements */}
      <div className="curtain-left" />
      <div className="curtain-right" />
      <Stars />

      {/* Auth Navigation */}
      <div className="absolute top-6 right-6 z-30 animate-fade-up">
        {!authLoading && (
          <div className="flex items-center gap-4">
            {isAuthenticated ? (
              <button
                onClick={handleLogout}
                className="text-[var(--cream)] opacity-60 hover:opacity-100 hover:text-[var(--gold)] transition-all text-sm"
              >
                Sign Out
              </button>
            ) : (
              <>
                <Link
                  href="/auth/login"
                  className="text-[var(--cream)] opacity-60 hover:opacity-100 hover:text-[var(--gold)] transition-all text-sm"
                >
                  Sign In
                </Link>
                <Link href="/auth/register" className="btn-secondary px-4 py-2 rounded-lg text-sm">
                  Create Account
                </Link>
              </>
            )}
          </div>
        )}
      </div>

      {/* Main content */}
      <div className="relative z-20 min-h-screen flex flex-col items-center justify-center px-6 py-12">
        <div className="spotlight w-full max-w-lg">
          {/* Logo/Title */}
          <div className="text-center mb-12 animate-fade-up">
            <div className="mb-4">
              <span className="text-6xl">🎭</span>
            </div>
            <h1 className="font-[Cormorant_Garamond] text-5xl md:text-6xl font-bold text-gold-gradient mb-4 tracking-wide">
              No Small Parts
            </h1>
            <div className="decorative-line w-48 mx-auto mb-4" />
            <p className="text-[var(--cream)] opacity-60 text-lg font-light tracking-wide">
              Take the stage. Perform together.
            </p>
          </div>

          {/* Form Card */}
          <div className="card-elegant rounded-2xl p-8 md:p-10 animate-fade-up animate-delay-1">
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Name Input */}
              <div className="animate-fade-up animate-delay-2">
                <label
                  htmlFor="name"
                  className="block text-[var(--gold)] text-sm font-medium mb-2 tracking-wide uppercase"
                >
                  Your Name
                </label>
                <input
                  ref={nameInputRef}
                  type="text"
                  id="name"
                  value={playerName}
                  onChange={(e) => setPlayerName(e.target.value)}
                  className="input-elegant w-full px-5 py-4 rounded-xl text-[var(--cream)] text-lg"
                  placeholder="Enter your stage name..."
                />
              </div>

              {/* Stage Code Input */}
              <div className="animate-fade-up animate-delay-3">
                <label
                  htmlFor="code"
                  className="block text-[var(--gold)] text-sm font-medium mb-2 tracking-wide uppercase"
                >
                  Stage Code
                  <span className="text-[var(--cream)] opacity-40 normal-case ml-2 tracking-normal">
                    (leave empty to create new)
                  </span>
                </label>
                <input
                  type="text"
                  id="code"
                  value={stageCode}
                  onChange={(e) => handleStageCodeChange(e.target.value)}
                  className={`input-elegant w-full px-5 py-4 rounded-xl text-[var(--cream)] text-lg font-mono tracking-widest uppercase ${
                    hasStageCode && !isValidStageCode ? 'border-red-500/50' : ''
                  }`}
                  placeholder="ABC123"
                />
                {hasStageCode && !isValidStageCode && (
                  <p className="text-red-400 text-sm mt-2 flex items-center gap-2">
                    <span>⚠</span> Stage code must be at least 6 characters
                  </p>
                )}
              </div>

              {/* Error Message */}
              {error && (
                <div className="bg-red-500/10 border border-red-500/30 rounded-xl px-4 py-3 text-red-400 text-sm text-center">
                  {error}
                </div>
              )}

              {/* Submit Button */}
              <div className="animate-fade-up animate-delay-4 pt-2">
                <button
                  type="submit"
                  disabled={isCreating || !playerName.trim() || (hasStageCode && !isValidStageCode)}
                  className="btn-primary w-full py-4 rounded-xl text-lg relative"
                >
                  <span>{buttonText}</span>
                </button>
              </div>
            </form>
          </div>

          {/* Scripts Library Link */}
          <div className="text-center mt-8 animate-fade-up animate-delay-4">
            <Link
              href="/scripts"
              className="text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors inline-flex items-center gap-2"
            >
              <span>📚</span>
              <span>Browse Scripts Library</span>
              <span>→</span>
            </Link>
          </div>

          {/* Footer text */}
          <p className="text-center text-[var(--cream)] opacity-40 text-sm mt-4 animate-fade-up animate-delay-4">
            Create a stage and share the code with your cast,
            <br />
            or enter an existing code to join the performance.
          </p>
        </div>
      </div>
    </main>
  );
}

export default function Home() {
  return (
    <Suspense
      fallback={
        <main className="min-h-screen stage-bg flex items-center justify-center">
          <div className="text-[var(--gold)] text-lg">Loading...</div>
        </main>
      }
    >
      <HomeContent />
    </Suspense>
  );
}
