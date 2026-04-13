package uploadedmedia

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/postgres"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pgcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	exampleQuantity = 5
)

func buildDatabaseClientForTest(t *testing.T) (*repository, *pgcontainers.PostgresContainer) {
	t.Helper()

	ctx := t.Context()
	container, db, config := pgtesting.BuildDatabaseContainerForTest(t)
	require.NoError(t, migrations.NewMigrator(loggingnoop.NewLogger()).Migrate(ctx, db))

	pgc, err := postgres.ProvideDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	c := ProvideUploadedMediaRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	require.NoError(t, err)

	return c.(*repository), container
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideUploadedMediaRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &database.MockClient{})

	return c.(*repository)
}

func TestProvideUploadedMediaRepository(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := ProvideUploadedMediaRepository(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			nil,
			&database.MockClient{},
		)

		assert.NotNil(t, actual)
	})
}
