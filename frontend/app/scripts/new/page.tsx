'use client';

import { useState, useEffect, useCallback, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { PlayLine, CastMember, saveDraft, getDraft, deleteDraft } from '@/lib/types';
import {
  groupLinesByIndex,
  getSortedLineIndices,
  getFirstAvailableCastMemberId,
  getNextLineIndex,
  getInvalidLineIndices,
  validatePlay,
} from '@/lib/scriptUtils';
import { TurnEditor } from '@/components/LineEditor';
import { useAuth } from '@/hooks/useAuth';
import { v4 as uuidv4 } from 'uuid';

function ScriptEditorContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const draftId = searchParams.get('draft');
  const { isAuthenticated, isLoading: authLoading } = useAuth();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [author, setAuthor] = useState('');
  const [cast, setCast] = useState<CastMember[]>([]);
  const [lines, setLines] = useState<PlayLine[]>([]);
  const [newCastName, setNewCastName] = useState('');
  const [newCastDescription, setNewCastDescription] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [isPublishing, setIsPublishing] = useState(false);
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [currentId, setCurrentId] = useState<string | null>(null);
  const [invalidLineIndices, setInvalidLineIndices] = useState<Set<number> | null>(null);

  // Load draft if editing
  useEffect(() => {
    if (draftId) {
      const draft = getDraft(draftId);
      if (draft) {
        setCurrentId(draft.id);
        setTitle(draft.title);
        setDescription(draft.description);
        setAuthor(draft.author || '');
        setCast(draft.cast);
        setLines(draft.lines);
      }
    }
  }, [draftId]);

  // Clear validation highlights when user fixes all invalid lines
  useEffect(() => {
    setInvalidLineIndices((prev) => {
      if (!prev) return prev;
      const stillInvalid = getInvalidLineIndices(lines);
      return stillInvalid.size === 0 ? null : stillInvalid;
    });
  }, [lines]);

  const handleAddCast = () => {
    if (!newCastName.trim()) return;

    const newMember: CastMember = {
      id: uuidv4(),
      name: newCastName.trim(),
      description: newCastDescription.trim() || undefined,
    };
    setCast([...cast, newMember]);
    setNewCastName('');
    setNewCastDescription('');
  };

  const handleUpdateCastMember = (
    id: string,
    updates: Partial<Pick<CastMember, 'name' | 'description'>>
  ) => {
    setCast(cast.map((c) => (c.id === id ? { ...c, ...updates } : c)));
  };

  const handleRemoveCast = (id: string) => {
    setCast(cast.filter((c) => c.id !== id));
    // Also remove lines for this cast member
    setLines(lines.filter((l) => l.castMemberId !== id));
  };

  // Add a new line
  const handleAddLine = () => {
    const newLine: PlayLine = {
      castMemberId: cast.length > 0 ? cast[0].id : '',
      text: '',
      index: getNextLineIndex(lines),
    };
    setLines([...lines, newLine]);
  };

  // Add a simultaneous speaker to an existing line
  const handleAddSimultaneousSpeaker = (lineIndex: number) => {
    const newLine: PlayLine = {
      castMemberId: getFirstAvailableCastMemberId(cast, lines, lineIndex),
      text: '',
      index: lineIndex,
    };
    setLines([...lines, newLine]);
  };

  const handleUpdateLine = (arrayIndex: number, updatedLine: PlayLine) => {
    const newLines = [...lines];
    newLines[arrayIndex] = updatedLine;
    setLines(newLines);
  };

  const handleDeleteSingleLine = (arrayIndex: number) => {
    setLines(lines.filter((_, i) => i !== arrayIndex));
  };

  // Delete all speakers in a line
  const handleDeleteLineGroup = (lineIndex: number) => {
    setLines(lines.filter((l) => l.index !== lineIndex));
  };

  // Move an entire line up or down
  const handleMoveLineGroup = (lineIndex: number, direction: 'up' | 'down') => {
    const sortedIndices = getSortedLineIndices(lines);
    const currentPosition = sortedIndices.indexOf(lineIndex);

    if (
      (direction === 'up' && currentPosition === 0) ||
      (direction === 'down' && currentPosition === sortedIndices.length - 1)
    ) {
      return;
    }

    const swapWithIndex =
      direction === 'up' ? sortedIndices[currentPosition - 1] : sortedIndices[currentPosition + 1];

    // Swap the indices of the two lines
    const newLines = lines.map((line) => {
      if (line.index === lineIndex) {
        return { ...line, index: swapWithIndex };
      }
      if (line.index === swapWithIndex) {
        return { ...line, index: lineIndex };
      }
      return line;
    });

    setLines(newLines);
  };

  const handleSaveDraft = useCallback(
    async (silent = false) => {
      if (!title.trim()) {
        if (!silent) alert('Please enter a title');
        return;
      }

      setIsSaving(true);
      try {
        const saved = saveDraft({
          id: currentId || undefined,
          title: title.trim(),
          description: description.trim(),
          author: author.trim() || undefined,
          cast,
          lines,
        });
        setCurrentId(saved.id);
        setLastSaved(new Date());
        if (!silent) {
          alert('Draft saved!');
        }
      } catch (err) {
        console.error('Failed to save draft:', err);
        if (!silent) alert('Failed to save draft');
      } finally {
        setIsSaving(false);
      }
    },
    [title, description, author, cast, lines, currentId]
  );

  // Auto-save draft every 30 seconds if there are changes
  useEffect(() => {
    if (!title) return;

    const timer = setInterval(() => {
      handleSaveDraft(true);
    }, 30000);

    return () => clearInterval(timer);
  }, [title, handleSaveDraft]);

  const handlePublish = async () => {
    const validation = validatePlay(title, cast, lines);
    if (!validation.valid) {
      setInvalidLineIndices(validation.invalidLineIndices ?? null);
      alert(validation.error);
      return;
    }
    setInvalidLineIndices(null);

    setIsPublishing(true);
    try {
      const res = await fetch('/api/play-scripts', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: title.trim(),
          description: description.trim(),
          author: author.trim() || '',
          isPublic: true, // Publishing makes it public
          cast,
          lines,
        }),
      });

      if (!res.ok) {
        throw new Error('Failed to publish');
      }

      // Delete the draft if we were editing one
      if (currentId && currentId.startsWith('draft-')) {
        deleteDraft(currentId);
      }

      alert('Script published!');
      router.push('/scripts');
    } catch (err) {
      console.error('Failed to publish script:', err);
      alert('Failed to publish script. Please try again.');
    } finally {
      setIsPublishing(false);
    }
  };

  return (
    <main className="min-h-screen stage-bg">
      {/* Header */}
      <header className="border-b border-[var(--gold)]/20 backdrop-blur-sm sticky top-0 z-10 bg-[var(--black)]/80">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link
                href="/scripts"
                className="text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors"
              >
                ← Back to Scripts
              </Link>
              <h1 className="font-[Cormorant_Garamond] text-2xl text-[var(--cream)]">
                {draftId ? 'Edit Script' : 'Create Script'}
              </h1>
            </div>
            <div className="flex items-center gap-3">
              {lastSaved && (
                <span className="text-[var(--cream)] opacity-40 text-sm">
                  Saved {lastSaved.toLocaleTimeString()}
                </span>
              )}
              <button
                onClick={() => handleSaveDraft(false)}
                disabled={isSaving}
                className="px-4 py-2 rounded-lg text-[var(--cream)] bg-[var(--charcoal)]/50 hover:bg-[var(--charcoal)] transition-colors disabled:opacity-50"
              >
                {isSaving ? 'Saving...' : 'Save Draft'}
              </button>
              {authLoading ? (
                <div className="px-5 py-2 rounded-lg bg-[var(--charcoal)]/30 text-[var(--cream)] opacity-50">
                  Loading...
                </div>
              ) : isAuthenticated ? (
                <button
                  onClick={handlePublish}
                  disabled={isPublishing}
                  className="btn-elegant px-5 py-2 rounded-lg font-medium disabled:opacity-50"
                >
                  {isPublishing ? 'Publishing...' : '🎭 Publish'}
                </button>
              ) : (
                <Link
                  href="/auth/login"
                  className="btn-elegant px-5 py-2 rounded-lg font-medium inline-flex items-center gap-2"
                >
                  <span>Sign in to Publish</span>
                </Link>
              )}
            </div>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Main Editor */}
          <div className="lg:col-span-2 space-y-8">
            {/* Script Details */}
            <section className="card-elegant rounded-xl p-6">
              <h2 className="font-[Cormorant_Garamond] text-xl text-[var(--gold)] mb-4">
                Script Details
              </h2>
              <div className="space-y-4">
                <div>
                  <label className="block text-[var(--cream)] opacity-60 text-sm mb-1">
                    Title *
                  </label>
                  <input
                    type="text"
                    value={title}
                    onChange={(e) => setTitle(e.target.value)}
                    placeholder="Enter script title..."
                    className="input-elegant w-full px-4 py-3 rounded-lg text-[var(--cream)] text-lg"
                    autoFocus
                  />
                </div>
                <div>
                  <label className="block text-[var(--cream)] opacity-60 text-sm mb-1">
                    Description
                  </label>
                  <textarea
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    placeholder="A brief description of the script..."
                    className="input-elegant w-full px-4 py-3 rounded-lg text-[var(--cream)] min-h-[100px] resize-y"
                  />
                </div>
                <div>
                  <label className="block text-[var(--cream)] opacity-60 text-sm mb-1">
                    Author
                  </label>
                  <input
                    type="text"
                    value={author}
                    onChange={(e) => setAuthor(e.target.value)}
                    placeholder="Author name (optional)"
                    className="input-elegant w-full px-4 py-3 rounded-lg text-[var(--cream)]"
                  />
                </div>
              </div>
            </section>

            {/* Lines */}
            <section className="card-elegant rounded-xl p-6">
              <div className="flex items-center justify-between mb-4">
                <h2 className="font-[Cormorant_Garamond] text-xl text-[var(--gold)]">
                  Script ({getSortedLineIndices(lines).length} line
                  {getSortedLineIndices(lines).length !== 1 ? 's' : ''})
                </h2>
                <button
                  onClick={handleAddLine}
                  disabled={cast.length === 0}
                  className="px-4 py-2 rounded-lg text-sm bg-[var(--gold)]/20 text-[var(--gold)] hover:bg-[var(--gold)]/30 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                  title={cast.length === 0 ? 'Add cast members first' : 'Add a new line'}
                >
                  + Add Line
                </button>
              </div>

              {cast.length === 0 ? (
                <p className="text-[var(--cream)] opacity-40 text-center py-8">
                  Add cast members first to start writing the script.
                </p>
              ) : lines.length === 0 ? (
                <p className="text-[var(--cream)] opacity-40 text-center py-8">
                  No lines yet. Click &quot;Add Line&quot; to start writing.
                </p>
              ) : (
                <div className="space-y-4">
                  {(() => {
                    const sortedIndices = getSortedLineIndices(lines);
                    const groupedLines = groupLinesByIndex(lines);

                    return sortedIndices.map((lineIdx, position) => {
                      const group = groupedLines.get(lineIdx)!;
                      return (
                        <TurnEditor
                          key={`line-${lineIdx}`}
                          lineIndex={position}
                          lineIdx={lineIdx}
                          lines={group.lines}
                          cast={cast}
                          lineArrayIndices={group.arrayIndices}
                          invalidLineIndices={invalidLineIndices}
                          onUpdateLine={handleUpdateLine}
                          onDeleteLine={handleDeleteSingleLine}
                          onAddSimultaneous={() => handleAddSimultaneousSpeaker(lineIdx)}
                          onDeleteLineGroup={() => handleDeleteLineGroup(lineIdx)}
                          onMoveLineGroup={(dir) => handleMoveLineGroup(lineIdx, dir)}
                          canMoveUp={position > 0}
                          canMoveDown={position < sortedIndices.length - 1}
                        />
                      );
                    });
                  })()}
                </div>
              )}
            </section>
          </div>

          {/* Sidebar - Cast Members */}
          <div className="lg:col-span-1">
            <section className="card-elegant rounded-xl p-6 sticky top-24">
              <h2 className="font-[Cormorant_Garamond] text-xl text-[var(--gold)] mb-4">
                Cast Members ({cast.length})
              </h2>

              {/* Add cast member */}
              <div className="space-y-2 mb-4">
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={newCastName}
                    onChange={(e) => setNewCastName(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter') {
                        e.preventDefault();
                        handleAddCast();
                      }
                    }}
                    placeholder="Character name..."
                    className="input-elegant flex-1 px-3 py-2 rounded-lg text-[var(--cream)] text-sm"
                  />
                  <button
                    onClick={handleAddCast}
                    disabled={!newCastName.trim()}
                    className="px-3 py-2 rounded-lg bg-[var(--gold)]/20 text-[var(--gold)] hover:bg-[var(--gold)]/30 transition-colors disabled:opacity-50"
                  >
                    +
                  </button>
                </div>
                <input
                  type="text"
                  value={newCastDescription}
                  onChange={(e) => setNewCastDescription(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      handleAddCast();
                    }
                  }}
                  placeholder="Description (e.g. Narrator, Villain) — optional"
                  className="input-elegant w-full px-3 py-2 rounded-lg text-[var(--cream)] text-sm opacity-80"
                />
              </div>

              {/* Cast list */}
              {cast.length === 0 ? (
                <p className="text-[var(--cream)] opacity-40 text-sm text-center py-4">
                  No cast members yet.
                </p>
              ) : (
                <ul className="space-y-2">
                  {cast.map((member) => {
                    const lineCount = lines.filter((l) => l.castMemberId === member.id).length;
                    return (
                      <li
                        key={member.id}
                        className="p-3 rounded-lg bg-[var(--charcoal)]/30 space-y-2"
                      >
                        <div className="flex items-center justify-between">
                          <span className="text-[var(--cream)] font-medium">{member.name}</span>
                          <div className="flex items-center gap-2">
                            <span className="text-[var(--cream)] opacity-40 text-xs">
                              {lineCount} line{lineCount !== 1 ? 's' : ''}
                            </span>
                            <button
                              onClick={() => handleRemoveCast(member.id)}
                              className="text-red-400 opacity-60 hover:opacity-100 transition-opacity text-sm"
                              title="Remove cast member"
                            >
                              ✕
                            </button>
                          </div>
                        </div>
                        <input
                          type="text"
                          value={member.description ?? ''}
                          onChange={(e) =>
                            handleUpdateCastMember(member.id, {
                              description: e.target.value || undefined,
                            })
                          }
                          placeholder="Add character description (optional)"
                          className="input-elegant w-full px-3 py-1.5 rounded text-[var(--cream)] text-sm opacity-80"
                        />
                      </li>
                    );
                  })}
                </ul>
              )}

              {/* Preview Info */}
              <div className="mt-6 pt-6 border-t border-[var(--cream)]/10">
                <h3 className="text-[var(--cream)] opacity-60 text-sm mb-2">Script Summary</h3>
                {(() => {
                  const uniqueLines =
                    lines.length > 0 ? new Set(lines.map((l) => l.index)).size : 0;
                  const totalSpeakers = lines.length;
                  const hasDuets = totalSpeakers > uniqueLines;
                  return (
                    <ul className="text-[var(--cream)] opacity-40 text-sm space-y-1">
                      <li>📝 {title || 'Untitled'}</li>
                      <li>
                        👥 {cast.length} character{cast.length !== 1 ? 's' : ''}
                      </li>
                      <li>
                        📜 {uniqueLines} line{uniqueLines !== 1 ? 's' : ''}
                        {hasDuets && ` (${totalSpeakers} total with duets)`}
                      </li>
                      {author && <li>✍️ by {author}</li>}
                    </ul>
                  );
                })()}
              </div>

              {/* Authentication Notice */}
              {!authLoading && !isAuthenticated && (
                <div className="mt-4 p-3 rounded-lg bg-[var(--gold)]/10 border border-[var(--gold)]/20">
                  <p className="text-[var(--cream)] text-sm opacity-80">
                    <span className="text-[var(--gold)]">Note:</span> You can save drafts locally,
                    but you&apos;ll need to{' '}
                    <Link href="/auth/login" className="text-[var(--gold)] hover:underline">
                      sign in
                    </Link>{' '}
                    or{' '}
                    <Link href="/auth/register" className="text-[var(--gold)] hover:underline">
                      create an account
                    </Link>{' '}
                    to publish your script.
                  </p>
                </div>
              )}
            </section>
          </div>
        </div>
      </div>
    </main>
  );
}

export default function ScriptEditorPage() {
  return (
    <Suspense
      fallback={
        <main className="min-h-screen stage-bg flex items-center justify-center">
          <div className="text-[var(--gold)] text-lg">Loading editor...</div>
        </main>
      }
    >
      <ScriptEditorContent />
    </Suspense>
  );
}
