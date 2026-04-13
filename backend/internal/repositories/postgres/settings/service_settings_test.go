package settings

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings/fakes"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createServiceSettingForTest(t *testing.T, ctx context.Context, exampleServiceSetting *types.ServiceSetting, dbc *repository) *types.ServiceSetting {
	t.Helper()

	// create
	if exampleServiceSetting == nil {
		exampleServiceSetting = fakes.BuildFakeServiceSetting()
	}
	dbInput := converters.ConvertServiceSettingToServiceSettingDatabaseCreationInput(exampleServiceSetting)

	created, err := dbc.CreateServiceSetting(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleServiceSetting.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleServiceSetting, created)

	serviceSetting, err := dbc.GetServiceSetting(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, serviceSetting)
	exampleServiceSetting.CreatedAt = serviceSetting.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, serviceSetting, exampleServiceSetting)

	return created
}

func TestQuerier_Integration_ServiceSettings(t *testing.T) {
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

	exampleServiceSetting := fakes.BuildFakeServiceSetting()
	createdServiceSettings := []*types.ServiceSetting{}

	// create
	createdServiceSettings = append(createdServiceSettings, createServiceSettingForTest(t, ctx, exampleServiceSetting, dbc))

	// create more
	for i := 0; i < exampleQuantity; i++ {
		input := fakes.BuildFakeServiceSetting()
		input.Name = fmt.Sprintf("%s %d", exampleServiceSetting.Name, i)
		createdServiceSettings = append(createdServiceSettings, createServiceSettingForTest(t, ctx, input, dbc))
	}

	// fetch as list
	serviceSettings, err := dbc.GetServiceSettings(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, serviceSettings.Data)
	assert.Equal(t, len(createdServiceSettings), len(serviceSettings.Data))

	// fetch via name search
	byName, err := dbc.SearchForServiceSettings(ctx, exampleServiceSetting.Name, nil)
	assert.NoError(t, err)
	assert.Equal(t, serviceSettings.Data, byName.Data)

	// delete
	for _, serviceSetting := range createdServiceSettings {
		assert.NoError(t, dbc.ArchiveServiceSetting(ctx, serviceSetting.ID))

		var exists bool
		exists, err = dbc.ServiceSettingExists(ctx, serviceSetting.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.ServiceSetting
		y, err = dbc.GetServiceSetting(ctx, serviceSetting.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ServiceSettingExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ServiceSettingExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetServiceSetting(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetServiceSetting(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_SearchForServiceSettings(T *testing.T) {
	T.Parallel()

	T.Run("with invalid query", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.SearchForServiceSettings(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateServiceSetting(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateServiceSetting(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_ArchiveServiceSetting(T *testing.T) {
	T.Parallel()

	T.Run("with invalid service setting ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveServiceSetting(ctx, ""))
	})
}

func TestQuerier_Integration_ServiceSettings_CursorBasedPagination(t *testing.T) {
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

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ServiceSetting]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "service setting",
		CreateItem: func(ctx context.Context, i int) *types.ServiceSetting {
			serviceSetting := fakes.BuildFakeServiceSetting()
			serviceSetting.Name = fmt.Sprintf("Service Setting %02d", i)
			return createServiceSettingForTest(t, ctx, serviceSetting, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSetting], error) {
			return dbc.GetServiceSettings(ctx, filter)
		},
		GetID: func(serviceSetting *types.ServiceSetting) string {
			return serviceSetting.ID
		},
		CleanupItem: func(ctx context.Context, serviceSetting *types.ServiceSetting) error {
			return dbc.ArchiveServiceSetting(ctx, serviceSetting.ID)
		},
	})
}
