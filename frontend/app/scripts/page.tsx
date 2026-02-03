'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Play, getConfig, getDrafts, deleteDraft, isDraft } from '@/lib/types';
import { ScriptCard } from '@/components/ScriptCard';
import { AdBannerPlaceholder } from '@/components/AdBanner';

type ScriptWithMeta = Play & { isUserScript?: boolean };

type FilterType = 'all' | 'sample' | 'published' | 'drafts';

export default function ScriptsPage() {
  const router = useRouter();
  const [scripts, setScripts] = useState<ScriptWithMeta[]>([]);
  const [drafts, setDrafts] = useState<Play[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [filter, setFilter] = useState<FilterType>('all');
  const [isLoading, setIsLoading] = useState(true);

  // Load scripts from API and drafts from localStorage
  useEffect(() => {
    async function fetchScripts() {
      try {
        const res = await fetch(`${getConfig().serverUrl}/api/scripts`);
        if (res.ok) {
          const data = await res.json();
          setScripts(data);
        }
      } catch (err) {
        console.error('Failed to fetch scripts:', err);
      } finally {
        setIsLoading(false);
      }
    }

    fetchScripts();
    setDrafts(getDrafts());
  }, []);

  type ScriptItem = { script: Play | ScriptWithMeta; isDraft: boolean };

  // Filter and search logic
  const filteredItems = (): ScriptItem[] => {
    let items: ScriptItem[] = [];

    if (filter === 'drafts' || filter === 'all') {
      items = [...items, ...drafts.map((d) => ({ script: d, isDraft: true }))];
    }

    if (filter !== 'drafts') {
      const filteredScripts = scripts.filter((s) => {
        if (filter === 'sample') return !s.isUserScript;
        if (filter === 'published') return s.isUserScript;
        return true;
      });
      items = [...items, ...filteredScripts.map((s) => ({ script: s, isDraft: false }))];
    }

    // Apply search
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      items = items.filter((item) => {
        return (
          item.script.title.toLowerCase().includes(query) ||
          item.script.description.toLowerCase().includes(query) ||
          (item.script.author && item.script.author.toLowerCase().includes(query)) ||
          item.script.cast.some((c) => c.name.toLowerCase().includes(query))
        );
      });
    }

    return items;
  };

  const handleDeleteDraft = (id: string) => {
    if (confirm('Delete this draft?')) {
      deleteDraft(id);
      setDrafts(getDrafts());
    }
  };

  const handleDeleteScript = async (id: string) => {
    if (!confirm('Delete this script? This cannot be undone.')) return;

    try {
      const res = await fetch(`${getConfig().serverUrl}/api/scripts/${id}`, {
        method: 'DELETE',
      });
      if (res.ok) {
        setScripts(scripts.filter((s) => s.id !== id));
      }
    } catch (err) {
      console.error('Failed to delete script:', err);
    }
  };

  const handleScriptClick = (script: Play, isDraftScript: boolean) => {
    if (isDraftScript) {
      router.push(`/scripts/new?draft=${script.id}`);
    } else {
      router.push(`/scripts/${script.id}`);
    }
  };

  const items = filteredItems();

  return (
    <main className="min-h-screen stage-bg">
      {/* Header */}
      <header className="border-b border-[var(--gold)]/20 backdrop-blur-sm sticky top-0 z-10 bg-[var(--black)]/80">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <Link
                href="/"
                className="text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors"
              >
                ← Home
              </Link>
              <h1 className="font-[Cormorant_Garamond] text-2xl text-[var(--cream)]">
                Scripts Library
              </h1>
            </div>
            <Link href="/scripts/new" className="btn-elegant px-5 py-2.5 rounded-lg font-medium">
              ✍️ Create Script
            </Link>
          </div>
        </div>
      </header>

      <div className="container mx-auto px-4 py-8">
        {/* Search and Filter */}
        <div className="flex flex-col md:flex-row gap-4 mb-8">
          {/* Search */}
          <div className="flex-1">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search by title, author, character..."
              className="input-elegant w-full px-4 py-3 rounded-lg text-[var(--cream)]"
            />
          </div>

          {/* Filter Tabs */}
          <div className="flex gap-2">
            {(
              [
                { value: 'all', label: 'All' },
                { value: 'sample', label: 'Samples' },
                { value: 'published', label: 'Published' },
                { value: 'drafts', label: 'My Drafts' },
              ] as { value: FilterType; label: string }[]
            ).map(({ value, label }) => (
              <button
                key={value}
                onClick={() => setFilter(value)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  filter === value
                    ? 'bg-[var(--gold)] text-[var(--black)]'
                    : 'bg-[var(--charcoal)]/50 text-[var(--cream)] hover:bg-[var(--charcoal)]'
                }`}
              >
                {label}
              </button>
            ))}
          </div>
        </div>

        {/* Scripts Grid */}
        {isLoading ? (
          <div className="text-center py-20">
            <div className="text-[var(--gold)] text-lg">Loading scripts...</div>
          </div>
        ) : items.length === 0 ? (
          <div className="text-center py-20">
            <div className="text-[var(--cream)] opacity-50 text-lg mb-4">
              {searchQuery ? 'No scripts found matching your search.' : 'No scripts found.'}
            </div>
            {filter === 'drafts' && (
              <Link
                href="/scripts/new"
                className="text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors"
              >
                Create your first script →
              </Link>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {items.map((item) => {
              return (
                <ScriptCard
                  key={item.script.id}
                  script={item.script}
                  isDraft={item.isDraft}
                  onClick={() => handleScriptClick(item.script, item.isDraft)}
                  onDelete={
                    item.isDraft
                      ? () => handleDeleteDraft(item.script.id)
                      : (item.script as ScriptWithMeta).isUserScript
                        ? () => handleDeleteScript(item.script.id)
                        : undefined
                  }
                />
              );
            })}
          </div>
        )}

        {/* Ad Banner */}
        <div className="mt-12">
          <AdBannerPlaceholder />
        </div>
      </div>
    </main>
  );
}
