package grpc

import (
	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"
	"github.com/primandproper/nosmallparts/backend/internal/proto/scriptspb"
)

// ConvertPlay is a play as the wire carries it, for the scripts service and
// for the stage that embeds the selected one.
func ConvertPlay(play *scripts.Play) *scriptspb.Play {
	out := &scriptspb.Play{
		Id:          play.ID,
		Title:       play.Title,
		Description: play.Description,
		Author:      play.Author,
		IsPublic:    play.IsPublic,
		Enabled:     play.Enabled,
		Cast:        make([]*scriptspb.CastMember, 0, len(play.Cast)),
		Lines:       make([]*scriptspb.PlayLine, 0, len(play.Lines)),
		Tags:        make([]*scriptspb.ScriptTag, 0, len(play.Tags)),
		RatingStats: convertRatingStats(play.RatingStats),
		UserRating:  play.UserRating,
	}

	if play.CreatedByUserID != "" {
		out.CreatedByUserId = new(play.CreatedByUserID)
	}

	for _, member := range play.Cast {
		out.Cast = append(out.Cast, &scriptspb.CastMember{Id: member.ID, Name: member.Name, Description: member.Description})
	}

	for _, line := range play.Lines {
		out.Lines = append(out.Lines, &scriptspb.PlayLine{
			CastMemberIds: line.CastMemberIDs,
			Text:          line.Text,
			Index:         line.Index,
			Direction:     line.Direction,
			ScrollMode:    convertScrollMode(line.ScrollMode),
			AutoScrollWpm: line.AutoScrollWPM,
		})
	}

	for _, tag := range play.Tags {
		out.Tags = append(out.Tags, convertTag(tag))
	}

	return out
}

func convertRatingStats(stats *scripts.RatingStats) *scriptspb.PlayRatingStats {
	if stats == nil {
		return nil
	}

	return &scriptspb.PlayRatingStats{AverageRating: stats.AverageRating, RatingCount: stats.RatingCount}
}

func convertTag(tag *scripts.Tag) *scriptspb.ScriptTag {
	return &scriptspb.ScriptTag{Id: tag.ID, Name: tag.Name, Description: tag.Description}
}

func convertScrollMode(mode scripts.ScrollMode) scriptspb.ScrollMode {
	switch mode {
	case scripts.ScrollModeManual:
		return scriptspb.ScrollMode_SCROLL_MODE_MANUAL
	case scripts.ScrollModeAuto:
		return scriptspb.ScrollMode_SCROLL_MODE_AUTO
	default:
		return scriptspb.ScrollMode_SCROLL_MODE_UNSPECIFIED
	}
}

// convertPlayInput reads a request's script. A request naming none is an empty
// input, which validation refuses for its missing title.
func convertPlayInput(in *scriptspb.PlayInput) *scripts.PlayInput {
	if in == nil {
		return &scripts.PlayInput{}
	}

	out := &scripts.PlayInput{
		Title:       in.GetTitle(),
		Description: in.GetDescription(),
		Author:      in.Author,
		IsPublic:    in.GetIsPublic(),
		TagIDs:      in.GetTagIds(),
		TagNames:    in.GetTagNames(),
	}

	for _, member := range in.GetCast() {
		out.Cast = append(out.Cast, &scripts.CastMemberInput{
			ID:          member.GetId(),
			Name:        member.GetName(),
			Description: member.Description,
		})
	}

	for _, line := range in.GetLines() {
		var mode scripts.ScrollMode

		switch line.GetScrollMode() {
		case scriptspb.ScrollMode_SCROLL_MODE_MANUAL:
			mode = scripts.ScrollModeManual
		case scriptspb.ScrollMode_SCROLL_MODE_AUTO:
			mode = scripts.ScrollModeAuto
		case scriptspb.ScrollMode_SCROLL_MODE_UNSPECIFIED:
		}

		out.Lines = append(out.Lines, &scripts.LineInput{
			CastMemberIDs: line.GetCastMemberIds(),
			Text:          line.GetText(),
			Index:         line.GetIndex(),
			Direction:     line.Direction,
			ScrollMode:    mode,
			AutoScrollWPM: line.AutoScrollWpm,
		})
	}

	return out
}
