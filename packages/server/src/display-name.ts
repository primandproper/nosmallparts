import { Filter } from 'bad-words';

export const MAX_DISPLAY_NAME_LENGTH = 32;

const RESERVED_NAMES = new Set(['host', 'admin', 'director', 'system', 'server']);

const profanity = new Filter();

// bad-words only matches whole words; also catch strong profanity embedded in
// a name ("fucknugget"). Curated to unambiguous strings — substring-matching
// the full bad-words list flags ordinary names ("spaced" contains "spac").
const embeddedWords = [
  'fuck',
  'shit',
  'cunt',
  'nigger',
  'nigga',
  'faggot',
  'bitch',
  'asshole',
  'wanker',
  'dickhead',
  'cocksuck',
  'motherfuck',
];

function isProfaneName(name: string): boolean {
  try {
    if (profanity.isProfane(name)) return true;
  } catch {
    // bad-words can throw on exotic input; fall through to substring check
  }
  const squashed = name.toLowerCase().replace(/[^a-z]/g, '');
  return embeddedWords.some((w) => squashed.includes(w));
}

export type DisplayNameResult =
  | { ok: true; name: string }
  | { ok: false; error: string };

/**
 * Display-name policy (§12.4): trimmed, non-empty, ≤32 chars, not a reserved
 * token, and passes the profanity filter. Returns the cleaned name.
 */
export function validateDisplayName(raw: string): DisplayNameResult {
  const name = raw.trim().replace(/\s+/g, ' ');
  if (!name) return { ok: false, error: 'Please enter a name.' };
  if (name.length > MAX_DISPLAY_NAME_LENGTH) {
    return { ok: false, error: `Name is too long (max ${MAX_DISPLAY_NAME_LENGTH} characters).` };
  }
  if (RESERVED_NAMES.has(name.toLowerCase())) {
    return { ok: false, error: 'Please choose a different name.' };
  }
  if (isProfaneName(name)) return { ok: false, error: 'Please choose a different name.' };
  return { ok: true, name };
}

/**
 * Per-stage uniqueness (case-insensitive). On collision the name is
 * auto-suffixed " (2)", " (3)", … keeping the total within the length cap.
 */
export function dedupeDisplayName(name: string, takenNames: string[]): string {
  const taken = new Set(takenNames.map((n) => n.trim().toLowerCase()));
  if (!taken.has(name.toLowerCase())) return name;
  for (let n = 2; ; n++) {
    const suffix = ` (${n})`;
    const base = name.slice(0, MAX_DISPLAY_NAME_LENGTH - suffix.length).trimEnd();
    const candidate = `${base}${suffix}`;
    if (!taken.has(candidate.toLowerCase())) return candidate;
  }
}
