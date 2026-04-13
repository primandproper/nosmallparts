package webhooks

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/fakes"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createWebhookForTest(t *testing.T, ctx context.Context, exampleWebhook *types.Webhook, dbc *repository) *types.Webhook {
	t.Helper()

	// create
	if exampleWebhook == nil {
		exampleWebhook = fakes.BuildFakeWebhook()
	}
	dbInput := converters.ConvertWebhookToWebhookDatabaseCreationInput(exampleWebhook)

	created, err := dbc.CreateWebhook(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleWebhook.CreatedAt = created.CreatedAt
	for i := range created.Events {
		exampleWebhook.Events[i].CreatedAt = created.Events[i].CreatedAt
	}
	assert.Equal(t, exampleWebhook, created)

	webhook, err := dbc.GetWebhook(ctx, created.ID, created.BelongsToAccount)
	exampleWebhook.CreatedAt = webhook.CreatedAt
	for i := range created.Events {
		exampleWebhook.Events[i].CreatedAt = webhook.Events[i].CreatedAt
	}

	assert.NoError(t, err)
	assert.Equal(t, webhook, exampleWebhook)

	return created
}

func TestQuerier_Integration_Webhooks(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhook.BelongsToAccount = account.ID
	createdWebhooks := []*types.Webhook{}

	// create
	createdWebhooks = append(createdWebhooks, createWebhookForTest(t, ctx, exampleWebhook, dbc))

	// create more
	for i := 0; i < exampleQuantity; i++ {
		input := fakes.BuildFakeWebhook()
		input.Name = fmt.Sprintf("%s %d", exampleWebhook.Name, i)
		input.BelongsToAccount = account.ID
		createdWebhooks = append(createdWebhooks, createWebhookForTest(t, ctx, input, dbc))
	}

	// fetch as list
	webhooks, err := dbc.GetWebhooks(ctx, account.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, webhooks.Data)
	assert.Equal(t, len(createdWebhooks), len(webhooks.Data))

	// fetch as list
	webhooksByAccountAndEvent, err := dbc.GetWebhooksForAccountAndEvent(ctx, account.ID, createdWebhooks[0].Events[0].TriggerEvent)
	assert.NoError(t, err)
	assert.NotEmpty(t, webhooksByAccountAndEvent)

	createdEvent, err := dbc.AddWebhookTriggerEvent(ctx, account.ID, &types.WebhookTriggerEventDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: createdWebhooks[0].ID,
		TriggerEvent:     types.WebhookArchivedServiceEventType,
	})
	assert.NoError(t, err)
	assert.NotNil(t, createdEvent)

	createdWebhooks[0].Events = append(createdWebhooks[0].Events, createdEvent)

	// delete
	for _, webhook := range createdWebhooks {
		for _, event := range webhook.Events {
			assert.NoError(t, dbc.ArchiveWebhookTriggerEvent(ctx, webhook.ID, event.ID))
		}

		assert.NoError(t, dbc.ArchiveWebhook(ctx, webhook.ID, account.ID))

		var exists bool
		exists, err = dbc.WebhookExists(ctx, webhook.ID, account.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.Webhook
		y, err = dbc.GetWebhook(ctx, webhook.ID, account.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhook(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhooks(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateWebhook(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_createWebhookTriggerEvent(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		created, err := c.createWebhookTriggerEvent(ctx, c.writeDB, fakes.BuildFakeID(), nil)
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("with missing account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		fakes.BuildFakeWebhookTriggerEvent()

		input := converters.ConvertWebhookTriggerEventToWebhookTriggerEventDatabaseCreationInput(fakes.BuildFakeWebhookTriggerEvent())

		created, err := c.createWebhookTriggerEvent(ctx, c.writeDB, "", input)
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestQuerier_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhook(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhookID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhookID, ""))
	})
}

func TestQuerier_ArchiveWebhookTriggerEvent(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhookTriggerEvent(ctx, "", exampleAccountID))
	})

	T.Run("with invalid webhook trigger event ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhookID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhookTriggerEvent(ctx, exampleWebhookID, ""))
	})
}

func TestQuerier_Integration_CursorBasedPagination(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.Webhook]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "webhook",
		CreateItem: func(ctx context.Context, i int) *types.Webhook {
			webhook := fakes.BuildFakeWebhook()
			webhook.Name = fmt.Sprintf("Webhook %02d", i) // Use zero-padded numbers for consistent sorting
			webhook.BelongsToAccount = account.ID
			return createWebhookForTest(t, ctx, webhook, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Webhook], error) {
			return dbc.GetWebhooks(ctx, account.ID, filter)
		},
		GetID: func(webhook *types.Webhook) string {
			return webhook.ID
		},
		CleanupItem: func(ctx context.Context, webhook *types.Webhook) error {
			// Cleanup webhook events first
			for _, event := range webhook.Events {
				if err = dbc.ArchiveWebhookTriggerEvent(ctx, webhook.ID, event.ID); err != nil {
					return err
				}
			}
			return dbc.ArchiveWebhook(ctx, webhook.ID, account.ID)
		},
	})
}
