'use client';

import { useState, useEffect } from 'react';
import { useRouter, useParams } from 'next/navigation';
import { usePlayState } from '@/hooks/usePlayState';
import { useLiveKit } from '@/hooks/useLiveKit';
import { Lobby } from '@/components/Lobby';
import { VideoRoom } from '@/components/VideoRoom';
import { Stage } from '@/components/Stage';
import { Countdown } from '@/components/Countdown';
import { Discussion } from '@/components/Discussion';
import { Chat } from '@/components/Chat';
import { AdBanner } from '@/components/AdBanner';

export default function StagePage() {
  const params = useParams();
  const stageId = params.stageId as string;
  const router = useRouter();
  const [playerName, setPlayerName] = useState<string | null>(null);
  const [isReady, setIsReady] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const name = sessionStorage.getItem('playerName');
    if (!name) {
      router.push('/');
      return;
    }
    setPlayerName(name);
    setIsReady(true);
  }, [router]);

  const {
    stageState,
    participantId,
    availablePlays,
    chatMessages,
    isConnected,
    error,
    toggleCastMember,
    setDirector,
    setAudience,
    setReady,
    selectPlay,
    sendChat,
    startPlay,
    nextLine,
    resetPlay,
  } = usePlayState({
    stageId: isReady ? stageId : '',
    participantName: playerName || '',
  });

  const { token: livekitToken, url: livekitUrl } = useLiveKit({
    stageId,
    participantId,
    participantName: playerName || '',
  });

  // Use the short code from stageState when available, fall back to URL param
  const displayCode = stageState?.code || stageId;

  const handleCopyStageCode = async () => {
    const shareableUrl = `${window.location.origin}/?stage=${displayCode}`;
    try {
      await navigator.clipboard.writeText(shareableUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  };

  if (!isReady || !playerName) {
    return (
      <main className="min-h-screen stage-bg flex items-center justify-center">
        <div className="text-center">
          <div className="text-4xl mb-4 animate-pulse">🎭</div>
          <p className="text-[var(--gold)] text-xl">Preparing the stage...</p>
        </div>
      </main>
    );
  }

  if (error) {
    return (
      <main className="min-h-screen stage-bg flex items-center justify-center">
        <div className="card-elegant rounded-2xl p-8 text-center max-w-md">
          <div className="text-4xl mb-4">😔</div>
          <h2 className="text-[var(--gold)] text-xl mb-2">Connection Error</h2>
          <p className="text-[var(--cream)] opacity-60">{error}</p>
          <button
            onClick={() => router.push('/')}
            className="btn-secondary px-6 py-2 rounded-lg mt-6"
          >
            Return Home
          </button>
        </div>
      </main>
    );
  }

  if (!isConnected || !stageState) {
    console.log('[Stage] Rendering "Connecting to stage"', {
      isConnected,
      hasStageState: !!stageState,
      timestamp: new Date().toISOString(),
    });
    return (
      <main className="min-h-screen stage-bg flex items-center justify-center">
        <div className="text-center">
          <div className="text-4xl mb-4 animate-pulse">🎭</div>
          <p className="text-[var(--gold)] text-xl">Connecting to stage...</p>
          <p className="text-[var(--cream)] opacity-40 text-sm mt-2">Stage: {displayCode}</p>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen stage-bg">
      {/* Header */}
      <header className="border-b border-[var(--gold)]/10 bg-[var(--velvet-deep)]/50 backdrop-blur-sm sticky top-0 z-50">
        <div className="container mx-auto px-4 py-4">
          <div className="flex justify-between items-center">
            {/* Logo */}
            <div className="flex items-center gap-3">
              <span className="text-2xl">🎭</span>
              <h1 className="font-[Cormorant_Garamond] text-2xl font-bold text-gold-gradient hidden sm:block">
                No Small Parts
              </h1>
            </div>

            {/* Stage Info */}
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <span className="text-[var(--cream)] opacity-40 text-sm hidden sm:inline">
                  Stage:
                </span>
                <code className="bg-[var(--charcoal)] px-3 py-1.5 rounded-lg text-[var(--gold)] font-mono tracking-wider text-sm">
                  {displayCode}
                </code>
                <button
                  onClick={handleCopyStageCode}
                  className="btn-secondary px-3 py-1.5 rounded-lg text-sm flex items-center gap-2"
                >
                  {copied ? (
                    <>
                      <span>✓</span>
                      <span className="hidden sm:inline">Copied!</span>
                    </>
                  ) : (
                    <>
                      <span>📋</span>
                      <span className="hidden sm:inline">Copy Link</span>
                    </>
                  )}
                </button>
              </div>

              {/* User */}
              <div className="flex items-center gap-2 pl-4 border-l border-[var(--gold)]/20">
                <div className="w-8 h-8 rounded-full bg-[var(--gold)]/20 flex items-center justify-center text-[var(--gold)] text-sm font-medium">
                  {playerName.charAt(0).toUpperCase()}
                </div>
                <span className="text-[var(--cream)] text-sm hidden sm:inline">{playerName}</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="container mx-auto px-4 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Video Area - on mobile: order-2 in lobby, order-1 during performance */}
          <div
            className={`lg:col-span-2 lg:order-1 ${
              stageState.phase === 'lobby' ? 'order-2' : 'order-1'
            }`}
          >
            <VideoRoom
              stageState={stageState}
              participantId={participantId}
              livekitToken={livekitToken}
              livekitUrl={livekitUrl}
            />
          </div>

          {/* Side Panel - on mobile: order-1 in lobby, order-2 during performance */}
          <div
            className={`space-y-4 lg:order-2 ${
              stageState.phase === 'lobby' ? 'order-1' : 'order-2'
            }`}
          >
            {stageState.phase === 'lobby' && (
              <Lobby
                stageState={stageState}
                participantId={participantId}
                availablePlays={availablePlays}
                onToggleCastMember={toggleCastMember}
                onSetDirector={setDirector}
                onSetAudience={setAudience}
                onSetReady={setReady}
                onSelectPlay={selectPlay}
                onStartPlay={startPlay}
              />
            )}

            {stageState.phase === 'countdown' && <Countdown onComplete={() => {}} />}

            {stageState.phase === 'performing' && (
              <Stage stageState={stageState} participantId={participantId} onNextLine={nextLine} />
            )}

            {stageState.phase === 'discussion' && (
              <Discussion stageState={stageState} onResetPlay={resetPlay} />
            )}

            {/* Chat - always visible */}
            <Chat
              messages={chatMessages}
              currentParticipant={
                (stageState.participants || []).find((p) => p.id === participantId) || null
              }
              onSendMessage={sendChat}
            />

            {/* Ad Banner */}
            <AdBanner slot="REPLACEME" />
          </div>
        </div>
      </div>
    </main>
  );
}
