import { describe, expect, it } from 'vitest';
import type { Participant } from '@nsp/shared';
import {
  guestIdentity,
  joinClient,
  makeStage,
  sleep,
  userIdentity,
  WsClient,
} from './helpers.js';

const GRACE_WAIT = 200; // NSP_HOST_GRACE_MS=80 in setup

function participants(state: Record<string, unknown> | undefined): Participant[] {
  return (state?.participants ?? []) as Participant[];
}

function hostOf(state: Record<string, unknown> | undefined): Participant | undefined {
  return participants(state).find((p) => p.isHost);
}

describe('host election (§4.1, §1.4)', () => {
  it('guests never become host; first signed-in joiner does', async () => {
    const stage = await makeStage();
    const g1 = await joinClient(guestIdentity(), stage.id, 'GuestOne');
    expect(hostOf(g1.lastState())).toBeUndefined();

    const u = await joinClient(await userIdentity(), stage.id, 'RealUser');
    expect(hostOf(u.lastState())?.name).toBe('RealUser');
  });

  it('host reconnect within grace cancels succession', async () => {
    const stage = await makeStage();
    const identity = await userIdentity();
    const host = await joinClient(identity, stage.id, 'Hosty');
    const other = await joinClient(await userIdentity(), stage.id, 'Other');
    await host.disconnect();

    // Within grace: no succession yet.
    await other.send({ type: 'heartbeat' });
    expect(hostOf(other.lastState())?.name).toBeUndefined;

    const back = await joinClient(identity, stage.id, 'Hosty');
    expect(other.systemMessages().some((m) => m.includes('(host) reconnected'))).toBe(true);
    await sleep(GRACE_WAIT);
    await back.send({ type: 'heartbeat' });
    expect(hostOf(back.lastState())?.name).toBe('Hosty');
  });

  it('after grace, oldest-joined signed-in user is promoted; guests skipped', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    await joinClient(guestIdentity(), stage.id, 'EarlyGuest');
    const successor = await joinClient(await userIdentity(), stage.id, 'Successor');
    const later = await joinClient(await userIdentity(), stage.id, 'LaterUser');
    await host.disconnect();
    await sleep(GRACE_WAIT);
    await later.send({ type: 'heartbeat' });
    expect(hostOf(later.lastState())?.name).toBe('Successor');
    expect(successor.systemMessages().some((m) => m.includes('Successor is now host'))).toBe(true);
  });

  it('if only guests remain, host is cleared and a later signed-in joiner takes over', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    const guest = await joinClient(guestIdentity(), stage.id, 'LoneGuest');
    await host.disconnect();
    await sleep(GRACE_WAIT);
    await guest.send({ type: 'heartbeat' });
    expect(hostOf(guest.lastState())).toBeUndefined();

    const newUser = await joinClient(await userIdentity(), stage.id, 'Saviour');
    expect(hostOf(newUser.lastState())?.name).toBe('Saviour');
  });

  it('closing one of two same-identity tabs does not trigger succession', async () => {
    const stage = await makeStage();
    const identity = await userIdentity();
    const tab1 = await joinClient(identity, stage.id, 'Hosty');
    const tab2 = await joinClient(identity, stage.id, 'Hosty');
    const other = await joinClient(await userIdentity(), stage.id, 'Other');
    // tab1 was displaced by tab2; closing it must not start grace.
    await tab1.disconnect();
    await sleep(GRACE_WAIT);
    await other.send({ type: 'heartbeat' });
    expect(hostOf(other.lastState())?.name).toBe('Hosty');
    expect(tab2.closedByServer).toBe(false);
  });
});

describe('host transfer (§4.2)', () => {
  it('transfers to a signed-in participant; old host loses gated actions', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'OldHost');
    const next = await joinClient(await userIdentity(), stage.id, 'NewHost');
    await host.send({ type: 'hostTransferHost', targetParticipantId: next.participantId });
    expect(hostOf(host.lastState())?.name).toBe('NewHost');
    expect(host.systemMessages().some((m) => m.includes('Host transferred to NewHost'))).toBe(true);

    await host.send({ type: 'hostPause' });
    expect(host.errors()).toContain('Host only.');
    await next.send({ type: 'hostPause' });
    expect(next.errors()).not.toContain('Host only.');
  });

  it('transfer to a guest is rejected', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    const guest = await joinClient(guestIdentity(), stage.id, 'JustAGuest');
    await host.send({ type: 'hostTransferHost', targetParticipantId: guest.participantId });
    expect(host.errors()).toContain('Target must be signed in.');
    expect(hostOf(host.lastState())?.name).toBe('Hosty');
  });
});

describe('kick / ban (§4.7)', () => {
  it('kick closes the target, posts a system message, and allows rejoin', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    const targetIdentity = guestIdentity();
    const target = await joinClient(targetIdentity, stage.id, 'Kickme');
    await host.send({ type: 'hostKick', targetParticipantId: target.participantId });
    expect(target.errors()).toContain('You were removed by the host.');
    expect(target.closedByServer).toBe(true);
    expect(host.systemMessages().some((m) => m.includes('Kickme was removed by the host'))).toBe(
      true
    );
    // Can rejoin freely.
    const again = await joinClient(targetIdentity, stage.id, 'Kickme');
    expect(again.participantId).toBeTruthy();
    expect(again.errors()).toHaveLength(0);
  });

  it('host cannot kick or ban self', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    await host.send({ type: 'hostKick', targetParticipantId: host.participantId });
    expect(host.errors()).toContain("Can't kick yourself.");
  });

  it('ban blocks the identity from rejoining', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    const targetIdentity = guestIdentity();
    const target = await joinClient(targetIdentity, stage.id, 'Banme');
    await host.send({ type: 'hostBan', targetParticipantId: target.participantId });
    expect(target.errors()).toContain('You were banned by the host.');

    const retry = new WsClient(targetIdentity);
    await retry.connect();
    await retry.join(stage.id, 'Banme');
    expect(retry.errors()).toContain('You have been removed from this stage.');
    expect(retry.closedByServer).toBe(true);
  });

  it('IP ban catches a banned guest who clears their cookie; scoped per stage', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    const bannedIp = '203.0.113.7';
    const target = await joinClient(guestIdentity(bannedIp), stage.id, 'Sneaky');
    await host.send({ type: 'hostBan', targetParticipantId: target.participantId });

    // Fresh guest cookie, same IP: still blocked.
    const freshCookie = new WsClient(guestIdentity(bannedIp));
    await freshCookie.connect();
    await freshCookie.join(stage.id, 'SneakyTwo');
    expect(freshCookie.errors()).toContain('You have been removed from this stage.');

    // Different IP, fresh identity: allowed (accepted per spec).
    const elsewhere = await joinClient(guestIdentity('198.51.100.9'), stage.id, 'Innocent');
    expect(elsewhere.participantId).toBeTruthy();

    // Same IP on a different stage: allowed (per-stage scope).
    const otherStage = await makeStage();
    const sameIpOtherStage = await joinClient(guestIdentity(bannedIp), otherStage.id, 'Sneaky');
    expect(sameIpOtherStage.participantId).toBeTruthy();
  });
});

describe('force mute (§4.8)', () => {
  it('sets and clears the forceMuted flag', async () => {
    const stage = await makeStage();
    const host = await joinClient(await userIdentity(), stage.id, 'Hosty');
    const target = await joinClient(guestIdentity(), stage.id, 'Loudmouth');

    await host.send({ type: 'hostMute', targetParticipantId: target.participantId, muted: true });
    let seen = participants(target.lastState()).find((p) => p.id === target.participantId);
    expect(seen?.forceMuted).toBe(true);
    expect(host.systemMessages().some((m) => m.includes('Host muted Loudmouth'))).toBe(true);

    await host.send({ type: 'hostMute', targetParticipantId: target.participantId, muted: false });
    seen = participants(target.lastState()).find((p) => p.id === target.participantId);
    expect(seen?.forceMuted).toBe(false);
  });

  it('mute is host-gated', async () => {
    const stage = await makeStage();
    await joinClient(await userIdentity(), stage.id, 'Hosty');
    const rando = await joinClient(guestIdentity(), stage.id, 'Rando');
    const victim = await joinClient(guestIdentity(), stage.id, 'Victim');
    await rando.send({ type: 'hostMute', targetParticipantId: victim.participantId, muted: true });
    expect(rando.errors()).toContain('Host only.');
  });
});

describe('role reassignment in lobby (§4.3)', () => {
  it('host reassign moves the role instantly; unassign clears it', async () => {
    const owner = await userIdentity();
    const { makePlay } = await import('./helpers.js');
    const play = await makePlay(owner.userId!);
    const stage = await makeStage(play.id);
    const host = await joinClient(owner, stage.id, 'Hosty');
    const p1 = await joinClient(guestIdentity(), stage.id, 'ActorOne');
    const p2 = await joinClient(guestIdentity(), stage.id, 'ActorTwo');
    const [a, b] = play.cast.map((c) => c.id);
    await p1.send({ type: 'toggleCastMember', castMemberId: a });

    // Reassign a from p1 to p2.
    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: p2.participantId,
      castMemberId: a,
    });
    let ps = participants(host.lastState());
    expect(ps.find((p) => p.id === p1.participantId)?.castMemberIds).toEqual([]);
    expect(ps.find((p) => p.id === p2.participantId)?.castMemberIds).toEqual([a]);

    // Doubling preserved: p2 also gets b.
    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: p2.participantId,
      castMemberId: b,
    });
    ps = participants(host.lastState());
    expect(ps.find((p) => p.id === p2.participantId)?.castMemberIds?.sort()).toEqual([a, b].sort());

    // Unassign clears.
    await host.send({
      type: 'hostUnassignRole',
      targetParticipantId: p2.participantId,
      castMemberId: a,
    });
    ps = participants(host.lastState());
    expect(ps.find((p) => p.id === p2.participantId)?.castMemberIds).toEqual([b]);
  });
});
