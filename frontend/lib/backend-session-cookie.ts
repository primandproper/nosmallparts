import { cookies } from 'next/headers';
import { createCipheriv, createDecipheriv, randomBytes } from 'crypto';

const BACKEND_SESSION_COOKIE_NAME = 'nosmallparts_backend_session';
const ALGORITHM = 'aes-256-gcm';
const IV_LENGTH = 16;
const AUTH_TAG_LENGTH = 16;

export interface BackendCookiePayload {
  backendCookie: string;
  exp?: number;
}

function getEncryptionKey(): Buffer {
  const raw = process.env.SESSION_COOKIE_ENCRYPTION_KEY;
  if (!raw?.trim()) {
    throw new Error(
      'SESSION_COOKIE_ENCRYPTION_KEY is not set. Set it in .env (see env.example).'
    );
  }
  const trimmed = raw.trim();
  if (trimmed.length >= 64 && /^[0-9a-fA-F]+$/.test(trimmed)) {
    return Buffer.from(trimmed.slice(0, 64), 'hex');
  }
  return Buffer.from(trimmed, 'base64');
}

/**
 * Encrypt a payload and return base64 string (iv + authTag + ciphertext).
 */
export async function encryptBackendCookiePayload(
  payload: BackendCookiePayload
): Promise<string> {
  const key = getEncryptionKey();
  if (key.length !== 32) {
    throw new Error('SESSION_COOKIE_ENCRYPTION_KEY must decode to 32 bytes.');
  }
  const iv = randomBytes(IV_LENGTH);
  const cipher = createCipheriv(ALGORITHM, key, iv, { authTagLength: AUTH_TAG_LENGTH });
  const plain = JSON.stringify(payload);
  const enc = Buffer.concat([cipher.update(plain, 'utf8'), cipher.final()]);
  const authTag = cipher.getAuthTag();
  return Buffer.concat([iv, authTag, enc]).toString('base64');
}

/**
 * Decrypt and parse payload. Returns null on any error.
 */
export async function decryptBackendCookiePayload(
  encryptedBase64: string
): Promise<BackendCookiePayload | null> {
  try {
    const key = getEncryptionKey();
    if (key.length !== 32) return null;
    const buf = Buffer.from(encryptedBase64, 'base64');
    if (buf.length < IV_LENGTH + AUTH_TAG_LENGTH) return null;
    const iv = buf.subarray(0, IV_LENGTH);
    const authTag = buf.subarray(IV_LENGTH, IV_LENGTH + AUTH_TAG_LENGTH);
    const ciphertext = buf.subarray(IV_LENGTH + AUTH_TAG_LENGTH);
    const decipher = createDecipheriv(ALGORITHM, key, iv, { authTagLength: AUTH_TAG_LENGTH });
    decipher.setAuthTag(authTag);
    const plain = decipher.update(ciphertext).toString('utf8') + decipher.final('utf8');
    return JSON.parse(plain) as BackendCookiePayload;
  } catch {
    return null;
  }
}

export async function setEncryptedBackendCookie(
  value: string,
  expiresAt: Date
): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.set(BACKEND_SESSION_COOKIE_NAME, value, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    path: '/',
    expires: expiresAt,
  });
}

export async function getEncryptedBackendCookie(): Promise<string | null> {
  const cookieStore = await cookies();
  const cookie = cookieStore.get(BACKEND_SESSION_COOKIE_NAME);
  return cookie?.value ?? null;
}

export async function clearEncryptedBackendCookie(): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.delete(BACKEND_SESSION_COOKIE_NAME);
}

/**
 * Read BFF cookie, decrypt, and return backend cookie value for X-Session-Cookie header.
 * On failure (missing, invalid, tampered), clears the BFF cookie and returns null.
 */
export async function getBackendCookieForRequest(): Promise<string | null> {
  const raw = await getEncryptedBackendCookie();
  if (!raw) return null;
  const payload = await decryptBackendCookiePayload(raw);
  if (!payload?.backendCookie) {
    await clearEncryptedBackendCookie();
    return null;
  }
  if (payload.exp != null && payload.exp < Date.now() / 1000) {
    await clearEncryptedBackendCookie();
    return null;
  }
  return payload.backendCookie;
}
