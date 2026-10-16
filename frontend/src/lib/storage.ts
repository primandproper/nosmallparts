/**
 * localStorage, read and written so that a browser refusing it (a private
 * window, storage cleared or blocked) degrades to having remembered nothing
 * rather than to a broken page.
 */

export function readItem(key: string): string | null {
  try {
    return window.localStorage.getItem(key);
  } catch {
    return null;
  }
}

export function writeItem(key: string, value: string): void {
  try {
    window.localStorage.setItem(key, value);
  } catch {
    // Remembering is a convenience; nothing depends on it having worked.
  }
}

export function readJSON<T>(key: string, fallback: T): T {
  const raw = readItem(key);
  if (raw === null) return fallback;

  try {
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

export function writeJSON(key: string, value: unknown): void {
  writeItem(key, JSON.stringify(value));
}
