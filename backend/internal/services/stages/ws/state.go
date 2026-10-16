package ws

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"
	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"
	scriptsgrpc "github.com/primandproper/nosmallparts/backend/internal/services/scripts/grpc"

	"github.com/oklog/ulid/v2"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// snapshot is a stage as it stands: the row, who is on it, and its play.
type snapshot struct {
	stage        *stages.Stage
	play         *scripts.Play
	participants []*stages.Participant
}

// snapshot reads the stage whole. A stage whose play has been deleted reads
// with no play; the foreign key has already cleared it.
func (h *Hub) snapshot(ctx context.Context, stageID string) (*snapshot, error) {
	q := h.client.Reader()

	stage, err := h.store.GetStage(ctx, q, stageID)
	if err != nil {
		return nil, err
	}

	participants, err := h.store.ActiveParticipants(ctx, q, stageID)
	if err != nil {
		return nil, err
	}

	s := &snapshot{stage: stage, participants: participants}

	if stage.PlayScriptID != "" {
		s.play, err = h.plays.GetPlay(ctx, q, stage.PlayScriptID)
		if err != nil && !errors.Is(err, scripts.ErrNotFound) {
			return nil, err
		}
	}

	return s, nil
}

// proto is the snapshot as StageState.
func (s *snapshot) proto() *stagepb.StageState {
	out := &stagepb.StageState{
		Id:               s.stage.ID,
		Code:             s.stage.Code,
		CurrentLineIndex: s.stage.LineIndex,
		Phase:            protoPhase(s.stage.Phase),
		PausedByHost:     s.stage.PausedByHost,
		DirectorMode:     s.stage.DirectorMode,
		Participants:     make([]*stagepb.Participant, 0, len(s.participants)),
	}

	if s.play != nil {
		out.SelectedPlay = scriptsgrpc.ConvertPlay(s.play)
	}

	for _, p := range s.participants {
		out.Participants = append(out.Participants, &stagepb.Participant{
			Id:            p.ID,
			Name:          p.DisplayName,
			CastMemberIds: slices.Clone(p.CastMemberIDs),
			IsReady:       p.IsReady,
			IsAudience:    p.IsAudience,
			IsHost:        s.isHost(p),
			ForceMuted:    p.ForceMuted,
		})
	}

	return out
}

// isHost reports whether a participant is the stage's host, by the identity
// the host was recorded under.
func (s *snapshot) isHost(p *stages.Participant) bool {
	host := s.stage.Host

	return (host.UserID != "" && host.UserID == p.Identity.UserID) ||
		(host.GuestID != "" && host.GuestID == p.Identity.GuestID)
}

func (s *snapshot) participant(id string) *stages.Participant {
	for _, p := range s.participants {
		if p.ID == id {
			return p
		}
	}

	return nil
}

func (s *snapshot) names(exceptID string) []string {
	var names []string

	for _, p := range s.participants {
		if p.ID != exceptID {
			names = append(names, p.DisplayName)
		}
	}

	return names
}

func (s *snapshot) performers() []*stages.Participant {
	var performers []*stages.Participant

	for _, p := range s.participants {
		if !p.IsAudience {
			performers = append(performers, p)
		}
	}

	return performers
}

func (s *snapshot) castMember(id string) *scripts.CastMember {
	if s.play == nil {
		return nil
	}

	for _, member := range s.play.Cast {
		if member.ID == id {
			return member
		}
	}

	return nil
}

// lineIndices are the play's distinct line indexes in order. The playhead
// counts positions in this list, so several lines sharing an index are one
// step.
func (s *snapshot) lineIndices() []int32 {
	if s.play == nil {
		return nil
	}

	var indices []int32

	for _, line := range s.play.Lines {
		if !slices.Contains(indices, line.Index) {
			indices = append(indices, line.Index)
		}
	}

	slices.Sort(indices)

	return indices
}

// currentSpeakers are the cast members on the current line while performing.
func (s *snapshot) currentSpeakers() []string {
	if s.play == nil || s.stage.Phase != stages.PhasePerforming {
		return nil
	}

	indices := s.lineIndices()
	if int(s.stage.LineIndex) >= len(indices) {
		return nil
	}

	current := indices[s.stage.LineIndex]

	var speakers []string

	for _, line := range s.play.Lines {
		if line.Index == current {
			speakers = append(speakers, line.CastMemberIDs...)
		}
	}

	return speakers
}

// isSpeaking reports whether a participant holds a role on the current line.
func (s *snapshot) isSpeaking(participantID string) bool {
	p := s.participant(participantID)
	if p == nil {
		return false
	}

	speakers := s.currentSpeakers()

	return slices.ContainsFunc(p.CastMemberIDs, func(id string) bool { return slices.Contains(speakers, id) })
}

// startProblems is what stands between the stage and starting the play, such
// as "3 unfilled roles, 2 not ready", or empty when nothing does: every role
// filled by a performer, and every performer ready.
func (s *snapshot) startProblems() string {
	if s.play == nil {
		return "no play selected"
	}

	var problems []string

	performers := s.performers()
	if len(performers) == 0 {
		problems = append(problems, "no performers")
	}

	assigned := map[string]bool{}

	notReady := 0

	for _, p := range performers {
		for _, id := range p.CastMemberIDs {
			assigned[id] = true
		}

		if !p.IsReady {
			notReady++
		}
	}

	unfilled := 0

	for _, member := range s.play.Cast {
		if !assigned[member.ID] {
			unfilled++
		}
	}

	if unfilled > 0 {
		problems = append(problems, fmt.Sprintf("%d unfilled role%s", unfilled, plural(unfilled)))
	}

	if notReady > 0 {
		problems = append(problems, fmt.Sprintf("%d not ready", notReady))
	}

	return strings.Join(problems, ", ")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}

	return "s"
}

func protoPhase(phase stages.Phase) stagepb.StagePhase {
	switch phase {
	case stages.PhaseLobby:
		return stagepb.StagePhase_STAGE_PHASE_LOBBY
	case stages.PhaseCountdown:
		return stagepb.StagePhase_STAGE_PHASE_COUNTDOWN
	case stages.PhasePerforming:
		return stagepb.StagePhase_STAGE_PHASE_PERFORMING
	case stages.PhaseDiscussion:
		return stagepb.StagePhase_STAGE_PHASE_DISCUSSION
	default:
		return stagepb.StagePhase_STAGE_PHASE_UNSPECIFIED
	}
}

// broadcast sends a message to every socket on the room but one.
func (r *room) broadcast(msg *stagepb.ServerMessage, exceptParticipantID string) {
	frame, err := protojson.Marshal(msg)
	if err != nil {
		return
	}

	for id, c := range r.conns {
		if id != exceptParticipantID {
			c.enqueue(frame)
		}
	}
}

// broadcastState reads the stage and sends it to everybody on it.
func (h *Hub) broadcastState(ctx context.Context, r *room) error {
	s, err := h.snapshot(ctx, r.id)
	if err != nil {
		return err
	}

	r.broadcast(stateMessage(s), "")

	return nil
}

// system sends everybody on the stage a line from the stage itself.
func (h *Hub) system(r *room, text string) {
	r.broadcast(&stagepb.ServerMessage{Message: &stagepb.ServerMessage_SystemMessage{SystemMessage: &stagepb.SystemMessage{
		Id:        ulid.Make().String(),
		Text:      text,
		Timestamp: timestamppb.New(h.now()),
	}}}, "")
}

func (h *Hub) now() time.Time {
	return h.client.CurrentTime().UTC()
}

func stateMessage(s *snapshot) *stagepb.ServerMessage {
	return &stagepb.ServerMessage{Message: &stagepb.ServerMessage_StageState{StageState: s.proto()}}
}
