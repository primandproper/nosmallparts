package integration

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/uploadedmedia/converters"
	uploadedmediafakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/uploadedmedia/fakes"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/filtering"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/uploaded_media"
	"github.com/verygoodsoftwarenotvirus/platform/v4/pointer"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkUploadedMediaEquality(t *testing.T, expected, actual *uploadedmedia.UploadedMedia) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected UploadedMedia to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected UploadedMedia to have CreatedAt")

	assert.Equal(t, expected.StoragePath, actual.StoragePath, "expected UploadedMedia StoragePath")
	assert.Equal(t, expected.MimeType, actual.MimeType, "expected UploadedMedia MimeType")
	assert.NotEmpty(t, actual.CreatedByUser, "expected UploadedMedia to have CreatedByUser")
}

func createUploadedMediaForTest(t *testing.T, testClient client.Client) *uploadedmedia.UploadedMedia {
	t.Helper()
	ctx := t.Context()

	exampleUploadedMedia := uploadedmediafakes.BuildFakeUploadedMedia()
	exampleUploadedMediaInput := converters.ConvertUploadedMediaToUploadedMediaCreationRequestInput(exampleUploadedMedia)

	input := grpcconverters.ConvertUploadedMediaCreationRequestInputToGRPCUploadedMediaCreationRequestInput(exampleUploadedMediaInput)

	createdUploadedMedia, err := testClient.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{Input: input})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCUploadedMediaToUploadedMedia(createdUploadedMedia.Created)
	checkUploadedMediaEquality(t, exampleUploadedMedia, converted)

	retrievedUploadedMedia, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, retrievedUploadedMedia)

	uploadedMediaItem := grpcconverters.ConvertGRPCUploadedMediaToUploadedMedia(retrievedUploadedMedia.Result)
	checkUploadedMediaEquality(t, converted, uploadedMediaItem)

	return uploadedMediaItem
}

func TestUploadedMedia_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		createUploadedMediaForTest(t, testClient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleUploadedMediaInput := &uploadedmedia.UploadedMediaCreationRequestInput{
			StoragePath: "", // empty storage path should fail validation
			MimeType:    "",
		}

		input := grpcconverters.ConvertUploadedMediaCreationRequestInputToGRPCUploadedMediaCreationRequestInput(exampleUploadedMediaInput)

		_, err := testClient.CreateUploadedMedia(ctx, &uploadedmediasvc.CreateUploadedMediaRequest{Input: input})
		assert.Error(t, err)
	})
}

func TestUploadedMedia_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient)

		retrieved, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{})
		assert.Error(t, err)
	})

	T.Run("cannot access other user's media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create uploaded media as user 1
		_, testClient1 := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient1)

		// Try to access as user 2
		_, testClient2 := createUserAndClientForTest(t)
		retrieved, err := testClient2.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})
}

func TestUploadedMedia_ReadingWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdUploadedMedia := []*uploadedmedia.UploadedMedia{}
		ids := []string{}
		for range exampleQuantity {
			created := createUploadedMediaForTest(t, testClient)
			createdUploadedMedia = append(createdUploadedMedia, created)
			ids = append(ids, created.ID)
		}

		results, err := testClient.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: ids,
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results.Results, len(createdUploadedMedia))
	})

	T.Run("filters out other users' media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create media as user 1
		_, testClient1 := createUserAndClientForTest(t)
		media1 := createUploadedMediaForTest(t, testClient1)

		// Create media as user 2
		_, testClient2 := createUserAndClientForTest(t)
		media2 := createUploadedMediaForTest(t, testClient2)

		// User 2 tries to fetch both IDs
		results, err := testClient2.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{media1.ID, media2.ID},
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		// Should only get their own media
		assert.Len(t, results.Results, 1)
		assert.Equal(t, media2.ID, results.Results[0].Id)
	})

	T.Run("empty IDs list", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{
			Ids: []string{},
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetUploadedMediaWithIDs(ctx, &uploadedmediasvc.GetUploadedMediaWithIDsRequest{Ids: []string{"id1", "id2"}})
		assert.Error(t, err)
	})
}

func TestUploadedMedia_ListingForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		createdUploadedMedia := []*uploadedmedia.UploadedMedia{}
		for range exampleQuantity {
			createdUploadedMedia = append(createdUploadedMedia, createUploadedMediaForTest(t, testClient))
		}

		results, err := testClient.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: user.ID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdUploadedMedia))
	})

	T.Run("cannot access other user's media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user1, testClient1 := createUserAndClientForTest(t)
		createUploadedMediaForTest(t, testClient1)

		_, testClient2 := createUserAndClientForTest(t)

		// User 2 tries to list user 1's media
		results, err := testClient2.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: user1.ID,
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{})
		assert.Error(t, err)
	})
}

func TestUploadedMedia_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient)

		newStoragePath := "updated/path/to/file.jpg"
		newMimeType := uploadedmediasvc.UploadedMediaMimeType_UPLOADED_MEDIA_MIME_TYPE_IMAGE_JPEG

		updateInput := &uploadedmediasvc.UploadedMediaUpdateRequestInput{
			StoragePath: &newStoragePath,
			MimeType:    &newMimeType,
		}

		updated, err := testClient.UpdateUploadedMedia(ctx, &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: createdUploadedMedia.ID,
			Input:           updateInput,
		})
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, newStoragePath, updated.Updated.StoragePath)
		assert.Equal(t, newMimeType, updated.Updated.MimeType)

		// Verify the update persisted
		retrieved, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		assert.NoError(t, err)
		assert.Equal(t, newStoragePath, retrieved.Result.StoragePath)
		assert.Equal(t, newMimeType, retrieved.Result.MimeType)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		newStoragePath := "updated/path.jpg"
		updateInput := &uploadedmediasvc.UploadedMediaUpdateRequestInput{
			StoragePath: &newStoragePath,
		}

		updated, err := testClient.UpdateUploadedMedia(ctx, &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: nonexistentID,
			Input:           updateInput,
		})
		assert.Error(t, err)
		assert.Nil(t, updated)
	})

	T.Run("cannot update other user's media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create media as user 1
		_, testClient1 := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient1)

		// Try to update as user 2
		_, testClient2 := createUserAndClientForTest(t)

		newStoragePath := "hacked/path.jpg"
		updateInput := &uploadedmediasvc.UploadedMediaUpdateRequestInput{
			StoragePath: &newStoragePath,
		}

		updated, err := testClient2.UpdateUploadedMedia(ctx, &uploadedmediasvc.UpdateUploadedMediaRequest{
			UploadedMediaId: createdUploadedMedia.ID,
			Input:           updateInput,
		})
		assert.Error(t, err)
		assert.Nil(t, updated)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateUploadedMedia(ctx, &uploadedmediasvc.UpdateUploadedMediaRequest{})
		assert.Error(t, err)
	})
}

func TestUploadedMedia_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient)

		archived, err := testClient.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: createdUploadedMedia.ID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, archived)

		// Verify it's been archived (should not be retrievable)
		retrieved, err := testClient.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		archived, err := testClient.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: nonexistentID,
		})
		assert.Error(t, err)
		assert.Nil(t, archived)
	})

	T.Run("cannot archive other user's media", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create media as user 1
		_, testClient1 := createUserAndClientForTest(t)
		createdUploadedMedia := createUploadedMediaForTest(t, testClient1)

		// Try to archive as user 2
		_, testClient2 := createUserAndClientForTest(t)

		archived, err := testClient2.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{
			UploadedMediaId: createdUploadedMedia.ID,
		})
		assert.Error(t, err)
		assert.Nil(t, archived)

		// Verify it's still accessible to user 1
		retrieved, err := testClient1.GetUploadedMedia(ctx, &uploadedmediasvc.GetUploadedMediaRequest{UploadedMediaId: createdUploadedMedia.ID})
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveUploadedMedia(ctx, &uploadedmediasvc.ArchiveUploadedMediaRequest{})
		assert.Error(t, err)
	})
}

func TestUploadedMedia_Pagination(T *testing.T) {
	T.Parallel()

	T.Run("respects limit", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		// Create more media than the limit
		for range 10 {
			createUploadedMediaForTest(t, testClient)
		}

		results, err := testClient.GetUploadedMediaForUser(ctx, &uploadedmediasvc.GetUploadedMediaForUserRequest{
			UserId: user.ID,
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: pointer.To(uint32(5)),
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.LessOrEqual(t, len(results.Results), 5)
	})
}
