'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';

interface LoginResponse {
  success: boolean;
  error?: string;
  requiresTOTP?: boolean;
  userId?: string;
}

export default function LoginPage() {
  const router = useRouter();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [totpToken, setTotpToken] = useState('');
  const [showTOTP, setShowTOTP] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username,
          password,
          totpToken: showTOTP ? totpToken : undefined,
        }),
      });

      const data: LoginResponse = await response.json();

      if (data.success) {
        // Redirect to home page on success
        router.push('/');
      } else if (data.requiresTOTP) {
        // Show TOTP input
        setShowTOTP(true);
        setError('');
      } else {
        setError(data.error || 'Login failed');
      }
    } catch {
      setError('Failed to connect to server. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="card-elegant rounded-2xl p-8 animate-fade-up animate-delay-1">
      <h2 className="text-2xl font-semibold text-[var(--cream)] text-center mb-6">Welcome Back</h2>

      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Username Input */}
        <div className="animate-fade-up animate-delay-2">
          <label
            htmlFor="username"
            className="block text-[var(--gold)] text-sm font-medium mb-2 tracking-wide uppercase"
          >
            Username
          </label>
          <input
            type="text"
            id="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            className="input-elegant w-full px-4 py-3 rounded-xl text-[var(--cream)]"
            placeholder="Enter your username"
            autoComplete="username"
            disabled={isLoading}
          />
        </div>

        {/* Password Input */}
        <div className="animate-fade-up animate-delay-3">
          <label
            htmlFor="password"
            className="block text-[var(--gold)] text-sm font-medium mb-2 tracking-wide uppercase"
          >
            Password
          </label>
          <input
            type="password"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="input-elegant w-full px-4 py-3 rounded-xl text-[var(--cream)]"
            placeholder="Enter your password"
            autoComplete="current-password"
            disabled={isLoading}
          />
        </div>

        {/* TOTP Input - only shown when required */}
        {showTOTP && (
          <div className="animate-fade-up">
            <label
              htmlFor="totpToken"
              className="block text-[var(--gold)] text-sm font-medium mb-2 tracking-wide uppercase"
            >
              Two-Factor Code
            </label>
            <input
              type="text"
              id="totpToken"
              value={totpToken}
              onChange={(e) => setTotpToken(e.target.value.replace(/\D/g, '').slice(0, 6))}
              className="input-elegant w-full px-4 py-3 rounded-xl text-[var(--cream)] font-mono tracking-widest text-center text-xl"
              placeholder="000000"
              maxLength={6}
              autoComplete="one-time-code"
              disabled={isLoading}
            />
            <p className="text-[var(--cream)] opacity-50 text-sm mt-2">
              Enter the 6-digit code from your authenticator app
            </p>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/30 rounded-xl px-4 py-3 text-red-400 text-sm text-center">
            {error}
          </div>
        )}

        {/* Submit Button */}
        <div className="animate-fade-up animate-delay-4 pt-2">
          <button
            type="submit"
            disabled={isLoading || !username || !password || (showTOTP && totpToken.length !== 6)}
            className="btn-primary w-full py-3 rounded-xl relative"
          >
            <span>{isLoading ? 'Signing in...' : 'Sign In'}</span>
          </button>
        </div>
      </form>

      {/* Links */}
      <div className="mt-6 space-y-3 text-center animate-fade-up animate-delay-4">
        <p className="text-[var(--cream)] opacity-60 text-sm">
          Don&apos;t have an account?{' '}
          <Link
            href="/auth/register"
            className="text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors"
          >
            Create one
          </Link>
        </p>
        <p className="text-[var(--cream)] opacity-40 text-xs">
          <Link href="/auth/forgot-password" className="hover:text-[var(--gold)] transition-colors">
            Forgot your password?
          </Link>
        </p>
        <p className="text-[var(--cream)] opacity-40 text-xs">
          <Link href="/terms" className="hover:text-[var(--gold)] transition-colors">
            Terms of Service
          </Link>
          {' · '}
          <Link href="/privacy" className="hover:text-[var(--gold)] transition-colors">
            Privacy Policy
          </Link>
        </p>
      </div>
    </div>
  );
}
