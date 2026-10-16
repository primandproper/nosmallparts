package api_test

// The cases in this file are ported from the TypeScript server's
// ws-host.test.ts, each subtest under its original name.

import (
	"slices"
	"testing"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

func hostOf(state *stagepb.StageState) *stagepb.Participant {
	for _, p := range state.GetParticipants() {
		if p.GetIsHost() {
			return p
		}
	}

	return nil
}

func hostPause() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostPause_{HostPause: &stagepb.ClientMessage_HostPause{}}}
}

func hostBackLine() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostBackLine_{HostBackLine: &stagepb.ClientMessage_HostBackLine{}}}
}

func hostSkipLine() *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostSkipLine_{HostSkipLine: &stagepb.ClientMessage_HostSkipLine{}}}
}

func hostTransferHost(targetID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostTransferHost_{
		HostTransferHost: &stagepb.ClientMessage_HostTransferHost{TargetParticipantId: targetID},
	}}
}

func hostKick(targetID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostKick_{HostKick: &stagepb.ClientMessage_HostKick{TargetParticipantId: targetID}}}
}

func hostBan(targetID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostBan_{HostBan: &stagepb.ClientMessage_HostBan{TargetParticipantId: targetID}}}
}

func hostMute(targetID string, muted bool) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostMute_{
		HostMute: &stagepb.ClientMessage_HostMute{TargetParticipantId: targetID, Muted: muted},
	}}
}

func hostReassignRole(targetID, castMemberID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostReassignRole_{
		HostReassignRole: &stagepb.ClientMessage_HostReassignRole{TargetParticipantId: targetID, CastMemberId: castMemberID},
	}}
}

func hostUnassignRole(targetID, castMemberID string) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostUnassignRole_{
		HostUnassignRole: &stagepb.ClientMessage_HostUnassignRole{TargetParticipantId: targetID, CastMemberId: castMemberID},
	}}
}

func hostToggleDirectorMode(enabled bool) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_HostToggleDirectorMode_{
		HostToggleDirectorMode: &stagepb.ClientMessage_HostToggleDirectorMode{Enabled: enabled},
	}}
}

func respondRoleOffer(offerID string, accept bool) *stagepb.ClientMessage {
	return &stagepb.ClientMessage{Message: &stagepb.ClientMessage_RespondRoleOffer_{
		RespondRoleOffer: &stagepb.ClientMessage_RespondRoleOffer{OfferId: offerID, Accept: accept},
	}}
}

func TestHostElection(T *testing.T) {
	T.Parallel()

	T.Run("guests never become host; first signed-in joiner does", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		g1 := h.joinClient(t, guestIdentity(), stage.ID, "GuestOne")
		test.Nil(t, hostOf(g1.lastState()))

		u := h.joinClient(t, h.userIdentity(t), stage.ID, "RealUser")
		test.EqOp(t, "RealUser", hostOf(u.lastState()).GetName())
	})

	T.Run("host reconnect within grace cancels succession", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")
		identity := h.userIdentity(t)

		host := h.joinClient(t, identity, stage.ID, "Hosty")
		other := h.joinClient(t, h.userIdentity(t), stage.ID, "Other")
		host.disconnect()

		back := h.joinClient(t, identity, stage.ID, "Hosty")
		other.waitFor("the host's return", func() bool { return other.hasSystemMessageContaining("(host) reconnected") })

		time.Sleep(graceWait)
		back.sync()
		test.EqOp(t, "Hosty", hostOf(back.lastState()).GetName())
	})

	T.Run("after grace, oldest-joined signed-in user is promoted; guests skipped", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		h.joinClient(t, guestIdentity(), stage.ID, "EarlyGuest")
		successor := h.joinClient(t, h.userIdentity(t), stage.ID, "Successor")
		later := h.joinClient(t, h.userIdentity(t), stage.ID, "LaterUser")
		host.disconnect()

		later.waitFor("the successor", func() bool { return hostOf(later.lastState()).GetName() == "Successor" })
		successor.waitFor("the announcement", func() bool { return successor.hasSystemMessageContaining("Successor is now host") })
	})

	T.Run("if only guests remain, host is cleared and a later signed-in joiner takes over", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		guest := h.joinClient(t, guestIdentity(), stage.ID, "LoneGuest")
		host.disconnect()

		time.Sleep(graceWait)
		guest.sync()
		test.Nil(t, hostOf(guest.lastState()))

		newUser := h.joinClient(t, h.userIdentity(t), stage.ID, "Savior")
		test.EqOp(t, "Savior", hostOf(newUser.lastState()).GetName())
	})

	T.Run("closing one of two same-identity tabs does not trigger succession", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")
		identity := h.userIdentity(t)

		tab1 := h.joinClient(t, identity, stage.ID, "Hosty")
		tab2 := h.joinClient(t, identity, stage.ID, "Hosty")
		other := h.joinClient(t, h.userIdentity(t), stage.ID, "Other")

		// tab2 displaced tab1; closing tab1 must not start the grace period.
		tab1.disconnect()

		time.Sleep(graceWait)
		other.sync()
		test.EqOp(t, "Hosty", hostOf(other.lastState()).GetName())
		test.False(t, tab2.closedByServer())
	})
}

func TestHostTransfer(T *testing.T) {
	T.Parallel()

	T.Run("transfers to a signed-in participant; old host loses gated actions", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "OldHost")
		next := h.joinClient(t, h.userIdentity(t), stage.ID, "NewHost")

		host.send(hostTransferHost(next.participantID()))
		host.waitFor("the transfer", func() bool {
			return hostOf(host.lastState()).GetName() == "NewHost" && host.hasSystemMessageContaining("Host transferred to NewHost")
		})

		host.send(hostPause())
		host.waitFor("the refusal", func() bool { return host.hasError(errHostOnly) })

		next.send(hostPause())
		next.waitFor("the pause", func() bool { return next.lastState().GetPausedByHost() })
		test.SliceNotContains(t, next.errors(), errHostOnly)
	})

	T.Run("transfer to a guest is rejected", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		guest := h.joinClient(t, guestIdentity(), stage.ID, "JustAGuest")

		host.send(hostTransferHost(guest.participantID()))
		host.waitFor("the refusal", func() bool { return host.hasError("Target must be signed in.") })
		test.EqOp(t, "Hosty", hostOf(host.lastState()).GetName())
	})
}

const errHostOnly = "Host only."

func TestKickAndBan(T *testing.T) {
	T.Parallel()

	T.Run("kick closes the target, posts a system message, and allows rejoin", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		targetIdentity := guestIdentity()
		target := h.joinClient(t, targetIdentity, stage.ID, "Kickme")

		host.send(hostKick(target.participantID()))
		target.waitFor("the removal", func() bool { return target.hasError("You were removed by the host.") })
		target.waitFor("the server to close the socket", target.closedByServer)
		host.waitFor("the announcement", func() bool { return host.hasSystemMessageContaining("Kickme was removed by the host") })

		// And out of the video room.
		test.True(t, h.video.wasRemoved(stage.ID, target.participantID()))

		// Free to come back.
		again := h.joinClient(t, targetIdentity, stage.ID, "Kickme")
		test.NotEq(t, "", again.participantID())
		test.SliceEmpty(t, again.errors())
	})

	T.Run("host cannot kick or ban self", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		host.send(hostKick(host.participantID()))
		host.waitFor("the refusal", func() bool { return host.hasError("Can't kick yourself.") })

		host.send(hostBan(host.participantID()))
		host.waitFor("the second refusal", func() bool {
			return len(slices.DeleteFunc(host.errors(), func(e string) bool { return e != "Can't kick yourself." })) == 2
		})
	})

	T.Run("ban blocks the identity from rejoining", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		targetIdentity := guestIdentity()
		target := h.joinClient(t, targetIdentity, stage.ID, "Banme")

		host.send(hostBan(target.participantID()))
		target.waitFor("the ban", func() bool { return target.hasError("You were banned by the host.") })

		retry := h.connect(t, targetIdentity)
		retry.join(stage.ID, "Banme")
		retry.waitFor("the refusal", func() bool { return retry.hasError("You have been removed from this stage.") })
		retry.waitFor("the server to close the socket", retry.closedByServer)
	})

	T.Run("IP ban catches a banned guest who clears their cookie; scoped per stage", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		const bannedIP = "203.0.113.7"
		target := h.joinClient(t, guestAt(bannedIP), stage.ID, "Sneaky")

		host.send(hostBan(target.participantID()))
		target.waitFor("the ban", func() bool { return target.hasError("You were banned by the host.") })

		// A fresh guest cookie from the same address is still refused.
		freshCookie := h.connect(t, guestAt(bannedIP))
		freshCookie.join(stage.ID, "SneakyTwo")
		freshCookie.waitFor("the refusal", func() bool { return freshCookie.hasError("You have been removed from this stage.") })

		// A fresh identity from elsewhere is let in.
		elsewhere := h.joinClient(t, guestAt("198.51.100.9"), stage.ID, "Innocent")
		test.NotEq(t, "", elsewhere.participantID())

		// The same address on another stage is let in: bans are per stage.
		otherStage := h.makeStage(t, "")
		sameIPOtherStage := h.joinClient(t, guestAt(bannedIP), otherStage.ID, "Sneaky")
		test.NotEq(t, "", sameIPOtherStage.participantID())
	})
}

func TestForceMute(T *testing.T) {
	T.Parallel()

	T.Run("sets and clears the forceMuted flag", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		target := h.joinClient(t, guestIdentity(), stage.ID, "Loudmouth")

		host.send(hostMute(target.participantID(), true))
		target.waitFor("the mute", func() bool { return target.me().GetForceMuted() })
		host.waitFor("the announcement", func() bool { return host.hasSystemMessageContaining("Host muted Loudmouth") })

		muted, set := h.video.mutedState(stage.ID, target.participantID())
		test.True(t, set)
		test.True(t, muted)

		host.send(hostMute(target.participantID(), false))
		target.waitFor("the unmute", func() bool { return !target.me().GetForceMuted() })

		muted, _ = h.video.mutedState(stage.ID, target.participantID())
		test.False(t, muted)
	})

	T.Run("mute is host-gated", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		rando := h.joinClient(t, guestIdentity(), stage.ID, "Rando")
		victim := h.joinClient(t, guestIdentity(), stage.ID, "Victim")

		rando.send(hostMute(victim.participantID(), true))
		rando.waitFor("the refusal", func() bool { return rando.hasError(errHostOnly) })

		_, set := h.video.mutedState(stage.ID, victim.participantID())
		test.False(t, set)
	})
}

func TestLobbyRoleReassignment(T *testing.T) {
	T.Parallel()

	T.Run("host reassign moves the role instantly; unassign clears it", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		a, b := castIDs(play)

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		p1 := h.joinClient(t, guestIdentity(), stage.ID, "ActorOne")
		p2 := h.joinClient(t, guestIdentity(), stage.ID, "ActorTwo")

		rolesOf := func(id string) []string {
			ids := slices.Clone(participantIn(host.lastState(), id).GetCastMemberIds())
			slices.Sort(ids)

			return ids
		}

		p1.send(toggleCastMember(a))
		host.waitFor("p1's claim", func() bool { return slices.Equal(rolesOf(p1.participantID()), []string{a}) })

		// Reassign a from p1 to p2.
		host.send(hostReassignRole(p2.participantID(), a))
		host.waitFor("the move", func() bool {
			return len(rolesOf(p1.participantID())) == 0 && slices.Equal(rolesOf(p2.participantID()), []string{a})
		})

		// Doubling holds: p2 takes b as well.
		both := []string{a, b}
		slices.Sort(both)
		host.send(hostReassignRole(p2.participantID(), b))
		host.waitFor("the second role", func() bool { return slices.Equal(rolesOf(p2.participantID()), both) })

		// Unassigning clears it.
		host.send(hostUnassignRole(p2.participantID(), a))
		host.waitFor("the unassign", func() bool { return slices.Equal(rolesOf(p2.participantID()), []string{b}) })
		must.NotNil(t, participantIn(host.lastState(), p2.participantID()))
	})
}
