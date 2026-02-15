package converters

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
)

// ConvertScrollModeToGRPC converts a domain ScrollMode to gRPC ScrollMode.
func ConvertScrollModeToGRPC(mode playscripts.ScrollMode) playscriptssvc.ScrollMode {
	switch mode {
	case playscripts.ScrollModeAuto:
		return playscriptssvc.ScrollMode_SCROLL_MODE_AUTO
	case playscripts.ScrollModeManual:
		return playscriptssvc.ScrollMode_SCROLL_MODE_MANUAL
	default:
		return playscriptssvc.ScrollMode_SCROLL_MODE_AUTO
	}
}

// ConvertGRPCScrollModeToScrollMode converts a gRPC ScrollMode to domain ScrollMode.
func ConvertGRPCScrollModeToScrollMode(mode playscriptssvc.ScrollMode) playscripts.ScrollMode {
	switch mode {
	case playscriptssvc.ScrollMode_SCROLL_MODE_AUTO:
		return playscripts.ScrollModeAuto
	case playscriptssvc.ScrollMode_SCROLL_MODE_MANUAL:
		return playscripts.ScrollModeManual
	default:
		return playscripts.ScrollModeAuto // Default to auto
	}
}

// ConvertPlayScriptToGRPCPlayScript converts a domain PlayScript to gRPC PlayScript.
func ConvertPlayScriptToGRPCPlayScript(ps *playscripts.PlayScript) *playscriptssvc.PlayScript {
	castMembers := make([]*playscriptssvc.PlayScriptCastMember, len(ps.CastMembers))
	for i, cm := range ps.CastMembers {
		castMembers[i] = &playscriptssvc.PlayScriptCastMember{
			CreatedAt:    grpcconverters.ConvertTimeToPBTimestamp(cm.CreatedAt),
			Id:           cm.ID,
			PlayScriptId: cm.PlayScriptID,
			Name:         cm.Name,
			Description:  cm.Description,
			SortOrder:    int32(cm.SortOrder),
		}
	}

	lines := make([]*playscriptssvc.PlayScriptLine, len(ps.Lines))
	for i, l := range ps.Lines {
		lines[i] = &playscriptssvc.PlayScriptLine{
			CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(l.CreatedAt),
			Id:            l.ID,
			PlayScriptId:  l.PlayScriptID,
			CastMemberId:  l.CastMemberID,
			LineText:      l.LineText,
			LineIndex:     int32(l.LineIndex),
			Direction:     l.Direction,
			ScrollMode:    ConvertScrollModeToGRPC(l.ScrollMode),
			AutoScrollWpm: int32(l.AutoScrollWpm),
		}
	}

	tags := make([]*playscriptssvc.ScriptTag, len(ps.Tags))
	for i, tag := range ps.Tags {
		tags[i] = ConvertScriptTagToGRPC(tag)
	}

	out := &playscriptssvc.PlayScript{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(ps.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(ps.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(ps.LastUpdatedAt),
		Id:               ps.ID,
		Title:            ps.Title,
		Description:      ps.Description,
		Author:           ps.Author,
		IsPublic:         ps.IsPublic,
		CastMembers:      castMembers,
		Lines:            lines,
		CreatedByUser:    ps.CreatedByUser,
		BelongsToAccount: ps.BelongsToAccount,
		Tags:             tags,
	}
	if ps.PrecededByID != nil {
		out.PrecededById = ps.PrecededByID
	}
	if ps.PrecedesID != nil {
		out.PrecedesId = ps.PrecedesID
	}
	if ps.PrecededBy != nil {
		out.PrecededBy = ConvertPlayScriptToGRPCPlayScript(ps.PrecededBy)
	}
	if ps.Precedes != nil {
		out.Precedes = ConvertPlayScriptToGRPCPlayScript(ps.Precedes)
	}
	return out
}

// ConvertScriptTagToGRPC converts a domain ScriptTag to gRPC ScriptTag.
func ConvertScriptTagToGRPC(tag *playscripts.ScriptTag) *playscriptssvc.ScriptTag {
	if tag == nil {
		return nil
	}
	return &playscriptssvc.ScriptTag{
		Id:          tag.ID,
		Name:        tag.Name,
		Description: tag.Description,
		CreatedAt:   grpcconverters.ConvertTimeToPBTimestamp(tag.CreatedAt),
	}
}

// ConvertGRPCPlayScriptToPlayScript converts a gRPC PlayScript to domain PlayScript.
func ConvertGRPCPlayScriptToPlayScript(ps *playscriptssvc.PlayScript) *playscripts.PlayScript {
	castMembers := make([]*playscripts.PlayScriptCastMember, len(ps.CastMembers))
	for i, cm := range ps.CastMembers {
		castMembers[i] = &playscripts.PlayScriptCastMember{
			CreatedAt:    grpcconverters.ConvertPBTimestampToTime(cm.CreatedAt),
			ID:           cm.Id,
			PlayScriptID: cm.PlayScriptId,
			Name:         cm.Name,
			Description:  cm.Description,
			SortOrder:    int(cm.SortOrder),
		}
	}

	lines := make([]*playscripts.PlayScriptLine, len(ps.Lines))
	for i, l := range ps.Lines {
		lines[i] = &playscripts.PlayScriptLine{
			CreatedAt:     grpcconverters.ConvertPBTimestampToTime(l.CreatedAt),
			ID:            l.Id,
			PlayScriptID:  l.PlayScriptId,
			CastMemberID:  l.CastMemberId,
			LineText:      l.LineText,
			LineIndex:     int(l.LineIndex),
			Direction:     l.Direction,
			ScrollMode:    ConvertGRPCScrollModeToScrollMode(l.ScrollMode),
			AutoScrollWpm: int(l.AutoScrollWpm),
		}
	}

	tags := make([]*playscripts.ScriptTag, len(ps.Tags))
	for i, tag := range ps.Tags {
		tags[i] = ConvertGRPCScriptTagToScriptTag(tag)
	}

	out := &playscripts.PlayScript{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(ps.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(ps.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertPBTimestampToTimePointer(ps.LastUpdatedAt),
		ID:               ps.Id,
		Title:            ps.Title,
		Description:      ps.Description,
		Author:           ps.Author,
		IsPublic:         ps.IsPublic,
		CastMembers:      castMembers,
		Lines:            lines,
		Tags:             tags,
		CreatedByUser:    ps.CreatedByUser,
		BelongsToAccount: ps.BelongsToAccount,
	}
	if ps.PrecededById != nil {
		out.PrecededByID = ps.PrecededById
	}
	if ps.PrecedesId != nil {
		out.PrecedesID = ps.PrecedesId
	}
	if ps.PrecededBy != nil {
		out.PrecededBy = ConvertGRPCPlayScriptToPlayScript(ps.PrecededBy)
	}
	if ps.Precedes != nil {
		out.Precedes = ConvertGRPCPlayScriptToPlayScript(ps.Precedes)
	}
	return out
}

// ConvertGRPCScriptTagToScriptTag converts a gRPC ScriptTag to domain ScriptTag.
func ConvertGRPCScriptTagToScriptTag(tag *playscriptssvc.ScriptTag) *playscripts.ScriptTag {
	if tag == nil {
		return nil
	}
	return &playscripts.ScriptTag{
		ID:          tag.Id,
		Name:        tag.Name,
		Description: tag.Description,
		CreatedAt:   grpcconverters.ConvertPBTimestampToTime(tag.CreatedAt),
	}
}

// ConvertGRPCPlayScriptCreationRequestInputToInputs converts a gRPC creation request input to domain inputs.
func ConvertGRPCPlayScriptCreationRequestInputToInputs(input *playscriptssvc.PlayScriptCreationRequestInput, userID, accountID string) (*playscripts.PlayScriptDatabaseCreationInput, []*playscripts.PlayScriptCastMemberDatabaseCreationInput, []*playscripts.PlayScriptLineDatabaseCreationInput) {
	playScriptID := identifiers.New()

	// Create the play script input
	playScriptInput := &playscripts.PlayScriptDatabaseCreationInput{
		ID:               playScriptID,
		Title:            input.Title,
		Description:      input.Description,
		Author:           input.Author,
		IsPublic:         input.IsPublic,
		CreatedByUser:    userID,
		BelongsToAccount: accountID,
	}

	// Create cast member inputs with ID mapping
	castMemberInputs := make([]*playscripts.PlayScriptCastMemberDatabaseCreationInput, len(input.CastMembers))
	castMemberIDMapping := make(map[string]string, len(input.CastMembers))
	for i, cm := range input.CastMembers {
		dbID := identifiers.New()
		castMemberIDMapping[cm.Id] = dbID
		castMemberInputs[i] = &playscripts.PlayScriptCastMemberDatabaseCreationInput{
			ID:           dbID,
			PlayScriptID: playScriptID,
			Name:         cm.Name,
			Description:  cm.Description,
			SortOrder:    int(cm.SortOrder),
		}
	}

	// Create line inputs using the cast member ID mapping
	lineInputs := make([]*playscripts.PlayScriptLineDatabaseCreationInput, len(input.Lines))
	for i, l := range input.Lines {
		scrollMode := ConvertGRPCScrollModeToScrollMode(l.ScrollMode)
		if scrollMode == "" {
			scrollMode = playscripts.DefaultScrollMode()
		}

		// Map the client-provided cast member ID to the database ID
		dbCastMemberID := castMemberIDMapping[l.CastMemberId]

		lineInputs[i] = &playscripts.PlayScriptLineDatabaseCreationInput{
			ID:            identifiers.New(),
			PlayScriptID:  playScriptID,
			CastMemberID:  dbCastMemberID,
			LineText:      l.LineText,
			LineIndex:     int(l.LineIndex),
			Direction:     l.Direction,
			ScrollMode:    scrollMode,
			AutoScrollWpm: int(l.AutoScrollWpm),
		}
	}

	return playScriptInput, castMemberInputs, lineInputs
}

// ConvertPlayScriptCreationRequestInputToGRPC converts domain creation input to gRPC.
func ConvertPlayScriptCreationRequestInputToGRPC(input *playscripts.PlayScriptCreationRequestInput) *playscriptssvc.PlayScriptCreationRequestInput {
	castMembers := make([]*playscriptssvc.PlayScriptCastMemberCreationInput, len(input.CastMembers))
	for i, cm := range input.CastMembers {
		castMembers[i] = &playscriptssvc.PlayScriptCastMemberCreationInput{
			Id:          cm.ID,
			Name:        cm.Name,
			Description: cm.Description,
			SortOrder:   int32(cm.SortOrder),
		}
	}

	lines := make([]*playscriptssvc.PlayScriptLineCreationInput, len(input.Lines))
	for i, l := range input.Lines {
		lines[i] = &playscriptssvc.PlayScriptLineCreationInput{
			CastMemberId:  l.CastMemberID,
			LineText:      l.LineText,
			LineIndex:     int32(l.LineIndex),
			Direction:     l.Direction,
			ScrollMode:    ConvertScrollModeToGRPC(l.ScrollMode),
			AutoScrollWpm: int32(l.AutoScrollWpm),
		}
	}

	return &playscriptssvc.PlayScriptCreationRequestInput{
		Title:       input.Title,
		Description: input.Description,
		Author:      input.Author,
		IsPublic:    input.IsPublic,
		CastMembers: castMembers,
		Lines:       lines,
	}
}

// ConvertGRPCPlayScriptUpdateRequestInputToPlayScriptUpdateRequestInput converts gRPC update input to domain.
func ConvertGRPCPlayScriptUpdateRequestInputToPlayScriptUpdateRequestInput(input *playscriptssvc.PlayScriptUpdateRequestInput) *playscripts.PlayScriptUpdateRequestInput {
	output := &playscripts.PlayScriptUpdateRequestInput{}

	if input.Title != nil {
		output.Title = input.Title
	}

	if input.Description != nil {
		output.Description = input.Description
	}

	if input.Author != nil {
		output.Author = input.Author
	}

	if input.IsPublic != nil {
		output.IsPublic = input.IsPublic
	}

	if input.PrecededById != nil {
		output.PrecededByID = input.PrecededById
	}
	if input.PrecedesId != nil {
		output.PrecedesID = input.PrecedesId
	}

	return output
}
