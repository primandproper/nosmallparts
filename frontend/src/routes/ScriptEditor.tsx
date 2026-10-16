import { useEffect, useMemo, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router';

import { Centered, Screen, Spinner } from '../components/Screen';
import { type CastMember, type PlayLine, ScrollMode } from '../generated/nosmallparts/scripts/v1/scripts';
import { useAuth } from '../lib/auth';
import { deleteDraft, getDraft, isDraftId, saveDraft } from '../lib/drafts';
import { messageOf } from '../lib/errors';
import { createPlayScript, getPlayScript, updatePlayScript } from '../lib/rpc';
import { validatePlay } from '../lib/script-utils';

interface CastRow extends CastMember {
  key: string;
}

interface LineRow extends PlayLine {
  key: string;
}

const AUTOSAVE_MS = 3000;

const fieldClass =
  'w-full rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text placeholder:text-stage-muted';
const bareFieldClass = 'w-full bg-transparent placeholder:text-stage-muted';

function withoutKey<T extends { key: string }>({ key: _key, ...rest }: T): Omit<T, 'key'> {
  return rest;
}

export default function ScriptEditor() {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const id = params.get('id');
  const draftId = params.get('draftId');
  const editingRemoteId = id && !isDraftId(id) ? id : null;
  const editingDraftId = draftId ?? (id && isDraftId(id) ? id : null);

  const { user } = useAuth();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [author, setAuthor] = useState('');
  const [isPublic, setIsPublic] = useState(false);
  const [cast, setCast] = useState<CastRow[]>([]);
  const [lines, setLines] = useState<LineRow[]>([]);
  const [currentDraftId, setCurrentDraftId] = useState<string | null>(editingDraftId);
  const [loading, setLoading] = useState(true);
  const [publishing, setPublishing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const autosaveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    let cancelled = false;

    const hydrate = (p: {
      title: string;
      description: string;
      author?: string;
      isPublic?: boolean;
      cast: CastMember[];
      lines: PlayLine[];
    }) => {
      setTitle(p.title);
      setDescription(p.description);
      setAuthor(p.author ?? '');
      setIsPublic(p.isPublic ?? false);
      setCast(p.cast.map((c, i) => ({ ...c, key: `c-${i}-${c.id}` })));
      setLines(p.lines.map((l, i) => ({ ...l, key: `l-${i}-${l.index}` })));
    };

    (async () => {
      try {
        if (editingRemoteId) {
          const p = await getPlayScript(editingRemoteId);
          if (!cancelled) hydrate(p);
        } else if (editingDraftId) {
          const d = getDraft(editingDraftId);
          if (d && !cancelled) hydrate(d);
        }
      } catch (e) {
        if (!cancelled) setError(messageOf(e, 'Failed to load'));
      } finally {
        if (!cancelled) setLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [editingRemoteId, editingDraftId]);

  // Autosave drafts every few seconds while editing. A published script is
  // never silently saved as a draft.
  useEffect(() => {
    if (loading || editingRemoteId) return;
    if (autosaveTimer.current) clearTimeout(autosaveTimer.current);
    autosaveTimer.current = setTimeout(() => {
      if (!title.trim() && cast.length === 0 && lines.length === 0) return;
      const saved = saveDraft({
        id: currentDraftId ?? undefined,
        title: title || 'Untitled',
        description,
        author: author || undefined,
        cast: cast.map(withoutKey),
        lines: lines.map(withoutKey),
      });
      if (!currentDraftId) setCurrentDraftId(saved.id);
    }, AUTOSAVE_MS);
    return () => {
      if (autosaveTimer.current) clearTimeout(autosaveTimer.current);
    };
  }, [title, description, author, cast, lines, editingRemoteId, currentDraftId, loading]);

  function addCastMember() {
    const n = cast.length + 1;
    setCast((prev) => [...prev, { id: `new-${Date.now()}-${n}`, name: `Character ${n}`, key: `c-${Date.now()}` }]);
  }
  function updateCast(key: string, patch: Partial<CastMember>) {
    setCast((prev) => prev.map((c) => (c.key === key ? { ...c, ...patch } : c)));
  }
  function removeCast(key: string) {
    const removed = cast.find((c) => c.key === key);
    setCast((prev) => prev.filter((c) => c.key !== key));
    if (removed) {
      setLines((prev) =>
        prev
          .map((l) => ({ ...l, castMemberIds: l.castMemberIds.filter((cid) => cid !== removed.id) }))
          .filter((l) => l.castMemberIds.length > 0),
      );
    }
  }

  function newLine(index: number, castMemberId: string | undefined): LineRow {
    return {
      key: `l-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
      castMemberIds: castMemberId ? [castMemberId] : [],
      text: '',
      index,
      scrollMode: ScrollMode.SCROLL_MODE_UNSPECIFIED,
    };
  }
  function addLine() {
    const nextIndex = lines.length === 0 ? 0 : Math.max(...lines.map((l) => l.index)) + 1;
    setLines((prev) => [...prev, newLine(nextIndex, cast[0]?.id)]);
  }
  function addOverlappingLine(atIndex: number) {
    const used = new Set(lines.filter((l) => l.index === atIndex).flatMap((l) => l.castMemberIds));
    const remaining = cast.find((c) => !used.has(c.id)) ?? cast[0];
    setLines((prev) => [...prev, newLine(atIndex, remaining?.id)]);
  }
  function updateLine(key: string, patch: Partial<PlayLine>) {
    setLines((prev) => prev.map((l) => (l.key === key ? { ...l, ...patch } : l)));
  }
  function toggleSpeaker(key: string, castMemberId: string) {
    setLines((prev) =>
      prev.map((l) => {
        if (l.key !== key) return l;
        const next = l.castMemberIds.includes(castMemberId)
          ? l.castMemberIds.filter((cid) => cid !== castMemberId)
          : [...l.castMemberIds, castMemberId];
        // A line keeps at least one speaker.
        return { ...l, castMemberIds: next.length === 0 ? l.castMemberIds : next };
      }),
    );
  }
  function removeLine(key: string) {
    setLines((prev) => prev.filter((l) => l.key !== key));
  }

  const validation = useMemo(
    () => validatePlay(title, cast.map(withoutKey), lines.map(withoutKey)),
    [title, cast, lines],
  );

  async function publish() {
    if (!validation.valid) {
      setError(validation.error ?? 'Invalid script');
      return;
    }
    if (!user) {
      window.alert('Sign in required\n\nPlease sign in to publish scripts.');
      return;
    }
    setPublishing(true);
    setError(null);
    try {
      const payload = {
        title,
        description,
        author: author.trim() || undefined,
        isPublic,
        cast: cast.map(withoutKey),
        lines: lines.map(withoutKey),
        tagIds: [],
        tagNames: [],
      };
      const saved = editingRemoteId
        ? await updatePlayScript(editingRemoteId, payload)
        : await createPlayScript(payload);
      if (currentDraftId) deleteDraft(currentDraftId);
      navigate(`/scripts/${saved.id}`, { replace: true });
    } catch (e) {
      setError(messageOf(e, 'Publish failed'));
    } finally {
      setPublishing(false);
    }
  }

  function discardDraft() {
    if (currentDraftId) deleteDraft(currentDraftId);
    navigate(-1);
  }

  const screenTitle = editingRemoteId ? 'Edit script' : 'New script';

  if (loading) {
    return (
      <Screen title={screenTitle}>
        <Centered>
          <Spinner />
        </Centered>
      </Screen>
    );
  }

  const noCast = cast.length === 0;

  return (
    <Screen title={screenTitle}>
      <div className="mx-auto flex w-full max-w-3xl flex-col gap-3.5 p-4">
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Title"
          className={`${fieldClass} text-lg`}
        />
        <input
          value={author}
          onChange={(e) => setAuthor(e.target.value)}
          placeholder="Author (optional)"
          className={fieldClass}
        />
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Description"
          rows={3}
          className={fieldClass}
        />
        <label className="flex items-center gap-3 text-stage-text">
          <input
            type="checkbox"
            role="switch"
            checked={isPublic}
            onChange={(e) => setIsPublic(e.target.checked)}
            className="h-5 w-5 accent-stage-gold"
          />
          Public
        </label>

        <section className="mt-2">
          <div className="flex items-center justify-between">
            <h2 className="font-serif text-xl text-stage-gold">Cast</h2>
            <button type="button" onClick={addCastMember} className="px-2 py-1 text-stage-gold">
              + Add character
            </button>
          </div>
          {noCast ? (
            <p className="mt-2 text-sm text-stage-muted">No characters yet.</p>
          ) : (
            <div className="mt-2 flex flex-col gap-2">
              {cast.map((c) => (
                <div key={c.key} className="flex flex-col rounded-md border border-stage-border bg-stage-surface p-3">
                  <input
                    value={c.name}
                    onChange={(e) => updateCast(c.key, { name: e.target.value })}
                    placeholder="Character name"
                    className={`${bareFieldClass} text-stage-text`}
                  />
                  <input
                    value={c.description ?? ''}
                    onChange={(e) => updateCast(c.key, { description: e.target.value })}
                    placeholder="Description (optional)"
                    className={`${bareFieldClass} mt-1 text-sm text-stage-muted`}
                  />
                  <button
                    type="button"
                    onClick={() => removeCast(c.key)}
                    className="mt-2 self-end text-sm text-red-400"
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>
          )}
        </section>

        <section className="mt-4">
          <div className="flex items-center justify-between">
            <h2 className="font-serif text-xl text-stage-gold">Lines</h2>
            <button
              type="button"
              onClick={addLine}
              disabled={noCast}
              className={`px-2 py-1 ${noCast ? 'text-stage-muted' : 'text-stage-gold'}`}
            >
              + Add line
            </button>
          </div>
          {lines.length === 0 ? (
            <p className="mt-2 text-sm text-stage-muted">{noCast ? 'Add a character first.' : 'No lines yet.'}</p>
          ) : (
            <div className="mt-2 flex flex-col gap-3">
              {Array.from(new Set(lines.map((l) => l.index)))
                .sort((a, b) => a - b)
                .map((turnIndex) => {
                  const groupLines = lines.filter((l) => l.index === turnIndex);
                  return (
                    <div key={`turn-${turnIndex}`} className="flex flex-col gap-1.5">
                      <div className="flex items-center justify-between">
                        <span className="text-xs text-stage-muted">
                          Turn {turnIndex + 1}
                          {groupLines.length > 1 ? ' · overlapping' : ''}
                        </span>
                        <button
                          type="button"
                          onClick={() => addOverlappingLine(turnIndex)}
                          disabled={noCast}
                          className={`px-2 py-0.5 text-xs ${noCast ? 'text-stage-muted' : 'text-stage-gold'}`}
                        >
                          + Overlap
                        </button>
                      </div>
                      {groupLines.map((l) => {
                        const selectedIds = new Set(l.castMemberIds);
                        return (
                          <div
                            key={l.key}
                            className="flex flex-col rounded-md border border-stage-border bg-stage-surface p-3"
                          >
                            <div className="flex flex-wrap items-center gap-1">
                              {selectedIds.size > 1 ? (
                                <span className="mr-1 text-[10px] font-semibold uppercase tracking-wider text-stage-gold">
                                  Unison
                                </span>
                              ) : null}
                              {cast.map((c) => {
                                const selected = selectedIds.has(c.id);
                                return (
                                  <button
                                    type="button"
                                    key={c.id}
                                    onClick={() => toggleSpeaker(l.key, c.id)}
                                    aria-pressed={selected}
                                    className={`rounded-full border px-2 py-0.5 text-xs ${
                                      selected
                                        ? 'border-stage-gold bg-stage-goldDim text-stage-bg'
                                        : 'border-stage-border text-stage-muted'
                                    }`}
                                  >
                                    {c.name}
                                  </button>
                                );
                              })}
                            </div>
                            <input
                              value={l.direction ?? ''}
                              onChange={(e) => updateLine(l.key, { direction: e.target.value || undefined })}
                              placeholder="Stage direction (optional)"
                              className={`${bareFieldClass} mt-2 italic text-stage-muted`}
                            />
                            <textarea
                              value={l.text}
                              onChange={(e) => updateLine(l.key, { text: e.target.value })}
                              placeholder="Line text"
                              rows={2}
                              className={`${bareFieldClass} mt-1 resize-y text-stage-text`}
                            />
                            <button
                              type="button"
                              onClick={() => removeLine(l.key)}
                              className="mt-2 self-end text-sm text-red-400"
                            >
                              Remove
                            </button>
                          </div>
                        );
                      })}
                    </div>
                  );
                })}
            </div>
          )}
        </section>

        {error ? <p className="text-red-400">{error}</p> : null}
        {!validation.valid && !error ? <p className="text-sm text-stage-muted">{validation.error}</p> : null}

        <button
          type="button"
          onClick={publish}
          disabled={publishing}
          className={`mt-4 rounded-md px-4 py-3 font-semibold text-stage-bg ${
            publishing ? 'bg-stage-goldDim' : 'bg-stage-gold'
          }`}
        >
          {publishing ? 'Publishing…' : editingRemoteId ? 'Save changes' : user ? 'Publish' : 'Sign in to publish'}
        </button>

        {!editingRemoteId ? (
          <button type="button" onClick={discardDraft} className="mt-2 text-center text-stage-muted">
            Discard draft
          </button>
        ) : null}
      </div>
    </Screen>
  );
}
