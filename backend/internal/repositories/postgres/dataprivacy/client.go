package dataprivacy

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/issuereports"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/dataprivacy/generated"
)

const (
	o11yName = "dataprivacy_db_client"
)

// repository is the data privacy repository client.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogRepo      audit.Repository
	identityRepo      identity.Repository
	issueReportsRepo  issuereports.Repository
	notificationsRepo notifications.Repository
	settingsRepo      settings.Repository
	uploadedMediaRepo uploadedmedia.Repository
	waitlistsRepo     waitlists.Repository
	webhooksRepo      webhooks.Repository
	readDB            *sql.DB
	writeDB           *sql.DB
}

// ProvideDataPrivacyRepository provides a new repository.
func ProvideDataPrivacyRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogRepo audit.Repository,
	identityRepo identity.Repository,
	issueReportsRepo issuereports.Repository,
	notificationsRepo notifications.Repository,
	settingsRepo settings.Repository,
	uploadedMediaRepo uploadedmedia.Repository,
	waitlistsRepo waitlists.Repository,
	webhooksRepo webhooks.Repository,
	client database.Client,
) dataprivacy.Repository {
	c := &repository{
		Client:            client,
		readDB:            client.ReadDB(),
		writeDB:           client.WriteDB(),
		tracer:            tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		logger:            logging.EnsureLogger(logger).WithName(o11yName),
		generatedQuerier:  generated.New(),
		auditLogRepo:      auditLogRepo,
		identityRepo:      identityRepo,
		issueReportsRepo:  issueReportsRepo,
		notificationsRepo: notificationsRepo,
		settingsRepo:      settingsRepo,
		uploadedMediaRepo: uploadedMediaRepo,
		waitlistsRepo:     waitlistsRepo,
		webhooksRepo:      webhooksRepo,
	}

	return c
}
