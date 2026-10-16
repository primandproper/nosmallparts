import {
  Code,
  createGrpcWebTransport,
  methodsOf,
  NotSignedInError,
  PlatformError,
  refreshCookieCredentialStore,
  Session,
  signIn as platformSignIn,
  signOut as platformSignOut,
  type SessionState,
} from '@primandproper/platform-client';
import { SignInServiceDefinition } from '@primandproper/platform-client/signin/v1';
import type { User } from '@primandproper/platform-client/identity/v1';

/**
 * The one login this tab holds. The server is this page's own origin, where it
 * mounts gRPC-web under /grpc; the refresh token lives in an httpOnly cookie
 * the server sets on that path, and the page keeps only the access token, in
 * memory, plus a hint in localStorage that a cookie should exist.
 */
export const session = new Session({
  transport: createGrpcWebTransport({ baseUrl: `${window.location.origin}/grpc` }),
  store: refreshCookieCredentialStore(),
});

export const SignInService = methodsOf(SignInServiceDefinition);

export type { SessionState, User };

/**
 * register creates an account and signs into it. The email address is the
 * username: the form asks for a name, an email and a password, as it always
 * has, and the name is the user's first name.
 */
export async function register(input: { name: string; email: string; password: string }): Promise<void> {
  const email = input.email.trim();
  const name = input.name.trim();

  await session.callAnonymous(SignInService.register, {
    user: { username: email, emailAddress: email, firstName: name, lastName: '' },
    account: { name, timeZone: '', billingAddress: undefined },
    password: input.password,
    invitation: undefined,
    agreements: [],
  });

  await signIn(email, input.password);
}

/** signIn signs in with an email address and password. */
export async function signIn(email: string, password: string): Promise<void> {
  const result = await platformSignIn(session, { handle: { emailAddress: email.trim() }, password });
  if (result.kind === 'second_factor_required') {
    // Nobody here enrolls a second factor, so a server asking for one is a
    // server this app does not know how to talk to.
    throw new Error('This account needs a second factor, which this app does not support.');
  }
}

/** signOut ends the login on the server and here. It never rejects. */
export function signOut(): Promise<void> {
  return platformSignOut(session);
}

/**
 * currentUser is who is signed in, or null for nobody. It always asks: a
 * session starts `anonymous` after a reload until its first call reads the
 * stored hint and exchanges the cookie, and holding nothing at all is a
 * NotSignedInError raised here, without a request.
 */
export async function currentUser(): Promise<User | null> {
  try {
    const res = await session.call(SignInService.getSelf, {});
    return res.user ?? null;
  } catch (err) {
    if (err instanceof NotSignedInError) return null;
    if (err instanceof PlatformError && err.code === Code.UNAUTHENTICATED) return null;
    throw err;
  }
}

/**
 * accessToken is a current access token for a signed-in session, refreshing
 * first if need be, or null for nobody. The stage socket carries it, since a
 * WebSocket cannot carry an Authorization header.
 */
export async function accessToken(): Promise<string | null> {
  // A call is what refreshes a token near expiry, or exchanges the cookie
  // after a reload; GetSelf is the cheapest one that needs a caller.
  if ((await currentUser()) === null) return null;

  const held = await session.held();

  return held?.token || null;
}
