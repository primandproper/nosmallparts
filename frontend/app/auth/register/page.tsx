'use client';

import { useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';

interface RegisterResponse {
  success: boolean;
  error?: string;
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
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [touched, setTouched] = useState<Record<string, boolean>>({});
  const [isLoading, setIsLoading] = useState(false);

  const validateField = (name: string, value: string, allData?: typeof formData): string => {
    const data = allData || formData;
    switch (name) {
      case 'username':
        if (!value.trim()) return 'Username is required';
        if (value.length < 3) return 'Must be at least 3 characters';
        return '';
      case 'emailAddress':
        if (!value.trim()) return 'Email is required';
        if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) return 'Invalid email format';
        return '';
      case 'password':
        if (!value) return 'Password is required';
        if (value.length < 8) return 'Must be at least 8 characters';
        if (!/[A-Z]/.test(value)) return 'Must contain at least one uppercase letter';
        if (!/[0-9]/.test(value)) return 'Must contain at least one number';
        if (!/[^A-Za-z0-9]/.test(value)) return 'Must contain at least one special character';
        return '';
      case 'confirmPassword':
        if (!value) return 'Please confirm your password';
        if (value !== data.password) return 'Passwords do not match';
        return '';
      case 'firstName':
        if (!value.trim()) return 'First name is required';
        return '';
      case 'lastName':
        if (!value.trim()) return 'Last name is required';
        return '';
      default:
        return '';
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    const newValue = type === 'checkbox' ? checked : value;

    const updated = { ...formData, [name]: newValue };
    setFormData(updated);
    setError('');

    // Re-validate touched fields (outside of setFormData to avoid nested state updates)
    const errorUpdates: Record<string, string> = {};

    if (name === 'password' && touched['confirmPassword'] && updated.confirmPassword) {
      errorUpdates['confirmPassword'] = validateField(
        'confirmPassword',
        updated.confirmPassword,
        updated
      );
    }

    if (touched[name] && type !== 'checkbox') {
      errorUpdates[name] = validateField(name, value);
    }

    if (Object.keys(errorUpdates).length > 0) {
      setFieldErrors((prev) => ({ ...prev, ...errorUpdates }));
    }
  };

  const handleBlur = (e: React.FocusEvent<HTMLInputElement>) => {
    const { name, value } = e.target;
    setTouched((prev) => ({ ...prev, [name]: true }));
    const fieldError = validateField(name, value);
    setFieldErrors((prev) => ({ ...prev, [name]: fieldError }));
  };

  const validateForm = (): string | null => {
    if (!formData.username.trim()) return 'Username is required';
    if (formData.username.length < 3) return 'Must be at least 3 characters';
    if (!formData.emailAddress.trim()) return 'Email is required';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.emailAddress)) return 'Invalid email format';
    if (!formData.firstName.trim()) return 'First name is required';
    if (!formData.lastName.trim()) return 'Last name is required';
    if (!formData.password) return 'Password is required';
    if (formData.password.length < 8) return 'Must be at least 8 characters';
    if (!/[A-Z]/.test(formData.password)) return 'Must contain at least one uppercase letter';
    if (!/[0-9]/.test(formData.password)) return 'Must contain at least one number';
    if (!/[^A-Za-z0-9]/.test(formData.password))
      return 'Must contain at least one special character';
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
        router.push('/auth/login');
      } else {
        setError(data.error || 'Registration failed');
      }
    } catch {
      setError('Failed to connect to server. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

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
            onBlur={handleBlur}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="Choose a unique username"
            autoComplete="username"
            disabled={isLoading}
          />
          {touched.username && fieldErrors.username && (
            <p className="text-red-400 text-xs mt-1">{fieldErrors.username}</p>
          )}
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
            onBlur={handleBlur}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="you@example.com"
            autoComplete="email"
            disabled={isLoading}
          />
          {touched.emailAddress && fieldErrors.emailAddress && (
            <p className="text-red-400 text-xs mt-1">{fieldErrors.emailAddress}</p>
          )}
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
              onBlur={handleBlur}
              className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
              placeholder="Jane"
              disabled={isLoading}
            />
            {touched.firstName && fieldErrors.firstName && (
              <p className="text-red-400 text-xs mt-1">{fieldErrors.firstName}</p>
            )}
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
              onBlur={handleBlur}
              className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
              placeholder="Doe"
              disabled={isLoading}
            />
            {touched.lastName && fieldErrors.lastName && (
              <p className="text-red-400 text-xs mt-1">{fieldErrors.lastName}</p>
            )}
          </div>
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
            onBlur={handleBlur}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="Strong password required"
            autoComplete="new-password"
            disabled={isLoading}
          />
          {touched.password && fieldErrors.password && (
            <p className="text-red-400 text-xs mt-1">{fieldErrors.password}</p>
          )}
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
            onBlur={handleBlur}
            className="input-elegant w-full px-3 py-2.5 rounded-lg text-[var(--cream)] text-sm"
            placeholder="Re-enter your password"
            autoComplete="new-password"
            disabled={isLoading}
          />
          {touched.confirmPassword && fieldErrors.confirmPassword && (
            <p className="text-red-400 text-xs mt-1">{fieldErrors.confirmPassword}</p>
          )}
          {touched.confirmPassword && !fieldErrors.confirmPassword && formData.confirmPassword && (
            <p className="text-green-400 text-xs mt-1">Passwords match</p>
          )}
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
