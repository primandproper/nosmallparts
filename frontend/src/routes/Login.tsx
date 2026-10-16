import { Code } from '@primandproper/platform-client';
import { type FormEvent, useState } from 'react';
import { useNavigate } from 'react-router';

import { inputClass, Screen, Spinner } from '../components/Screen';
import { useAuth } from '../lib/auth';
import { hasCode, messageOf } from '../lib/errors';
import { signIn } from '../lib/session';

export default function Login() {
  const navigate = useNavigate();
  const { refresh } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setBusy(true);
    try {
      await signIn(email, password);
      await refresh();
      navigate('/', { replace: true });
    } catch (e) {
      setError(hasCode(e, Code.UNAUTHENTICATED) ? 'Wrong email or password.' : messageOf(e, 'Sign-in failed'));
    } finally {
      setBusy(false);
    }
  }

  return (
    <Screen title="Sign in">
      <div className="flex flex-1 items-center justify-center px-6">
        <form onSubmit={onSubmit} className="flex w-full max-w-sm flex-col gap-3">
          <h2 className="mb-4 text-center font-serif text-2xl text-stage-gold">Sign in</h2>
          <input
            className={inputClass}
            placeholder="Email"
            type="email"
            autoCapitalize="none"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
          />
          <input
            className={inputClass}
            placeholder="Password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          {error ? <p className="text-red-400">{error}</p> : null}
          <button
            type="submit"
            className="flex items-center justify-center gap-2 rounded-md bg-stage-gold px-4 py-3 font-semibold text-stage-bg disabled:opacity-50"
            disabled={busy || !email || !password}
          >
            {busy ? <Spinner dark /> : null}
            {busy ? 'Signing in…' : 'Sign in'}
          </button>
        </form>
      </div>
    </Screen>
  );
}
