package integration

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/fakes"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/webhooks/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkWebhookEquality(t *testing.T, expected, actual *webhooks.Webhook) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected Webhook to have MealPlanTaskID")
	assert.NotZero(t, actual.CreatedAt, "expected Webhook to have CreatedAt")

	assert.Equal(t, expected.Name, actual.Name, "expected Webhook Name")
	assert.Equal(t, expected.URL, actual.URL, "expected Webhook URL")
	assert.Equal(t, expected.Method, actual.Method, "expected Webhook Method")
	assert.Equal(t, expected.ContentType, actual.ContentType, "expected Webhook ContentType")
	assert.NotEmpty(t, actual.BelongsToAccount, "expected Webhook to have BelongsToAccount")

	require.Equal(t, len(expected.Events), len(actual.Events), "expected Webhook Events length")
	for i, expectedEvent := range expected.Events {
		if i > len(actual.Events) {
			continue
		}

		actualEvent := actual.Events[i]
		assert.NotEmpty(t, actualEvent.ID, "expected Webhook Event %d to have MealPlanTaskID", i)
		assert.NotZero(t, actualEvent.CreatedAt, "expected Webhook Event %d to have CreatedAt", i)
		assert.Equal(t, expectedEvent.TriggerEvent, actualEvent.TriggerEvent, "expected Webhook Event %d TriggerEvent", i)
		assert.Equal(t, actual.ID, actualEvent.BelongsToWebhook, "expected Webhook Event %d BelongsToWebhook", i)
	}
}

func createWebhookForTest(t *testing.T, testClient client.Client) *webhooks.Webhook {
	t.Helper()
	ctx := t.Context()

	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhookInput := converters.ConvertWebhookToWebhookCreationRequestInput(exampleWebhook)

	input := grpcconverters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(exampleWebhookInput)

	createdWebhook, err := testClient.CreateWebhook(ctx, &webhookssvc.CreateWebhookRequest{Input: input})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCWebhookToWebhook(createdWebhook.Created)
	checkWebhookEquality(t, exampleWebhook, converted)

	retrievedWebhook, err := testClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, retrievedWebhook)

	webhook := grpcconverters.ConvertGRPCWebhookToWebhook(retrievedWebhook.Result)
	checkWebhookEquality(t, converted, webhook)

	return webhook
}

func TestWebhooks_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedHTTPClientForTest(t)

		_, err := c.CreateWebhook(ctx, &webhookssvc.CreateWebhookRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleWebhookInput := &webhooks.WebhookCreationRequestInput{
			ContentType: "application/whatever",
			Method:      "UNRECOGNIZED",
			Name:        t.Name(),
			URL:         "invalid protocol :\\ neato.ai",
			Events:      []string{},
		}

		input := grpcconverters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(exampleWebhookInput)

		_, err := testClient.CreateWebhook(ctx, &webhookssvc.CreateWebhookRequest{Input: input})
		assert.Error(t, err)
	})
}

func TestWebhooks_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)

		retrieved, err := testClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.ID})
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent MealPlanTaskID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedHTTPClientForTest(t)
		_, err := c.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{})
		assert.Error(t, err)
	})
}

func TestWebhooks_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdWebhooks := []*webhooks.Webhook{}
		for range exampleQuantity {
			createdWebhooks = append(createdWebhooks, createWebhookForTest(t, testClient))
		}

		results, err := testClient.GetWebhooks(ctx, &webhookssvc.GetWebhooksRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.True(t, len(results.Results) >= len(createdWebhooks))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedHTTPClientForTest(t)
		_, err := c.GetWebhooks(ctx, &webhookssvc.GetWebhooksRequest{})
		assert.Error(t, err)
	})
}

func TestWebhooks_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)

		_, err := testClient.ArchiveWebhook(ctx, &webhookssvc.ArchiveWebhookRequest{WebhookId: createdWebhook.ID})
		assert.NoError(t, err)
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.ArchiveWebhook(ctx, &webhookssvc.ArchiveWebhookRequest{WebhookId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedHTTPClientForTest(t)
		_, err := c.ArchiveWebhook(ctx, &webhookssvc.ArchiveWebhookRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerEvents_Adding(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)

		_, err := testClient.AddWebhookTriggerEvent(ctx, &webhookssvc.AddWebhookTriggerEventRequest{
			WebhookId: createdWebhook.ID,
			Input: &webhookssvc.WebhookTriggerEventCreationRequestInput{
				BelongsToWebhook: createdWebhook.ID,
				TriggerEvent:     webhooks.WebhookArchivedTriggerEvent,
			},
		})
		assert.NoError(t, err)
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.AddWebhookTriggerEvent(ctx, &webhookssvc.AddWebhookTriggerEventRequest{WebhookId: nonexistentID, Input: &webhookssvc.WebhookTriggerEventCreationRequestInput{}})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedHTTPClientForTest(t)
		_, err := c.AddWebhookTriggerEvent(ctx, &webhookssvc.AddWebhookTriggerEventRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookTriggerEvents_Removing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdWebhook := createWebhookForTest(t, testClient)

		createdTriggerEvent, err := testClient.AddWebhookTriggerEvent(ctx, &webhookssvc.AddWebhookTriggerEventRequest{
			WebhookId: createdWebhook.ID,
			Input: &webhookssvc.WebhookTriggerEventCreationRequestInput{
				BelongsToWebhook: createdWebhook.ID,
				TriggerEvent:     webhooks.WebhookArchivedTriggerEvent,
			},
		})
		require.NoError(t, err)

		_, err = testClient.ArchiveWebhookTriggerEvent(ctx, &webhookssvc.ArchiveWebhookTriggerEventRequest{
			WebhookId:             createdWebhook.ID,
			WebhookTriggerEventId: createdTriggerEvent.Created.Id,
		})
		assert.NoError(t, err)
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.ArchiveWebhookTriggerEvent(ctx, &webhookssvc.ArchiveWebhookTriggerEventRequest{WebhookId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedHTTPClientForTest(t)
		_, err := c.ArchiveWebhookTriggerEvent(ctx, &webhookssvc.ArchiveWebhookTriggerEventRequest{})
		assert.Error(t, err)
	})
}
