import type { CastMember, PlayLine } from '../generated/nosmallparts/scripts/v1/scripts';
import { readJSON, writeJSON } from './storage';

const DRAFTS_KEY = 'nsp:drafts';

/** Draft is a script being written that has not been saved to the server. */
export interface Draft {
  id: string;
  title: string;
  description: string;
  author?: string;
  cast: CastMember[];
  lines: PlayLine[];
}

export function isDraftId(id: string): boolean {
  return id.startsWith('draft-');
}

function newDraftId(): string {
  return `draft-${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;
}

export function getDrafts(): Draft[] {
  return readJSON<Draft[]>(DRAFTS_KEY, []);
}

export function getDraft(id: string): Draft | null {
  return getDrafts().find((d) => d.id === id) ?? null;
}

/** saveDraft writes a draft, minting its ID the first time, and answers it. */
export function saveDraft(draft: Omit<Draft, 'id'> & { id?: string }): Draft {
  const drafts = getDrafts();
  const saved: Draft = { ...draft, id: draft.id || newDraftId() };

  const existing = drafts.findIndex((d) => d.id === saved.id);
  if (existing >= 0) drafts[existing] = saved;
  else drafts.push(saved);

  writeJSON(DRAFTS_KEY, drafts);

  return saved;
}

export function deleteDraft(id: string): void {
  writeJSON(
    DRAFTS_KEY,
    getDrafts().filter((d) => d.id !== id),
  );
}
