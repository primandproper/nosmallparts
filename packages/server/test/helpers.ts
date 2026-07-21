import type { Context } from 'hono';
import type { WSEvents } from 'hono/ws';
import { ulid } from 'ulid';
import type { AppVariables } from '../src/app.js';
import { createStageWsHandler } from '../src/ws/stage-handler.js';

export const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

// ---------------------------------------------------------------------------
// HTTP helpers (real Better Auth against the test DB)
// ---------------------------------------------------------------------------

let userCounter = 0;

export interface TestUser {
  email: string;
  password: string;
  name: string;
  cookie: string;
  id: string;
}

/** Register a fresh user through Better Auth and return its session cookie. */
export async function signUpUser(
  app: { request: (path: string, init?: RequestInit) => Promise<Response> },
  opts: { email?: string; name?: string; password?: string } = {}
): Promise<TestUser> {
  const email = opts.email ?? `user${++userCounter}@test.example`;
  const password = opts.password ?? 'password-123';
  const name = opts.name ?? `User ${userCounter}`;
  const res = await app.request('/api/auth/sign-up/email', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Origin: 'http://localhost:8081' },
    body: JSON.stringify({ email, password, name }),
  });
  if (res.status >= 400) {
    throw new Error(`sign-up failed (${res.status}): ${await res.text()}`);
  }
  const cookie = extractCookies(res);
  const meRes = await app.request('/api/me', { headers: { Cookie: cookie } });
  const me = (await meRes.json()) as { user: { id: string } | null };
  if (!me.user) throw new Error('sign-up did not produce a session');
  return { email, password, name, cookie, id: me.user.id };
}

export function extractCookies(res: Response): string {
  const setCookies = res.headers.getSetCookie?.() ?? [];
  return setCookies.map((c) => c.split(';')[0]).join('; ');
}

export function jsonInit(body: unknown, cookie?: string): RequestInit {
  return {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(cookie ? { Cookie: cookie } : {}),
    },
    body: JSON.stringify(body),
  };
}

export const SAMPLE_SCRIPT_INPUT = {
  title: 'Two Hander',
  description: 'A short scene',
  author: 'Test Author',
  isPublic: true,
  cast: [
    { id: 'a', name: 'Alice' },
    { id: 'b', name: 'Bob' },
  ],
  lines: [
    { castMemberIds: ['a'], text: 'Line one', index: 0 },
    { castMemberIds: ['b'], text: 'Line two', index: 1 },
    { castMemberIds: ['a'], text: 'Line three', index: 2 },
  ],
};

// ---------------------------------------------------------------------------
// WebSocket harness — drives the WSEvents object returned by the handler
// directly, with fake sockets. Covers the full server logic without TCP.
// ---------------------------------------------------------------------------

export interface WsIdentity {
  userId?: string | null;
  guestId?: string | null;
  ip?: string;
}

function fakeWsContext(vars: WsIdentity): Context<{ Variables: AppVariables }> {
  const headers = new Headers();
  if (vars.guestId) headers.set('cookie', `nsp_guest_id=${vars.guestId}`);
  if (vars.ip) headers.set('x-forwarded-for', vars.ip);
  const user = vars.userId ? { id: vars.userId, email: `${vars.userId}@ws.example` } : null;
  return {
    get: (key: string) => (key === 'user' ? user : null),
    req: {
      query: () => undefined,
      header: (name: string) => headers.get(name) ?? undefined,
      raw: { headers },
    },
    env: {},
  } as unknown as Context<{ Variables: AppVariables }>;
}

export class WsClient {
  received: Array<Record<string, unknown>> = [];
  closedByServer = false;
  private events: WSEvents | null = null;
  private closed = false;
  private wsCtx: { send: (d: string) => void; close: () => void };

  constructor(public identity: WsIdentity) {
    this.wsCtx = {
      send: (data: string) => {
        this.received.push(JSON.parse(data));
      },
      close: () => {
        // Server-initiated close (kick/ban/displace): run onClose like a real
        // socket teardown would.
        this.closedByServer = true;
        if (!this.closed) {
          this.closed = true;
          void this.events?.onClose?.(
            undefined as never,
            this.wsCtx as never
          );
        }
      },
    };
  }

  async connect(): Promise<void> {
    const handler = createStageWsHandler();
    this.events = await handler(fakeWsContext(this.identity));
  }

  async send(msg: Record<string, unknown>): Promise<void> {
    await this.sendRaw(JSON.stringify(msg));
  }

  async sendRaw(data: string): Promise<void> {
    await this.events?.onMessage?.({ data } as never, this.wsCtx as never);
  }

  /** Client-initiated disconnect. Awaits the server's onClose handling. */
  async disconnect(): Promise<void> {
    if (this.closed) return;
    this.closed = true;
    await this.events?.onClose?.(undefined as never, this.wsCtx as never);
  }

  async join(stageIdOrCode: string, name: string): Promise<void> {
    await this.send({ type: 'join', stageId: stageIdOrCode, participantName: name });
  }

  /** Messages of a given type, in arrival order. */
  ofType<T = Record<string, unknown>>(type: string): T[] {
    return this.received.filter((m) => m.type === type) as T[];
  }

  last<T = Record<string, unknown>>(type: string): T | undefined {
    const all = this.ofType<T>(type);
    return all[all.length - 1];
  }

  get participantId(): string | undefined {
    return (this.last<{ participantId: string }>('joined') ?? {}).participantId;
  }

  lastState(): Record<string, unknown> | undefined {
    const states = this.received.filter((m) => m.type === 'stageState' || m.type === 'joined');
    const last = states[states.length - 1] as
      | { type: string; state?: Record<string, unknown>; stageState?: Record<string, unknown> }
      | undefined;
    return last?.state ?? last?.stageState;
  }

  errors(): string[] {
    return this.ofType<{ message: string }>('error').map((e) => e.message);
  }

  systemMessages(): string[] {
    return this.ofType<{ message: { text: string } }>('systemMessage').map((m) => m.message.text);
  }
}

export function guestIdentity(ip?: string): WsIdentity {
  return { guestId: ulid(), ip };
}

export async function makeStage(playScriptId?: string | null): Promise<{ id: string; code: string }> {
  const { createStage } = await import('../src/repos/stages.js');
  return createStage(playScriptId ?? null);
}

export async function makePlay(ownerUserId: string) {
  const { createPlayScript } = await import('../src/repos/play-scripts.js');
  return createPlayScript(ownerUserId, SAMPLE_SCRIPT_INPUT);
}

/** Connect a WsClient and join a stage in one call. */
export async function joinClient(
  identity: WsIdentity,
  stageId: string,
  name: string
): Promise<WsClient> {
  const c = new WsClient(identity);
  await c.connect();
  await c.join(stageId, name);
  return c;
}

/**
 * WS-test users need a real `user` row (stage_participants.user_id and
 * stages.host_user_id have FK constraints).
 */
export async function userIdentity(ip?: string): Promise<WsIdentity> {
  const { db } = await import('../src/db/index.js');
  const { user } = await import('../src/db/schema.js');
  const id = `user_${ulid()}`;
  const now = new Date();
  await db.insert(user).values({
    id,
    name: id,
    email: `${id}@ws.example`,
    emailVerified: false,
    createdAt: now,
    updatedAt: now,
  });
  return { userId: id, ip };
}
