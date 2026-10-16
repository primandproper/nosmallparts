import { useState } from 'react';
import { Link } from 'react-router';

import type { StageState } from '../generated/nosmallparts/stage/v1/stage';
import { useAuth } from '../lib/auth';
import { ratePlayScript } from '../lib/rpc';

interface Props {
  state: StageState;
  participantId?: string | null;
  onReset: () => void;
}

export function Discussion({ state, participantId, onReset }: Props) {
  const me = participantId ? state.participants.find((p) => p.id === participantId) : null;
  const { user } = useAuth();
  const [rating, setRating] = useState<number | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [rateError, setRateError] = useState<string | null>(null);

  const play = state.selectedPlay;
  const isAuthor = !!user && !!play?.createdByUserId && user.id === play.createdByUserId;

  async function submit(r: number) {
    if (!play) return;
    setRating(r);
    setSaving(true);
    setRateError(null);
    try {
      await ratePlayScript(play.id, r);
      setSaved(true);
    } catch {
      setRateError('Could not save your rating.');
      setRating(null);
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="flex flex-1 flex-col items-center justify-center gap-5 bg-stage-bg px-6 py-10">
      {me?.forceMuted ? <p className="text-xs text-red-400">🔇 You are muted by the host.</p> : null}
      <p className="font-serif text-5xl text-stage-gold">Bravo!</p>
      {play ? <p className="text-stage-muted">That was "{play.title}"</p> : null}

      <div className="mt-4 flex flex-col items-center gap-2">
        <p className="text-stage-muted">Cast</p>
        {state.participants
          .filter((p) => !p.isAudience && p.castMemberIds.length > 0)
          .map((p) => (
            <p key={p.id} className="text-stage-text">
              {p.name}
              {play
                ? `: ${p.castMemberIds
                    .map((id) => play.cast.find((c) => c.id === id)?.name)
                    .filter(Boolean)
                    .join(', ')}`
                : ''}
            </p>
          ))}
      </div>

      {play ? (
        user ? (
          isAuthor ? (
            <div className="flex flex-col items-center gap-1">
              <div className="flex gap-2 opacity-40">
                {[1, 2, 3, 4, 5].map((n) => (
                  <span key={n} className="rounded-md border border-stage-border px-3 py-1 text-stage-muted">
                    ★ {n}
                  </span>
                ))}
              </div>
              <p className="text-xs text-stage-muted">Authors can't rate their own work.</p>
            </div>
          ) : (
            <div className="flex flex-col items-center gap-2">
              <p className="text-stage-muted">Rate this script:</p>
              <div className="flex gap-2">
                {[1, 2, 3, 4, 5].map((n) => (
                  <button
                    type="button"
                    key={n}
                    onClick={() => submit(n)}
                    disabled={saving || saved}
                    className={`rounded-md border border-stage-border px-3 py-1 ${
                      rating && n <= rating ? 'text-stage-gold' : 'text-stage-muted'
                    }`}
                  >
                    ★ {n}
                  </button>
                ))}
              </div>
              {saved ? <p className="text-xs text-stage-muted">Thanks!</p> : null}
              {rateError ? <p className="text-xs text-red-400">{rateError}</p> : null}
            </div>
          )
        ) : (
          <div className="flex flex-col items-center gap-1">
            <Link
              to="/login"
              className="rounded-md border border-stage-gold px-4 py-2 font-semibold text-stage-gold active:opacity-70"
            >
              Sign in to rate
            </Link>
            <p className="text-xs text-stage-muted">Only signed-in users can rate scripts.</p>
          </div>
        )
      ) : null}

      <button
        type="button"
        onClick={onReset}
        className="mt-6 rounded-md bg-stage-gold px-5 py-3 font-semibold text-stage-bg"
      >
        Perform again
      </button>
    </div>
  );
}
