package dataprivacy

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"
	issue_reports "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/issuereports"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/waitlists"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/webhooks"

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
	require.NoError(t, migrations.NewMigrator(logging.NewNoopLogger()).Migrate(ctx, db))

	pgc, err := postgres.ProvideDatabaseClient(ctx, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), config)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), pgc)
	identityRepo := identity.ProvideIdentityRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, pgc)
	issueReportsRepo := issue_reports.ProvideIssueReportsRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, pgc)
	notificationsRepo := notifications.ProvideNotificationsRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, pgc)
	settingsRepo := settings.ProvideSettingsRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, pgc)
	uploadedMediaRepo := uploadedmedia.ProvideUploadedMediaRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, pgc)
	waitlistsRepo := waitlists.ProvideWaitlistsRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), pgc)
	webhooksRepo := webhooks.ProvideWebhooksRepository(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), auditLogEntryRepo, pgc)

	c := ProvideDataPrivacyRepository(
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
		auditLogEntryRepo,
		identityRepo,
		issueReportsRepo,
		notificationsRepo,
		settingsRepo,
		uploadedMediaRepo,
		waitlistsRepo,
		webhooksRepo,
		pgc,
	)
	require.NoError(t, err)

	return c.(*repository), container
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideDataPrivacyRepository(
		logging.NewNoopLogger(),
		tracing.NewNoopTracerProvider(),
		nil, // auditLogRepo
		nil, // identityRepo
		nil, // issueReportsRepo
		nil, // notificationsRepo
		nil, // settingsRepo
		nil, // uploadedMediaRepo
		nil, // waitlistsRepo
		nil, // webhooksRepo
		&database.MockClient{},
	)

	return c.(*repository)
}
