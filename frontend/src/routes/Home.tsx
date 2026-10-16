import { Code, NotSignedInError } from '@primandproper/platform-client';
import { type FormEvent, useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router';

import { inputClass, Screen, Spinner } from '../components/Screen';
import { displayName, useAuth } from '../lib/auth';
import { hasCode, messageOf } from '../lib/errors';
import { getStoredPlayerName, storePlayerName } from '../lib/player-name';
import { createStage, stageExists } from '../lib/rpc';
import { signOut } from '../lib/session';

export default function Home() {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const script = params.get('s');
  const { state, user, refresh } = useAuth();

  const [name, setName] = useState(() => getStoredPlayerName() ?? '');
  const [code, setCode] = useState('');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (user && !name) setName(displayName(user));
  }, [user, name]);

  async function onEnter(event: FormEvent) {
    event.preventDefault();
    if (!name.trim()) {
      setError('Enter your name');
      return;
    }
    storePlayerName(name.trim());
    setBusy(true);
    setError(null);
    try {
      const joinCode = code.trim().toUpperCase();
      if (joinCode) {
        const res = await stageExists(joinCode);
        if (!res.exists) {
          setError('Stage not found. Leave code blank to create one.');
          return;
        }
        navigate(`/stage/${res.code}`, { replace: true });
      } else {
        const res = await createStage(script ?? undefined);
        navigate(`/stage/${res.code}`, { replace: true });
      }
    } catch (e) {
      // Creating a stage needs an account (§1.4); joining one does not.
      setError(
        e instanceof NotSignedInError || hasCode(e, Code.UNAUTHENTICATED)
          ? 'Sign in to create a stage, or enter a code to join one as a guest.'
          : messageOf(e, 'Failed'),
      );
    } finally {
      setBusy(false);
    }
  }

  async function onSignOut() {
    await signOut();
    await refresh();
  }

  return (
    <Screen title="No Small Parts">
      <div className="flex flex-1 flex-col items-center px-6 pt-16 lg:flex-row lg:items-center lg:justify-center lg:gap-16 lg:px-12 lg:pt-0">
        <div className="lg:max-w-md lg:flex-1">
          <h2 className="mb-2 text-center font-serif text-4xl text-stage-gold lg:text-left lg:text-6xl">
            No Small Parts
          </h2>
          <p className="mb-10 text-center text-stage-muted lg:mb-0 lg:text-left lg:text-lg">
            Perform short scripts together in real time.
          </p>
        </div>

        <form onSubmit={onEnter} className="flex w-full max-w-sm flex-col gap-3 lg:max-w-md">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Your name"
            className={inputClass}
          />
          <input
            value={code}
            onChange={(e) => setCode(e.target.value.toUpperCase())}
            placeholder="Stage code (leave blank to create)"
            autoCapitalize="characters"
            className={inputClass}
          />
          {error ? <p className="text-sm text-red-400">{error}</p> : null}
          <button
            type="submit"
            disabled={busy}
            className={`flex items-center justify-center gap-2 rounded-md px-4 py-3 font-semibold text-stage-bg ${
              busy ? 'bg-stage-goldDim' : 'bg-stage-gold'
            }`}
          >
            {busy ? <Spinner dark /> : null}
            {busy ? 'Working…' : code ? 'Join stage' : 'Create stage'}
          </button>

          <div className="mt-6 flex flex-col gap-2">
            <Link to="/scripts" className="rounded-md border border-stage-border px-4 py-3 text-center text-stage-text">
              Browse scripts
            </Link>
            {state === 'unreachable' ? (
              // A login the server could not be asked about is not a
              // signed-out one: offering a sign-in form here would mislead.
              <button
                type="button"
                className="rounded-md border border-stage-border px-4 py-2 text-center text-sm text-stage-muted"
                onClick={() => void refresh()}
              >
                Couldn't reach the server · Retry
              </button>
            ) : user === undefined ? null : user ? (
              <button
                type="button"
                className="rounded-md border border-stage-border px-4 py-2 text-center text-sm text-stage-muted"
                onClick={onSignOut}
              >
                Signed in as {user.emailAddress} · Sign out
              </button>
            ) : (
              <div className="flex gap-2">
                <Link
                  to="/login"
                  className="flex-1 rounded-md border border-stage-border px-3 py-2 text-center text-stage-text"
                >
                  Sign in
                </Link>
                <Link
                  to="/register"
                  className="flex-1 rounded-md border border-stage-border px-3 py-2 text-center text-stage-text"
                >
                  Register
                </Link>
              </div>
            )}
          </div>
        </form>
      </div>
    </Screen>
  );
}
