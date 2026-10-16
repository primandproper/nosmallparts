import { type MouseEvent, useEffect, useLayoutEffect, useRef, useState } from 'react';

interface Props {
  text: string;
  direction?: string;
  wpm: number;
  autoScroll: boolean;
  startDelaySec?: number;
  variant?: 'card' | 'overlay';
}

const SPEEDS = [0.75, 1, 1.25] as const;
const SEEK_STEP_SEC = 10;

export function TeleprompterLine({ text, direction, wpm, autoScroll, startDelaySec = 5, variant = 'card' }: Props) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const [contentHeight, setContentHeight] = useState(0);
  const [viewportHeight, setViewportHeight] = useState(0);
  const [scrollY, setScrollY] = useState(0);
  const [paused, setPaused] = useState(false);
  const [speed, setSpeed] = useState<number>(1);
  const [delayRemaining, setDelayRemaining] = useState(autoScroll ? startDelaySec : 0);

  const maxScroll = Math.max(0, contentHeight - viewportHeight);
  const wordCount = text.trim().split(/\s+/).filter(Boolean).length;
  const baseDurationSec = Math.max(1, (wordCount / Math.max(60, wpm)) * 60);
  const pxPerSec = maxScroll / baseDurationSec;

  // Measure the box and its text, and again whenever either resizes.
  useLayoutEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    const measure = () => {
      setContentHeight(el.scrollHeight);
      setViewportHeight(el.clientHeight);
    };
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(el);
    if (el.firstElementChild) observer.observe(el.firstElementChild);
    return () => observer.disconnect();
  }, [text]);

  const scrollTo = (y: number) => {
    if (scrollRef.current) scrollRef.current.scrollTop = y;
  };

  // Reset everything when the line changes.
  useEffect(() => {
    setScrollY(0);
    setPaused(false);
    setSpeed(1);
    setDelayRemaining(autoScroll ? startDelaySec : 0);
    scrollTo(0);
  }, [text, autoScroll, startDelaySec]);

  // Countdown before autoscroll begins.
  useEffect(() => {
    if (!autoScroll || delayRemaining <= 0) return;
    const t = setTimeout(() => setDelayRemaining((d) => Math.max(0, d - 1)), 1000);
    return () => clearTimeout(t);
  }, [autoScroll, delayRemaining]);

  // Animation loop.
  useEffect(() => {
    if (!autoScroll || paused || delayRemaining > 0 || maxScroll <= 0) return;
    let raf = 0;
    let last = performance.now();
    const step = (now: number) => {
      const dt = (now - last) / 1000;
      last = now;
      setScrollY((prev) => {
        if (prev >= maxScroll) return prev;
        const next = Math.min(maxScroll, prev + pxPerSec * speed * dt);
        scrollTo(next);
        return next;
      });
      raf = requestAnimationFrame(step);
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [autoScroll, paused, delayRemaining, maxScroll, pxPerSec, speed]);

  const seekBy = (deltaSec: number) => {
    if (maxScroll <= 0) return;
    setScrollY((prev) => {
      const next = Math.max(0, Math.min(maxScroll, prev + pxPerSec * deltaSec));
      scrollTo(next);
      return next;
    });
  };

  const onTrackClick = (e: MouseEvent<HTMLButtonElement>) => {
    if (maxScroll <= 0) return;
    const rect = e.currentTarget.getBoundingClientRect();
    if (rect.width <= 0) return;
    const next = Math.max(0, Math.min(maxScroll, (maxScroll * (e.clientX - rect.left)) / rect.width));
    setScrollY(next);
    scrollTo(next);
  };

  const progress = maxScroll > 0 ? scrollY / maxScroll : autoScroll ? 0 : 1;
  const showControls = autoScroll && maxScroll > 0;
  const isWaiting = autoScroll && delayRemaining > 0;

  const isOverlay = variant === 'overlay';
  // About four lines show at a time; autoscroll handles the rest.
  const boxClass = isOverlay
    ? 'max-h-[14vh]'
    : 'max-h-[18vh] rounded-md border border-stage-border bg-stage-surface px-3 py-2';
  const lineTextClass = isOverlay
    ? 'text-stage-gold leading-snug text-base md:text-lg xl:text-xl 2xl:text-2xl'
    : 'text-stage-gold leading-snug text-sm xl:text-base 2xl:text-lg';

  return (
    <div>
      {direction ? <p className="mb-2 text-xs italic text-stage-muted">({direction})</p> : null}
      <div
        ref={scrollRef}
        className={`overflow-y-scroll ${boxClass}`}
        onScroll={(e) => setScrollY(e.currentTarget.scrollTop)}
      >
        <p className={lineTextClass}>{text}</p>
      </div>

      {showControls ? (
        <div className="mt-2">
          <button
            type="button"
            onClick={onTrackClick}
            aria-label="Seek line position"
            className="block h-1.5 w-full rounded-full bg-stage-border"
          >
            <span
              className="block h-full rounded-full bg-stage-gold"
              style={{ width: `${Math.round(progress * 100)}%` }}
            />
          </button>

          <div className="mt-2 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <button
                type="button"
                onClick={() => seekBy(-SEEK_STEP_SEC)}
                aria-label="Back 10 seconds"
                className="rounded border border-stage-border px-2 py-1 text-xs text-stage-text"
              >
                ⟲ 10s
              </button>
              <button
                type="button"
                onClick={() => setPaused((p) => !p)}
                disabled={isWaiting}
                aria-label={paused ? 'Resume autoscroll' : 'Pause autoscroll'}
                className={`rounded border px-3 py-1 text-xs font-semibold ${
                  isWaiting
                    ? 'border-stage-border text-stage-muted'
                    : 'border-stage-gold bg-stage-goldDim/30 text-stage-text'
                }`}
              >
                {isWaiting ? `Auto-scroll in ${delayRemaining}s` : paused ? '▶ Play' : '⏸ Pause'}
              </button>
              <button
                type="button"
                onClick={() => seekBy(SEEK_STEP_SEC)}
                aria-label="Forward 10 seconds"
                className="rounded border border-stage-border px-2 py-1 text-xs text-stage-text"
              >
                ⟳ 10s
              </button>
            </div>

            <div className="flex items-center gap-1">
              {SPEEDS.map((s) => (
                <button
                  type="button"
                  key={s}
                  onClick={() => setSpeed(s)}
                  aria-label={`Speed ${s}x`}
                  aria-pressed={speed === s}
                  className={`rounded border px-2 py-1 text-xs ${
                    speed === s
                      ? 'border-stage-gold bg-stage-goldDim/30 font-semibold text-stage-gold'
                      : 'border-stage-border text-stage-muted'
                  }`}
                >
                  {s}×
                </button>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
