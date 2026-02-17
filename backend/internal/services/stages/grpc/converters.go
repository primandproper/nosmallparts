package grpc

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	play_scripts "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"
	stagessvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/stages"
	playscriptsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts/grpc/converters"
)

func convertStagePhaseToGRPC(p stagesdomain.StagePhase) stagessvc.StagePhase {
	switch p {
	case stagesdomain.StagePhaseLobby:
		return stagessvc.StagePhase_STAGE_PHASE_LOBBY
	case stagesdomain.StagePhaseCountdown:
		return stagessvc.StagePhase_STAGE_PHASE_COUNTDOWN
	case stagesdomain.StagePhasePerforming:
		return stagessvc.StagePhase_STAGE_PHASE_PERFORMING
	case stagesdomain.StagePhaseDiscussion:
		return stagessvc.StagePhase_STAGE_PHASE_DISCUSSION
	default:
		return stagessvc.StagePhase_STAGE_PHASE_UNSPECIFIED
	}
}

func convertParticipantToGRPC(p *stagesdomain.StageParticipant) *stagessvc.StageParticipant {
	if p == nil {
		return nil
	}
	return &stagessvc.StageParticipant{
		Id:            p.ID,
		StageId:       p.StageID,
		Name:          p.DisplayName,
		CastMemberIds: p.CastMemberIDs,
		IsDirector:    p.IsDirector,
		IsReady:       p.IsReady,
		IsAudience:    p.IsAudience,
	}
}

func convertChatMessageToGRPC(m *stagesdomain.ChatMessage) *stagessvc.ChatMessage {
	if m == nil {
		return nil
	}
	return &stagessvc.ChatMessage{
		Id:         m.ID,
		SenderId:   m.SenderID,
		SenderName: m.SenderName,
		Text:       m.Text,
		CastOnly:   m.CastOnly,
		Timestamp:  m.Timestamp,
	}
}

func convertStageStateToGRPC(s *stagesdomain.StageState) *stagessvc.StageState {
	if s == nil {
		return nil
	}
	out := &stagessvc.StageState{
		Id:               s.ID,
		Code:             s.Code,
		Phase:            convertStagePhaseToGRPC(s.Phase),
		CurrentLineIndex: int32(s.CurrentLineIndex),
	}
	for _, p := range s.Participants {
		out.Participants = append(out.Participants, convertParticipantToGRPC(p))
	}
	if s.SelectedPlay != nil {
		out.SelectedPlay = playscriptsgrpc.ConvertPlayScriptToGRPCPlayScript(s.SelectedPlay)
	}
	return out
}

func convertPlayScriptsToGRPC(plays []*playscripts.PlayScript) []*play_scripts.PlayScript {
	if len(plays) == 0 {
		return nil
	}
	out := make([]*play_scripts.PlayScript, len(plays))
	for i, p := range plays {
		out[i] = playscriptsgrpc.ConvertPlayScriptToGRPCPlayScript(p)
	}
	return out
}
