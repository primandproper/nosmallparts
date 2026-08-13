import { describe, expect, it } from 'vitest';
import type { Participant, Play, StageState } from '@nsp/shared';
import {
  guestIdentity,
  joinClient,
  makePlay,
  makeStage,
  sleep,
  userIdentity,
  WsClient,
} from './helpers.js';

const COUNTDOWN_WAIT = 120; // NSP_COUNTDOWN_MS=40 in setup

function participants(state: Record<string, unknown> | undefined): Participant[] {
  return (state?.participants ?? []) as Participant[];
}

describe('WS join basics (§3.1, §8)', () => {
  it('rejects a connection with no identity at join time', async () => {
    const stage = await makeStage();
    const c = new WsClient({});
    await c.connect();
    await c.join(stage.id, 'NoIdentity');
    expect(c.errors().some((e) => e.includes('Identity required'))).toBe(true);
    expect(c.closedByServer).toBe(true);
  });

  it('rejects malformed JSON without crashing, keeps connection open', async () => {
    const stage = await makeStage();
    const c = await joinClient(guestIdentity(), stage.id, 'Guesty');
    await c.sendRaw('{not json');
    expect(c.errors()).toContain('Invalid JSON');
    // Connection still usable afterwards.
    await c.send({ type: 'heartbeat' });
    expect(c.ofType('stageState').length).toBeGreaterThan(0);
  });

  it('rejects unknown message types but keeps the connection open', async () => {
    const stage = await makeStage();
    const c = await joinClient(guestIdentity(), stage.id, 'Guesty');
    await c.send({ type: 'flyToTheMoon' });
    expect(c.errors().some((e) => e.includes('Unknown message type'))).toBe(true);
    expect(c.closedByServer).toBe(false);
  });

  it('rejects messages before join', async () => {
    const c = new WsClient(guestIdentity());
    await c.connect();
    await c.send({ type: 'heartbeat' });
    expect(c.errors()).toContain('Must join before sending messages');
  });

  it('joining a non-existent code reports Stage not found', async () => {
    const c = new WsClient(guestIdentity());
    await c.connect();
    await c.join('ZZZZ99', 'Nobody');
    expect(c.errors()).toContain('Stage not found');
  });

  it('join by code is case-insensitive', async () => {
    const stage = await makeStage();
    const c = await joinClient(guestIdentity(), stage.code.toLowerCase(), 'CaseTest');
    expect(c.participantId).toBeTruthy();
  });

  it('enforces MAX_PARTICIPANTS_PER_STAGE = 12', async () => {
    const stage = await makeStage();
    const clients: WsClient[] = [];
    for (let i = 0; i < 12; i++) {
      clients.push(await joinClient(guestIdentity(), stage.id, `Player ${i}`));
    }
    expect(clients.every((c) => c.participantId)).toBe(true);
    const thirteenth = new WsClient(guestIdentity());
    await thirteenth.connect();
    await thirteenth.join(stage.id, 'Unlucky');
    expect(thirteenth.errors().some((e) => e.includes('Stage is full'))).toBe(true);
    expect(thirteenth.closedByServer).toBe(true);
  });
});

describe('display-name policy (§12.4)', () => {
  it('blocks empty names', async () => {
    const stage = await makeStage();
    const c = new WsClient(guestIdentity());
    await c.connect();
    await c.join(stage.id, '   ');
    expect(c.errors()).toContain('Please enter a name.');
  });

  it('blocks names over 32 chars', async () => {
    const stage = await makeStage();
    const c = new WsClient(guestIdentity());
    await c.connect();
    await c.join(stage.id, 'x'.repeat(33));
    expect(c.errors().some((e) => e.includes('too long'))).toBe(true);
  });

  it('blocks reserved tokens case-insensitively', async () => {
    const stage = await makeStage();
    for (const name of ['Host', 'ADMIN', 'director', 'System', 'sErVeR']) {
      const c = new WsClient(guestIdentity());
      await c.connect();
      await c.join(stage.id, name);
      expect(c.errors()).toContain('Please choose a different name.');
    }
  });

  it('blocks profanity', async () => {
    const stage = await makeStage();
    const c = new WsClient(guestIdentity());
    await c.connect();
    await c.join(stage.id, 'shithead');
    expect(c.errors()).toContain('Please choose a different name.');
  });

  it('auto-suffixes duplicate names case-insensitively', async () => {
    const stage = await makeStage();
    await joinClient(guestIdentity(), stage.id, 'Sam');
    const second = await joinClient(guestIdentity(), stage.id, 'sam');
    const third = await joinClient(guestIdentity(), stage.id, 'SAM');
    const names = participants(third.lastState()).map((p) => p.name);
    expect(names).toContain('Sam');
    expect(names).toContain('sam (2)');
    expect(names).toContain('SAM (3)');
    expect(second.participantId).toBeTruthy();
  });
});

describe('same identity, multiple connections (§3.2)', () => {
  it('second tab displaces the first; participant appears once', async () => {
    const stage = await makeStage();
    const identity = guestIdentity();
    const tab1 = await joinClient(identity, stage.id, 'TabUser');
    const tab2 = await joinClient(identity, stage.id, 'TabUser');
    expect(tab1.ofType('sessionReplaced')).toHaveLength(1);
    expect(tab1.closedByServer).toBe(true);
    const ps = participants(tab2.lastState()).filter((p) => p.name === 'TabUser');
    expect(ps).toHaveLength(1);
    expect(tab2.participantId).toBe(tab1.participantId);
  });

  it('reconnect after drop reuses the participant row and keeps roles', async () => {
    const owner = await userIdentity();
    const play = await makePlay(owner.userId!);
    const stage = await makeStage(play.id);
    const identity = guestIdentity();
    const first = await joinClient(identity, stage.id, 'Reconnector');
    const cast0 = play.cast[0].id;
    await first.send({ type: 'toggleCastMember', castMemberId: cast0 });
    await first.disconnect();

    const again = await joinClient(identity, stage.id, 'Reconnector');
    expect(again.participantId).toBe(first.participantId);
    const me = participants(again.lastState()).find((p) => p.id === again.participantId);
    expect(me?.castMemberIds).toContain(cast0);
  });
});

describe('lobby actions (§3.3)', () => {
  async function lobbySetup() {
    const owner = await userIdentity();
    const play = await makePlay(owner.userId!);
    const stage = await makeStage(play.id);
    const host = await joinClient(owner, stage.id, 'Hosty');
    const other = await joinClient(guestIdentity(), stage.id, 'Other');
    return { play, stage, host, other };
  }

  it('claim/release/steal roles; doubling allowed; audience clears roles', async () => {
    const { play, host, other } = await lobbySetup();
    const [a, b] = play.cast.map((c) => c.id);

    // Claim.
    await host.send({ type: 'toggleCastMember', castMemberId: a });
    let me = participants(host.lastState()).find((p) => p.id === host.participantId);
    expect(me?.castMemberIds).toEqual([a]);

    // Doubling: same participant can hold multiple roles.
    await host.send({ type: 'toggleCastMember', castMemberId: b });
    me = participants(host.lastState()).find((p) => p.id === host.participantId);
    expect(me?.castMemberIds?.sort()).toEqual([a, b].sort());

    // Steal: exclusive across participants.
    await other.send({ type: 'toggleCastMember', castMemberId: a });
    const stolen = participants(other.lastState());
    expect(stolen.find((p) => p.id === other.participantId)?.castMemberIds).toEqual([a]);
    expect(stolen.find((p) => p.id === host.participantId)?.castMemberIds).toEqual([b]);

    // Click again -> release.
    await other.send({ type: 'toggleCastMember', castMemberId: a });
    expect(
      participants(other.lastState()).find((p) => p.id === other.participantId)?.castMemberIds
    ).toEqual([]);

    // Toggling audience clears roles.
    await host.send({ type: 'setAudience', isAudience: true });
    expect(
      participants(host.lastState()).find((p) => p.id === host.participantId)?.castMemberIds
    ).toEqual([]);
  });

  it('selecting a different play clears all roles for everyone', async () => {
    const { play, host, other } = await lobbySetup();
    const owner2 = await userIdentity();
    const play2 = await makePlay(owner2.userId!);
    await host.send({ type: 'toggleCastMember', castMemberId: play.cast[0].id });
    await other.send({ type: 'toggleCastMember', castMemberId: play.cast[1].id });
    await host.send({ type: 'selectPlay', playId: play2.id });
    const ps = participants(host.lastState());
    expect(ps.every((p) => p.castMemberIds.length === 0)).toBe(true);
    expect((host.lastState()?.selectedPlay as Play | null)?.id).toBe(play2.id);
  });

  it('ready toggle is visible to others in real time', async () => {
    const { host, other } = await lobbySetup();
    await other.send({ type: 'setReady', isReady: true });
    const seen = participants(host.lastState()).find((p) => p.id === other.participantId);
    expect(seen?.isReady).toBe(true);
  });
});

describe('start-play preconditions (§3.4) and flow (§3.5, §3.6, §3.7)', () => {
  async function readyStage() {
    const owner = await userIdentity();
    const play = await makePlay(owner.userId!);
    const stage = await makeStage(play.id);
    const host = await joinClient(owner, stage.id, 'Hosty');
    const partner = await joinClient(guestIdentity(), stage.id, 'Partner');
    const [a, b] = play.cast.map((c) => c.id);
    await host.send({ type: 'toggleCastMember', castMemberId: a });
    await partner.send({ type: 'toggleCastMember', castMemberId: b });
    return { play, stage, host, partner, a, b };
  }

  it('rejects start with unfilled roles and unready performers, with details', async () => {
    const owner = await userIdentity();
    const play = await makePlay(owner.userId!);
    const stage = await makeStage(play.id);
    const host = await joinClient(owner, stage.id, 'Hosty');
    await host.send({ type: 'startPlay' });
    const err = host.errors().find((e) => e.startsWith('Cannot start'));
    expect(err).toBeTruthy();
    expect(err).toContain('unfilled role');
    expect(err).toContain('not ready');
    expect(host.lastState()?.phase).toBe('lobby');
  });

  it('starts when all roles filled and everyone ready; countdown then performing', async () => {
    const { host, partner } = await readyStage();
    await host.send({ type: 'setReady', isReady: true });
    await partner.send({ type: 'setReady', isReady: true });
    await host.send({ type: 'startPlay' });
    expect(host.ofType('countdownStart')).toHaveLength(1);
    expect(host.lastState()?.phase).toBe('countdown');
    await sleep(COUNTDOWN_WAIT);
    await host.send({ type: 'heartbeat' });
    expect(host.ofType('playStarted').length).toBeGreaterThan(0);
    expect(host.lastState()?.phase).toBe('performing');
    expect(host.lastState()?.currentLineIndex).toBe(0);
  });

  it('speaker-only advancement; audience blocked; ends in discussion', async () => {
    const { host, partner, stage } = await readyStage();
    await host.send({ type: 'setReady', isReady: true });
    await partner.send({ type: 'setReady', isReady: true });
    const watcher = await joinClient(guestIdentity(), stage.id, 'Watcher');
    await watcher.send({ type: 'setAudience', isAudience: true });
    await watcher.send({ type: 'setReady', isReady: true });
    await host.send({ type: 'startPlay' });
    await sleep(COUNTDOWN_WAIT);

    // Line 0 belongs to cast[0] (host). Partner cannot advance it.
    await partner.send({ type: 'nextLine' });
    expect(partner.errors()).toContain('Not your line.');

    // Audience cannot advance in any mode.
    await watcher.send({ type: 'nextLine' });
    expect(watcher.errors()).toContain('Audience cannot advance lines.');

    // The speaker can.
    await host.send({ type: 'nextLine' });
    expect(host.last<{ currentLineIndex: number }>('lineAdvanced')?.currentLineIndex).toBe(1);

    // Line 1 is partner's; host now rejected.
    await host.send({ type: 'nextLine' });
    expect(host.errors()).toContain('Not your line.');
    await partner.send({ type: 'nextLine' });

    // Line 2 (last) is host's; advancing past it ends the play.
    await host.send({ type: 'nextLine' });
    expect(host.ofType('playEnded')).toHaveLength(1);
    expect(host.lastState()?.phase).toBe('discussion');
  });

  it('pause blocks advancement; resume restores; back/skip are host-gated', async () => {
    const { host, partner } = await readyStage();
    await host.send({ type: 'setReady', isReady: true });
    await partner.send({ type: 'setReady', isReady: true });
    await host.send({ type: 'startPlay' });
    await sleep(COUNTDOWN_WAIT);

    await host.send({ type: 'hostPause' });
    await host.send({ type: 'nextLine' });
    expect(host.errors()).toContain('Play is paused by host.');

    await partner.send({ type: 'hostResume' });
    expect(partner.errors()).toContain('Host only.');
    await host.send({ type: 'hostResume' });

    // Back clamps at 0.
    await host.send({ type: 'hostBackLine' });
    expect(host.lastState()?.currentLineIndex).toBe(0);
    await partner.send({ type: 'hostBackLine' });
    expect(partner.errors()).toContain('Host only.');

    // Skip advances; skipping past the last line ends the play.
    await host.send({ type: 'hostSkipLine' });
    await host.send({ type: 'hostSkipLine' });
    await host.send({ type: 'hostSkipLine' });
    expect(host.ofType('playEnded')).toHaveLength(1);
    expect(host.lastState()?.phase).toBe('discussion');

    // Reset from discussion returns everyone to lobby un-ready.
    await partner.send({ type: 'resetPlay' });
    expect(partner.lastState()?.phase).toBe('lobby');
    expect(participants(partner.lastState()).every((p) => !p.isReady)).toBe(true);
  });

  it('director mode: only director starts and advances; host cannot claim roles', async () => {
    const owner = await userIdentity();
    const play = await makePlay(owner.userId!);
    const stage = await makeStage(play.id);
    const host = await joinClient(owner, stage.id, 'Directed');
    const p1 = await joinClient(guestIdentity(), stage.id, 'ActorOne');
    const p2 = await joinClient(guestIdentity(), stage.id, 'ActorTwo');
    const [a, b] = play.cast.map((c) => c.id);

    await host.send({ type: 'hostToggleDirectorMode', enabled: true });
    expect((host.lastState() as StageState | undefined)?.directorMode).toBe(true);

    // Host cannot claim a role while directing.
    await host.send({ type: 'toggleCastMember', castMemberId: a });
    expect(host.errors()).toContain('Exit director mode to take a role.');

    await p1.send({ type: 'toggleCastMember', castMemberId: a });
    await p2.send({ type: 'toggleCastMember', castMemberId: b });
    for (const c of [host, p1, p2]) await c.send({ type: 'setReady', isReady: true });

    // Non-director start rejected.
    await p1.send({ type: 'startPlay' });
    expect(p1.errors()).toContain('Only the director can start the play.');

    await host.send({ type: 'startPlay' });
    await sleep(COUNTDOWN_WAIT);
    await host.send({ type: 'heartbeat' });
    expect(host.lastState()?.phase).toBe('performing');

    // Speaker cannot advance; only the director.
    await p1.send({ type: 'nextLine' });
    expect(p1.errors()).toContain('Director controls line advance.');
    await host.send({ type: 'nextLine' });
    expect(host.last<{ currentLineIndex: number }>('lineAdvanced')?.currentLineIndex).toBe(1);

    // Director mode cannot be toggled outside the lobby.
    await host.send({ type: 'hostToggleDirectorMode', enabled: false });
    expect(host.errors()).toContain('Director mode can only be toggled from the lobby.');
  });
});

describe('late joiners (§3.9)', () => {
  it('joining mid-performance lands in audience, locked until lobby', async () => {
    const owner = await userIdentity();
    const play = await makePlay(owner.userId!);
    const stage = await makeStage(play.id);
    const host = await joinClient(owner, stage.id, 'Hosty');
    const partner = await joinClient(guestIdentity(), stage.id, 'Partner');
    const [a, b] = play.cast.map((c) => c.id);
    await host.send({ type: 'toggleCastMember', castMemberId: a });
    await partner.send({ type: 'toggleCastMember', castMemberId: b });
    await host.send({ type: 'setReady', isReady: true });
    await partner.send({ type: 'setReady', isReady: true });
    await host.send({ type: 'startPlay' });
    await sleep(COUNTDOWN_WAIT);

    const late = await joinClient(guestIdentity(), stage.id, 'Latecomer');
    const me = participants(late.lastState()).find((p) => p.id === late.participantId);
    expect(me?.isAudience).toBe(true);
    // Current line index in sync immediately.
    expect(late.lastState()?.currentLineIndex).toBe(0);
    // Announced via system message.
    expect(host.systemMessages().some((m) => m.includes('Latecomer joined the audience'))).toBe(
      true
    );

    // Cannot self-toggle out of audience mid-play.
    await late.send({ type: 'setAudience', isAudience: false });
    expect(late.errors().some((e) => e.includes('returns to the lobby'))).toBe(true);

    // After reset to lobby, they can.
    await host.send({ type: 'hostEndPlay' });
    await late.send({ type: 'setAudience', isAudience: false });
    const after = participants(late.lastState()).find((p) => p.id === late.participantId);
    expect(after?.isAudience).toBe(false);
  });
});

describe('chat (§6)', () => {
  it('public chat reaches everyone; cast-only excludes audience; truncation applies', async () => {
    const stage = await makeStage();
    const a = await joinClient(guestIdentity(), stage.id, 'Speaker');
    const b = await joinClient(guestIdentity(), stage.id, 'Listener');
    const aud = await joinClient(guestIdentity(), stage.id, 'Watcher');
    await aud.send({ type: 'setAudience', isAudience: true });

    await a.send({ type: 'sendChat', text: 'hello all', castOnly: false });
    expect(b.last<{ message: { text: string } }>('chatMessage')?.message.text).toBe('hello all');
    expect(aud.last<{ message: { text: string } }>('chatMessage')?.message.text).toBe('hello all');

    await a.send({ type: 'sendChat', text: 'cast secret', castOnly: true });
    expect(b.last<{ message: { text: string } }>('chatMessage')?.message.text).toBe('cast secret');
    expect(aud.ofType<{ message: { text: string } }>('chatMessage').map((m) => m.message.text)).not.toContain(
      'cast secret'
    );

    // Server-side truncation at 256 with ellipsis.
    await a.send({ type: 'sendChat', text: 'y'.repeat(400), castOnly: false });
    const long = b.last<{ message: { text: string } }>('chatMessage')?.message.text ?? '';
    expect(long.length).toBe(256);
    expect(long.endsWith('…')).toBe(true);
  });
});
