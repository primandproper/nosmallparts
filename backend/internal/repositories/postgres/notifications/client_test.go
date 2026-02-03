package notifications

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"

	"github.com/stretchr/testify/require"
	pgcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	exampleQuantity = 3
)

func buildDatabaseClientForTest(t *testing.T) (*repository, *pgcontainers.PostgresContainer) {
	t.Helper()

	ctx := t.Context()
	container, db, config := pgtesting.BuildDatabaseContainerForTest(t)
	require.NoError(t, migrations.NewMigrator(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), db, config).Migrate(ctx))

	pgc, err := postgres.ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), config)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), pgc)
	require.NoError(t, err)

	c := ProvideNotificationsRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, pgc)
	require.NoError(t, err)

	return c.(*repository), container
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideNotificationsRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, &database.MockClient{})

	return c.(*repository)
}
