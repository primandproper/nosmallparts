package ws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/primandproper/nosmallparts/backend/internal/displaynames"
	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"
	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/proto/scriptspb"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"
	scriptsgrpc "github.com/primandproper/nosmallparts/backend/internal/services/scripts/grpc"

	"github.com/primandproper/primitives-go/v2/database"

	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MaxChatLength is the longest a chat message is broadcast, in characters; the
// server cuts a longer one and marks the cut.
const MaxChatLength = 256

// What a participant is told when something they asked for did not happen,
// for the messages more than one handler sends.
const (
	errInternal  = "Something went wrong."
	errHostOnly  = "Host only."
	errNotJoined = "Must join before sending messages"
)

// vacancy is a role whose holder disconnected mid-performance. Outside
// director mode the audience may claim it until the deadline, and the play
// ends if nobody does.
type vacancy struct {
	deadline       *time.Time
	timer          *time.Timer
	castMemberID   string
	castMemberName string
	previousHolder string
}

// handle decodes one frame and does what it asks.
func (h *Hub) handle(ctx context.Context, c *conn, frame []byte) {
	ctx, op := h.o11y.Begin(ctx)
	defer op.End()

	msg, problem := decode(frame)
	if problem != "" {
		c.sendError(problem)

		return
	}

	if join := msg.GetJoin(); join != nil {
		if err := h.join(ctx, c, join); err != nil {
			op.Logger().Error("joining a stage", err)
			c.sendError(errInternal)
		}

		return
	}

	if c.stageID == "" {
		c.sendError(errNotJoined)

		return
	}

	r, ok := h.existingRoom(c.stageID)
	if !ok {
		c.sendError("Stage not found")

		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// A socket that was displaced, or removed by the host, is no longer the
	// one its participant is speaking through.
	if r.gone || r.conns[c.participantID] != c {
		c.sendError(errNotJoined)

		return
	}

	if err := h.dispatch(ctx, r, c, msg); err != nil {
		op.Logger().WithValue("stage.id", r.id).Error("handling a stage message", err)
		c.sendError(errInternal)
	}
}

// decode reads a frame as a ClientMessage, or says what is wrong with it.
func decode(frame []byte) (msg *stagepb.ClientMessage, problem string) {
	if !json.Valid(frame) {
		return nil, "Invalid JSON"
	}

	var keys map[string]json.RawMessage
	if err := json.Unmarshal(frame, &keys); err != nil {
		return nil, "Unknown message type"
	}

	known := (&stagepb.ClientMessage{}).ProtoReflect().Descriptor().Oneofs().ByName("message").Fields()

	for key := range keys {
		if known.ByJSONName(key) == nil && known.ByTextName(key) == nil {
			return nil, "Unknown message type: " + key
		}
	}

	msg = &stagepb.ClientMessage{}
	if err := protojson.Unmarshal(frame, msg); err != nil {
		return nil, "Invalid message"
	}

	if msg.GetMessage() == nil {
		return nil, "Unknown message type"
	}

	return msg, ""
}

// dispatch does what one message asks. The caller holds the room's lock and
// has checked the socket is joined.
func (h *Hub) dispatch(ctx context.Context, r *room, c *conn, msg *stagepb.ClientMessage) error {
	switch m := msg.GetMessage().(type) {
	case *stagepb.ClientMessage_Heartbeat_:
		s, err := h.snapshot(ctx, r.id)
		if err != nil {
			return err
		}

		c.sendMessage(stateMessage(s))

		return nil
	case *stagepb.ClientMessage_ToggleCastMember_:
		return h.toggleCastMember(ctx, r, c, m.ToggleCastMember)
	case *stagepb.ClientMessage_SetAudience_:
		return h.setAudience(ctx, r, c, m.SetAudience)
	case *stagepb.ClientMessage_SetReady_:
		return h.setReady(ctx, r, c, m.SetReady)
	case *stagepb.ClientMessage_SelectPlay_:
		return h.selectPlay(ctx, r, m.SelectPlay)
	case *stagepb.ClientMessage_SendChat_:
		return h.sendChat(ctx, r, c, m.SendChat)
	case *stagepb.ClientMessage_StartPlay_:
		return h.startPlay(ctx, r, c)
	case *stagepb.ClientMessage_NextLine_:
		return h.nextLine(ctx, r, c)
	case *stagepb.ClientMessage_ResetPlay_:
		return h.resetPlay(ctx, r, c)
	case *stagepb.ClientMessage_AudienceClaimRole_:
		return h.audienceClaimRole(ctx, r, c, m.AudienceClaimRole)
	case *stagepb.ClientMessage_HostPause_:
		return h.hostPause(ctx, r, c, true)
	case *stagepb.ClientMessage_HostResume_:
		return h.hostPause(ctx, r, c, false)
	case *stagepb.ClientMessage_HostEndPlay_:
		return h.hostEndPlay(ctx, r, c)
	case *stagepb.ClientMessage_HostBackLine_:
		return h.hostMoveLine(ctx, r, c, false)
	case *stagepb.ClientMessage_HostSkipLine_:
		return h.hostMoveLine(ctx, r, c, true)
	case *stagepb.ClientMessage_HostReassignRole_:
		return h.hostReassignRole(ctx, r, c, m.HostReassignRole)
	case *stagepb.ClientMessage_HostUnassignRole_:
		return h.hostUnassignRole(ctx, r, c, m.HostUnassignRole)
	case *stagepb.ClientMessage_RespondRoleOffer_:
		return h.respondRoleOffer(ctx, r, c, m.RespondRoleOffer)
	case *stagepb.ClientMessage_HostTransferHost_:
		return h.hostTransferHost(ctx, r, c, m.HostTransferHost)
	case *stagepb.ClientMessage_HostKick_:
		return h.hostRemove(ctx, r, c, m.HostKick.GetTargetParticipantId(), false)
	case *stagepb.ClientMessage_HostBan_:
		return h.hostRemove(ctx, r, c, m.HostBan.GetTargetParticipantId(), true)
	case *stagepb.ClientMessage_HostMute_:
		return h.hostMute(ctx, r, c, m.HostMute)
	case *stagepb.ClientMessage_HostToggleDirectorMode_:
		return h.hostToggleDirectorMode(ctx, r, c, m.HostToggleDirectorMode)
	default:
		c.sendError("Unknown message type: " + string(msg.ProtoReflect().WhichOneof(msg.ProtoReflect().Descriptor().Oneofs().ByName("message")).Name()))

		return nil
	}
}

// selfOnly refuses a message naming a participant other than the socket's own.
// A message may name one, for the client's convenience, but only its own.
func selfOnly(c *conn, named string) bool {
	if named != "" && named != c.participantID {
		c.sendError("You can only change your own participant.")

		return false
	}

	return true
}

// tx runs writes in one transaction.
func (h *Hub) tx(ctx context.Context, fn func(tx database.Tx) error) error {
	return h.client.WithTransaction(ctx, fn)
}

// join admits a socket to a stage: as the participant this identity already
// is there, reconnecting, or as a new one.
func (h *Hub) join(ctx context.Context, c *conn, j *stagepb.ClientMessage_Join) error {
	if c.identity.IsZero() {
		c.sendError("Identity required. Call /api/identity before connecting.")
		c.close()

		return nil
	}

	if c.stageID != "" {
		c.sendError("Already joined a stage.")

		return nil
	}

	name, err := displaynames.Validate(j.GetParticipantName())
	if err != nil {
		c.sendError(err.Error())

		return nil
	}

	stage, err := h.findStage(ctx, j.GetStageId())
	if errors.Is(err, stages.ErrStageNotFound) {
		c.sendError("Stage not found")

		return nil
	} else if err != nil {
		return err
	}

	banned, err := h.store.IsBanned(ctx, h.client.Reader(), stage.ID, c.identity, c.ip)
	if err != nil {
		return err
	}

	if banned {
		c.sendError("You have been removed from this stage.")
		c.close()

		return nil
	}

	r := h.room(stage.ID)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.gone {
		c.sendError("Stage not found")

		return nil
	}

	// A join cancels a pending deletion of the emptied stage.
	if r.cleanup != nil {
		r.cleanup.Stop()
		r.cleanup = nil
	}

	before, err := h.snapshot(ctx, stage.ID)
	if err != nil {
		return err
	}

	participantID, announcement, ok, err := h.admit(ctx, r, c, before, name)
	if err != nil || !ok {
		return err
	}

	r.conns[participantID] = c
	c.stageID, c.participantID = stage.ID, participantID

	// A participant who left roles vacant mid-play takes them back on return;
	// their role rows survived the disconnect.
	for id, v := range r.vacancies {
		if v.previousHolder == participantID {
			r.clearVacancy(id)
			r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_RoleVacancyCleared_{
				RoleVacancyCleared: &stagepb.ServerMessage_RoleVacancyCleared{CastMemberId: id},
			}}, "")
			h.system(r, fmt.Sprintf("%s reconnected and reclaimed %s.", name, v.castMemberName))
		}
	}

	hostAnnouncement, err := h.claimHost(ctx, r, c, before.stage, name)
	if err != nil {
		return err
	}

	if hostAnnouncement == "" {
		hostAnnouncement = r.returningHost(before.stage, c.identity, name)
	}

	after, err := h.snapshot(ctx, stage.ID)
	if err != nil {
		return err
	}

	c.sendMessage(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_Joined_{Joined: &stagepb.ServerMessage_Joined{
		ParticipantId: participantID,
		StageState:    after.proto(),
	}}})

	if err = h.sendAvailablePlays(ctx, c); err != nil {
		return err
	}

	// Somebody arriving during a vacancy countdown sees it with its real
	// deadline.
	for _, v := range r.vacancies {
		c.sendMessage(vacatedMessage(v))
	}

	r.broadcast(stateMessage(after), participantID)

	for _, text := range []string{announcement, hostAnnouncement} {
		if text != "" {
			h.system(r, text)
		}
	}

	return nil
}

// findStage resolves what a joiner typed: a stage ID, or a join code in any
// case.
func (h *Hub) findStage(ctx context.Context, idOrCode string) (*stages.Stage, error) {
	idOrCode = strings.TrimSpace(idOrCode)
	if idOrCode == "" {
		return nil, stages.ErrStageNotFound
	}

	stage, err := h.store.GetStage(ctx, h.client.Reader(), idOrCode)
	if !errors.Is(err, stages.ErrStageNotFound) {
		return stage, err
	}

	return h.store.GetStageByCode(ctx, h.client.Reader(), strings.ToUpper(idOrCode))
}

// admit finds or makes the joiner's participant row. A second socket from an
// identity already on the stage takes over from the first, which is told so
// and closed, so nobody appears twice. ok is false when the joiner was refused
// and told why.
func (h *Hub) admit(ctx context.Context, r *room, c *conn, s *snapshot, name string) (id, announcement string, ok bool, err error) {
	existing, err := h.store.LatestParticipantFor(ctx, h.client.Reader(), r.id, c.identity)
	if err != nil && !errors.Is(err, stages.ErrParticipantNotFound) {
		return "", "", false, err
	}

	if existing != nil {
		if previous := r.conns[existing.ID]; previous != nil && previous != c {
			previous.displaced = true
			previous.sendMessage(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_SessionReplaced_{
				SessionReplaced: &stagepb.ServerMessage_SessionReplaced{},
			}})
			previous.close()
		}

		if existing.DisplayName != name {
			name = displaynames.Dedupe(name, s.names(existing.ID))
		}

		err = h.tx(ctx, func(tx database.Tx) error {
			if existing.DisplayName != name {
				if renameErr := h.store.Rename(ctx, tx, existing.ID, name); renameErr != nil {
					return renameErr
				}
			}

			// Re-admit a participant whose earlier socket's close already
			// marked them gone, which would hide them from the stage.
			return h.store.SetLeftAt(ctx, tx, existing.ID, nil)
		})

		return existing.ID, "", err == nil, err
	}

	if len(r.conns) >= MaxParticipants {
		c.sendError(fmt.Sprintf("Stage is full (%d max).", MaxParticipants))
		c.close()

		return "", "", false, nil
	}

	name = displaynames.Dedupe(name, s.names(""))

	// Somebody arriving after the lobby lands in the audience, and stays there
	// until the stage is back in the lobby.
	lateJoin := s.stage.Phase != stages.PhaseLobby

	participant := &stages.Participant{
		ID:          ulid.Make().String(),
		StageID:     r.id,
		Identity:    c.identity,
		DisplayName: name,
		IsAudience:  lateJoin,
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.AddParticipant(ctx, tx, participant)
	}); err != nil {
		return "", "", false, err
	}

	if lateJoin {
		announcement = name + " joined the audience."
	}

	return participant.ID, announcement, true, nil
}

// claimHost makes the joiner host of a stage nobody hosts, when they are
// signed in; a guest never hosts. It answers what to announce.
func (h *Hub) claimHost(ctx context.Context, r *room, c *conn, stage *stages.Stage, name string) (string, error) {
	if stage.HasHost() || !c.identity.SignedIn() {
		return "", nil
	}

	if err := h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetHost(ctx, tx, r.id, c.identity)
	}); err != nil {
		return "", err
	}

	return name + " is host.", nil
}

// sendAvailablePlays sends a joiner the scripts they may select: the public
// ones and their own.
func (h *Hub) sendAvailablePlays(ctx context.Context, c *conn) error {
	plays, err := h.library.List(ctx, &scripts.Viewer{UserID: c.identity.UserID}, "")
	if err != nil {
		return err
	}

	out := make([]*scriptspb.Play, 0, len(plays))
	for _, play := range plays {
		out = append(out, scriptsgrpc.ConvertPlay(play))
	}

	c.sendMessage(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_AvailablePlays_{
		AvailablePlays: &stagepb.ServerMessage_AvailablePlays{Plays: out},
	}})

	return nil
}

// isHost reports whether the socket's identity hosts the stage.
func isHost(s *snapshot, c *conn) bool {
	return s.stage.HasHost() && s.stage.Host.Matches(c.identity)
}

func (h *Hub) toggleCastMember(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_ToggleCastMember) error {
	if !selfOnly(c, m.GetParticipantId()) {
		return nil
	}

	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	if s.stage.DirectorMode && isHost(s, c) {
		c.sendError("Exit director mode to take a role.")

		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.ToggleRole(ctx, tx, r.id, c.participantID, m.GetCastMemberId())
	}); err != nil {
		return err
	}

	return h.broadcastState(ctx, r)
}

func (h *Hub) setAudience(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_SetAudience) error {
	if !selfOnly(c, m.GetParticipantId()) {
		return nil
	}

	if !m.GetIsAudience() {
		s, err := h.snapshot(ctx, r.id)
		if err != nil {
			return err
		}

		if s.stage.Phase != stages.PhaseLobby {
			c.sendError("You can join the cast when the stage returns to the lobby.")

			return nil
		}
	}

	if err := h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetAudience(ctx, tx, c.participantID, m.GetIsAudience())
	}); err != nil {
		return err
	}

	return h.broadcastState(ctx, r)
}

func (h *Hub) setReady(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_SetReady) error {
	if !selfOnly(c, m.GetParticipantId()) {
		return nil
	}

	if err := h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetReady(ctx, tx, c.participantID, m.GetIsReady())
	}); err != nil {
		return err
	}

	return h.broadcastState(ctx, r)
}

// selectPlay picks the stage's play, which clears every role on it.
func (h *Hub) selectPlay(ctx context.Context, r *room, m *stagepb.ClientMessage_SelectPlay) error {
	if err := h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetPlay(ctx, tx, r.id, m.GetPlayId())
	}); err != nil {
		return err
	}

	return h.broadcastState(ctx, r)
}

// sendChat sends a message to the stage, or only to its performers when it is
// cast-only. A message is cut to MaxChatLength here rather than refused.
func (h *Hub) sendChat(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_SendChat) error {
	if !selfOnly(c, m.GetParticipantId()) {
		return nil
	}

	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	sender := s.participant(c.participantID)

	senderName := "Unknown"
	if sender != nil {
		senderName = sender.DisplayName
	}

	msg := &stagepb.ServerMessage{Message: &stagepb.ServerMessage_ChatMessage{ChatMessage: &stagepb.ChatMessage{
		Id:         ulid.Make().String(),
		SenderId:   c.participantID,
		SenderName: senderName,
		Text:       TruncateChat(m.GetText()),
		CastOnly:   m.GetCastOnly(),
		Timestamp:  timestamppb.New(h.now()),
	}}}

	if !m.GetCastOnly() {
		r.broadcast(msg, "")

		return nil
	}

	for _, performer := range s.performers() {
		if recipient := r.conns[performer.ID]; recipient != nil {
			recipient.sendMessage(msg)
		}
	}

	return nil
}

// TruncateChat cuts a message longer than MaxChatLength characters, marking
// the cut with an ellipsis that keeps it at the limit.
func TruncateChat(text string) string {
	if utf8.RuneCountInString(text) <= MaxChatLength {
		return text
	}

	return string([]rune(text)[:MaxChatLength-1]) + "…"
}

// startPlay counts down and then starts the performance, once every role is
// filled and every performer is ready. In director mode only the director may
// start it.
func (h *Hub) startPlay(ctx context.Context, r *room, c *conn) error {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	switch {
	case s.stage.Phase != stages.PhaseLobby:
		c.sendError("Play already started.")

		return nil
	case s.play == nil:
		c.sendError("Select a play first.")

		return nil
	case s.stage.DirectorMode && !isHost(s, c):
		c.sendError("Only the director can start the play.")

		return nil
	}

	if problems := s.startProblems(); problems != "" {
		c.sendError("Cannot start: " + problems)

		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetPhaseAndLine(ctx, tx, r.id, stages.PhaseCountdown, 0)
	}); err != nil {
		return err
	}

	seconds := int32(math.Max(1, math.Ceil(h.timings.Countdown.Seconds())))
	r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_CountdownStart_{
		CountdownStart: &stagepb.ServerMessage_CountdownStart{DurationSeconds: seconds},
	}}, "")

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	r.countdown = h.later(r, h.timings.Countdown, func(ctx context.Context) {
		r.countdown = nil

		if countdownErr := h.finishCountdown(ctx, r); countdownErr != nil {
			h.o11y.Logger().WithValue("stage.id", r.id).Error("starting a performance", countdownErr)
		}
	})

	return nil
}

// finishCountdown starts the performance, unless the stage left the countdown
// while it ran.
func (h *Hub) finishCountdown(ctx context.Context, r *room) error {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	if s.stage.Phase != stages.PhaseCountdown {
		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetPhaseAndLine(ctx, tx, r.id, stages.PhasePerforming, 0)
	}); err != nil {
		return err
	}

	r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_PlayStarted_{PlayStarted: &stagepb.ServerMessage_PlayStarted{}}}, "")

	return h.broadcastState(ctx, r)
}

// nextLine advances the playhead. In the default mode only somebody speaking
// the current line may; in director mode only the director may; the audience
// never may.
func (h *Hub) nextLine(ctx context.Context, r *room, c *conn) error {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	me := s.participant(c.participantID)

	switch {
	case s.stage.PausedByHost:
		c.sendError("Play is paused by host.")

		return nil
	case me != nil && me.IsAudience:
		c.sendError("Audience cannot advance lines.")

		return nil
	case s.stage.DirectorMode && !isHost(s, c):
		c.sendError("Director controls line advance.")

		return nil
	case !s.stage.DirectorMode && !s.isSpeaking(c.participantID):
		c.sendError("Not your line.")

		return nil
	}

	return h.advance(ctx, r, s)
}

// advance moves the playhead one step, and past the last line ends the play.
func (h *Hub) advance(ctx context.Context, r *room, s *snapshot) error {
	if s.play == nil {
		return nil
	}

	next := s.stage.LineIndex + 1
	ended := int(next) >= len(s.lineIndices())

	if err := h.tx(ctx, func(tx database.Tx) error {
		if ended {
			return h.store.SetPhase(ctx, tx, r.id, stages.PhaseDiscussion)
		}

		return h.store.SetLine(ctx, tx, r.id, next)
	}); err != nil {
		return err
	}

	if ended {
		r.clearVacancies()
		r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_PlayEnded_{PlayEnded: &stagepb.ServerMessage_PlayEnded{}}}, "")
	} else {
		r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_LineAdvanced_{
			LineAdvanced: &stagepb.ServerMessage_LineAdvanced{CurrentLineIndex: next},
		}}, "")
	}

	return h.broadcastState(ctx, r)
}

// resetPlay sends everybody back to the lobby, unready, with their roles kept.
// From the discussion anybody may ask, to perform again; from anywhere else
// only the host may, so a performance is not reset by accident.
func (h *Hub) resetPlay(ctx context.Context, r *room, c *conn) error {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	if !isHost(s, c) && s.stage.Phase != stages.PhaseDiscussion {
		c.sendError("Only the host can reset the play.")

		return nil
	}

	return h.reset(ctx, r)
}

// reset returns the stage to the lobby, unpaused, with no vacancies open.
func (h *Hub) reset(ctx context.Context, r *room) error {
	r.clearVacancies()

	if r.countdown != nil {
		r.countdown.Stop()
		r.countdown = nil
	}

	if err := h.tx(ctx, func(tx database.Tx) error {
		if err := h.store.Reset(ctx, tx, r.id); err != nil {
			return err
		}

		return h.store.SetPaused(ctx, tx, r.id, false)
	}); err != nil {
		return err
	}

	return h.broadcastState(ctx, r)
}

// hostPause pauses or resumes the performance. Host only.
func (h *Hub) hostPause(ctx context.Context, r *room, c *conn, paused bool) error {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	if !isHost(s, c) {
		c.sendError(errHostOnly)

		return nil
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetPaused(ctx, tx, r.id, paused)
	}); err != nil {
		return err
	}

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	if paused {
		h.system(r, "Host paused the play.")
	} else {
		h.system(r, "Host resumed the play.")
	}

	return nil
}

// hostEndPlay stops the performance and returns everybody to the lobby. Host
// only.
func (h *Hub) hostEndPlay(ctx context.Context, r *room, c *conn) error {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	if !isHost(s, c) {
		c.sendError(errHostOnly)

		return nil
	}

	if err = h.reset(ctx, r); err != nil {
		return err
	}

	h.system(r, "Host ended the play.")

	return nil
}

// audienceClaimRole gives an open vacancy to the audience member who claims
// it, during its countdown. The play stays paused for the host to resume.
func (h *Hub) audienceClaimRole(ctx context.Context, r *room, c *conn, m *stagepb.ClientMessage_AudienceClaimRole) error {
	v, open := r.vacancies[m.GetCastMemberId()]
	if !open || v.deadline == nil {
		c.sendError("That role is not open to claims.")

		return nil
	}

	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	me := s.participant(c.participantID)
	if me == nil || !me.IsAudience {
		c.sendError("Only audience members can claim a vacant role.")

		return nil
	}

	r.clearVacancy(v.castMemberID)

	if err = h.tx(ctx, func(tx database.Tx) error {
		if audienceErr := h.store.SetAudience(ctx, tx, c.participantID, false); audienceErr != nil {
			return audienceErr
		}

		return h.store.AssignRole(ctx, tx, r.id, c.participantID, v.castMemberID)
	}); err != nil {
		return err
	}

	r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_RoleVacancyCleared_{
		RoleVacancyCleared: &stagepb.ServerMessage_RoleVacancyCleared{CastMemberId: v.castMemberID},
	}}, "")

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	h.system(r, fmt.Sprintf("%s claimed %s. Host can resume when ready.", me.DisplayName, v.castMemberName))

	return nil
}

// disconnected is a socket's departure: the participant is marked gone, any
// roles they held mid-performance fall vacant, and a stage left empty is
// deleted after CleanupDelay unless somebody joins first. A socket another
// took over from, or one closed because the process is stopping, leaves its
// participant as they were.
func (h *Hub) disconnected(ctx context.Context, c *conn) {
	if c.stageID == "" || h.isClosing() {
		return
	}

	r, ok := h.existingRoom(c.stageID)
	if !ok {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.gone || c.displaced || r.conns[c.participantID] != c {
		return
	}

	delete(r.conns, c.participantID)

	if err := h.depart(ctx, r, c); err != nil {
		h.o11y.Logger().WithValue("stage.id", r.id).Error("handling a departure", err)
	}

	if len(r.conns) == 0 {
		r.cleanup = h.later(r, h.timings.CleanupDelay, func(ctx context.Context) {
			r.cleanup = nil

			if len(r.conns) > 0 {
				return
			}

			if err := h.deleteStage(ctx, r); err != nil {
				h.o11y.Logger().WithValue("stage.id", r.id).Error("deleting an empty stage", err)
			}
		})
	}
}

// depart marks a participant gone, withdraws offers awaiting their answer,
// opens vacancies for the roles they held, and, when they were the host's
// last socket, starts the host's grace period.
func (h *Hub) depart(ctx context.Context, r *room, c *conn) error {
	before, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	leaving := before.participant(c.participantID)

	name := "Host"
	if leaving != nil {
		name = leaving.DisplayName
	}

	if err = h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetLeftAt(ctx, tx, c.participantID, new(h.now()))
	}); err != nil {
		return err
	}

	h.withdrawOffersTo(r, c.participantID, name)

	if err = h.broadcastState(ctx, r); err != nil {
		return err
	}

	if leaving != nil {
		if err = h.openVacancies(ctx, r, before, leaving); err != nil {
			return err
		}
	}

	h.hostLeft(r, before, c, name)

	return nil
}

// openVacancies pauses a performance whose role-holder left, and opens a
// vacancy for each role they held. Outside director mode the audience gets
// VacancyClaim to claim each before the play ends; in director mode there is
// no countdown, and the director reassigns.
func (h *Hub) openVacancies(ctx context.Context, r *room, s *snapshot, leaving *stages.Participant) error {
	if s.stage.Phase != stages.PhasePerforming || len(leaving.CastMemberIDs) == 0 {
		return nil
	}

	var vacated []*scripts.CastMember

	for _, id := range leaving.CastMemberIDs {
		if member := s.castMember(id); member != nil {
			vacated = append(vacated, member)
		}
	}

	if len(vacated) == 0 {
		return nil
	}

	if err := h.tx(ctx, func(tx database.Tx) error {
		return h.store.SetPaused(ctx, tx, r.id, true)
	}); err != nil {
		return err
	}

	for _, member := range vacated {
		if _, open := r.vacancies[member.ID]; open {
			continue
		}

		v := &vacancy{castMemberID: member.ID, castMemberName: member.Name, previousHolder: leaving.ID}

		if !s.stage.DirectorMode {
			v.deadline = new(h.now().Add(h.timings.VacancyClaim))
			v.timer = h.later(r, h.timings.VacancyClaim, func(ctx context.Context) {
				if err := h.expireVacancy(ctx, r, v); err != nil {
					h.o11y.Logger().WithValue("stage.id", r.id).Error("expiring a vacancy", err)
				}
			})
		}

		r.vacancies[member.ID] = v
		r.broadcast(vacatedMessage(v), "")

		if s.stage.DirectorMode {
			h.system(r, fmt.Sprintf("%s disconnected, so %s is vacant. The director can reassign the role.", leaving.DisplayName, member.Name))
		} else {
			h.system(r, fmt.Sprintf("%s disconnected, so %s is vacant. Any audience member can claim it.", leaving.DisplayName, member.Name))
		}
	}

	return h.broadcastState(ctx, r)
}

// expireVacancy ends the play when a vacancy's countdown ran out unclaimed.
func (h *Hub) expireVacancy(ctx context.Context, r *room, v *vacancy) error {
	if r.vacancies[v.castMemberID] != v {
		return nil
	}

	r.clearVacancy(v.castMemberID)

	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	if s.stage.Phase != stages.PhasePerforming {
		return nil
	}

	r.clearVacancies()

	if err = h.tx(ctx, func(tx database.Tx) error {
		if phaseErr := h.store.SetPhase(ctx, tx, r.id, stages.PhaseDiscussion); phaseErr != nil {
			return phaseErr
		}

		return h.store.SetPaused(ctx, tx, r.id, false)
	}); err != nil {
		return err
	}

	r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_PlayEnded_{PlayEnded: &stagepb.ServerMessage_PlayEnded{}}}, "")
	h.system(r, fmt.Sprintf("Play ended: %s went vacant.", v.castMemberName))

	return h.broadcastState(ctx, r)
}

func vacatedMessage(v *vacancy) *stagepb.ServerMessage {
	vacancy := &stagepb.RoleVacancy{CastMemberId: v.castMemberID, CastMemberName: v.castMemberName}
	if v.deadline != nil {
		vacancy.ClaimDeadline = timestamppb.New(*v.deadline)
	}

	return &stagepb.ServerMessage{Message: &stagepb.ServerMessage_RoleVacated{RoleVacated: vacancy}}
}

// clearVacancy closes one vacancy and stops its countdown.
func (r *room) clearVacancy(castMemberID string) {
	if v, ok := r.vacancies[castMemberID]; ok {
		if v.timer != nil {
			v.timer.Stop()
		}

		delete(r.vacancies, castMemberID)
	}
}

func (r *room) clearVacancies() {
	for id := range r.vacancies {
		r.clearVacancy(id)
	}
}
