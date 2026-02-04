'use client';

import { useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';

interface RegisterResponse {
  success: boolean;
  error?: string;
  userId?: string;
  twoFactorQrCode?: string;
  twoFactorSecret?: string;
}

function RegisterForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const invitationToken = searchParams.get('invitation') || '';

  const [formData, setFormData] = useState({
    username: '',
    password: '',
    confirmPassword: '',
    emailAddress: '',
    firstName: '',
    lastName: '',
    birthday: '',
    acceptedTos: false,
    acceptedPrivacyPolicy: false,
  });
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [registrationComplete, setRegistrationComplete] = useState(false);
  const [twoFactorData, setTwoFactorData] = useState<{
    qrCode: string;
    secret: string;
  } | null>(null);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
    setError('');
  };

  const validateForm = (): string | null => {
    if (!formData.username.trim()) return 'Username is required';
    if (formData.username.length < 3) return 'Username must be at least 3 characters';
    if (!formData.emailAddress.trim()) return 'Email is required';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.emailAddress)) return 'Invalid email format';
    if (!formData.firstName.trim()) return 'First name is required';
    if (!formData.lastName.trim()) return 'Last name is required';
    if (!formData.password) return 'Password is required';
    if (formData.password.length < 8) return 'Password must be at least 8 characters';
    if (formData.password !== formData.confirmPassword) return 'Passwords do not match';
    if (!formData.acceptedTos) return 'You must accept the Terms of Service';
    if (!formData.acceptedPrivacyPolicy) return 'You must accept the Privacy Policy';
    return null;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const validationError = validateForm();
    if (validationError) {
      setError(validationError);
      return;
    }

    setError('');
    setIsLoading(true);

    try {
      const response = await fetch('/api/auth/register', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: formData.username,
          password: formData.password,
          emailAddress: formData.emailAddress,
          firstName: formData.firstName,
          lastName: formData.lastName,
          birthday: formData.birthday || undefined,
          invitationToken: invitationToken || undefined,
          acceptedTos: formData.acceptedTos,
          acceptedPrivacyPolicy: formData.acceptedPrivacyPolicy,
        }),
      });

      const data: RegisterResponse = await response.json();

      if (data.success) {
        // Show 2FA setup if provided
        if (data.twoFactorQrCode && data.twoFactorSecret) {
          setTwoFactorData({
            qrCode: data.twoFactorQrCode,
            secret: data.twoFactorSecret,
          });
        }
        setRegistrationComplete(true);
      } else {
        setError(data.error || 'Registration failed');
      }
    } catch {
      setError('Failed to connect to server. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleContinueToLogin = () => {
    router.push('/auth/login');
  };

  // Show 2FA setup after successful registration
  if (registrationComplete) {
    return (
      <div className="card-elegant rounded-2xl p-8 animate-fade-up">
        <div className="text-center mb-6">
          <div className="text-4xl mb-4">✅</div>
          <h2 className="text-2xl font-semibold text-[var(--cream)] mb-2">
            Registration Complete!
          </h2>
          <p className="text-[var(--cream)] opacity-60">
            Your account has been created successfully.
          </p>
        </div>

        {twoFactorData && (
          <div className="mb-6 p-4 bg-[var(--black)] rounded-xl border border-[var(--gold)]/20">
            <h3 className="text-[var(--gold)] font-medium mb-3 text-center">
              Set Up Two-Factor Authentication
            </h3>
            <p className="text-[var(--cream)] opacity-60 text-sm mb-4 text-center">
              Scan this QR code with your authenticator app to enable 2FA for your account.
            </p>

            {/* QR Code Display */}
            <div className="flex justify-center mb-4">
              {twoFactorData.qrCode.startsWith('data:image') ? (
                // If it's a data URL, display as image
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={twoFactorData.qrCode}
                  alt="2FA QR Code"
                  className="w-48 h-48 bg-white p-2 rounded-lg"
                />
              ) : (
                // Otherwise show as base64 image
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={`data:image/png;base64,${twoFactorData.qrCode}`}
                  alt="2FA QR Code"
                  className="w-48 h-48 bg-white p-2 rounded-lg"
                />
              )}
            </div>

            {/* Manual Secret */}
            <div className="text-center">
              <p className="text-[var(--cream)] opacity-50 text-xs mb-2">
                Or enter this code manually:
              </p>
              <code className="block bg-[var(--black-lighter)] px-4 py-2 rounded-lg font-mono text-sm text-[var(--gold)] select-all">
                {twoFactorData.secret}
              </code>
            </div>
          </div>
        )}

        <button onClick={handleContinueToLogin} className="btn-primary w-full py-3 rounded-xl">
          <span>Continue to Login</span>
        </button>

        <p className="text-center text-[var(--cream)] opacity-40 text-xs mt-4">
          {twoFactorData
            ? "You'll need to verify your 2FA after logging in."
            : 'Please check your email to verify your account.'}
        </p>
      </div>
    );
  }

  return (
    <div className="card-elegant rounded-2xl p-8 animate-fade-up animate-delay-1">
      <h2 className="text-2xl font-semibold text-[var(--cream)] text-center mb-6">
        Create Your Account
      </h2>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Username */}
        <div>
          <label
            htmlFor="username"
            className="block text-[var(--gold)] text-xs font-medium mb-1.5 tracking-wide uppercase"
          >
            Username
          </label>
          <input
            type="text"
            id="username"
            name="username"
            value={formData.username}
            onChange={handleChange}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="Choose a unique username"
            autoComplete="username"
            disabled={isLoading}
          />
        </div>

        {/* Password */}
        <div>
          <label
            htmlFor="password"
            className="block text-[var(--gold)] text-xs font-medium mb-1.5 tracking-wide uppercase"
          >
            Password
          </label>
          <input
            type="password"
            id="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="At least 8 characters"
            autoComplete="new-password"
            disabled={isLoading}
          />
        </div>

        {/* Confirm Password */}
        <div>
          <label
            htmlFor="confirmPassword"
            className="block text-[var(--gold)] text-xs font-medium mb-1.5 tracking-wide uppercase"
          >
            Confirm Password
          </label>
          <input
            type="password"
            id="confirmPassword"
            name="confirmPassword"
            value={formData.confirmPassword}
            onChange={handleChange}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="Re-enter your password"
            autoComplete="new-password"
            disabled={isLoading}
          />
        </div>

        {/* Email */}
        <div>
          <label
            htmlFor="emailAddress"
            className="block text-[var(--gold)] text-xs font-medium mb-1.5 tracking-wide uppercase"
          >
            Email Address
          </label>
          <input
            type="email"
            id="emailAddress"
            name="emailAddress"
            value={formData.emailAddress}
            onChange={handleChange}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="you@example.com"
            autoComplete="email"
            disabled={isLoading}
          />
        </div>
        
        {/* Name Fields */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label
              htmlFor="firstName"
              className="block text-[var(--gold)] text-xs font-medium mb-1.5 tracking-wide uppercase"
            >
              First Name
            </label>
            <input
              type="text"
              id="firstName"
              name="firstName"
              value={formData.firstName}
              onChange={handleChange}
              className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
              placeholder="Jane"
              disabled={isLoading}
            />
          </div>
          <div>
            <label
              htmlFor="lastName"
              className="block text-[var(--gold)] text-xs font-medium mb-1.5 tracking-wide uppercase"
            >
              Last Name
            </label>
            <input
              type="text"
              id="lastName"
              name="lastName"
              value={formData.lastName}
              onChange={handleChange}
              className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
              placeholder="Doe"
              disabled={isLoading}
            />
          </div>
        </div>

        {/* Birthday (optional) */}
        <div>
          <label
            htmlFor="birthday"
            className="block text-[var(--gold)] text-xs font-medium mb-1.5 tracking-wide uppercase"
          >
            Birthday <span className="text-[var(--cream)] opacity-40 normal-case">(optional)</span>
          </label>
          <input
            type="date"
            id="birthday"
            name="birthday"
            value={formData.birthday}
            onChange={handleChange}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            disabled={isLoading}
          />
        </div>

        {/* Terms and Privacy */}
        <div className="space-y-3 pt-2">
          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              name="acceptedTos"
              checked={formData.acceptedTos}
              onChange={handleChange}
              className="mt-1 w-4 h-4 rounded border-[var(--gold)]/30 bg-[var(--black)] text-[var(--gold)] focus:ring-[var(--gold)] focus:ring-offset-0"
              disabled={isLoading}
            />
            <span className="text-[var(--cream)] opacity-70 text-sm">
              I accept the{' '}
              <Link href="/terms" className="text-[var(--gold)] hover:underline">
                Terms of Service
              </Link>
            </span>
          </label>

          <label className="flex items-start gap-3 cursor-pointer">
            <input
              type="checkbox"
              name="acceptedPrivacyPolicy"
              checked={formData.acceptedPrivacyPolicy}
              onChange={handleChange}
              className="mt-1 w-4 h-4 rounded border-[var(--gold)]/30 bg-[var(--black)] text-[var(--gold)] focus:ring-[var(--gold)] focus:ring-offset-0"
              disabled={isLoading}
            />
            <span className="text-[var(--cream)] opacity-70 text-sm">
              I accept the{' '}
              <Link href="/privacy" className="text-[var(--gold)] hover:underline">
                Privacy Policy
              </Link>
            </span>
          </label>
        </div>

        {/* Error Message */}
        {error && (
          <div className="bg-red-500/10 border border-red-500/30 rounded-xl px-4 py-3 text-red-400 text-sm text-center">
            {error}
          </div>
        )}

        {/* Submit Button */}
        <div className="pt-2">
          <button
            type="submit"
            disabled={isLoading}
            className="btn-primary w-full py-3 rounded-xl relative"
          >
            <span>{isLoading ? 'Creating Account...' : 'Create Account'}</span>
          </button>
        </div>
      </form>

      {/* Login Link */}
      <p className="text-center text-[var(--cream)] opacity-60 text-sm mt-6">
        Already have an account?{' '}
        <Link
          href="/auth/login"
          className="text-[var(--gold)] hover:text-[var(--gold-light)] transition-colors"
        >
          Sign in
        </Link>
      </p>
    </div>
  );
}

export default function RegisterPage() {
  return (
    <Suspense
      fallback={
        <div className="card-elegant rounded-2xl p-8 animate-fade-up">
          <div className="text-center text-[var(--cream)]">Loading...</div>
        </div>
      }
    >
      <RegisterForm />
    </Suspense>
  );
}
