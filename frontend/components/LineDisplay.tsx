'use client';

import { useState, useEffect, useRef } from 'react';
import { PlayLine, Play, getCastMemberNameForLine } from '@/lib/types';

interface LineDisplayProps {
  line: PlayLine;
  play: Play;
  isCurrentLine: boolean;
  isPreview?: boolean;
}

const LONG_LINE_THRESHOLD = 100; // characters
const DEFAULT_WPM = 150;

export function LineDisplay({ line, play, isCurrentLine, isPreview = false }: LineDisplayProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [scrollSpeed, setScrollSpeed] = useState(0); // pixels per frame, negative = up, positive = down
  const [isPaused, setIsPaused] = useState(true);
  const animationRef = useRef<number | null>(null);

  const isLongLine = line.text.length > LONG_LINE_THRESHOLD;
  const hasScrollMode = !!line.scrollMode || isLongLine;
  const wordCount = line.text.split(/\s+/).length;
  const wpm = line.autoScrollWpm || DEFAULT_WPM;
  const castMemberName = getCastMemberNameForLine(play, line);

  // Calculate total scroll duration for auto mode
  const scrollDurationMs = (wordCount / wpm) * 60 * 1000;

  // Continuous scroll animation
  useEffect(() => {
    if (isPaused || scrollSpeed === 0 || !containerRef.current) {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
        animationRef.current = null;
      }
      return;
    }

    const animate = () => {
      if (!containerRef.current) return;

      const newScrollTop = containerRef.current.scrollTop + scrollSpeed;
      const maxScroll = containerRef.current.scrollHeight - containerRef.current.clientHeight;

      // Clamp scroll position
      containerRef.current.scrollTop = Math.max(0, Math.min(newScrollTop, maxScroll));

      animationRef.current = requestAnimationFrame(animate);
    };

    animationRef.current = requestAnimationFrame(animate);

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current);
      }
    };
  }, [isPaused, scrollSpeed]);

  // Auto-start scrolling for auto mode OR very long lines
  const VERY_LONG_THRESHOLD = 50; // words - auto-start for lines longer than this
  const isVeryLong = wordCount > VERY_LONG_THRESHOLD;
  const shouldAutoStart = line.scrollMode === 'auto' || isVeryLong;

  useEffect(() => {
    if (isCurrentLine && shouldAutoStart && hasScrollMode) {
      // Delay before auto-starting (gives performer time to read the beginning)
      const delay = setTimeout(
        () => {
          // Calculate scroll speed based on WPM
          const baseSpeed = wpm / 60 / 2; // Adjust divisor to tune speed
          setScrollSpeed(baseSpeed);
          setIsPaused(false);
        },
        isVeryLong && line.scrollMode !== 'auto' ? 3000 : 1000
      ); // 3s delay for auto-detected long lines, 1s for explicit auto

      return () => clearTimeout(delay);
    } else {
      setIsPaused(true);
      setScrollSpeed(0);
    }
  }, [isCurrentLine, shouldAutoStart, hasScrollMode, wpm, isVeryLong, line.scrollMode]);

  // Reset scroll position when line changes
  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = 0;
    }
    setIsPaused(true);
    setScrollSpeed(0);
  }, [line.text]);

  const adjustSpeed = (delta: number) => {
    setScrollSpeed((prev) => prev + delta);
    if (isPaused) setIsPaused(false);
  };

  const togglePause = () => {
    setIsPaused(!isPaused);
  };

  // Preview mode - show abbreviated version
  if (isPreview) {
    return (
      <div className="space-y-1">
        {line.direction && (
          <div className="text-[var(--gold)] text-xs italic opacity-80">({line.direction})</div>
        )}
        <div className="text-[var(--cream)] text-sm opacity-70 line-clamp-2">
          &quot;{line.text}&quot;
        </div>
        {isLongLine && (
          <div className="text-[var(--cream)] opacity-40 text-xs">
            {wordCount} words •{' '}
            {line.scrollMode === 'auto'
              ? `Auto-scroll ${wpm} WPM`
              : line.scrollMode === 'manual'
                ? 'Manual scroll'
                : 'Long line'}
          </div>
        )}
      </div>
    );
  }

  // Full display mode for current line
  return (
    <div className="space-y-3">
      {/* Direction badge */}
      {line.direction && (
        <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-[var(--gold)]/10 border border-[var(--gold)]/30">
          <span className="text-lg">🎭</span>
          <span className="text-[var(--gold)] text-sm italic font-medium">{line.direction}</span>
        </div>
      )}

      {/* Line text - scrollable container for long lines */}
      {hasScrollMode ? (
        <div className="relative">
          {/* Scrollable text container */}
          <div
            ref={containerRef}
            className="max-h-48 overflow-y-auto rounded-xl bg-[var(--charcoal)]/30 p-4 scroll-smooth"
          >
            <div className="text-[var(--cream)] text-2xl leading-relaxed font-[Cormorant_Garamond] italic">
              &quot;{line.text}&quot;
            </div>
          </div>

          {/* Scroll speed controls - always visible */}
          <div className="mt-3 p-3 rounded-xl bg-[var(--charcoal)]/50 border border-[var(--cream)]/10">
            <div className="flex items-center justify-between gap-2">
              {/* Scroll up / slower */}
              <button
                onClick={() => adjustSpeed(-0.5)}
                className="flex-1 px-3 py-2.5 rounded-lg bg-[var(--gold)]/20 text-[var(--gold)] hover:bg-[var(--gold)]/30 transition-colors font-medium text-sm"
              >
                ↑ Slower
              </button>

              {/* Pause/Play */}
              <button
                onClick={togglePause}
                className={`px-4 py-2.5 rounded-lg font-medium text-sm transition-colors ${
                  isPaused
                    ? 'bg-green-500/20 text-green-400 hover:bg-green-500/30'
                    : 'bg-red-500/20 text-red-400 hover:bg-red-500/30'
                }`}
              >
                {isPaused ? '▶ Play' : '⏸ Pause'}
              </button>

              {/* Scroll down / faster */}
              <button
                onClick={() => adjustSpeed(0.5)}
                className="flex-1 px-3 py-2.5 rounded-lg bg-[var(--gold)]/20 text-[var(--gold)] hover:bg-[var(--gold)]/30 transition-colors font-medium text-sm"
              >
                ↓ Faster
              </button>
            </div>

            {/* Speed indicator */}
            <div className="text-center text-[var(--cream)] opacity-40 text-xs mt-2">
              {wordCount} words • Speed:{' '}
              {scrollSpeed > 0 ? `+${scrollSpeed.toFixed(1)}` : scrollSpeed.toFixed(1)}
            </div>
          </div>
        </div>
      ) : (
        <div className="text-[var(--cream)] text-2xl leading-relaxed font-[Cormorant_Garamond] italic">
          &quot;{line.text}&quot;
        </div>
      )}
    </div>
  );
}

// Simplified version just for showing direction in line previews
export function DirectionBadge({ direction }: { direction?: string }) {
  if (!direction) return null;

  return <span className="text-[var(--gold)] text-xs italic opacity-80 mr-1">({direction})</span>;
}
