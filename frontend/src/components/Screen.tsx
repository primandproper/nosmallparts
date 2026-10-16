import { type ReactNode, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router';

/**
 * Screen is a page with the header the Expo stack drew: a back link everywhere
 * but home, the screen's title, and an optional action on the right.
 */
export function Screen({ title, right, children }: { title: string; right?: ReactNode; children: ReactNode }) {
  const navigate = useNavigate();
  const location = useLocation();
  const isHome = location.pathname === '/';

  useEffect(() => {
    document.title = isHome ? 'No Small Parts' : `${title} · No Small Parts`;
  }, [title, isHome]);

  return (
    <div className="flex min-h-dvh flex-col bg-stage-bg">
      <header className="flex h-14 shrink-0 items-center gap-3 border-b border-stage-border px-4">
        {isHome ? null : (
          <button
            type="button"
            onClick={() => (window.history.length > 1 ? navigate(-1) : navigate('/'))}
            className="text-stage-gold"
            aria-label="Back"
          >
            ‹ Back
          </button>
        )}
        <h1 className="flex-1 truncate font-semibold text-stage-gold">
          {isHome ? (
            <Link to="/" className="hover:underline">
              {title}
            </Link>
          ) : (
            title
          )}
        </h1>
        {right}
      </header>
      <main className="flex flex-1 flex-col">{children}</main>
    </div>
  );
}

/** Spinner is the activity indicator, in the stage gold unless told otherwise. */
export function Spinner({ dark = false, className = '' }: { dark?: boolean; className?: string }) {
  return (
    <span
      role="status"
      aria-label="Loading"
      className={`inline-block h-4 w-4 animate-spin rounded-full border-2 border-t-transparent ${
        dark ? 'border-stage-bg' : 'border-stage-gold'
      } ${className}`}
    />
  );
}

/** Centered fills the page with one thing in its middle. */
export function Centered({ children }: { children: ReactNode }) {
  return <div className="flex flex-1 items-center justify-center px-6 py-16">{children}</div>;
}

/** inputClass is the bordered text field every form uses. */
export const inputClass =
  'w-full rounded-md border border-stage-border bg-stage-surface px-3 py-3 text-stage-text placeholder:text-stage-muted';
