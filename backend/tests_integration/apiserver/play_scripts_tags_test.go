package integration

import (
	"testing"

	playscriptsfakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/fakes"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createPlayScriptForTest(t *testing.T, testClient client.Client) *playscriptssvc.PlayScript {
	t.Helper()
	ctx := t.Context()

	input := playscriptsfakes.BuildFakePlayScriptCreationRequestInput()
	input.IsPublic = true // Make public so GetPlayScriptsWithTag can find it
	grpcInput := grpcconverters.ConvertPlayScriptCreationRequestInputToGRPC(input)

	created, err := testClient.CreatePlayScript(ctx, &playscriptssvc.CreatePlayScriptRequest{Input: grpcInput})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.Created)

	return created.Created
}

func TestPlayScriptTags_GetPlayScriptsWithTag(T *testing.T) {
	T.Parallel()

	T.Run("returns scripts with tag and excludes after tag removed", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		ps := createPlayScriptForTest(t, testClient)
		require.NotEmpty(t, ps.Id)

		// Create tag (service admin only)
		createTagResp, err := adminClient.CreateScriptTag(ctx, &playscriptssvc.CreateScriptTagRequest{
			Name:        "profanity",
			Description: "Contains strong language",
		})
		require.NoError(t, err)
		require.NotNil(t, createTagResp.Created)
		tagID := createTagResp.Created.Id

		// Add tag to play script (service admin only)
		_, err = adminClient.AddTagToPlayScript(ctx, &playscriptssvc.AddTagToPlayScriptRequest{
			PlayScriptId: ps.Id,
			ScriptTagId:  tagID,
		})
		require.NoError(t, err)

		// GetPlayScriptsWithTag should return the script (ReadPlayScriptsPermission)
		withTagResp, err := testClient.GetPlayScriptsWithTag(ctx, &playscriptssvc.GetPlayScriptsWithTagRequest{
			TagName: "profanity",
		})
		require.NoError(t, err)
		require.NotNil(t, withTagResp)
		require.NotEmpty(t, withTagResp.Results)
		assert.Equal(t, ps.Id, withTagResp.Results[0].Id)

		// Remove tag
		_, err = adminClient.RemoveTagFromPlayScript(ctx, &playscriptssvc.RemoveTagFromPlayScriptRequest{
			PlayScriptId: ps.Id,
			ScriptTagId:  tagID,
		})
		require.NoError(t, err)

		// GetPlayScriptsWithTag should no longer return the script
		afterRemoveResp, err := testClient.GetPlayScriptsWithTag(ctx, &playscriptssvc.GetPlayScriptsWithTagRequest{
			TagName: "profanity",
		})
		require.NoError(t, err)
		require.NotNil(t, afterRemoveResp)
		assert.Empty(t, afterRemoveResp.Results)
	})
}

func TestPlayScriptTags_GetPlayScriptReturnsTags(T *testing.T) {
	T.Parallel()

	T.Run("GetPlayScript returns script with tags populated", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		ps := createPlayScriptForTest(t, testClient)

		// Create and add tag
		createTagResp, err := adminClient.CreateScriptTag(ctx, &playscriptssvc.CreateScriptTagRequest{
			Name:        "one-act",
			Description: "One-act play",
		})
		require.NoError(t, err)
		require.NotNil(t, createTagResp.Created)

		_, err = adminClient.AddTagToPlayScript(ctx, &playscriptssvc.AddTagToPlayScriptRequest{
			PlayScriptId: ps.Id,
			ScriptTagId:  createTagResp.Created.Id,
		})
		require.NoError(t, err)

		// GetPlayScript should return script with tags
		getResp, err := testClient.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{
			PlayScriptId: ps.Id,
		})
		require.NoError(t, err)
		require.NotNil(t, getResp)
		require.NotNil(t, getResp.Result)
		require.Len(t, getResp.Result.Tags, 1)
		assert.Equal(t, "one-act", getResp.Result.Tags[0].Name)
		assert.Equal(t, "One-act play", getResp.Result.Tags[0].Description)
	})
}
