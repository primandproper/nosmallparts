import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router';

import { Centered, inputClass, Screen, Spinner } from '../components/Screen';
import type { Play } from '../generated/nosmallparts/scripts/v1/scripts';
import { type Draft, getDrafts } from '../lib/drafts';
import { messageOf } from '../lib/errors';
import { listPlayScripts } from '../lib/rpc';

type Entry = { kind: 'draft'; draft: Draft } | { kind: 'play'; play: Play };

export default function ScriptsLibrary() {
  const [search, setSearch] = useState('');
  const [plays, setPlays] = useState<Play[]>([]);
  const [drafts, setDrafts] = useState<Draft[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async (q: string) => {
    setLoading(true);
    setError(null);
    try {
      setPlays(await listPlayScripts(q));
      setDrafts(getDrafts());
    } catch (e) {
      setError(messageOf(e, 'Failed to load'));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load(search);
  }, [load, search]);

  const entries: Entry[] = [
    ...drafts.map((draft) => ({ kind: 'draft' as const, draft })),
    ...plays.map((play) => ({ kind: 'play' as const, play })),
  ];

  return (
    <Screen
      title="Scripts"
      right={
        <Link to="/scripts/new" className="font-semibold text-stage-gold">
          + New
        </Link>
      }
    >
      <div className="mx-auto w-full max-w-3xl px-4 pt-3">
        <input
          placeholder="Search title, author, cast, tag…"
          className={`${inputClass} py-2`}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>
      {loading ? (
        <Centered>
          <Spinner />
        </Centered>
      ) : error ? (
        <Centered>
          <p className="text-red-400">{error}</p>
        </Centered>
      ) : entries.length === 0 ? (
        <p className="mt-12 text-center text-stage-muted">No scripts yet.</p>
      ) : (
        <ul className="mx-auto flex w-full max-w-3xl flex-col gap-2.5 p-4">
          {entries.map((entry) =>
            entry.kind === 'draft' ? (
              <li key={entry.draft.id}>
                <ScriptCard
                  to={`/scripts/new?draftId=${encodeURIComponent(entry.draft.id)}`}
                  title={entry.draft.title}
                  author={entry.draft.author}
                  description={entry.draft.description}
                  castCount={entry.draft.cast.length}
                  lineCount={entry.draft.lines.length}
                  badge="draft"
                />
              </li>
            ) : (
              <li key={entry.play.id}>
                <ScriptCard
                  to={`/scripts/${entry.play.id}`}
                  title={entry.play.title}
                  author={entry.play.author}
                  description={entry.play.description}
                  castCount={entry.play.cast.length}
                  lineCount={entry.play.lines.length}
                  badge={entry.play.isPublic ? 'public' : undefined}
                  disabled={!entry.play.enabled}
                  rating={entry.play.ratingStats}
                />
              </li>
            ),
          )}
        </ul>
      )}
    </Screen>
  );
}

function ScriptCard(props: {
  to: string;
  title: string;
  author?: string;
  description: string;
  castCount: number;
  lineCount: number;
  badge?: 'draft' | 'public';
  disabled?: boolean;
  rating?: { averageRating: number; ratingCount: number };
}) {
  return (
    <Link to={props.to} className="block rounded-lg border border-stage-border bg-stage-surface p-4">
      <div className="flex items-center justify-between gap-2">
        <span className="flex-1 truncate text-lg font-semibold text-stage-text">{props.title || 'Untitled'}</span>
        {props.disabled ? <span className="rounded bg-red-900 px-2 py-0.5 text-xs text-red-200">disabled</span> : null}
        {props.badge === 'draft' ? (
          <span className="rounded bg-stage-border px-2 py-0.5 text-xs text-stage-muted">draft</span>
        ) : props.badge === 'public' ? (
          <span className="rounded bg-stage-goldDim px-2 py-0.5 text-xs text-stage-bg">public</span>
        ) : null}
      </div>
      {props.author ? <p className="mt-1 text-sm text-stage-muted">by {props.author}</p> : null}
      {props.description ? <p className="mt-2 line-clamp-2 text-sm text-stage-text">{props.description}</p> : null}
      <div className="mt-3 flex gap-3 text-xs text-stage-muted">
        <span>{props.castCount} cast</span>
        <span>{props.lineCount} lines</span>
        {props.rating ? (
          <span>
            ★ {props.rating.averageRating.toFixed(1)} ({props.rating.ratingCount})
          </span>
        ) : null}
      </div>
    </Link>
  );
}
