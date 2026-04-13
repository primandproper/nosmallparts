package dataprivacy

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/postgres"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
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
	require.NoError(t, migrations.NewMigrator(loggingnoop.NewLogger()).Migrate(ctx, db))

	pgc, err := postgres.ProvideDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	identityRepo := identity.ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	issueReportsRepo := issue_reports.ProvideIssueReportsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	notificationsRepo := notifications.ProvideNotificationsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	settingsRepo := settings.ProvideSettingsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	uploadedMediaRepo := uploadedmedia.ProvideUploadedMediaRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	waitlistsRepo := waitlists.ProvideWaitlistsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	webhooksRepo := webhooks.ProvideWebhooksRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)

	c := ProvideDataPrivacyRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
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
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
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
