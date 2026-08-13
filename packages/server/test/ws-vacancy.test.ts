import { describe, expect, it } from 'vitest';
import type { Participant, RoleVacancy } from '@nsp/shared';
import {
  guestIdentity,
  joinClient,
  makePlay,
  makeStage,
  sleep,
  userIdentity,
  WsClient,
  type WsIdentity,
} from './helpers.js';

const COUNTDOWN_WAIT = 120; // NSP_COUNTDOWN_MS=40
const VACANCY_WAIT = 300; // NSP_VACANCY_CLAIM_MS=150

function participants(state: Record<string, unknown> | undefined): Participant[] {
  return (state?.participants ?? []) as Participant[];
}

async function performingStage(opts: { directorMode?: boolean } = {}) {
  const owner = await userIdentity();
  const play = await makePlay(owner.userId!);
  const stage = await makeStage(play.id);
  const host = await joinClient(owner, stage.id, 'Hosty');
  const partnerIdentity = guestIdentity();
  const partner = await joinClient(partnerIdentity, stage.id, 'Partner');
  const audience = await joinClient(guestIdentity(), stage.id, 'Watcher');
  await audience.send({ type: 'setAudience', isAudience: true });
  const [a, b] = play.cast.map((c) => c.id);

  if (opts.directorMode) {
    await host.send({ type: 'hostToggleDirectorMode', enabled: true });
    const third = await joinClient(guestIdentity(), stage.id, 'ThirdActor');
    await third.send({ type: 'toggleCastMember', castMemberId: a });
    await partner.send({ type: 'toggleCastMember', castMemberId: b });
    for (const c of [host, partner, third, audience]) {
      await c.send({ type: 'setReady', isReady: true });
    }
    await host.send({ type: 'startPlay' });
    await sleep(COUNTDOWN_WAIT);
    return { play, stage, host, partner, partnerIdentity, audience, third, a, b };
  }

  await host.send({ type: 'toggleCastMember', castMemberId: a });
  await partner.send({ type: 'toggleCastMember', castMemberId: b });
  for (const c of [host, partner, audience]) {
    await c.send({ type: 'setReady', isReady: true });
  }
  await host.send({ type: 'startPlay' });
  await sleep(COUNTDOWN_WAIT);
  return { play, stage, host, partner, partnerIdentity, audience, a, b, third: undefined };
}

describe('mid-play role vacancy (§4.5)', () => {
  it('disconnecting role-holder pauses the play and starts a claim countdown', async () => {
    const { host, partner, b } = await performingStage();
    await partner.disconnect();

    expect(host.lastState()?.pausedByHost).toBe(true);
    const vacated = host.last<{ vacancy: RoleVacancy }>('roleVacated');
    expect(vacated?.vacancy.castMemberId).toBe(b);
    expect(vacated?.vacancy.castMemberName).toBe('Bob');
    expect(vacated?.vacancy.claimDeadline).toBeGreaterThan(Date.now());
    expect(host.systemMessages().some((m) => m.includes('Bob is vacant'))).toBe(true);
  });

  it('audience member claims the vacant role with one click; play stays paused', async () => {
    const { host, partner, audience, b } = await performingStage();
    await partner.disconnect();

    await audience.send({ type: 'audienceClaimRole', castMemberId: b });
    expect(audience.errors()).toHaveLength(0);
    const me = participants(audience.lastState()).find((p) => p.id === audience.participantId);
    expect(me?.isAudience).toBe(false);
    expect(me?.castMemberIds).toContain(b);
    expect(host.ofType('roleVacancyCleared').length).toBeGreaterThan(0);
    expect(host.lastState()?.pausedByHost).toBe(true); // host must resume manually

    // Host resumes; play continues from same line.
    await host.send({ type: 'hostResume' });
    expect(host.lastState()?.pausedByHost).toBe(false);
    expect(host.lastState()?.phase).toBe('performing');

    // Expired timer must not fire later.
    await sleep(VACANCY_WAIT);
    await host.send({ type: 'heartbeat' });
    expect(host.lastState()?.phase).toBe('performing');
  });

  it('non-audience participants cannot claim', async () => {
    const { host, partner, b } = await performingStage();
    await partner.disconnect();
    await host.send({ type: 'audienceClaimRole', castMemberId: b });
    expect(host.errors()).toContain('Only audience members can claim a vacant role.');
  });

  it('original holder reconnecting within the countdown reclaims the role', async () => {
    const { host, partner, partnerIdentity, stage, b } = await performingStage();
    await partner.disconnect();

    const back = await joinClient(partnerIdentity, stage.id, 'Partner');
    expect(back.participantId).toBe(partner.participantId);
    const me = participants(back.lastState()).find((p) => p.id === back.participantId);
    expect(me?.castMemberIds).toContain(b);
    expect(host.ofType('roleVacancyCleared').length).toBeGreaterThan(0);
    expect(host.systemMessages().some((m) => m.includes('reclaimed Bob'))).toBe(true);

    // Countdown cancelled: no auto-end.
    await sleep(VACANCY_WAIT);
    await host.send({ type: 'heartbeat' });
    expect(host.lastState()?.phase).toBe('performing');
  });

  it('host fast-track via hostReassignRole bypasses the countdown', async () => {
    const { host, partner, audience, b } = await performingStage();
    await partner.disconnect();

    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: audience.participantId,
      castMemberId: b,
    });
    expect(host.ofType('roleVacancyCleared').length).toBeGreaterThan(0);
    const offer = audience.last<{ offer: { offerId: string } }>('roleOffer');
    expect(offer).toBeTruthy();
    await audience.send({ type: 'respondRoleOffer', offerId: offer!.offer.offerId, accept: true });
    const me = participants(audience.lastState()).find((p) => p.id === audience.participantId);
    expect(me?.castMemberIds).toContain(b);

    // Timer must not auto-end the play later.
    await sleep(VACANCY_WAIT);
    await host.send({ type: 'heartbeat' });
    expect(host.lastState()?.phase).toBe('performing');
  });

  it('expiry with no claim auto-ends the play to discussion', async () => {
    const { host, partner } = await performingStage();
    await partner.disconnect();

    await sleep(VACANCY_WAIT);
    await host.send({ type: 'heartbeat' });
    expect(host.ofType('playEnded').length).toBe(1);
    expect(host.lastState()?.phase).toBe('discussion');
    expect(host.systemMessages().some((m) => m.includes('Play ended: Bob went vacant'))).toBe(true);
  });

  it('late joiner during the countdown receives the live claimDeadline', async () => {
    const { partner, stage } = await performingStage();
    await partner.disconnect();

    const late = await joinClient(guestIdentity(), stage.id, 'JustArrived');
    const vac = late.last<{ vacancy: RoleVacancy }>('roleVacated');
    expect(vac?.vacancy.castMemberId).toBeTruthy();
    expect(vac?.vacancy.claimDeadline).toBeGreaterThan(Date.now());
  });

  it('director mode: pause without countdown; claims rejected; no auto-end', async () => {
    const { host, partner, audience, b } = await performingStage({ directorMode: true });
    await partner.disconnect();

    expect(host.lastState()?.pausedByHost).toBe(true);
    const vac = host.last<{ vacancy: RoleVacancy }>('roleVacated');
    expect(vac?.vacancy.claimDeadline).toBeNull();

    await audience.send({ type: 'audienceClaimRole', castMemberId: b });
    expect(audience.errors()).toContain('That role is not open to claims.');

    // No 45s auto-end in director mode.
    await sleep(VACANCY_WAIT);
    await host.send({ type: 'heartbeat' });
    expect(host.lastState()?.phase).toBe('performing');

    // Director reassigns via the offer flow.
    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: audience.participantId,
      castMemberId: b,
    });
    const offer = audience.last<{ offer: { offerId: string } }>('roleOffer');
    await audience.send({ type: 'respondRoleOffer', offerId: offer!.offer.offerId, accept: true });
    const me = participants(audience.lastState()).find((p) => p.id === audience.participantId);
    expect(me?.castMemberIds).toContain(b);
  });
});

describe('role offers (§4.4, §9)', () => {
  it('decline keeps the role unchanged and the play paused', async () => {
    const { host, audience, partner, b } = await performingStage();
    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: audience.participantId,
      castMemberId: b,
    });
    expect(host.lastState()?.pausedByHost).toBe(true);
    const offer = audience.last<{ offer: { offerId: string } }>('roleOffer');
    await audience.send({ type: 'respondRoleOffer', offerId: offer!.offer.offerId, accept: false });
    expect(host.systemMessages().some((m) => m.includes('declined Bob'))).toBe(true);
    const holder = participants(host.lastState()).find((p) => p.id === partner.participantId);
    expect(holder?.castMemberIds).toContain(b);
    expect(host.lastState()?.pausedByHost).toBe(true);
  });

  it('only the target can respond to an offer', async () => {
    const { host, audience, partner, b } = await performingStage();
    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: audience.participantId,
      castMemberId: b,
    });
    const offer = audience.last<{ offer: { offerId: string } }>('roleOffer');
    await partner.send({ type: 'respondRoleOffer', offerId: offer!.offer.offerId, accept: true });
    expect(partner.errors()).toContain('Not your offer to respond to.');
  });

  it('target disconnecting before responding clears the offer; play stays paused', async () => {
    const { host, audience, b } = await performingStage();
    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: audience.participantId,
      castMemberId: b,
    });
    await audience.disconnect();
    expect(host.ofType('roleOfferCleared').length).toBeGreaterThan(0);
    expect(
      host.systemMessages().some((m) => m.includes('disconnected before responding'))
    ).toBe(true);
    expect(host.lastState()?.pausedByHost).toBe(true);
  });

  it('host disconnecting during a pending offer auto-clears it; play stays paused', async () => {
    const { host, audience, partner, b } = await performingStage();
    await host.send({
      type: 'hostReassignRole',
      targetParticipantId: audience.participantId,
      castMemberId: b,
    });
    await host.disconnect();
    expect(audience.ofType('roleOfferCleared').length).toBeGreaterThan(0);
    expect(partner.lastState()?.pausedByHost).toBe(true);
  });
});
