package integration

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications/fakes"
	notificationssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/notifications/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUserNotificationForTest(t *testing.T, forUser string) *notifications.UserNotification {
	t.Helper()

	ctx := t.Context()

	creationRequestInput := fakes.BuildFakeUserNotification()
	input := converters.ConvertUserNotificationToUserNotificationDatabaseCreationInput(creationRequestInput)
	input.BelongsToUser = forUser

	created, err := notifsRepo.CreateUserNotification(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, created)

	return created
}

func TestUserNotifications_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		retrieved, err := testClient.GetUserNotification(ctx, &notificationssvc.GetUserNotificationRequest{UserNotificationId: created.ID})
		assert.NoError(t, err)

		converted := grpcconverters.ConvertGRPCUserNotificationToUserNotification(retrieved.Result)

		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		c := buildUnauthenticatedHTTPClientForTest(t)

		_, err := c.GetUserNotification(ctx, &notificationssvc.GetUserNotificationRequest{UserNotificationId: created.ID})
		assert.Error(t, err)
	})

	T.Run("invalid MealPlanTaskID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetUserNotification(ctx, &notificationssvc.GetUserNotificationRequest{UserNotificationId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestUserNotifications_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		updateInput := fakes.BuildFakeUserNotificationUpdateRequestInput()
		created.Update(updateInput)

		response, err := testClient.UpdateUserNotification(ctx, &notificationssvc.UpdateUserNotificationRequest{
			UserNotificationId: created.ID,
			Input:              grpcconverters.ConvertUserNotificationUpdateRequestInputToGRPCUserNotificationUpdateRequestInput(updateInput),
		})
		assert.NoError(t, err)

		updated := grpcconverters.ConvertGRPCUserNotificationToUserNotification(response.Updated)
		// Ensure UpdatedAt was set
		require.NotNil(t, updated.LastUpdatedAt)

		assertRoughEquality(t, created, updated, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		created := createUserNotificationForTest(t, user.ID)

		updateInput := fakes.BuildFakeUserNotificationUpdateRequestInput()
		created.Update(updateInput)

		c := buildUnauthenticatedHTTPClientForTest(t)

		_, err := c.UpdateUserNotification(ctx, &notificationssvc.UpdateUserNotificationRequest{
			UserNotificationId: created.ID,
			Input:              grpcconverters.ConvertUserNotificationUpdateRequestInputToGRPCUserNotificationUpdateRequestInput(updateInput),
		})
		assert.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()

		/*
			there's no way to provide invalid input to this method, but
			I want to make it explicit that tests should be written the moment that changes
		*/
	})
}

func TestUserNotifications_Listing(T *testing.T) {
	T.Parallel()

	u, testClient := createUserAndClientForTest(T)
	createdUserNotifications := []*notifications.UserNotification{}
	for range exampleQuantity {
		created := createUserNotificationForTest(T, u.ID)
		createdUserNotifications = append(createdUserNotifications, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetUserNotifications(ctx, &notificationssvc.GetUserNotificationsRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.True(t, len(retrieved.Results) >= len(createdUserNotifications))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedHTTPClientForTest(t)

		_, err := c.GetUserNotifications(ctx, &notificationssvc.GetUserNotificationsRequest{})
		assert.Error(t, err)
	})
}
