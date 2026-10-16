package api_test

// The cases in this file are ported from the TypeScript server's
// ws-stage.test.ts, each subtest under its original name. Messages are the
// protojson ClientMessage and ServerMessage now, so a type the old suite named
// as {"type": "x"} is the oneof field x.

import (
	"slices"
	"strings"
	"testing"

	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

func TestWSJoinBasics(T *testing.T) {
	T.Parallel()

	T.Run("rejects a connection with no identity at join time", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		c := h.connect(t, wsIdentity{})
		c.join(stage.ID, "NoIdentity")

		c.waitFor("the identity refusal", func() bool { return c.hasErrorContaining("Identity required") })
		c.waitFor("the server to close the socket", c.closedByServer)
	})

	T.Run("rejects malformed JSON without crashing, keeps connection open", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")
		c := h.joinClient(t, guestIdentity(), stage.ID, "Guesty")

		c.sendRaw("{not json")
		c.waitFor("the refusal", func() bool { return c.hasError("Invalid JSON") })

		// The socket still answers afterwards.
		c.sync()
		test.False(t, c.closedByServer())
	})

	T.Run("rejects unknown message types but keeps the connection open", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")
		c := h.joinClient(t, guestIdentity(), stage.ID, "Guesty")

		c.sendRaw(`{"flyToTheMoon": {}}`)
		c.waitFor("the refusal", func() bool { return c.hasErrorContaining("Unknown message type") })

		c.sync()
		test.False(t, c.closedByServer())
	})

	T.Run("rejects messages before join", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		c := h.connect(t, guestIdentity())
		c.send(heartbeat())

		c.waitFor("the refusal", func() bool { return c.hasError("Must join before sending messages") })
	})

	T.Run("joining a non-existent code reports Stage not found", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		c := h.connect(t, guestIdentity())
		c.join("ZZZZ99", "Nobody")

		c.waitFor("the refusal", func() bool { return c.hasError("Stage not found") })
	})

	T.Run("join by code is case-insensitive", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		c := h.joinClient(t, guestIdentity(), strings.ToLower(stage.Code), "CaseTest")
		test.NotEq(t, "", c.participantID())
	})

	T.Run("enforces MAX_PARTICIPANTS_PER_STAGE = 12", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		for n := range 12 {
			h.joinClient(t, guestIdentity(), stage.ID, "Player "+string(rune('A'+n)))
		}

		unlucky := h.connect(t, guestIdentity())
		unlucky.join(stage.ID, "Unlucky")

		unlucky.waitFor("the refusal", func() bool { return unlucky.hasErrorContaining("Stage is full") })
		unlucky.waitFor("the server to close the socket", unlucky.closedByServer)
	})
}

func TestDisplayNamePolicy(T *testing.T) {
	T.Parallel()

	T.Run("blocks empty names", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		c := h.connect(t, guestIdentity())
		c.join(stage.ID, "   ")
		c.waitFor("the refusal", func() bool { return c.hasError("Please enter a name.") })
	})

	T.Run("blocks names over 32 chars", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		c := h.connect(t, guestIdentity())
		c.join(stage.ID, strings.Repeat("x", 33))
		c.waitFor("the refusal", func() bool { return c.hasErrorContaining("too long") })
	})

	T.Run("blocks reserved tokens case-insensitively", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		for _, name := range []string{"Host", "ADMIN", "director", "System", "sErVeR"} {
			c := h.connect(t, guestIdentity())
			c.join(stage.ID, name)
			c.waitFor("the refusal of "+name, func() bool { return c.hasError("Please choose a different name.") })
		}
	})

	T.Run("blocks profanity", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		c := h.connect(t, guestIdentity())
		c.join(stage.ID, "shithead")
		c.waitFor("the refusal", func() bool { return c.hasError("Please choose a different name.") })
	})

	T.Run("auto-suffixes duplicate names case-insensitively", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		h.joinClient(t, guestIdentity(), stage.ID, "Sam")
		second := h.joinClient(t, guestIdentity(), stage.ID, "sam")
		third := h.joinClient(t, guestIdentity(), stage.ID, "SAM")

		var names []string
		for _, p := range third.lastState().GetParticipants() {
			names = append(names, p.GetName())
		}

		test.SliceContains(t, names, "Sam")
		test.SliceContains(t, names, "sam (2)")
		test.SliceContains(t, names, "SAM (3)")
		test.NotEq(t, "", second.participantID())
	})
}

func TestSameIdentityMultipleConnections(T *testing.T) {
	T.Parallel()

	T.Run("second tab displaces the first; participant appears once", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")
		identity := guestIdentity()

		tab1 := h.joinClient(t, identity, stage.ID, "TabUser")
		tab2 := h.joinClient(t, identity, stage.ID, "TabUser")

		tab1.waitFor("the session to be replaced", func() bool {
			return tab1.count(isKind((*stagepb.ServerMessage).GetSessionReplaced)) == 1
		})
		tab1.waitFor("the server to close the first tab", tab1.closedByServer)

		matches := 0

		for _, p := range tab2.lastState().GetParticipants() {
			if p.GetName() == "TabUser" {
				matches++
			}
		}

		test.EqOp(t, 1, matches)
		test.EqOp(t, tab1.participantID(), tab2.participantID())
	})

	T.Run("reconnect after drop reuses the participant row and keeps roles", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		identity := guestIdentity()
		cast0, _ := castIDs(play)

		first := h.joinClient(t, identity, stage.ID, "Reconnector")
		first.send(toggleCastMember(cast0))
		first.waitFor("the role", func() bool { return slices.Contains(first.me().GetCastMemberIds(), cast0) })
		first.disconnect()

		again := h.joinClient(t, identity, stage.ID, "Reconnector")
		test.EqOp(t, first.participantID(), again.participantID())
		test.SliceContains(t, again.me().GetCastMemberIds(), cast0)
	})
}

func TestLobbyActions(T *testing.T) {
	T.Parallel()

	type lobby struct {
		h     *harness
		host  *wsClient
		other *wsClient
		a, b  string
	}

	setUp := func(t *testing.T) *lobby {
		t.Helper()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		a, b := castIDs(play)

		return &lobby{
			h:     h,
			host:  h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty"),
			other: h.joinClient(t, guestIdentity(), stage.ID, "Other"),
			a:     a,
			b:     b,
		}
	}

	T.Run("claim/release/steal roles; doubling allowed; audience clears roles", func(t *testing.T) {
		t.Parallel()

		l := setUp(t)
		host, other := l.host, l.other

		rolesOf := func(c *wsClient, id string) []string {
			ids := slices.Clone(participantIn(c.lastState(), id).GetCastMemberIds())
			slices.Sort(ids)

			return ids
		}

		// Claim.
		host.send(toggleCastMember(l.a))
		host.waitFor("the claim", func() bool { return slices.Equal(rolesOf(host, host.participantID()), []string{l.a}) })

		// Doubling: one participant may hold several roles.
		host.send(toggleCastMember(l.b))
		both := []string{l.a, l.b}
		slices.Sort(both)
		host.waitFor("the second role", func() bool { return slices.Equal(rolesOf(host, host.participantID()), both) })

		// Steal: roles are exclusive across participants.
		other.send(toggleCastMember(l.a))
		other.waitFor("the steal", func() bool {
			return slices.Equal(rolesOf(other, other.participantID()), []string{l.a}) &&
				slices.Equal(rolesOf(other, host.participantID()), []string{l.b})
		})

		// Clicking again releases.
		other.send(toggleCastMember(l.a))
		other.waitFor("the release", func() bool { return len(rolesOf(other, other.participantID())) == 0 })

		// Moving to the audience gives up every role.
		host.send(setAudience(true))
		host.waitFor("the roles to clear", func() bool {
			me := host.me()

			return me.GetIsAudience() && len(me.GetCastMemberIds()) == 0
		})
	})

	T.Run("selecting a different play clears all roles for everyone", func(t *testing.T) {
		t.Parallel()

		l := setUp(t)
		play2 := l.h.makePlay(t)

		l.host.send(toggleCastMember(l.a))
		l.other.send(toggleCastMember(l.b))
		l.host.waitFor("both roles taken", func() bool {
			return len(l.host.me().GetCastMemberIds()) == 1 && len(participantIn(l.host.lastState(), l.other.participantID()).GetCastMemberIds()) == 1
		})

		l.host.send(selectPlay(play2.GetId()))
		l.host.waitFor("the new play", func() bool { return l.host.lastState().GetSelectedPlay().GetId() == play2.GetId() })

		for _, p := range l.host.lastState().GetParticipants() {
			test.SliceEmpty(t, p.GetCastMemberIds())
		}
	})

	T.Run("ready toggle is visible to others in real time", func(t *testing.T) {
		t.Parallel()

		l := setUp(t)

		l.other.send(setReady(true))
		l.host.waitFor("the other's readiness", func() bool {
			return participantIn(l.host.lastState(), l.other.participantID()).GetIsReady()
		})
	})
}

func TestStartPlay(T *testing.T) {
	T.Parallel()

	type ready struct {
		h       *harness
		host    *wsClient
		partner *wsClient
		stageID string
	}

	// setUp seats the host in Alice and a guest in Bob, nobody ready.
	setUp := func(t *testing.T) *ready {
		t.Helper()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		a, b := castIDs(play)

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		partner := h.joinClient(t, guestIdentity(), stage.ID, "Partner")

		host.send(toggleCastMember(a))
		partner.send(toggleCastMember(b))
		host.waitFor("both roles taken", func() bool {
			return len(host.me().GetCastMemberIds()) == 1 && len(participantIn(host.lastState(), partner.participantID()).GetCastMemberIds()) == 1
		})

		return &ready{h: h, stageID: stage.ID, host: host, partner: partner}
	}

	readyUp := func(t *testing.T, clients ...*wsClient) {
		t.Helper()

		for _, c := range clients {
			c.send(setReady(true))
			c.waitFor("readiness", func() bool { return c.me().GetIsReady() })
		}
	}

	T.Run("rejects start with unfilled roles and unready performers, with details", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")

		host.send(startPlay())
		host.waitFor("the refusal", func() bool { return host.hasErrorContaining("Cannot start") })

		idx := slices.IndexFunc(host.errors(), func(e string) bool { return strings.HasPrefix(e, "Cannot start") })
		must.NotEq(t, -1, idx)
		test.StrContains(t, host.errors()[idx], "unfilled role")
		test.StrContains(t, host.errors()[idx], "not ready")
		test.EqOp(t, stagepb.StagePhase_STAGE_PHASE_LOBBY, phaseOf(host))
	})

	T.Run("starts when all roles filled and everyone ready; countdown then performing", func(t *testing.T) {
		t.Parallel()

		r := setUp(t)
		readyUp(t, r.host, r.partner)

		r.host.send(startPlay())
		r.host.waitFor("the countdown", func() bool {
			return r.host.count(isKind((*stagepb.ServerMessage).GetCountdownStart)) == 1 &&
				phaseOf(r.host) == stagepb.StagePhase_STAGE_PHASE_COUNTDOWN
		})

		r.host.waitFor("the performance", func() bool {
			return r.host.count(isKind((*stagepb.ServerMessage).GetPlayStarted)) > 0 &&
				phaseOf(r.host) == stagepb.StagePhase_STAGE_PHASE_PERFORMING
		})
		test.EqOp(t, int32(0), r.host.lastState().GetCurrentLineIndex())
	})

	T.Run("speaker-only advancement; audience blocked; ends in discussion", func(t *testing.T) {
		t.Parallel()

		r := setUp(t)
		readyUp(t, r.host, r.partner)

		watcher := r.h.joinClient(t, guestIdentity(), r.stageID, "Watcher")
		watcher.send(setAudience(true))
		watcher.waitFor("the audience", func() bool { return watcher.me().GetIsAudience() })
		readyUp(t, watcher)

		r.host.send(startPlay())
		r.host.waitFor("the performance", func() bool { return phaseOf(r.host) == stagepb.StagePhase_STAGE_PHASE_PERFORMING })

		// Line 0 is Alice's, the host's. The partner cannot advance it.
		r.partner.send(nextLine())
		r.partner.waitFor("the refusal", func() bool { return r.partner.hasError("Not your line.") })

		// The audience cannot advance in any mode.
		watcher.send(nextLine())
		watcher.waitFor("the refusal", func() bool { return watcher.hasError("Audience cannot advance lines.") })

		// The speaker can.
		r.host.send(nextLine())
		r.host.waitFor("line 1", func() bool { return r.host.lastState().GetCurrentLineIndex() == 1 })

		// Line 1 is the partner's; the host is now refused.
		r.host.send(nextLine())
		r.host.waitFor("the refusal", func() bool { return r.host.hasError("Not your line.") })

		r.partner.send(nextLine())
		r.host.waitFor("line 2", func() bool { return r.host.lastState().GetCurrentLineIndex() == 2 })

		// Line 2, the last, is the host's; advancing past it ends the play.
		r.host.send(nextLine())
		r.host.waitFor("the end", func() bool {
			return r.host.count(isKind((*stagepb.ServerMessage).GetPlayEnded)) == 1 &&
				phaseOf(r.host) == stagepb.StagePhase_STAGE_PHASE_DISCUSSION
		})
	})
}

func TestStartPlayHostControls(T *testing.T) {
	T.Parallel()

	T.Run("pause blocks advancement; resume restores; back/skip are host-gated", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		a, b := castIDs(play)

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		partner := h.joinClient(t, guestIdentity(), stage.ID, "Partner")

		host.send(toggleCastMember(a))
		partner.send(toggleCastMember(b))
		host.send(setReady(true))
		partner.send(setReady(true))
		host.waitFor("everybody ready", func() bool {
			p := participantIn(host.lastState(), partner.participantID())

			return host.me().GetIsReady() && p.GetIsReady() && len(p.GetCastMemberIds()) == 1 && len(host.me().GetCastMemberIds()) == 1
		})

		host.send(startPlay())
		host.waitFor("the performance", func() bool { return phaseOf(host) == stagepb.StagePhase_STAGE_PHASE_PERFORMING })

		host.send(hostPause())
		host.waitFor("the pause", func() bool { return host.lastState().GetPausedByHost() })

		host.send(nextLine())
		host.waitFor("the refusal", func() bool { return host.hasError("Play is paused by host.") })

		partner.send(hostResume())
		partner.waitFor("the refusal", func() bool { return partner.hasError(errHostOnly) })

		host.send(hostResume())
		host.waitFor("the resume", func() bool { return !host.lastState().GetPausedByHost() })

		// Back clamps at the first line.
		host.send(hostBackLine())
		host.waitFor("the step back", func() bool { return host.hasSystemMessageContaining("Host went back a line") })
		test.EqOp(t, int32(0), host.lastState().GetCurrentLineIndex())

		partner.send(hostBackLine())
		partner.waitFor("the refusal", func() bool {
			return len(slices.DeleteFunc(partner.errors(), func(e string) bool { return e != errHostOnly })) == 2
		})

		// Skip advances; skipping past the last line ends the play.
		for range 3 {
			host.send(hostSkipLine())
		}

		host.waitFor("the end", func() bool {
			return host.count(isKind((*stagepb.ServerMessage).GetPlayEnded)) == 1 &&
				phaseOf(host) == stagepb.StagePhase_STAGE_PHASE_DISCUSSION
		})

		// A reset from the discussion returns everybody to the lobby, unready.
		partner.send(resetPlay())
		partner.waitFor("the lobby", func() bool { return phaseOf(partner) == stagepb.StagePhase_STAGE_PHASE_LOBBY })

		for _, p := range partner.lastState().GetParticipants() {
			test.False(t, p.GetIsReady())
		}
	})

	T.Run("director mode: only director starts and advances; host cannot claim roles", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		a, b := castIDs(play)

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Directed")
		p1 := h.joinClient(t, guestIdentity(), stage.ID, "ActorOne")
		p2 := h.joinClient(t, guestIdentity(), stage.ID, "ActorTwo")

		host.send(hostToggleDirectorMode(true))
		host.waitFor("director mode", func() bool { return host.lastState().GetDirectorMode() })

		// The director cannot take a role.
		host.send(toggleCastMember(a))
		host.waitFor("the refusal", func() bool { return host.hasError("Exit director mode to take a role.") })

		p1.send(toggleCastMember(a))
		p2.send(toggleCastMember(b))

		for _, c := range []*wsClient{host, p1, p2} {
			c.send(setReady(true))
		}

		host.waitFor("everybody ready", func() bool {
			state := host.lastState()
			one := participantIn(state, p1.participantID())
			two := participantIn(state, p2.participantID())

			return host.me().GetIsReady() && one.GetIsReady() && two.GetIsReady() &&
				len(one.GetCastMemberIds()) == 1 && len(two.GetCastMemberIds()) == 1
		})

		// Somebody else starting is refused.
		p1.send(startPlay())
		p1.waitFor("the refusal", func() bool { return p1.hasError("Only the director can start the play.") })

		host.send(startPlay())
		host.waitFor("the performance", func() bool { return phaseOf(host) == stagepb.StagePhase_STAGE_PHASE_PERFORMING })

		// The speaker cannot advance; only the director.
		p1.send(nextLine())
		p1.waitFor("the refusal", func() bool { return p1.hasError("Director controls line advance.") })

		host.send(nextLine())
		host.waitFor("line 1", func() bool { return host.lastState().GetCurrentLineIndex() == 1 })

		// Director mode is switched from the lobby only.
		host.send(hostToggleDirectorMode(false))
		host.waitFor("the refusal", func() bool { return host.hasError("Director mode can only be toggled from the lobby.") })
	})
}

func TestLateJoiners(T *testing.T) {
	T.Parallel()

	T.Run("joining mid-performance lands in audience, locked until lobby", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		play := h.makePlay(t)
		stage := h.makeStage(t, play.GetId())
		a, b := castIDs(play)

		host := h.joinClient(t, h.userIdentity(t), stage.ID, "Hosty")
		partner := h.joinClient(t, guestIdentity(), stage.ID, "Partner")

		host.send(toggleCastMember(a))
		partner.send(toggleCastMember(b))
		host.send(setReady(true))
		partner.send(setReady(true))
		host.waitFor("everybody ready", func() bool {
			p := participantIn(host.lastState(), partner.participantID())

			return host.me().GetIsReady() && p.GetIsReady() && len(p.GetCastMemberIds()) == 1 && len(host.me().GetCastMemberIds()) == 1
		})

		host.send(startPlay())
		host.waitFor("the performance", func() bool { return phaseOf(host) == stagepb.StagePhase_STAGE_PHASE_PERFORMING })

		late := h.joinClient(t, guestIdentity(), stage.ID, "Latecomer")
		test.True(t, late.me().GetIsAudience())
		// The playhead is in step from the first state.
		test.EqOp(t, int32(0), late.lastState().GetCurrentLineIndex())
		host.waitFor("the announcement", func() bool { return host.hasSystemMessageContaining("Latecomer joined the audience") })

		// No leaving the audience mid-play.
		late.send(setAudience(false))
		late.waitFor("the refusal", func() bool { return late.hasErrorContaining("returns to the lobby") })

		// Once the stage is back in the lobby, they can.
		host.send(hostEndPlay())
		host.waitFor("the lobby", func() bool { return phaseOf(host) == stagepb.StagePhase_STAGE_PHASE_LOBBY })

		late.send(setAudience(false))
		late.waitFor("the cast", func() bool { return !late.me().GetIsAudience() })
	})
}

func TestChat(T *testing.T) {
	T.Parallel()

	T.Run("public chat reaches everyone; cast-only excludes audience; truncation applies", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		a := h.joinClient(t, guestIdentity(), stage.ID, "Speaker")
		b := h.joinClient(t, guestIdentity(), stage.ID, "Listener")
		aud := h.joinClient(t, guestIdentity(), stage.ID, "Watcher")

		aud.send(setAudience(true))
		a.waitFor("the watcher in the audience", func() bool { return participantIn(a.lastState(), aud.participantID()).GetIsAudience() })

		a.send(sendChat("hello all", false))
		b.waitFor("the public message", func() bool { return slices.Contains(b.chats(), "hello all") })
		aud.waitFor("the public message", func() bool { return slices.Contains(aud.chats(), "hello all") })

		a.send(sendChat("cast secret", true))
		b.waitFor("the cast-only message", func() bool { return slices.Contains(b.chats(), "cast secret") })

		// A public message sent after it reaches the audience after anything
		// that would have been sent to them before it.
		a.send(sendChat("after the secret", false))
		aud.waitFor("the later message", func() bool { return slices.Contains(aud.chats(), "after the secret") })
		test.SliceNotContains(t, aud.chats(), "cast secret")

		// Cut at 256 with an ellipsis.
		long := strings.Repeat("y", 400)
		a.send(sendChat(long, false))
		b.waitFor("the long message", func() bool {
			return slices.ContainsFunc(b.chats(), func(m string) bool { return strings.HasPrefix(m, "yyyy") })
		})

		idx := slices.IndexFunc(b.chats(), func(m string) bool { return strings.HasPrefix(m, "yyyy") })
		cut := []rune(b.chats()[idx])
		test.EqOp(t, 256, len(cut))
		test.EqOp(t, '…', cut[len(cut)-1])
	})
}
