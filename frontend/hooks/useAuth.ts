'use client';

import { useState, useEffect, useCallback } from 'react';

interface SessionData {
  authenticated: boolean;
  userId?: string;
  accountId?: string;
}

interface UseAuthReturn {
  isLoading: boolean;
  isAuthenticated: boolean;
  userId: string | null;
  accountId: string | null;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}

export function useAuth(): UseAuthReturn {
  const [session, setSession] = useState<SessionData | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refreshSession = useCallback(async () => {
    try {
      const response = await fetch('/api/auth/session');
      const data: SessionData = await response.json();
      setSession(data);
    } catch {
      setSession({ authenticated: false });
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshSession();
  }, [refreshSession]);

  const logout = useCallback(async () => {
    try {
      await fetch('/api/auth/logout', { method: 'POST' });
      setSession({ authenticated: false });
    } catch {
      // Still clear local state even if the API call fails
      setSession({ authenticated: false });
    }
  }, []);

  return {
    isLoading,
    isAuthenticated: session?.authenticated ?? false,
    userId: session?.userId ?? null,
    accountId: session?.accountId ?? null,
    logout,
    refreshSession,
  };
}
