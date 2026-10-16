import { type FormEvent, lazy, Suspense, useEffect, useState } from 'react';
import { Link, useParams } from 'react-router';

import { ChatPanel } from '../components/ChatPanel';
import { Countdown } from '../components/Countdown';
import { Discussion } from '../components/Discussion';
import { HostControls } from '../components/HostControls';
import { Lobby } from '../components/Lobby';
import { RoleOfferPrompt } from '../components/RoleOfferPrompt';
import { inputClass, Screen, Spinner } from '../components/Screen';
import { Stage } from '../components/Stage';
import { ToastHost } from '../components/Toasts';
import { VacancyBanner } from '../components/VacancyBanner';
import { StagePhase, type StageState } from '../generated/nosmallparts/stage/v1/stage';
import { usePlayState } from '../hooks/usePlayState';
import { getStoredPlayerName, storePlayerName } from '../lib/player-name';

// The server's display-name policy refusals (§12.4) send the user back to the
// name prompt, rather than reconnecting over and over with a refused name.
const NAME_ERRORS = ['Please enter a name.', 'Please choose a different name.', 'Name is too long'];

// LiveKit is most of the bundle and only a performance needs it.
const VideoRoom = lazy(() => import('../components/VideoRoom').then((m) => ({ default: m.VideoRoom })));

function titleFor(state: StageState): string {
  switch (state.phase) {
    case StagePhase.STAGE_PHASE_LOBBY:
      return `Lobby · ${state.code}`;
    case StagePhase.STAGE_PHASE_PERFORMING:
      return state.selectedPlay?.title || 'Performing';
    case StagePhase.STAGE_PHASE_COUNTDOWN:
      return 'Starting…';
    default:
      return 'Discussion';
  }
}

export default function StageScreen() {
  const { stageId = '' } = useParams();
  const [playerName, setPlayerName] = useState<string | null>(() => getStoredPlayerName());
  const [nameInput, setNameInput] = useState('');
  const [nameError, setNameError] = useState<string | null>(null);
  const [chatOpen, setChatOpen] = useState(false);

  const play = usePlayState({
    stageIdOrCode: stageId,
    participantName: playerName ?? '',
    enabled: !!playerName,
  });
  const { stageState, participantId, error, isConnected, wasReplaced, reconnect } = play;

  // A name the server's policy refuses sends the user back to the prompt.
  useEffect(() => {
    if (playerName && error && NAME_ERRORS.some((e) => error.startsWith(e))) {
      setNameInput(playerName);
      setPlayerName(null);
      setNameError(error);
    }
  }, [error, playerName]);

  function onJoin(event: FormEvent) {
    event.preventDefault();
    const trimmed = nameInput.trim();
    if (!trimmed) {
      setNameError('Enter your name to join');
      return;
    }
    setNameError(null);
    // Clear a stale connection error (a name refused before, say) so the
    // name-policy effect does not bounce straight back here.
    reconnect();
    storePlayerName(trimmed);
    setPlayerName(trimmed);
  }

  if (!playerName) {
    return (
      <Screen title="Join Stage">
        <div className="flex flex-1 flex-col items-center justify-center px-6">
          <h2 className="mb-6 font-serif text-2xl text-stage-gold">Enter your name to join</h2>
          <form onSubmit={onJoin} className="flex w-full max-w-sm flex-col gap-3">
            <input
              value={nameInput}
              onChange={(e) => setNameInput(e.target.value)}
              placeholder="Your name"
              autoFocus
              maxLength={32}
              className={inputClass}
            />
            {nameError ? <p className="text-sm text-red-400">{nameError}</p> : null}
            <button
              type="submit"
              className="rounded-md bg-stage-gold px-4 py-3 text-center font-semibold text-stage-bg"
            >
              Join
            </button>
            {/* Optional: guests join without an account (§5.3). */}
            <Link to="/login" className="self-center py-1 text-xs text-stage-muted underline active:opacity-70">
              Sign in (optional, needed to host or rate)
            </Link>
          </form>
        </div>
      </Screen>
    );
  }

  if (!stageState) {
    return (
      <Screen title="Stage">
        <div className="flex flex-1 flex-col items-center justify-center px-6">
          {wasReplaced ? null : <Spinner />}
          <p className="mt-3 text-center text-stage-muted">
            {wasReplaced
              ? 'This stage was opened in another window.'
              : isConnected
                ? 'Joining stage…'
                : (error ?? 'Connecting…')}
          </p>
          {wasReplaced ? (
            <button
              type="button"
              onClick={reconnect}
              className="mt-4 rounded-md border border-stage-gold px-4 py-2 font-semibold text-stage-gold active:opacity-70"
            >
              Take over here
            </button>
          ) : null}
        </div>
      </Screen>
    );
  }

  const me = stageState.participants.find((p) => p.id === participantId);
  const hostControls = (
    <div className="px-3 pt-3">
      <HostControls
        state={stageState}
        participantId={participantId}
        onPause={play.hostPause}
        onResume={play.hostResume}
        onBackLine={play.hostBackLine}
        onSkipLine={play.hostSkipLine}
        onEndPlay={play.hostEndPlay}
        onToggleDirectorMode={play.hostToggleDirectorMode}
        onReassignRole={play.hostReassignRole}
        onUnassignRole={play.hostUnassignRole}
        onMute={play.hostMute}
      />
    </div>
  );

  return (
    <Screen title={titleFor(stageState)}>
      <div className="relative flex min-h-0 flex-1 flex-col">
        <ConnectionBanner isConnected={isConnected} wasReplaced={wasReplaced} error={error} onReconnect={reconnect} />
        <RoleOfferPrompt offers={play.roleOffers} onRespond={play.respondRoleOffer} />
        <VacancyBanner vacancies={play.roleVacancies} isAudience={!!me?.isAudience} onClaim={play.audienceClaimRole} />
        {stageState.phase === StagePhase.STAGE_PHASE_LOBBY ? (
          <Lobby
            state={stageState}
            availablePlays={play.availablePlays}
            participantId={participantId}
            systemMessages={play.systemMessages}
            onSelectPlay={play.selectPlay}
            onToggleRole={play.toggleCastMember}
            onToggleAudience={play.setAudience}
            onToggleReady={play.setReady}
            onStart={play.startPlay}
            onHostPause={play.hostPause}
            onHostResume={play.hostResume}
            onHostBackLine={play.hostBackLine}
            onHostSkipLine={play.hostSkipLine}
            onHostEndPlay={play.hostEndPlay}
            onHostTransferHost={play.hostTransferHost}
            onHostKick={play.hostKick}
            onHostBan={play.hostBan}
            onHostMute={play.hostMute}
            onHostToggleDirectorMode={play.hostToggleDirectorMode}
            onHostReassignRole={play.hostReassignRole}
            onHostUnassignRole={play.hostUnassignRole}
          />
        ) : null}
        {stageState.phase === StagePhase.STAGE_PHASE_COUNTDOWN ? <Countdown /> : null}
        {stageState.phase === StagePhase.STAGE_PHASE_PERFORMING ? (
          <div className="flex min-h-0 flex-1 flex-col">
            {hostControls}
            <Stage
              state={stageState}
              participantId={participantId}
              onNextLine={play.nextLine}
              videoSlot={
                <Suspense fallback={<VideoPlaceholder />}>
                  <VideoRoom stageId={stageState.id} participantId={participantId} state={stageState} />
                </Suspense>
              }
            />
          </div>
        ) : null}
        {stageState.phase === StagePhase.STAGE_PHASE_DISCUSSION ? (
          <div className="flex flex-1 flex-col">
            {hostControls}
            <Discussion state={stageState} participantId={participantId} onReset={play.resetPlay} />
          </div>
        ) : null}

        {/* Chat: a floating toggle and a docked panel (§6). */}
        {chatOpen ? (
          <div className="fixed bottom-16 right-3 z-20 flex h-80 w-80 max-w-[calc(100vw-1.5rem)] flex-col overflow-hidden rounded-md border border-stage-border bg-stage-bg shadow-lg">
            <div className="flex items-center justify-between border-b border-stage-border bg-stage-surface px-3 py-2">
              <span className="text-xs uppercase tracking-widest text-stage-gold">Chat</span>
              <button type="button" onClick={() => setChatOpen(false)} className="text-xs text-stage-muted">
                Close
              </button>
            </div>
            <ChatPanel
              messages={play.chatMessages}
              systemMessages={play.systemMessages}
              isAudience={!!me?.isAudience}
              onSend={play.sendChat}
            />
          </div>
        ) : null}
        <button
          type="button"
          onClick={() => setChatOpen((v) => !v)}
          aria-label={chatOpen ? 'Close chat' : 'Open chat'}
          className="fixed bottom-3 right-3 z-20 rounded-full border border-stage-border bg-stage-surface px-4 py-2 text-xs font-semibold text-stage-text shadow-lg active:opacity-80"
        >
          {chatOpen ? '💬 Hide chat' : '💬 Chat'}
        </button>
        <ToastHost toasts={play.toasts} onDismiss={play.dismissToast} />
      </div>
    </Screen>
  );
}

function ConnectionBanner({
  isConnected,
  wasReplaced,
  error,
  onReconnect,
}: {
  isConnected: boolean;
  wasReplaced: boolean;
  error: string | null;
  onReconnect: () => void;
}) {
  if (isConnected && !wasReplaced) return null;

  return (
    <div className={`flex items-center justify-between gap-3 px-4 py-2 ${wasReplaced ? 'bg-red-900' : 'bg-amber-900'}`}>
      <p className="flex-1 text-xs text-white">
        {wasReplaced
          ? 'This stage was opened in another window. Controls here are inactive.'
          : (error ?? 'Connection lost. Reconnecting…')}
      </p>
      {wasReplaced ? (
        <button
          type="button"
          onClick={onReconnect}
          className="rounded border border-white/40 px-2 py-1 text-xs font-semibold text-white active:opacity-70"
        >
          Take over here
        </button>
      ) : null}
    </div>
  );
}

function VideoPlaceholder() {
  return (
    <div className="flex flex-1 items-center justify-center rounded-md border border-stage-border p-3">
      <p className="text-xs text-stage-muted">Connecting video…</p>
    </div>
  );
}
