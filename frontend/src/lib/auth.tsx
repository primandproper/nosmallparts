import { createContext, type ReactNode, useCallback, useContext, useEffect, useState } from 'react';

import { currentUser, session, type SessionState, type User } from './session';

interface Auth {
  /** state is the session's: `unreachable` is a login the server could not be asked about. */
  state: SessionState;
  /** user is who is signed in, or null for nobody; undefined until first known. */
  user: User | null | undefined;
  /** refresh reads who is signed in again, after a sign-in, a sign-out or a retry. */
  refresh: () => Promise<void>;
}

const AuthContext = createContext<Auth | null>(null);

/** AuthProvider tracks the session and who it is signed in as. */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<SessionState>(session.state);
  const [user, setUser] = useState<User | null | undefined>(undefined);

  const refresh = useCallback(async () => {
    try {
      setUser(await currentUser());
    } catch {
      // The session says why, as `unreachable`; the user is not known.
      setUser(null);
    }
  }, []);

  useEffect(() => {
    const stop = session.onStateChange((next) => {
      setState(next);

      if (next === 'anonymous') setUser(null);
    });

    void refresh();

    return stop;
  }, [refresh]);

  return <AuthContext.Provider value={{ state, user, refresh }}>{children}</AuthContext.Provider>;
}

export function useAuth(): Auth {
  const auth = useContext(AuthContext);
  if (!auth) throw new Error('useAuth needs an AuthProvider above it');

  return auth;
}

/** displayName is how a user is addressed: their first name, or their address. */
export function displayName(user: User): string {
  return user.firstName || user.emailAddress;
}
