'use client';
import { StageState, getCastMemberById } from '@/lib/types';
import { RatingModal } from '@/components/RatingModal';

interface DiscussionProps {
  stageState: StageState;
  onResetPlay: () => void;
  playId?: string | null;
  playTitle?: string;
  onRatingComplete?: () => void;
  isAuthenticated?: boolean;
}

export function Discussion({
  stageState,
  onResetPlay,
  playId,
  playTitle,
  onRatingComplete,
  isAuthenticated,
}: DiscussionProps) {
  const handleSubmitRating = async (rating: number) => {
    if (!playId) return;
    const res = await fetch(`/api/play-scripts/${playId}/rate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ rating }),
      credentials: 'include',
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data?.error || 'Failed to submit rating');
    }
  };
  return (
    <div className="card-elegant rounded-2xl p-6 space-y-6">
      {/* Header */}
      <div className="text-center">
        <div className="text-5xl mb-4">🎭</div>
        <h2 className="font-[Cormorant_Garamond] text-3xl font-bold text-gold-gradient mb-2">
          Bravo!
        </h2>
        <p className="text-[var(--cream)] opacity-60">A wonderful performance. Take a bow!</p>
      </div>

      {/* Decorative line */}
      <div className="decorative-line" />

      {/* Cast List */}
      {stageState.selectedPlay && (
        <div>
          <h3 className="text-[var(--gold)] text-xs font-medium mb-3 tracking-wider uppercase text-center">
            Tonight&apos;s Cast
          </h3>
          <div className="space-y-2">
            {(stageState.participants || [])
              .filter((p) => (p.castMemberIds || []).length > 0)
              .map((participant) => (
                <div
                  key={participant.id}
                  className="flex justify-between items-center p-3 rounded-xl bg-[var(--charcoal)]/50"
                >
                  <span className="text-[var(--cream)]">{participant.name}</span>
                  <div className="text-right">
                    {(participant.castMemberIds || [])
                      .map((id) => getCastMemberById(stageState.selectedPlay!, id))
                      .filter(Boolean)
                      .map((cm) => (
                        <div key={cm!.id}>
                          <span className="text-[var(--gold)] text-sm">{cm!.name}</span>
                          {cm!.description && (
                            <p className="text-[var(--cream)] opacity-50 text-xs">
                              {cm!.description}
                            </p>
                          )}
                        </div>
                      ))}
                  </div>
                </div>
              ))}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="space-y-3 pt-2">
        {playId && isAuthenticated && (
          <RatingModal
            playId={playId}
            playTitle={playTitle || stageState.selectedPlay?.title || 'This play'}
            onSubmit={handleSubmitRating}
            onDismiss={onRatingComplete ?? (() => {})}
          />
        )}
        <button onClick={onResetPlay} className="btn-primary w-full py-4 rounded-xl text-lg">
          <span>🎭 Perform Again</span>
        </button>
        <button onClick={onResetPlay} className="btn-secondary w-full py-3 rounded-xl">
          Return to Lobby
        </button>
      </div>
    </div>
  );
}
