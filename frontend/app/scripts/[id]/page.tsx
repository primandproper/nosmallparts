'use client';

import { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import { Play, getConfig, getCastMemberById } from '@/lib/types';

type ScriptWithMeta = Play & { isUserScript?: boolean };

export default function ScriptDetailPage() {
  const params = useParams();
  const router = useRouter();
  const scriptId = params.id as string;

  const [script, setScript] = useState<ScriptWithMeta | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchScript() {
      try {
        const res = await fetch(`${getConfig().serverUrl}/api/scripts/${scriptId}`);
        if (!res.ok) {
          if (res.status === 404) {
            setError('Script not found');
          } else {
            setError('Failed to load script');
          }
          return;
        }
        const data = await res.json();
        setScript(data);
      } catch (err) {
        console.error('Failed to fetch script:', err);
        setError('Failed to load script');
      } finally {
        setIsLoading(false);
      }
    }

    if (scriptId) {
      fetchScript();
    }
  }, [scriptId]);

  const handleDelete = async () => {
    if (!script?.isUserScript) return;
    if (!confirm('Delete this script? This cannot be undone.')) return;

    try {
      const res = await fetch(`${getConfig().serverUrl}/api/scripts/${scriptId}`, {
        method: 'DELETE',
      });
      if (res.ok) {
        router.push('/scripts');
      }
    } catch (err) {
      console.error('Failed to delete script:', err);
      alert('Failed to delete script');
    }
  };

  if (isLoading) {
    return (
      <main className="min-h-screen stage-bg flex items-center justify-center">
        <div className="text-[var(--gold)] text-lg">Loading script...</div>
      </main>
    );
  }

  if (error || !script) {
    return (
      <main className="min-h-screen stage-bg flex items-center justify-center">
        <div className="text-center">
          <div className="text-red-400 text-lg mb-4">{error || 'Script not found'}</div>
          <Link href="/scripts" className="text-[var(--gold)] hover:text-[var(--gold-light)]">
            ← Back to Scripts
          </Link>
        </div>
      </main>
    );
  }

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
            </div>
            <div className="flex items-center gap-3">
              {script.isUserScript && (
                <button
                  onClick={handleDelete}
                  className="px-4 py-2 rounded-lg text-red-400 bg-red-500/10 hover:bg-red-500/20 transition-colors"
                >
                  Delete
                </button>
              )}
              <Link href="/" className="btn-elegant px-5 py-2 rounded-lg font-medium">
                🎭 Use in Stage
              </Link>
            </div>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-8">
        {/* Script Info */}
        <div className="card-elegant rounded-xl p-8 mb-8">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h1 className="font-[Cormorant_Garamond] text-4xl text-[var(--cream)] mb-2">
                {script.title}
              </h1>
              {script.author && <p className="text-[var(--gold)] text-lg">by {script.author}</p>}
            </div>
            <span
              className={`px-3 py-1 rounded-full text-sm font-medium border ${
                script.isUserScript
                  ? 'bg-green-500/20 text-green-400 border-green-500/30'
                  : 'bg-purple-500/20 text-purple-400 border-purple-500/30'
              }`}
            >
              {script.isUserScript ? 'Published' : 'Sample'}
            </span>
          </div>

          {script.description && (
            <p className="text-[var(--cream)] opacity-70 text-lg mb-6">{script.description}</p>
          )}

          <div className="flex items-center gap-6 text-[var(--cream)] opacity-50">
            <span>
              👥 {script.cast.length} role{script.cast.length !== 1 ? 's' : ''}
            </span>
            <span>
              📜 {script.lines.length} line{script.lines.length !== 1 ? 's' : ''}
            </span>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Lines */}
          <div className="lg:col-span-2">
            <h2 className="font-[Cormorant_Garamond] text-2xl text-[var(--gold)] mb-4">Script</h2>
            <div className="space-y-4">
              {script.lines.map((line, index) => {
                const castMember = getCastMemberById(script, line.castMemberId);
                return (
                  <div key={index} className="card-elegant rounded-lg p-4">
                    <div className="flex items-start gap-4">
                      <span className="text-[var(--gold)] font-semibold min-w-[120px]">
                        {castMember?.name || 'Unknown'}:
                      </span>
                      <div className="flex-1">
                        {line.direction && (
                          <span className="text-purple-400 text-sm italic mr-2">
                            ({line.direction})
                          </span>
                        )}
                        <span className="text-[var(--cream)]">{line.text}</span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Cast */}
          <div className="lg:col-span-1">
            <div className="card-elegant rounded-xl p-6 sticky top-24">
              <h2 className="font-[Cormorant_Garamond] text-2xl text-[var(--gold)] mb-4">Cast</h2>
              <ul className="space-y-3">
                {script.cast.map((member) => {
                  const lineCount = script.lines.filter((l) => l.castMemberId === member.id).length;
                  return (
                    <li
                      key={member.id}
                      className="flex items-center justify-between p-3 rounded-lg bg-[var(--charcoal)]/30"
                    >
                      <span className="text-[var(--cream)] font-medium">{member.name}</span>
                      <span className="text-[var(--cream)] opacity-40 text-sm">
                        {lineCount} line{lineCount !== 1 ? 's' : ''}
                      </span>
                    </li>
                  );
                })}
              </ul>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
