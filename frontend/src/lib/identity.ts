/**
 * ensureGuestIdentity makes sure this browser holds a guest identity before it
 * opens a stage socket: the server mints the nsp_guest_id cookie on its first
 * answer, and every socket and video token after that carries it. A signed-in
 * user is answered as themselves and minted nothing.
 */
export async function ensureGuestIdentity(): Promise<void> {
  const res = await fetch('/api/identity', { credentials: 'same-origin' });
  if (!res.ok) throw new Error(`Could not establish an identity (${res.status}).`);
}
