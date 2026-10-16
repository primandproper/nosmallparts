import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router';

import { Centered, Screen, Spinner } from '../components/Screen';
import type { Play } from '../generated/nosmallparts/scripts/v1/scripts';
import { useAuth } from '../lib/auth';
import { messageOf } from '../lib/errors';
import { getStoredPlayerName } from '../lib/player-name';
import { createStage, getPlayScript, ratePlayScript } from '../lib/rpc';

export default function ScriptDetail() {
  const { id = '' } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [play, setPlay] = useState<Play | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [rating, setRating] = useState<number | null>(null);
  const [savingRating, setSavingRating] = useState(false);
  const [ratingError, setRatingError] = useState<string | null>(null);
  const [startingStage, setStartingStage] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        setLoading(true);
        const p = await getPlayScript(id);
        if (!cancelled) {
          setPlay(p);
          setRating(p.userRating ?? null);
        }
      } catch (e) {
        if (!cancelled) setError(messageOf(e, 'Failed'));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [id]);

  async function submitRating(r: number) {
    if (!play) return;
    setSavingRating(true);
    setRatingError(null);
    try {
      const stats = await ratePlayScript(play.id, r);
      setRating(r);
      setPlay({ ...play, userRating: r, ratingStats: stats });
    } catch (e) {
      setRatingError(messageOf(e, 'Rating failed'));
    } finally {
      setSavingRating(false);
    }
  }

  async function startStage() {
    if (!play || startingStage) return;
    setStartingStage(true);
    try {
      if (!getStoredPlayerName()) {
        navigate(`/?s=${encodeURIComponent(play.id)}`);
        return;
      }
      const res = await createStage(play.id);
      navigate(`/stage/${res.code}`, { replace: true });
    } catch (e) {
      setError(messageOf(e, 'Could not start a stage'));
    } finally {
      setStartingStage(false);
    }
  }

  if (loading) {
    return (
      <Screen title="Script">
        <Centered>
          <Spinner />
        </Centered>
      </Screen>
    );
  }
  if (!play) {
    return (
      <Screen title="Script">
        <Centered>
          <p className="text-red-400">{error || 'Not found'}</p>
        </Centered>
      </Screen>
    );
  }

  const castById = new Map(play.cast.map((c) => [c.id, c]));
  const sortedLines = [...play.lines].sort((a, b) => a.index - b.index);

  return (
    <Screen title={play.title}>
      <div className="mx-auto w-full max-w-3xl p-4">
        <h2 className="font-serif text-3xl text-stage-gold">{play.title}</h2>
        {play.author ? <p className="mt-1 text-stage-muted">by {play.author}</p> : null}
        {play.description ? <p className="mt-3 text-stage-text">{play.description}</p> : null}

        <div className="mt-4 flex flex-wrap gap-2">
          {play.cast.map((c) => (
            <span
              key={c.id}
              className="rounded-full border border-stage-border bg-stage-surface px-3 py-1 text-xs text-stage-text"
            >
              {c.name}
            </span>
          ))}
        </div>

        {user ? (
          <div className="mt-6">
            <div className="mb-2 flex items-center gap-2">
              <span className="text-stage-muted">Your rating:</span>
              {savingRating ? <Spinner /> : null}
            </div>
            <div className="flex gap-2">
              {[1, 2, 3, 4, 5].map((n) => (
                <button
                  type="button"
                  key={n}
                  disabled={savingRating}
                  onClick={() => submitRating(n)}
                  className={`rounded-md border border-stage-border px-3 py-1 ${
                    rating && n <= rating ? 'text-stage-gold' : 'text-stage-muted'
                  }`}
                >
                  ★ {n}
                </button>
              ))}
            </div>
            {ratingError ? <p className="mt-2 text-sm text-red-400">{ratingError}</p> : null}
            {play.ratingStats ? (
              <p className="mt-2 text-xs text-stage-muted">
                Avg {play.ratingStats.averageRating.toFixed(1)} across {play.ratingStats.ratingCount} rating
                {play.ratingStats.ratingCount === 1 ? '' : 's'}
              </p>
            ) : null}
          </div>
        ) : null}

        {error ? <p className="mt-4 text-red-400">{error}</p> : null}

        <button
          type="button"
          onClick={startStage}
          disabled={startingStage}
          className={`mt-6 flex w-full items-center justify-center gap-2 rounded-md px-4 py-3 font-semibold text-stage-bg ${
            startingStage ? 'bg-stage-goldDim' : 'bg-stage-gold'
          }`}
        >
          {startingStage ? <Spinner dark /> : null}
          {startingStage ? 'Starting…' : 'Start a stage with this script'}
        </button>

        <h3 className="mb-2 mt-8 font-serif text-xl text-stage-gold">Script</h3>
        {sortedLines.map((l, i) => {
          const names = l.castMemberIds.map((cid) => castById.get(cid)?.name).filter((n): n is string => !!n);
          const isUnison = names.length > 1;
          const prev = i > 0 ? sortedLines[i - 1] : null;
          const isOverlap = prev != null && prev.index === l.index;
          return (
            <div key={`${l.index}-${i}`} className="mb-3">
              <div className="flex items-center gap-2">
                <span className="text-stage-gold">{names.join(' & ') || 'Unknown'}</span>
                {isUnison ? (
                  <span className="text-[10px] uppercase tracking-wider text-stage-muted">unison</span>
                ) : null}
                {isOverlap ? (
                  <span className="text-[10px] uppercase tracking-wider text-stage-muted">overlapping</span>
                ) : null}
              </div>
              {l.direction ? <p className="italic text-stage-muted">({l.direction})</p> : null}
              <p className="text-stage-text">{l.text}</p>
            </div>
          );
        })}
      </div>
    </Screen>
  );
}
