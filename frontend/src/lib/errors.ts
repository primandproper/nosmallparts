import { Code, isTransient, PlatformError } from '@primandproper/platform-client';

/** messageOf is what to show a person for a failed call. */
export function messageOf(err: unknown, fallback = 'Something went wrong.'): string {
  if (err instanceof PlatformError) {
    if (isTransient(err)) return "Can't reach the server. Check your connection and try again.";

    return err.serverMessage || fallback;
  }

  if (err instanceof Error && err.message) return err.message;

  return fallback;
}

/** hasCode reports whether a call failed with the given gRPC code. */
export function hasCode(err: unknown, code: Code): boolean {
  return err instanceof PlatformError && err.code === code;
}
