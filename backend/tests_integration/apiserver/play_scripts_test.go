package integration

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts"
	domainconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/playscripts/fakes"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/playscripts"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkPlayScriptEquality(t *testing.T, expected, actual *playscripts.PlayScript) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected PlayScript to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected PlayScript to have CreatedAt")

	assert.Equal(t, expected.Title, actual.Title, "expected PlayScript Title")
	assert.Equal(t, expected.Description, actual.Description, "expected PlayScript Description")
	assert.Equal(t, expected.Author, actual.Author, "expected PlayScript Author")
	assert.Equal(t, expected.IsPublic, actual.IsPublic, "expected PlayScript IsPublic")
	assert.NotEmpty(t, actual.CreatedByUser, "expected PlayScript to have CreatedByUser")
	assert.NotEmpty(t, actual.BelongsToAccount, "expected PlayScript to have BelongsToAccount")
}

func createPlayScriptForTest(t *testing.T, testClient client.Client) *playscripts.PlayScript {
	t.Helper()
	ctx := t.Context()

	examplePlayScript := fakes.BuildFakePlayScript()
	examplePlayScriptInput := domainconverters.ConvertPlayScriptToPlayScriptCreationRequestInput(examplePlayScript)

	input := grpcconverters.ConvertPlayScriptCreationRequestInputToGRPC(examplePlayScriptInput)

	createdPlayScript, err := testClient.CreatePlayScript(ctx, &playscriptssvc.CreatePlayScriptRequest{Input: input})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCPlayScriptToPlayScript(createdPlayScript.Created)

	// Check main fields match
	assert.Equal(t, examplePlayScriptInput.Title, converted.Title)
	assert.Equal(t, examplePlayScriptInput.Description, converted.Description)
	assert.Equal(t, examplePlayScriptInput.Author, converted.Author)
	assert.Equal(t, examplePlayScriptInput.IsPublic, converted.IsPublic)

	retrievedPlayScript, err := testClient.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{PlayScriptId: createdPlayScript.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, retrievedPlayScript)

	playScript := grpcconverters.ConvertGRPCPlayScriptToPlayScript(retrievedPlayScript.Result)
	checkPlayScriptEquality(t, converted, playScript)

	return playScript
}

func TestPlayScripts_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		createPlayScriptForTest(t, testClient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreatePlayScript(ctx, &playscriptssvc.CreatePlayScriptRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		// Empty title should fail validation
		input := &playscriptssvc.PlayScriptCreationRequestInput{
			Title:       "", // empty title should fail validation
			Description: "test description",
		}

		_, err := testClient.CreatePlayScript(ctx, &playscriptssvc.CreatePlayScriptRequest{Input: input})
		assert.Error(t, err)
	})
}

func TestPlayScripts_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdPlayScript := createPlayScriptForTest(t, testClient)

		retrieved, err := testClient.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{PlayScriptId: createdPlayScript.ID})
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{PlayScriptId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{})
		assert.Error(t, err)
	})
}

func TestPlayScripts_ListingForAccount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdPlayScripts := []*playscripts.PlayScript{}
		for range exampleQuantity {
			createdPlayScripts = append(createdPlayScripts, createPlayScriptForTest(t, testClient))
		}

		results, err := testClient.GetPlayScriptsForAccount(ctx, &playscriptssvc.GetPlayScriptsForAccountRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdPlayScripts))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetPlayScriptsForAccount(ctx, &playscriptssvc.GetPlayScriptsForAccountRequest{})
		assert.Error(t, err)
	})
}

func TestPlayScripts_ListingPublic(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		// Create some public play scripts
		for range exampleQuantity {
			input := fakes.BuildFakePlayScriptCreationRequestInput()
			input.IsPublic = true
			grpcInput := grpcconverters.ConvertPlayScriptCreationRequestInputToGRPC(input)

			_, err := testClient.CreatePlayScript(ctx, &playscriptssvc.CreatePlayScriptRequest{Input: grpcInput})
			require.NoError(t, err)
		}

		results, err := testClient.GetPublicPlayScripts(ctx, &playscriptssvc.GetPublicPlayScriptsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= exampleQuantity)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetPublicPlayScripts(ctx, &playscriptssvc.GetPublicPlayScriptsRequest{})
		assert.Error(t, err)
	})
}

func TestPlayScripts_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdPlayScript := createPlayScriptForTest(t, testClient)

		newTitle := "Updated Play Script Title"
		updated, err := testClient.UpdatePlayScript(ctx, &playscriptssvc.UpdatePlayScriptRequest{
			PlayScriptId: createdPlayScript.ID,
			Input: &playscriptssvc.PlayScriptUpdateRequestInput{
				Title: &newTitle,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, newTitle, updated.Updated.Title)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		newTitle := "Updated Title"
		_, err := testClient.UpdatePlayScript(ctx, &playscriptssvc.UpdatePlayScriptRequest{
			PlayScriptId: nonexistentID,
			Input: &playscriptssvc.PlayScriptUpdateRequestInput{
				Title: &newTitle,
			},
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdatePlayScript(ctx, &playscriptssvc.UpdatePlayScriptRequest{})
		assert.Error(t, err)
	})
}

func TestPlayScripts_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdPlayScript := createPlayScriptForTest(t, testClient)

		_, err := testClient.ArchivePlayScript(ctx, &playscriptssvc.ArchivePlayScriptRequest{PlayScriptId: createdPlayScript.ID})
		assert.NoError(t, err)
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createPlayScriptForTest(t, testClient)

		_, err := testClient.ArchivePlayScript(ctx, &playscriptssvc.ArchivePlayScriptRequest{PlayScriptId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchivePlayScript(ctx, &playscriptssvc.ArchivePlayScriptRequest{})
		assert.Error(t, err)
	})
}
