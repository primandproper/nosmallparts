'use client';

import { useState } from 'react';

interface RatingModalProps {
  playId: string;
  playTitle: string;
  onSubmit: (rating: number) => Promise<void>;
  onDismiss: () => void;
  initialRating?: number;
}

export function RatingModal({
  playId,
  playTitle,
  onSubmit,
  onDismiss,
  initialRating = 0,
}: RatingModalProps) {
  const [rating, setRating] = useState(initialRating);
  const [hoverRating, setHoverRating] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const displayRating = hoverRating || rating;

  const handleSubmit = async () => {
    if (rating < 1 || rating > 5) return;
    setIsSubmitting(true);
    setError(null);
    try {
      await onSubmit(rating);
      onDismiss();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit rating');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
      onClick={(e) => e.target === e.currentTarget && onDismiss()}
    >
      <div
        className="card-elegant rounded-2xl p-6 max-w-md w-full space-y-6"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="font-[Cormorant_Garamond] text-2xl font-bold text-gold-gradient text-center">
          Rate this play
        </h3>
        <p className="text-[var(--cream)] opacity-70 text-center text-sm">{playTitle}</p>

        {/* Star rating */}
        <div className="flex justify-center gap-2">
          {[1, 2, 3, 4, 5].map((value) => (
            <button
              key={value}
              type="button"
              className="text-3xl transition-transform hover:scale-110 focus:outline-none focus:ring-2 focus:ring-[var(--gold)] rounded"
              onMouseEnter={() => setHoverRating(value)}
              onMouseLeave={() => setHoverRating(0)}
              onClick={() => setRating(value)}
              aria-label={`${value} star${value > 1 ? 's' : ''}`}
            >
              {value <= displayRating ? (
                <span className="text-[var(--gold)]">★</span>
              ) : (
                <span className="text-[var(--cream)]/40">☆</span>
              )}
            </button>
          ))}
        </div>

        {error && (
          <p className="text-red-400 text-sm text-center">{error}</p>
        )}

        <div className="flex flex-col gap-2">
          <button
            onClick={handleSubmit}
            disabled={rating < 1 || rating > 5 || isSubmitting}
            className="btn-primary w-full py-3 rounded-xl disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isSubmitting ? 'Submitting...' : 'Submit Rating'}
          </button>
          <button
            onClick={onDismiss}
            disabled={isSubmitting}
            className="btn-secondary w-full py-2 rounded-xl text-sm"
          >
            Skip
          </button>
        </div>
      </div>
    </div>
  );
}
