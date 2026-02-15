package converters

import (
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
)

// ConvertPlayScriptToPlayScriptUpdateRequestInput creates a PlayScriptUpdateRequestInput from a PlayScript.
func ConvertPlayScriptToPlayScriptUpdateRequestInput(x *types.PlayScript) *types.PlayScriptUpdateRequestInput {
	out := &types.PlayScriptUpdateRequestInput{
		Title:       &x.Title,
		Description: &x.Description,
		Author:      &x.Author,
		IsPublic:    &x.IsPublic,
	}

	return out
}

// ConvertPlayScriptCreationRequestInputToPlayScriptDatabaseCreationInput creates a PlayScriptDatabaseCreationInput from a PlayScriptCreationRequestInput.
func ConvertPlayScriptCreationRequestInputToPlayScriptDatabaseCreationInput(x *types.PlayScriptCreationRequestInput, userID, accountID string) *types.PlayScriptDatabaseCreationInput {
	id := x.ID
	if id == "" {
		id = identifiers.New()
	}
	out := &types.PlayScriptDatabaseCreationInput{
		ID:               id,
		Title:            x.Title,
		Description:      x.Description,
		Author:           x.Author,
		IsPublic:         x.IsPublic,
		CreatedByUser:    userID,
		BelongsToAccount: accountID,
		PrecededByID:     x.PrecededByID,
		PrecedesID:       x.PrecedesID,
	}

	return out
}

// ConvertPlayScriptToPlayScriptCreationRequestInput builds a PlayScriptCreationRequestInput from a PlayScript.
func ConvertPlayScriptToPlayScriptCreationRequestInput(x *types.PlayScript) *types.PlayScriptCreationRequestInput {
	castMembers := make([]*types.PlayScriptCastMemberCreationInput, len(x.CastMembers))
	for i, cm := range x.CastMembers {
		castMembers[i] = &types.PlayScriptCastMemberCreationInput{
			ID:        cm.ID,
			Name:      cm.Name,
			SortOrder: cm.SortOrder,
		}
	}

	lines := make([]*types.PlayScriptLineCreationRequestInput, len(x.Lines))
	for i, l := range x.Lines {
		lines[i] = &types.PlayScriptLineCreationRequestInput{
			CastMemberID:  l.CastMemberID,
			LineText:      l.LineText,
			LineIndex:     l.LineIndex,
			Direction:     l.Direction,
			ScrollMode:    l.ScrollMode,
			AutoScrollWpm: l.AutoScrollWpm,
		}
	}

	return &types.PlayScriptCreationRequestInput{
		Title:       x.Title,
		Description: x.Description,
		Author:      x.Author,
		IsPublic:    x.IsPublic,
		CastMembers: castMembers,
		Lines:       lines,
	}
}

// ConvertPlayScriptToPlayScriptDatabaseCreationInput builds a PlayScriptDatabaseCreationInput from a PlayScript.
func ConvertPlayScriptToPlayScriptDatabaseCreationInput(x *types.PlayScript) *types.PlayScriptDatabaseCreationInput {
	return &types.PlayScriptDatabaseCreationInput{
		ID:               x.ID,
		Title:            x.Title,
		Description:      x.Description,
		Author:           x.Author,
		IsPublic:         x.IsPublic,
		CreatedByUser:    x.CreatedByUser,
		BelongsToAccount: x.BelongsToAccount,
	}
}

// ConvertCastMemberCreationInputsToDatabaseInputs converts a slice of PlayScriptCastMemberCreationInput to database creation inputs.
func ConvertCastMemberCreationInputsToDatabaseInputs(playScriptID string, inputs []*types.PlayScriptCastMemberCreationInput) (dbInputs []*types.PlayScriptCastMemberDatabaseCreationInput, idMapping map[string]string) {
	dbInputs = make([]*types.PlayScriptCastMemberDatabaseCreationInput, len(inputs))
	// Map from client-provided ID to generated database ID
	idMapping = make(map[string]string, len(inputs))

	for i, input := range inputs {
		dbID := identifiers.New()
		idMapping[input.ID] = dbID
		dbInputs[i] = &types.PlayScriptCastMemberDatabaseCreationInput{
			ID:           dbID,
			PlayScriptID: playScriptID,
			Name:         input.Name,
			SortOrder:    input.SortOrder,
		}
	}

	return dbInputs, idMapping
}

// ConvertLineCreationInputsToDatabaseInputs converts a slice of PlayScriptLineCreationRequestInput to database creation inputs.
func ConvertLineCreationInputsToDatabaseInputs(playScriptID string, inputs []*types.PlayScriptLineCreationRequestInput, castMemberIDMapping map[string]string) []*types.PlayScriptLineDatabaseCreationInput {
	result := make([]*types.PlayScriptLineDatabaseCreationInput, len(inputs))

	for i, input := range inputs {
		scrollMode := input.ScrollMode
		if scrollMode == "" {
			scrollMode = types.DefaultScrollMode()
		}

		// Map the client-provided cast member ID to the database ID
		dbCastMemberID := castMemberIDMapping[input.CastMemberID]

		result[i] = &types.PlayScriptLineDatabaseCreationInput{
			ID:            identifiers.New(),
			PlayScriptID:  playScriptID,
			CastMemberID:  dbCastMemberID,
			LineText:      input.LineText,
			LineIndex:     input.LineIndex,
			Direction:     input.Direction,
			ScrollMode:    scrollMode,
			AutoScrollWpm: input.AutoScrollWpm,
		}
	}

	return result
}
