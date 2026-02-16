'use client';

import { Play } from '@/lib/types';

interface ScriptCardProps {
  script: Play & { isUserScript?: boolean };
  isDraft?: boolean;
  onClick?: () => void;
  onDelete?: () => void;
}

export function ScriptCard({ script, isDraft = false, onClick, onDelete }: ScriptCardProps) {
  const lineCount = (script.lines || []).length;
  const castCount = (script.cast || []).length;

  // Determine badge type
  const getBadge = () => {
    if (isDraft) {
      return { text: 'Draft', className: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30' };
    }
    if (script.isUserScript) {
      return { text: 'Published', className: 'bg-green-500/20 text-green-400 border-green-500/30' };
    }
    return { text: 'Sample', className: 'bg-purple-500/20 text-purple-400 border-purple-500/30' };
  };

  const badge = getBadge();

  return (
    <div
      className={`card-elegant rounded-xl p-5 transition-all hover:border-[var(--gold)]/30 ${
        onClick ? 'cursor-pointer' : ''
      }`}
      onClick={onClick}
    >
      {/* Header */}
      <div className="flex items-start justify-between gap-3 mb-3">
        <div className="flex-1 min-w-0">
          <h3 className="font-[Cormorant_Garamond] text-xl font-semibold text-[var(--cream)] truncate">
            {script.title}
          </h3>
          {script.author && (
            <p className="text-[var(--gold)] text-sm opacity-80">by {script.author}</p>
          )}
        </div>

        {/* Badge */}
        <span
          className={`px-2 py-1 rounded-full text-xs font-medium border shrink-0 ${badge.className}`}
        >
          {badge.text}
        </span>
      </div>

      {/* Description */}
      {script.description && (
        <p className="text-[var(--cream)] opacity-60 text-sm line-clamp-2 mb-4">
          {script.description}
        </p>
      )}

      {/* Stats */}
      <div className="flex items-center gap-4 text-sm flex-wrap">
        <div className="flex items-center gap-1.5 text-[var(--cream)] opacity-50">
          <span>👥</span>
          <span>
            {castCount} role{castCount !== 1 ? 's' : ''}
          </span>
        </div>
        <div className="flex items-center gap-1.5 text-[var(--cream)] opacity-50">
          <span>📜</span>
          <span>
            {lineCount} line{lineCount !== 1 ? 's' : ''}
          </span>
        </div>
        {script.ratingStats && script.ratingStats.ratingCount > 0 && (
          <div className="flex items-center gap-1.5 text-[var(--gold)] opacity-80">
            <span>★</span>
            <span>
              {script.ratingStats.averageRating.toFixed(1)} ({script.ratingStats.ratingCount})
            </span>
            {typeof script.userRating === 'number' && script.userRating > 0 && (
              <span className="text-[var(--cream)] opacity-60 text-xs">
                · you: {script.userRating}★
              </span>
            )}
          </div>
        )}
      </div>

      {/* Cast preview */}
      {castCount > 0 && (
        <div className="mt-3 pt-3 border-t border-[var(--cream)]/10">
          <p className="text-[var(--cream)] opacity-40 text-xs">
            {(script.cast || []).map((c) => c.name).join(', ')}
          </p>
        </div>
      )}

      {/* Delete button for drafts/user scripts */}
      {onDelete && (isDraft || script.isUserScript) && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="mt-3 text-red-400 hover:text-red-300 text-xs opacity-60 hover:opacity-100 transition-opacity"
        >
          Delete
        </button>
      )}
    </div>
  );
}
