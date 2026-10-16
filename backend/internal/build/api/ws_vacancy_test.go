package api_test

// The cases in this file are ported from the TypeScript server's
// ws-vacancy.test.ts, each subtest under its original name.

import (
	"slices"
	"testing"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

// performing is a stage mid-performance: the host as Alice, a guest partner as
// Bob, and a guest watching from the audience.
type performing struct {
	h               *harness
	stageID         string
	host            *wsClient
	partner         *wsClient
	partnerIdentity wsIdentity
	audience        *wsClient
	a, b            string
}

func performingStage(t *testing.T) *performing {
	t.Helper()

	h := newHarness(t)
	play := h.makePlay(t)
	stage := h.makeStage(t, play.GetId())
	a, b := castIDs(play)

	p := &performing{h: h, stageID: stage.ID, partnerIdentity: guestIdentity(), a: a, b: b}
	p.host = h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
	p.partner = h.joinClient(t, p.partnerIdentity, stage.ID, "Partner")
	p.audience = h.joinClient(t, guestIdentity(), stage.ID, "Watcher")

	p.audience.send(setAudience(true))
	p.host.send(toggleCastMember(a))
	p.partner.send(toggleCastMember(b))

	for _, c := range []*wsClient{p.host, p.partner, p.audience} {
		c.send(setReady(true))
	}

	p.host.waitFor("everybody seated and ready", func() bool {
		state := p.host.lastState()
		host := participantIn(state, p.host.participantID())
		partner := participantIn(state, p.partner.participantID())
		watcher := participantIn(state, p.audience.participantID())

		return slices.Equal(host.GetCastMemberIds(), []string{a}) && host.GetIsReady() &&
			slices.Equal(partner.GetCastMemberIds(), []string{b}) && partner.GetIsReady() &&
			watcher.GetIsAudience() && watcher.GetIsReady()
	})

	p.host.send(startPlay())
	p.host.waitFor("the performance", func() bool { return phaseOf(p.host) == stagepb.StagePhase_STAGE_PHASE_PERFORMING })

	return p
}

func lastVacancy(c *wsClient) *stagepb.RoleVacancy {
	var last *stagepb.RoleVacancy

	for _, msg := range c.messages() {
		if v := msg.GetRoleVacated(); v != nil {
			last = v
		}
	}

	return last
}

func vacancyCleared(msg *stagepb.ServerMessage) bool {
	return msg.GetRoleVacancyCleared() != nil
}

func TestMidPlayRoleVacancy(T *testing.T) {
	T.Parallel()

	T.Run("disconnecting role-holder pauses the play and starts a claim countdown", func(t *testing.T) {
		t.Parallel()

		p := performingStage(t)
		p.partner.disconnect()

		p.host.waitFor("the pause and the vacancy", func() bool {
			return p.host.lastState().GetPausedByHost() && lastVacancy(p.host) != nil && p.host.hasSystemMessageContaining("Bob "+vacancyMessage)
		})

		vacated := lastVacancy(p.host)
		test.EqOp(t, p.b, vacated.GetCastMemberId())
		test.EqOp(t, "Bob", vacated.GetCastMemberName())
		must.NotNil(t, vacated.GetClaimDeadline())
		test.True(t, vacated.GetClaimDeadline().AsTime().After(time.Now()))
	})

	T.Run("audience member claims the vacant role with one click; play stays paused", func(t *testing.T) {
		t.Parallel()

		p := performingStage(t)
		p.partner.disconnect()
		p.audience.waitFor("the vacancy", func() bool { return lastVacancy(p.audience) != nil })

		p.audience.send(audienceClaimRole(p.b))
		p.audience.waitFor("the role", func() bool {
			me := p.audience.me()

			return !me.GetIsAudience() && slices.Contains(me.GetCastMemberIds(), p.b)
		})
		test.SliceEmpty(t, p.audience.errors())

		p.host.waitFor("the vacancy to clear", func() bool { return p.host.count(vacancyCleared) > 0 })
		// The host resumes by hand.
		test.True(t, p.host.lastState().GetPausedByHost())

		// The host resumes; the play goes on from the same line.
		p.host.send(hostResume())
		p.host.waitFor("the resume", func() bool { return !p.host.lastState().GetPausedByHost() })
		test.EqOp(t, stagepb.StagePhase_STAGE_PHASE_PERFORMING, phaseOf(p.host))

		// The claim stopped the countdown; it must not end the play later.
		time.Sleep(vacancyWait)
		p.host.sync()
		test.EqOp(t, stagepb.StagePhase_STAGE_PHASE_PERFORMING, phaseOf(p.host))
	})

	T.Run("non-audience participants cannot claim", func(t *testing.T) {
		t.Parallel()

		p := performingStage(t)
		p.partner.disconnect()
		p.host.waitFor("the vacancy", func() bool { return lastVacancy(p.host) != nil })

		p.host.send(audienceClaimRole(p.b))
		p.host.waitFor("the refusal", func() bool { return p.host.hasError("Only audience members can claim a vacant role.") })
	})

	T.Run("original holder reconnecting within the countdown reclaims the role", func(t *testing.T) {
		t.Parallel()

		p := performingStage(t)
		p.partner.disconnect()
		p.host.waitFor("the vacancy", func() bool { return lastVacancy(p.host) != nil })

		back := p.h.joinClient(t, p.partnerIdentity, p.stageID, "Partner")
		test.EqOp(t, p.partner.participantID(), back.participantID())
		test.SliceContains(t, back.me().GetCastMemberIds(), p.b)

		p.host.waitFor("the reclaim", func() bool {
			return p.host.count(vacancyCleared) > 0 && p.host.hasSystemMessageContaining("reclaimed Bob")
		})

		// The countdown was cancelled: no auto-end.
		time.Sleep(vacancyWait)
		p.host.sync()
		test.EqOp(t, stagepb.StagePhase_STAGE_PHASE_PERFORMING, phaseOf(p.host))
	})

	T.Run("expiry with no claim auto-ends the play to discussion", func(t *testing.T) {
		t.Parallel()

		p := performingStage(t)
		p.partner.disconnect()

		p.host.waitFor("the play to end", func() bool {
			return p.host.count(isKind((*stagepb.ServerMessage).GetPlayEnded)) == 1 &&
				phaseOf(p.host) == stagepb.StagePhase_STAGE_PHASE_DISCUSSION
		})
		test.True(t, p.host.hasSystemMessageContaining("Play ended: Bob went vacant"))
	})

	T.Run("late joiner during the countdown receives the live claimDeadline", func(t *testing.T) {
		t.Parallel()

		p := performingStage(t)
		p.partner.disconnect()
		p.host.waitFor("the vacancy", func() bool { return lastVacancy(p.host) != nil })

		late := p.h.joinClient(t, guestIdentity(), p.stageID, "JustArrived")
		late.waitFor("the vacancy", func() bool { return lastVacancy(late) != nil })

		vacancy := lastVacancy(late)
		test.NotEq(t, "", vacancy.GetCastMemberId())
		must.NotNil(t, vacancy.GetClaimDeadline())
		test.True(t, vacancy.GetClaimDeadline().AsTime().After(time.Now()))
	})
}
