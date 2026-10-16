import { Code } from '@primandproper/platform-client';
import { type FormEvent, useState } from 'react';
import { useNavigate } from 'react-router';

import { inputClass, Screen, Spinner } from '../components/Screen';
import { useAuth } from '../lib/auth';
import { hasCode, messageOf } from '../lib/errors';
import { register } from '../lib/session';

const MIN_PASSWORD_LENGTH = 8;

export default function Register() {
  const navigate = useNavigate();
  const { refresh } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setBusy(true);
    try {
      await register({ name, email, password });
      await refresh();
      navigate('/', { replace: true });
    } catch (e) {
      setError(
        hasCode(e, Code.ALREADY_EXISTS) ? 'An account with that email already exists.' : messageOf(e, 'Sign-up failed'),
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <Screen title="Create account">
      <div className="flex flex-1 items-center justify-center px-6">
        <form onSubmit={onSubmit} className="flex w-full max-w-sm flex-col gap-3">
          <h2 className="mb-4 text-center font-serif text-2xl text-stage-gold">Create account</h2>
          <input className={inputClass} placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} />
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
            placeholder={`Password (min ${MIN_PASSWORD_LENGTH} chars)`}
            type="password"
            autoComplete="new-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
          {error ? <p className="text-red-400">{error}</p> : null}
          <button
            type="submit"
            className="flex items-center justify-center gap-2 rounded-md bg-stage-gold px-4 py-3 font-semibold text-stage-bg disabled:opacity-50"
            disabled={busy || !name || !email || password.length < MIN_PASSWORD_LENGTH}
          >
            {busy ? <Spinner dark /> : null}
            {busy ? 'Creating…' : 'Create account'}
          </button>
        </form>
      </div>
    </Screen>
  );
}
