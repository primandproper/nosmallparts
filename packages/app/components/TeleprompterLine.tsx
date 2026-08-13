import { useEffect, useRef, useState } from 'react';
import { Pressable, ScrollView, Text, View } from 'react-native';
import type { GestureResponderEvent, LayoutChangeEvent } from 'react-native';

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

export function TeleprompterLine({
  text,
  direction,
  wpm,
  autoScroll,
  startDelaySec = 5,
  variant = 'card',
}: Props) {
  const scrollRef = useRef<ScrollView>(null);
  const [contentHeight, setContentHeight] = useState(0);
  const [viewportHeight, setViewportHeight] = useState(0);
  const [trackWidth, setTrackWidth] = useState(0);
  const [scrollY, setScrollY] = useState(0);
  const [paused, setPaused] = useState(false);
  const [speed, setSpeed] = useState<number>(1);
  const [delayRemaining, setDelayRemaining] = useState(autoScroll ? startDelaySec : 0);

  const maxScroll = Math.max(0, contentHeight - viewportHeight);
  const wordCount = text.trim().split(/\s+/).filter(Boolean).length;
  const baseDurationSec = Math.max(1, (wordCount / Math.max(60, wpm)) * 60);
  const pxPerSec = maxScroll / baseDurationSec;

  // Reset all state when the line text changes.
  useEffect(() => {
    setScrollY(0);
    setPaused(false);
    setSpeed(1);
    setDelayRemaining(autoScroll ? startDelaySec : 0);
    scrollRef.current?.scrollTo({ y: 0, animated: false });
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
    let last = typeof performance !== 'undefined' ? performance.now() : Date.now();
    const step = (now: number) => {
      const dt = (now - last) / 1000;
      last = now;
      setScrollY((prev) => {
        if (prev >= maxScroll) return prev;
        const next = Math.min(maxScroll, prev + pxPerSec * speed * dt);
        scrollRef.current?.scrollTo({ y: next, animated: false });
        return next;
      });
      raf = requestAnimationFrame(step);
    };
    raf = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf);
  }, [autoScroll, paused, delayRemaining, maxScroll, pxPerSec, speed]);

  const seekBy = (deltaSec: number) => {
    if (maxScroll <= 0) return;
    const delta = pxPerSec * deltaSec;
    setScrollY((prev) => {
      const next = Math.max(0, Math.min(maxScroll, prev + delta));
      scrollRef.current?.scrollTo({ y: next, animated: false });
      return next;
    });
  };

  const seekToFraction = (frac: number) => {
    if (maxScroll <= 0) return;
    const next = Math.max(0, Math.min(maxScroll, maxScroll * frac));
    setScrollY(next);
    scrollRef.current?.scrollTo({ y: next, animated: false });
  };

  const onTrackPress = (e: GestureResponderEvent) => {
    if (trackWidth <= 0) return;
    const x = e.nativeEvent.locationX;
    seekToFraction(x / trackWidth);
  };

  const onTrackLayout = (e: LayoutChangeEvent) => setTrackWidth(e.nativeEvent.layout.width);

  const progress = maxScroll > 0 ? scrollY / maxScroll : autoScroll ? 0 : 1;
  const showControls = autoScroll && maxScroll > 0;
  const isWaiting = autoScroll && delayRemaining > 0;

  const isOverlay = variant === 'overlay';
  // Roughly 4 lines visible at a time — autoscroll handles the rest.
  const scrollViewClass = isOverlay
    ? 'max-h-[14vh]'
    : 'max-h-[18vh] rounded-md border border-stage-border bg-stage-surface px-3 py-2';
  const lineTextClass = isOverlay
    ? 'text-stage-gold leading-snug text-base md:text-lg xl:text-xl 2xl:text-2xl'
    : 'text-stage-gold leading-snug text-sm xl:text-base 2xl:text-lg';

  return (
    <View>
      {direction ? (
        <Text className="italic text-stage-muted text-xs mb-2">({direction})</Text>
      ) : null}
      <ScrollView
        ref={scrollRef}
        className={scrollViewClass}
        showsVerticalScrollIndicator
        persistentScrollbar
        onContentSizeChange={(_, h) => setContentHeight(h)}
        onLayout={(e) => setViewportHeight(e.nativeEvent.layout.height)}
        onScroll={(e) => setScrollY(e.nativeEvent.contentOffset.y)}
        scrollEventThrottle={32}
        // RN-web honors overflowY; native ignores unknown style keys.
        style={{ overflowY: 'scroll' } as object}
      >
        <Text className={lineTextClass}>{text}</Text>
      </ScrollView>

      {showControls ? (
        <View className="mt-2">
          <Pressable
            onPress={onTrackPress}
            onLayout={onTrackLayout}
            accessibilityRole="adjustable"
            accessibilityLabel="Seek line position"
            className="h-1.5 w-full rounded-full bg-stage-border"
          >
            <View
              className="h-full rounded-full bg-stage-gold"
              style={{ width: `${Math.round(progress * 100)}%` }}
            />
          </Pressable>

          <View className="mt-2 flex-row items-center justify-between">
            <View className="flex-row items-center gap-2">
              <Pressable
                onPress={() => seekBy(-SEEK_STEP_SEC)}
                accessibilityRole="button"
                accessibilityLabel="Back 10 seconds"
                className="rounded border border-stage-border px-2 py-1"
              >
                <Text className="text-stage-text text-xs">⟲ 10s</Text>
              </Pressable>
              <Pressable
                onPress={() => setPaused((p) => !p)}
                disabled={isWaiting}
                accessibilityRole="button"
                accessibilityLabel={paused ? 'Resume autoscroll' : 'Pause autoscroll'}
                className={`rounded border px-3 py-1 ${
                  isWaiting
                    ? 'border-stage-border'
                    : 'border-stage-gold bg-stage-goldDim/30'
                }`}
              >
                <Text
                  className={`text-xs font-semibold ${
                    isWaiting ? 'text-stage-muted' : 'text-stage-text'
                  }`}
                >
                  {isWaiting
                    ? `Auto-scroll in ${delayRemaining}s`
                    : paused
                      ? '▶ Play'
                      : '⏸ Pause'}
                </Text>
              </Pressable>
              <Pressable
                onPress={() => seekBy(SEEK_STEP_SEC)}
                accessibilityRole="button"
                accessibilityLabel="Forward 10 seconds"
                className="rounded border border-stage-border px-2 py-1"
              >
                <Text className="text-stage-text text-xs">⟳ 10s</Text>
              </Pressable>
            </View>

            <View className="flex-row items-center gap-1">
              {SPEEDS.map((s) => (
                <Pressable
                  key={s}
                  onPress={() => setSpeed(s)}
                  accessibilityRole="button"
                  accessibilityLabel={`Speed ${s}x`}
                  className={`rounded border px-2 py-1 ${
                    speed === s
                      ? 'border-stage-gold bg-stage-goldDim/30'
                      : 'border-stage-border'
                  }`}
                >
                  <Text
                    className={`text-xs ${
                      speed === s ? 'text-stage-gold font-semibold' : 'text-stage-muted'
                    }`}
                  >
                    {s}×
                  </Text>
                </Pressable>
              ))}
            </View>
          </View>
        </View>
      ) : null}
    </View>
  );
}
