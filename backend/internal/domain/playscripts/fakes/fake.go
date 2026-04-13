package fakes

import (
	"time"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"

	fake "github.com/brianvoe/gofakeit/v7"
)

func init() {
	if err := fake.Seed(time.Now().UnixNano()); err != nil {
		panic(err)
	}
}

// BuildFakeID builds a fake ID.
func BuildFakeID() string {
	return identifiers.New()
}

// BuildFakeTime builds a fake time.
func BuildFakeTime() time.Time {
	return fake.Date().Add(0).Truncate(time.Second).UTC()
}

// BuildFakePlayScript builds a fake play script.
func BuildFakePlayScript() *types.PlayScript {
	castMembers := BuildFakePlayScriptCastMembers(3)
	lines := BuildFakePlayScriptLines(5, castMembers)

	return &types.PlayScript{
		ID:               BuildFakeID(),
		Title:            fake.BookTitle(),
		Description:      fake.Sentence(10),
		Author:           fake.Name(),
		IsPublic:         fake.Bool(),
		CastMembers:      castMembers,
		Lines:            lines,
		CreatedAt:        BuildFakeTime(),
		LastUpdatedAt:    nil,
		ArchivedAt:       nil,
		CreatedByUser:    BuildFakeID(),
		BelongsToAccount: BuildFakeID(),
		PrecededByID:     nil,
		PrecedesID:       nil,
	}
}

// BuildFakePlayScriptCastMember builds a fake play script cast member.
func BuildFakePlayScriptCastMember() *types.PlayScriptCastMember {
	return &types.PlayScriptCastMember{
		ID:           BuildFakeID(),
		PlayScriptID: BuildFakeID(),
		Name:         fake.FirstName(),
		SortOrder:    fake.Number(0, 10),
		CreatedAt:    BuildFakeTime(),
	}
}

// BuildFakePlayScriptCastMembers builds a slice of fake play script cast members.
func BuildFakePlayScriptCastMembers(count int) []*types.PlayScriptCastMember {
	playScriptID := BuildFakeID()
	result := make([]*types.PlayScriptCastMember, count)
	for i := 0; i < count; i++ {
		result[i] = &types.PlayScriptCastMember{
			ID:           BuildFakeID(),
			PlayScriptID: playScriptID,
			Name:         fake.FirstName(),
			SortOrder:    i,
			CreatedAt:    BuildFakeTime(),
		}
	}
	return result
}

// BuildFakePlayScriptLine builds a fake play script line.
func BuildFakePlayScriptLine() *types.PlayScriptLine {
	return &types.PlayScriptLine{
		ID:            BuildFakeID(),
		PlayScriptID:  BuildFakeID(),
		CastMemberID:  BuildFakeID(),
		LineText:      fake.Sentence(10),
		LineIndex:     fake.Number(0, 100),
		Direction:     fake.RandomString([]string{"", "angrily", "sadly", "with excitement", "whispering"}),
		ScrollMode:    types.ScrollModeAuto,
		AutoScrollWpm: fake.Number(100, 200),
		CreatedAt:     BuildFakeTime(),
	}
}

// BuildFakePlayScriptLines builds a slice of fake play script lines.
func BuildFakePlayScriptLines(count int, castMembers []*types.PlayScriptCastMember) []*types.PlayScriptLine {
	if len(castMembers) == 0 {
		return nil
	}

	playScriptID := castMembers[0].PlayScriptID
	result := make([]*types.PlayScriptLine, count)
	for i := 0; i < count; i++ {
		castMember := castMembers[i%len(castMembers)]
		result[i] = &types.PlayScriptLine{
			ID:            BuildFakeID(),
			PlayScriptID:  playScriptID,
			CastMemberID:  castMember.ID,
			LineText:      fake.Sentence(10),
			LineIndex:     i,
			Direction:     fake.RandomString([]string{"", "angrily", "sadly", "with excitement", "whispering"}),
			ScrollMode:    types.ScrollMode(fake.RandomString([]string{"auto", "manual"})),
			AutoScrollWpm: fake.Number(100, 200),
			CreatedAt:     BuildFakeTime(),
		}
	}
	return result
}

// BuildFakePlayScriptCreationRequestInput builds a fake PlayScriptCreationRequestInput.
func BuildFakePlayScriptCreationRequestInput() *types.PlayScriptCreationRequestInput {
	castMembers := []*types.PlayScriptCastMemberCreationInput{
		{ID: BuildFakeID(), Name: fake.FirstName(), SortOrder: 0},
		{ID: BuildFakeID(), Name: fake.FirstName(), SortOrder: 1},
	}

	lines := []*types.PlayScriptLineCreationRequestInput{
		{CastMemberID: castMembers[0].ID, LineText: fake.Sentence(10), LineIndex: 0, ScrollMode: types.ScrollModeAuto},
		{CastMemberID: castMembers[1].ID, LineText: fake.Sentence(10), LineIndex: 1, ScrollMode: types.ScrollModeAuto},
	}

	return &types.PlayScriptCreationRequestInput{
		Title:       fake.BookTitle(),
		Description: fake.Sentence(10),
		Author:      fake.Name(),
		IsPublic:    fake.Bool(),
		CastMembers: castMembers,
		Lines:       lines,
	}
}

// BuildFakePlayScriptDatabaseCreationInput builds a fake PlayScriptDatabaseCreationInput.
func BuildFakePlayScriptDatabaseCreationInput() *types.PlayScriptDatabaseCreationInput {
	return &types.PlayScriptDatabaseCreationInput{
		ID:               BuildFakeID(),
		Title:            fake.BookTitle(),
		Description:      fake.Sentence(10),
		Author:           fake.Name(),
		IsPublic:         fake.Bool(),
		CreatedByUser:    BuildFakeID(),
		BelongsToAccount: BuildFakeID(),
		PrecededByID:     nil,
		PrecedesID:       nil,
	}
}

// BuildFakePlayScriptRating builds a fake PlayScriptRating.
func BuildFakePlayScriptRating() *types.PlayScriptRating {
	return &types.PlayScriptRating{
		ID:            BuildFakeID(),
		PlayScriptID:  BuildFakeID(),
		BelongsToUser: BuildFakeID(),
		Rating:        fake.Number(1, 5),
		CreatedAt:     BuildFakeTime(),
		LastUpdatedAt: nil,
		ArchivedAt:    nil,
	}
}

// BuildFakePlayScriptRatingStats builds a fake PlayScriptRatingStats.
func BuildFakePlayScriptRatingStats() *types.PlayScriptRatingStats {
	return &types.PlayScriptRatingStats{
		AverageRating: float64(fake.Number(10, 50)) / 10,
		RatingCount:   fake.Number(0, 100),
	}
}

// BuildFakePlayScriptUpdateRequestInput builds a fake PlayScriptUpdateRequestInput.
func BuildFakePlayScriptUpdateRequestInput() *types.PlayScriptUpdateRequestInput {
	title := fake.BookTitle()
	description := fake.Sentence(10)
	author := fake.Name()
	isPublic := fake.Bool()

	return &types.PlayScriptUpdateRequestInput{
		Title:       &title,
		Description: &description,
		Author:      &author,
		IsPublic:    &isPublic,
	}
}
