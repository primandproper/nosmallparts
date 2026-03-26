// HTTP API client for stage actions.
// All actions that were previously WebSocket messages are now HTTP POST/GET calls.

function getApiUrl(): string {
  const envUrl = typeof process !== 'undefined' ? process.env.NEXT_PUBLIC_API_URL : undefined;
  if (envUrl) return envUrl;

  if (typeof window === 'undefined') return 'http://localhost:8000';

  const { protocol, hostname } = window.location;
  // In production, the API is at api.nosmall.parts; in dev, same host port 8000
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return `${protocol}//${hostname}:8000`;
  }
  return `${protocol}//api.${hostname.replace(/^www\./, '')}`;
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const baseUrl = getApiUrl();
  const res = await fetch(`${baseUrl}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });

  if (!res.ok) {
    const body = await res.text();
    let message: string;
    try {
      message = JSON.parse(body).error || body;
    } catch {
      message = body;
    }
    throw new Error(message);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export interface JoinStageResponse {
  participantId: string;
  stageState: any;
  availablePlays: any[];
  pusherKey: string;
  pusherCluster: string;
}

export function joinStage(stageId: string, participantName: string): Promise<JoinStageResponse> {
  return apiFetch('/api/stages/join', {
    method: 'POST',
    body: JSON.stringify({ stageId, participantName }),
  });
}

export function leaveStage(stageId: string, participantId: string): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/leave`, {
    method: 'POST',
    body: JSON.stringify({ participantId }),
  });
}

export function toggleCastMember(stageId: string, participantId: string, castMemberId: string): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/toggle-cast-member`, {
    method: 'POST',
    body: JSON.stringify({ participantId, castMemberId }),
  });
}

export function setDirector(stageId: string, participantId: string, isDirector: boolean): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/set-director`, {
    method: 'POST',
    body: JSON.stringify({ participantId, isDirector }),
  });
}

export function setAudience(stageId: string, participantId: string, isAudience: boolean): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/set-audience`, {
    method: 'POST',
    body: JSON.stringify({ participantId, isAudience }),
  });
}

export function setReady(stageId: string, participantId: string, isReady: boolean): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/set-ready`, {
    method: 'POST',
    body: JSON.stringify({ participantId, isReady }),
  });
}

export function selectPlay(stageId: string, playId: string): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/select-play`, {
    method: 'POST',
    body: JSON.stringify({ playId }),
  });
}

export function sendChat(stageId: string, participantId: string, text: string, castOnly: boolean): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/chat`, {
    method: 'POST',
    body: JSON.stringify({ participantId, text, castOnly }),
  });
}

export function startPlay(stageId: string): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/start`, {
    method: 'POST',
    body: JSON.stringify({}),
  });
}

export function nextLine(stageId: string): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/next-line`, {
    method: 'POST',
    body: JSON.stringify({}),
  });
}

export function resetPlay(stageId: string): Promise<void> {
  return apiFetch(`/api/stages/${stageId}/reset`, {
    method: 'POST',
    body: JSON.stringify({}),
  });
}

export function getStageState(stageId: string): Promise<any> {
  return apiFetch(`/api/stages/${stageId}/state`);
}
