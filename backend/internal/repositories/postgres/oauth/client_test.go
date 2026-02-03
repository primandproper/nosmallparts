package oauth

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
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

	c := ProvideOAuthRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, config, pgc)
	require.NoError(t, err)

	return c.(*repository), container
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	config := &databasecfg.Config{
		Provider:                 databasecfg.ProviderPostgres,
		ConnectionDetails:        databasecfg.ConnectionDetails{},
		Debug:                    false,
		LogQueries:               false,
		RunMigrations:            true,
		MaxPingAttempts:          10,
		PingWaitPeriod:           time.Second,
		OAuth2TokenEncryptionKey: "blahblahblahblahblahblahblahblah",
	}

	c := ProvideOAuthRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), nil, config, &database.MockClient{})

	return c.(*repository)
}
