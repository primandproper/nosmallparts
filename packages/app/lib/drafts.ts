import AsyncStorage from '@react-native-async-storage/async-storage';
import type { Play } from '@nsp/shared';

const DRAFTS_KEY = 'nsp:drafts';

function generateDraftId(): string {
  return `draft-${Date.now()}-${Math.random().toString(36).slice(2, 11)}`;
}

export function isDraftId(id: string): boolean {
  return id.startsWith('draft-');
}

export async function getDrafts(): Promise<Play[]> {
  try {
    const raw = await AsyncStorage.getItem(DRAFTS_KEY);
    return raw ? (JSON.parse(raw) as Play[]) : [];
  } catch {
    return [];
  }
}

export async function getDraft(id: string): Promise<Play | null> {
  const drafts = await getDrafts();
  return drafts.find((d) => d.id === id) ?? null;
}

export async function saveDraft(play: Partial<Play> & { title: string }): Promise<Play> {
  const drafts = await getDrafts();
  const existingIndex = play.id ? drafts.findIndex((d) => d.id === play.id) : -1;

  const saved: Play = {
    id: play.id || generateDraftId(),
    title: play.title,
    description: play.description ?? '',
    ...(play.author !== undefined ? { author: play.author } : {}),
    cast: play.cast ?? [],
    lines: play.lines ?? [],
  };

  if (existingIndex >= 0) drafts[existingIndex] = saved;
  else drafts.push(saved);

  await AsyncStorage.setItem(DRAFTS_KEY, JSON.stringify(drafts));
  return saved;
}

export async function deleteDraft(id: string): Promise<boolean> {
  const drafts = await getDrafts();
  const next = drafts.filter((d) => d.id !== id);
  if (next.length === drafts.length) return false;
  await AsyncStorage.setItem(DRAFTS_KEY, JSON.stringify(next));
  return true;
}
