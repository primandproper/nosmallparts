package ws

import (
	"context"
	"fmt"
	"slices"

	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	"github.com/primandproper/primitives-go/v2/database"

	"github.com/oklog/ulid/v2"
)

// offer is a role the host offered somebody mid-performance, awaiting their
// answer while the play is paused.
type offer struct {
	id             string
	castMemberID   string
	castMemberName string
	target         string // participant ID
}

// hostOnly reads the stage and refuses the message unless the socket's
// identity hosts it. ok is false when it refused.
func (h *Hub) hostOnly(ctx context.Context, r *room, c *conn) (*snapshot, bool, error) {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return nil, false, err
	}

	if !isHost(s, c) {
		c.sendError(errHostOnly)

		return s, false, nil
	}

	return s, true, nil
}

// hostMoveLine is the host stepping the playhead back, clamped at the first
// line, or skipping it forward, which past the last line ends the play.
func (h *Hub) hostMoveLine(ctx context.Context, r *room, c *conn, forward bool) error {
	s, ok, err := h.hostOnly(ctx, r, c)
	if err != nil || !ok {
		return err
	}

	if s.stage.Phase != stages.PhasePerforming {
		c.sendError("Not currently performing.")

		return nil
	}

	if forward {
		if err = h.advance(ctx, r, s); err != nil {
			return err
		}

		h.system(r, "Host skipped a line.")

		return nil
	}

	line := max(0, s.stage.LineIndex-1)

	if line != s.stage.LineIndex {
		if err = h.tx(ctx, func(tx database.Tx) error {
			return h.store.SetLine(ctx, tx, r.id, line)
		}); err != nil {
			return err
		}
	}

	r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_LineAdvanced_{
		LineAdvanced: &stagepb.ServerMessage_LineAdvanced{CurrentLineIndex: line},
	}}, "")

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	h.system(r, "Host went back a line.")

	return nil
}

// hostReassignRole gives a role to a participant. In the lobby it moves at
// once; mid-performance the play pauses and the target is offered it, and any
// vacancy countdown on the role is cancelled, since the host is filling it.
func (h *Hub) hostReassignRole(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_HostReassignRole) error {
	s, ok, err := h.hostOnly(ctx, r, c)
	if err != nil || !ok {
		return err
	}

	target := s.participant(m.GetTargetParticipantId())
	if target == nil {
		c.sendError("Target participant not found.")

		return nil
	}

	member := s.castMember(m.GetCastMemberId())
	if member == nil {
		c.sendError("Cast member not found in current play.")

		return nil
	}

	if s.stage.Phase != stages.PhasePerforming {
		if err = h.tx(ctx, func(tx database.Tx) error {
			return h.store.AssignRole(ctx, tx, r.id, target.ID, member.ID)
		}); err != nil {
			return err
		}

		if err = h.broadcastState(ctx, r); err != nil {
			return err
		}

		h.system(r, fmt.Sprintf("Host assigned %s to %s.", member.Name, target.DisplayName))

		return nil
	}

	targetConn := r.conns[target.ID]
	if targetConn == nil {
		c.sendError("Target is no longer connected.")

		return nil
	}

	if _, open := r.vacancies[member.ID]; open {
		r.clearVacancy(member.ID)
		r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_RoleVacancyCleared_{
			RoleVacancyCleared: &stagepb.ServerMessage_RoleVacancyCleared{CastMemberId: member.ID},
		}}, "")
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetPaused(ctx, tx, r.id, true)
	}); err != nil {
		return err
	}

	o := &offer{id: ulid.Make().String(), castMemberID: member.ID, castMemberName: member.Name, target: target.ID}
	r.offers[o.id] = o

	targetConn.sendMessage(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_RoleOffer{RoleOffer: &stagepb.RoleOffer{
		OfferId:        o.id,
		CastMemberId:   o.castMemberID,
		CastMemberName: o.castMemberName,
	}}})

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	h.system(r, fmt.Sprintf("Host offered %s to %s. Play is paused.", member.Name, target.DisplayName))

	return nil
}

// respondRoleOffer is the target's answer to an offer. Accepting takes the
// role; declining leaves it with whoever held it. Either way the play stays
// paused for the host to resume.
func (h *Hub) respondRoleOffer(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_RespondRoleOffer) error {
	o, ok := r.offers[m.GetOfferId()]
	if !ok {
		c.sendError("Offer not found or already resolved.")

		return nil
	}

	if o.target != c.participantID {
		c.sendError("Not your offer to respond to.")

		return nil
	}

	delete(r.offers, o.id)
	r.broadcast(offerClearedMessage(o.id), "")

	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	name := "Someone"
	if me := s.participant(c.participantID); me != nil {
		name = me.DisplayName
	}

	if !m.GetAccept() {
		h.system(r, fmt.Sprintf("%s declined %s. Play stays paused.", name, o.castMemberName))

		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		if audienceErr := h.store.SetAudience(ctx, tx, c.participantID, false); audienceErr != nil {
			return audienceErr
		}

		return h.store.AssignRole(ctx, tx, r.id, c.participantID, o.castMemberID)
	}); err != nil {
		return err
	}

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	h.system(r, fmt.Sprintf("%s accepted %s. Resume when ready.", name, o.castMemberName))

	return nil
}

func offerClearedMessage(offerID string) *stagepb.ServerMessage {
	return &stagepb.ServerMessage{Message: &stagepb.ServerMessage_RoleOfferCleared_{
		RoleOfferCleared: &stagepb.ServerMessage_RoleOfferCleared{OfferId: offerID},
	}}
}

// hostUnassignRole takes a role away from a participant on the stage.
func (h *Hub) hostUnassignRole(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_HostUnassignRole) error {
	s, ok, err := h.hostOnly(ctx, r, c)
	if err != nil || !ok {
		return err
	}

	if s.participant(m.GetTargetParticipantId()) == nil {
		c.sendError("Target participant not found.")

		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.UnassignRole(ctx, tx, m.GetTargetParticipantId(), m.GetCastMemberId())
	}); err != nil {
		return err
	}

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	h.system(r, "Host unassigned a role.")

	return nil
}

// hostTransferHost hands hosting to a connected, signed-in participant. A
// guest never hosts.
func (h *Hub) hostTransferHost(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_HostTransferHost) error {
	s, ok, err := h.hostOnly(ctx, r, c)
	if err != nil || !ok {
		return err
	}

	targetConn := r.conns[m.GetTargetParticipantId()]
	if targetConn == nil {
		c.sendError("Target is no longer connected.")

		return nil
	}

	if !targetConn.identity.SignedIn() {
		c.sendError("Target must be signed in.")

		return nil
	}

	r.stopHostGrace()

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetHost(ctx, tx, r.id, targetConn.identity)
	}); err != nil {
		return err
	}

	name := "Someone"
	if target := s.participant(m.GetTargetParticipantId()); target != nil {
		name = target.DisplayName
	}

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	h.system(r, "Host transferred to "+name+".")

	return nil
}

// hostRemove is a kick, or with ban a ban: the target is told, closed, marked
// gone and stripped of their roles, and removed from the video room. A ban
// also refuses the identity, and the address they were connected from, on this
// stage from now on; a kick lets them straight back in.
func (h *Hub) hostRemove(ctx context.Context, r *room, c *conn, targetID string, ban bool) error {
	s, ok, err := h.hostOnly(ctx, r, c)
	if err != nil || !ok {
		return err
	}

	if targetID == c.participantID {
		c.sendError("Can't kick yourself.")

		return nil
	}

	target := s.participant(targetID)
	if target == nil {
		c.sendError("Target not found.")

		return nil
	}

	targetConn := r.conns[targetID]

	if err = h.tx(ctx, func(tx database.Tx) error {
		if ban {
			if banErr := h.store.AddBan(ctx, tx, r.id, target.Identity); banErr != nil {
				return banErr
			}

			// The address is known only while they are connected.
			if targetConn != nil {
				if banErr := h.store.AddIPBan(ctx, tx, r.id, targetConn.ip); banErr != nil {
					return banErr
				}
			}
		}

		if leftErr := h.store.SetLeftAt(ctx, tx, targetID, new(h.now())); leftErr != nil {
			return leftErr
		}

		return h.store.ClearRoles(ctx, tx, targetID)
	}); err != nil {
		return err
	}

	if targetConn != nil {
		if ban {
			targetConn.sendError("You were banned by the host.")
		} else {
			targetConn.sendError("You were removed by the host.")
		}

		delete(r.conns, targetID)
		targetConn.close()
	}

	h.video.RemoveParticipant(ctx, r.id, targetID)

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	if ban {
		h.system(r, target.DisplayName+" was banned by the host.")
	} else {
		h.system(r, target.DisplayName+" was removed by the host.")
	}

	return nil
}

// hostMute forces a participant's microphone off, or lets it back on.
func (h *Hub) hostMute(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_HostMute) error {
	s, ok, err := h.hostOnly(ctx, r, c)
	if err != nil || !ok {
		return err
	}

	target := s.participant(m.GetTargetParticipantId())
	if target == nil {
		c.sendError("Target not found.")

		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetForceMuted(ctx, tx, target.ID, m.GetMuted())
	}); err != nil {
		return err
	}

	h.video.SetMuted(ctx, r.id, target.ID, m.GetMuted())

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	if m.GetMuted() {
		h.system(r, "Host muted "+target.DisplayName+".")
	} else {
		h.system(r, "Host unmuted "+target.DisplayName+".")
	}

	return nil
}

// hostToggleDirectorMode switches director mode, from the lobby only. A host
// entering it leaves the cast: the director runs the play rather than being in
// it.
func (h *Hub) hostToggleDirectorMode(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_HostToggleDirectorMode) error {
	s, ok, err := h.hostOnly(ctx, r, c)
	if err != nil || !ok {
		return err
	}

	if s.stage.Phase != stages.PhaseLobby {
		c.sendError("Director mode can only be toggled from the lobby.")

		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		if modeErr := h.store.SetDirectorMode(ctx, tx, r.id, m.GetEnabled()); modeErr != nil {
			return modeErr
		}

		if !m.GetEnabled() {
			return nil
		}

		for id, other := range r.conns {
			if other.identity.Matches(c.identity) {
				if clearErr := h.store.ClearRoles(ctx, tx, id); clearErr != nil {
					return clearErr
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	if m.GetEnabled() {
		h.system(r, "Director mode enabled. Host will advance lines.")
	} else {
		h.system(r, "Director mode disabled.")
	}

	return nil
}

// returningHost cancels a pending succession when the host comes back within
// the grace period, and answers what to announce.
func (r *room) returningHost(stage *stages.Stage, who stages.Identity, name string) string {
	if r.hostGrace == nil || !stage.HasHost() || !stage.Host.Matches(who) {
		return ""
	}

	r.stopHostGrace()

	return name + " (host) reconnected."
}

func (r *room) stopHostGrace() {
	if r.hostGrace != nil {
		r.hostGrace.Stop()
		r.hostGrace = nil
	}
}

// identityConnected reports whether any socket on the stage is this identity.
func (r *room) identityConnected(who stages.Identity) bool {
	for _, c := range r.conns {
		if c.identity.Matches(who) {
			return true
		}
	}

	return false
}

// hostLeft runs when a socket leaves: if it was the host's last, offers the
// host made are withdrawn and, after HostGrace, the role passes on unless the
// host came back.
func (h *Hub) hostLeft(r *room, s *snapshot, leaving *conn, leavingName string) {
	if !s.stage.HasHost() || !s.stage.Host.Matches(leaving.identity) || r.identityConnected(leaving.identity) {
		return
	}

	for id := range r.offers {
		delete(r.offers, id)
		r.broadcast(offerClearedMessage(id), "")
	}

	leavingHost := leaving.identity

	r.stopHostGrace()
	r.hostGrace = h.later(r, h.timings.HostGrace, func(ctx context.Context) {
		r.hostGrace = nil

		if err := h.promoteSuccessor(ctx, r, leavingHost, leavingName); err != nil {
			h.o11y.Logger().WithValue("stage.id", r.id).Error("passing on the host", err)
		}
	})
}

// promoteSuccessor makes the longest-connected signed-in participant host. A
// guest never hosts, so a stage of guests is left with nobody hosting until a
// signed-in user joins and claims it.
func (h *Hub) promoteSuccessor(ctx context.Context, r *room, leavingHost stages.Identity, leavingName string) error {
	if r.identityConnected(leavingHost) {
		return nil
	}

	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	// participants are in the order they joined.
	idx := slices.IndexFunc(s.participants, func(p *stages.Participant) bool {
		return r.conns[p.ID] != nil && p.Identity.SignedIn() && !p.Identity.Matches(leavingHost)
	})

	var successor *stages.Participant
	if idx >= 0 {
		successor = s.participants[idx]
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		if successor == nil {
			return h.store.SetHost(ctx, tx, r.id, stages.Identity{})
		}

		return h.store.SetHost(ctx, tx, r.id, successor.Identity)
	}); err != nil {
		return err
	}

	if successor != nil {
		h.system(r, fmt.Sprintf("%s left. %s is now host.", leavingName, successor.DisplayName))
	}

	return h.broadcastState(ctx, r)
}

// withdrawOffersTo clears the offers awaiting a participant who left before
// answering. The play stays paused.
func (h *Hub) withdrawOffersTo(r *room, participantID, name string) {
	for id, o := range r.offers {
		if o.target != participantID {
			continue
		}

		delete(r.offers, id)
		r.broadcast(offerClearedMessage(id), "")
		h.system(r, fmt.Sprintf("%s disconnected before responding to %s. Play stays paused.", name, o.castMemberName))
	}
}
