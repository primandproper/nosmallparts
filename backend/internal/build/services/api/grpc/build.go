//go:build wireinject
// +build wireinject

package grpcapi

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	tokenscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	authmgr "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth/managers"
	identitymgr "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/manager"
	oauthmgr "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth/manager"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	loggingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/config"
	metricscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/config"
	tracingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/qrcodes"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/random"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/grpc"
	uploadscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/objectstorage"
	auditrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/auth"
	identityrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/identity"
	issuereportsrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/issuereports"
	notificationsrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/notifications"
	oauthrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/oauth"
	settingsrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/settings"
	uploadedmediarepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/waitlists"
	webhooksrepo "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/webhooks"
	auditsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/audit/grpc"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/grpc"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/grpc/interceptors"
	authhttpsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/dataprivacy/grpc"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc"
	internalopssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/internalops/grpc"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/issuereports/grpc"
	notificationssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/notifications/grpc"
	oauthsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/oauth/grpc"
	settingssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/settings/grpc"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/config"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/grpc"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/waitlists/grpc"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/webhooks/grpc"

	"github.com/google/wire"
)

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.APIServiceConfig,
) (*GRPCService, error) {
	wire.Build(ConfigProviders,
		// core
		metricscfg.MetricsConfigProviders,
		loggingcfg.LogConfigProviders,
		tracingcfg.TracingConfigProviders,
		msgconfig.MessageQueueProviders,
		authentication.AuthProviders,
		sessions.SessionProviders,
		observability.O11yProviders,
		random.RandProviders,
		postgres.PGProviders,
		grpc.ProvidersGRPC,
		qrcodes.QRCodeProviders,
		tokenscfg.TokenIssuerProviders,
		interceptors.InterceptorProviders,
		uploadscfg.Providers,
		objectstorage.Providers,
		// repos
		auditrepo.AuditRepoProviders,
		authrepo.AuthRepoProviders,
		identityrepo.IDRepoProviders,
		issuereportsrepo.IssueReportsRepoProviders,
		notificationsrepo.NotifRepoProviders,
		settingsrepo.SettingsRepoProviders,
		uploadedmediarepo.UploadedMediaRepoProviders,
		webhooksrepo.WebhookProviders,
		oauthrepo.OAuthRepoProviders,
		waitlistsrepo.WaitlistsRepoProviders,
		// services
		authhttpsvc.AuthHTTPServiceProviders,
		auditsvc.AuditSvcProviders,
		authsvc.AuthSvcProviders,
		dataprivacysvc.DataPrivSvcProviders,
		identitysvc.IDSvcProviders,
		internalopssvc.InternalOpsSvcProviders,
		issuereportssvc.IssueReportSvcProviders,
		notificationssvc.NotifsSvcProviders,
		settingssvc.SettingSvcProviders,
		uploadedmediasvc.UploadedMediaSvcProviders,
		webhookssvc.WebhookSvcProviders,
		oauthsvc.OAuthSvcProviders,
		waitlistssvc.WaitlistsSvcProviders,
		uploadedmediacfg.UploadedMediaConfigProviders,
		// manager
		identitymgr.IDManagerProviders,
		oauthmgr.OAuthManagerProviders,
		authmgr.AuthManagerProviders,
		// misc
		ProvideUserTextSearcher,
		AggregateMethodPermissions,
		BuildUnaryServerInterceptors,
		BuildStreamServerInterceptors,
		BuildRegistrationFuncs,
		NewGRPCService,
	)

	return nil, nil
}
