package grpc

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	stagesdomain "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/stages"
	stagessvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/stages"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	playscriptsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts/grpc/converters"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const o11yName = "stages_service"

var _ stagessvc.StagesServiceServer = (*serviceImpl)(nil)

type serviceImpl struct {
	stagessvc.UnimplementedStagesServiceServer
	tracer          tracing.Tracer
	logger          logging.Logger
	stagesRepo      stagesdomain.Repository
	playScriptsRepo playscripts.Repository
}

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	stagesRepo stagesdomain.Repository,
	playScriptsRepo playscripts.Repository,
) stagessvc.StagesServiceServer {
	return &serviceImpl{
		logger:          logging.NewNamedLogger(logger, o11yName),
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		stagesRepo:      stagesRepo,
		playScriptsRepo: playScriptsRepo,
	}
}

func looksLikeStageCode(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

func generateStageCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		for i := range randomBytes {
			randomBytes[i] = byte(time.Now().UnixNano() >> (i * 8))
		}
	}
	code := make([]byte, 6)
	for i := 0; i < 6; i++ {
		code[i] = chars[int(randomBytes[i])%len(chars)]
	}
	return string(code)
}

func (s *serviceImpl) resolveStageID(ctx context.Context, id string) string {
	id = strings.TrimSpace(id)
	codeUpper := strings.ToUpper(id)
	var stage *stagesdomain.Stage
	var err error
	if looksLikeStageCode(codeUpper) {
		stage, err = s.stagesRepo.GetStageByCode(ctx, codeUpper)
		if errors.Is(err, sql.ErrNoRows) {
			stage, err = s.stagesRepo.GetStageByID(ctx, id)
		}
	} else {
		stage, err = s.stagesRepo.GetStageByID(ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			stage, err = s.stagesRepo.GetStageByCode(ctx, codeUpper)
		}
	}
	if err != nil || stage == nil {
		return ""
	}
	return stage.ID
}

func (s *serviceImpl) buildStageState(ctx context.Context, stage *stagesdomain.Stage) *stagesdomain.StageState {
	state := &stagesdomain.StageState{
		ID:               stage.ID,
		Code:             stage.Code,
		Participants:     stage.Participants,
		CurrentLineIndex: stage.CurrentLineIndex,
		Phase:            stage.Phase,
	}
	if stage.PlayScriptID != nil && *stage.PlayScriptID != "" {
		play, err := s.playScriptsRepo.GetPlayScript(ctx, *stage.PlayScriptID, false)
		if err == nil {
			state.SelectedPlay = play
		}
	}
	return state
}

func (s *serviceImpl) JoinStage(ctx context.Context, req *stagessvc.JoinStageRequest) (*stagessvc.JoinStageResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := strings.TrimSpace(req.StageId)
	participantName := strings.TrimSpace(req.ParticipantName)
	if participantName == "" {
		return nil, status.Error(codes.InvalidArgument, "participant name is required")
	}

	var stage *stagesdomain.Stage
	var err error
	if stageID != "" {
		codeUpper := strings.ToUpper(stageID)
		if looksLikeStageCode(codeUpper) {
			stage, err = s.stagesRepo.GetStageByCode(ctx, codeUpper)
			if errors.Is(err, sql.ErrNoRows) {
				stage, err = s.stagesRepo.GetStageByID(ctx, stageID)
			}
		} else {
			stage, err = s.stagesRepo.GetStageByID(ctx, stageID)
			if errors.Is(err, sql.ErrNoRows) {
				stage, err = s.stagesRepo.GetStageByCode(ctx, codeUpper)
			}
		}
	}

	if stage == nil || errors.Is(err, sql.ErrNoRows) {
		stage, err = s.stagesRepo.CreateStage(ctx, &stagesdomain.StageCreationInput{
			ID:   identifiers.New(),
			Code: generateStageCode(),
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to create stage")
		}
	} else if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	participant, err := s.stagesRepo.RejoinParticipant(ctx, stage.ID, participantName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, status.Error(codes.Internal, "failed to join stage")
	}
	if participant == nil {
		participant, err = s.stagesRepo.AddParticipant(ctx, &stagesdomain.ParticipantCreationInput{
			ID:          identifiers.New(),
			StageID:     stage.ID,
			DisplayName: participantName,
		})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to join stage")
		}
	}

	stage, err = s.stagesRepo.GetStageByID(ctx, stage.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	stageState := s.buildStageState(ctx, stage)

	filter := filtering.DefaultQueryFilter()
	playsResult, err := s.playScriptsRepo.GetPublicPlayScripts(ctx, filter)
	if err != nil {
		playsResult = &filtering.QueryFilteredResult[playscripts.PlayScript]{Data: []*playscripts.PlayScript{}}
	}
	plays := playsResult.Data
	if plays == nil {
		plays = []*playscripts.PlayScript{}
	}

	return &stagessvc.JoinStageResponse{
		ParticipantId: participant.ID,
		StageState:    convertStageStateToGRPC(stageState),
		Plays:         convertPlayScriptsToGRPC(plays),
	}, nil
}

func (s *serviceImpl) GetStageState(ctx context.Context, req *stagessvc.GetStageStateRequest) (*stagessvc.GetStageStateResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	stage, err := s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	return &stagessvc.GetStageStateResponse{
		StageState: convertStageStateToGRPC(s.buildStageState(ctx, stage)),
	}, nil
}

func (s *serviceImpl) UpdateParticipant(ctx context.Context, req *stagessvc.UpdateParticipantRequest) (*stagessvc.UpdateParticipantResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	input := &stagesdomain.ParticipantUpdateInput{
		IsDirector: req.IsDirector,
		IsReady:    req.IsReady,
		IsAudience: req.IsAudience,
	}
	if err := s.stagesRepo.UpdateParticipant(ctx, req.ParticipantId, input); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "participant not found")
		}
		return nil, status.Error(codes.Internal, "failed to update participant")
	}

	stage, err := s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	return &stagessvc.UpdateParticipantResponse{
		StageState: convertStageStateToGRPC(s.buildStageState(ctx, stage)),
	}, nil
}

func (s *serviceImpl) ToggleRole(ctx context.Context, req *stagessvc.ToggleRoleRequest) (*stagessvc.ToggleRoleResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	roles, err := s.stagesRepo.GetParticipantRoles(ctx, req.ParticipantId)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, span, "failed to get participant roles")
		return nil, status.Error(codes.Internal, "failed to get participant roles")
	}

	hasRole := false
	for _, role := range roles {
		if role == req.CastMemberId {
			hasRole = true
			break
		}
	}

	if hasRole {
		if err = s.stagesRepo.RemoveParticipantRole(ctx, req.ParticipantId, req.CastMemberId); err != nil {
			return nil, status.Error(codes.Internal, "failed to update role")
		}
	} else {
		if err = s.stagesRepo.AddParticipantRole(ctx, req.ParticipantId, req.CastMemberId); err != nil {
			return nil, status.Error(codes.Internal, "failed to update role")
		}
	}

	participants, err := s.stagesRepo.GetActiveParticipantsForStage(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get participants")
	}

	stage, err := s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	stageState := s.buildStageState(ctx, stage)

	resp := &stagessvc.ToggleRoleResponse{
		StageState: convertStageStateToGRPC(stageState),
	}
	for _, p := range participants {
		resp.Participants = append(resp.Participants, convertParticipantToGRPC(p))
	}
	return resp, nil
}

func (s *serviceImpl) UpdateStage(ctx context.Context, req *stagessvc.UpdateStageRequest) (*stagessvc.UpdateStageResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	input := &stagesdomain.StageUpdateInput{}
	if req.PlayScriptId != nil {
		input.PlayScriptID = req.PlayScriptId
	}
	if req.CurrentLineIndex != nil {
		idx := int(*req.CurrentLineIndex)
		input.CurrentLineIndex = &idx
	}
	if req.Phase != nil && *req.Phase != stagessvc.StagePhase_STAGE_PHASE_UNSPECIFIED {
		switch *req.Phase {
		case stagessvc.StagePhase_STAGE_PHASE_LOBBY:
			p := stagesdomain.StagePhaseLobby
			input.Phase = &p
		case stagessvc.StagePhase_STAGE_PHASE_COUNTDOWN:
			p := stagesdomain.StagePhaseCountdown
			input.Phase = &p
		case stagessvc.StagePhase_STAGE_PHASE_PERFORMING:
			p := stagesdomain.StagePhasePerforming
			input.Phase = &p
		case stagessvc.StagePhase_STAGE_PHASE_DISCUSSION:
			p := stagesdomain.StagePhaseDiscussion
			input.Phase = &p
		}
	}

	if err := s.stagesRepo.UpdateStage(ctx, stageID, input); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "stage not found")
		}
		return nil, status.Error(codes.Internal, "failed to update stage")
	}

	stage, err := s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	return &stagessvc.UpdateStageResponse{
		StageState: convertStageStateToGRPC(s.buildStageState(ctx, stage)),
	}, nil
}

func (s *serviceImpl) SendChat(ctx context.Context, req *stagessvc.SendChatRequest) (*stagessvc.SendChatResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	if req.ParticipantId == "" || strings.TrimSpace(req.Text) == "" {
		return nil, status.Error(codes.InvalidArgument, "participantId and text are required")
	}

	participant, err := s.stagesRepo.GetParticipant(ctx, req.ParticipantId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "participant not found")
		}
		return nil, status.Error(codes.Internal, "failed to get participant")
	}

	chatMessage := &stagesdomain.ChatMessage{
		ID:         identifiers.New(),
		SenderID:   participant.ID,
		SenderName: participant.DisplayName,
		Text:       strings.TrimSpace(req.Text),
		CastOnly:   req.CastOnly,
		Timestamp:  time.Now().UnixMilli(),
	}

	return &stagessvc.SendChatResponse{
		Message: convertChatMessageToGRPC(chatMessage),
	}, nil
}

func (s *serviceImpl) LeaveStage(ctx context.Context, req *stagessvc.LeaveStageRequest) (*stagessvc.LeaveStageResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	if err := s.stagesRepo.SetParticipantLeft(ctx, req.ParticipantId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "participant not found")
		}
		return nil, status.Error(codes.Internal, "failed to leave stage")
	}

	return &stagessvc.LeaveStageResponse{}, nil
}

func (s *serviceImpl) StartPlay(ctx context.Context, req *stagessvc.StartPlayRequest) (*stagessvc.StartPlayResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	countdownPhase := stagesdomain.StagePhaseCountdown
	if err := s.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
		Phase: &countdownPhase,
	}); err != nil {
		return nil, status.Error(codes.Internal, "failed to start play")
	}

	stage, err := s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	return &stagessvc.StartPlayResponse{
		StageState: convertStageStateToGRPC(s.buildStageState(ctx, stage)),
	}, nil
}

func (s *serviceImpl) NextLine(ctx context.Context, req *stagessvc.NextLineRequest) (*stagessvc.NextLineResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	stage, err := s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	if stage.PlayScriptID == nil || *stage.PlayScriptID == "" {
		return nil, status.Error(codes.InvalidArgument, "no play selected on stage")
	}

	play, err := s.playScriptsRepo.GetPlayScript(ctx, *stage.PlayScriptID, false)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get play")
	}

	turnIndices := playscripts.GetSortedUniqueLineIndices(play)
	newTurnIndex := stage.CurrentLineIndex + 1

	if newTurnIndex >= len(turnIndices) {
		discussionPhase := stagesdomain.StagePhaseDiscussion
		if err = s.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
			Phase: &discussionPhase,
		}); err != nil {
			return nil, status.Error(codes.Internal, "failed to advance line")
		}

		stage, err = s.stagesRepo.GetStageByID(ctx, stageID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get stage")
		}
		resp := &stagessvc.NextLineResponse{
			StageState:   convertStageStateToGRPC(s.buildStageState(ctx, stage)),
			LineAdvanced: false,
			PlayEnded:    true,
		}
		if play.PrecedesID != nil && *play.PrecedesID != "" {
			if nextPlay, nextPlayErr := s.playScriptsRepo.GetPlayScript(ctx, *play.PrecedesID, false); nextPlayErr == nil {
				resp.NextPlay = playscriptsgrpc.ConvertPlayScriptToGRPCPlayScript(nextPlay)
			}
		}
		return resp, nil
	}

	if err = s.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
		CurrentLineIndex: &newTurnIndex,
	}); err != nil {
		return nil, status.Error(codes.Internal, "failed to advance line")
	}

	stage, err = s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}
	return &stagessvc.NextLineResponse{
		StageState:       convertStageStateToGRPC(s.buildStageState(ctx, stage)),
		LineAdvanced:     true,
		CurrentLineIndex: int32(newTurnIndex),
		PlayEnded:        false,
	}, nil
}

func (s *serviceImpl) ResetPlay(ctx context.Context, req *stagessvc.ResetPlayRequest) (*stagessvc.ResetPlayResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	stageID := s.resolveStageID(ctx, req.StageId)
	if stageID == "" {
		return nil, status.Error(codes.NotFound, "stage not found")
	}

	lobbyPhase := stagesdomain.StagePhaseLobby
	lineIndex := 0
	if err := s.stagesRepo.UpdateStage(ctx, stageID, &stagesdomain.StageUpdateInput{
		Phase:            &lobbyPhase,
		CurrentLineIndex: &lineIndex,
	}); err != nil {
		return nil, status.Error(codes.Internal, "failed to reset play")
	}

	participants, err := s.stagesRepo.GetActiveParticipantsForStage(ctx, stageID)
	if err != nil {
		observability.AcknowledgeError(err, s.logger, span, "failed to get active participants")
	}
	for _, p := range participants {
		isReady := false
		if updateErr := s.stagesRepo.UpdateParticipant(ctx, p.ID, &stagesdomain.ParticipantUpdateInput{
			IsReady: &isReady,
		}); updateErr != nil {
			observability.AcknowledgeError(updateErr, s.logger, span, "failed to update participant ready state")
		}
	}

	stage, err := s.stagesRepo.GetStageByID(ctx, stageID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get stage")
	}

	return &stagessvc.ResetPlayResponse{
		StageState: convertStageStateToGRPC(s.buildStageState(ctx, stage)),
	}, nil
}
