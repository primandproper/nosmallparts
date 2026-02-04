package grpcapi

import (
	"context"

	auditsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	internalopssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/internalops"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"
	notificationssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	oauthsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/oauth"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/playscripts"
	settingssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/settings"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/uploaded_media"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	textsearchcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/config"
	platformgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/grpc"
	authgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/grpc"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/grpc/interceptors"
	identitygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc"
	identityindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"
	issuereportsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/issuereports/grpc"
	notificationsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/notifications/grpc"
	oauthgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/oauth/grpc"
	playscriptsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts/grpc"
	settingsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/settings/grpc"
	uploadedmediagrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/grpc"
	waitlistsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/waitlists/grpc"
	webhooksgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/webhooks/grpc"

	grpc "google.golang.org/grpc"
)

func BuildRegistrationFuncs(
	auditLogService auditsvc.AuditServiceServer,
	authService authsvc.AuthServiceServer,
	dataPrivacyServer dataprivacysvc.DataPrivacyServiceServer,
	identityServiceServer identitysvc.IdentityServiceServer,
	internalOpsService internalopssvc.InternalOperationsServer,
	issueReportsService issuereportssvc.IssueReportsServiceServer,
	notificationsService notificationssvc.UserNotificationsServiceServer,
	oauthService oauthsvc.OAuthServiceServer,
	playScriptsService playscriptssvc.PlayScriptsServiceServer,
	settingsService settingssvc.SettingsServiceServer,
	uploadedMediaService uploadedmediasvc.UploadedMediaServiceServer,
	waitlistsService waitlistssvc.WaitlistsServiceServer,
	webhooksService webhookssvc.WebhooksServiceServer,
) []platformgrpc.RegistrationFunc {
	return []platformgrpc.RegistrationFunc{
		func(server *grpc.Server) {
			auditsvc.RegisterAuditServiceServer(server, auditLogService)
			authsvc.RegisterAuthServiceServer(server, authService)
			dataprivacysvc.RegisterDataPrivacyServiceServer(server, dataPrivacyServer)
			identitysvc.RegisterIdentityServiceServer(server, identityServiceServer)
			internalopssvc.RegisterInternalOperationsServer(server, internalOpsService)
			issuereportssvc.RegisterIssueReportsServiceServer(server, issueReportsService)
			notificationssvc.RegisterUserNotificationsServiceServer(server, notificationsService)
			oauthsvc.RegisterOAuthServiceServer(server, oauthService)
			playscriptssvc.RegisterPlayScriptsServiceServer(server, playScriptsService)
			settingssvc.RegisterSettingsServiceServer(server, settingsService)
			uploadedmediasvc.RegisterUploadedMediaServiceServer(server, uploadedMediaService)
			waitlistssvc.RegisterWaitlistsServiceServer(server, waitlistsService)
			webhookssvc.RegisterWebhooksServiceServer(server, webhooksService)
		},
	}
}

func BuildUnaryServerInterceptors(authInterceptor *interceptors.AuthInterceptor) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		authInterceptor.UnaryServerInterceptor(),
	}
}

func BuildStreamServerInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		//
	}
}

func ProvideUserTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (identityindexing.UserTextSearcher, error) {
	return textsearchcfg.ProvideIndex[identityindexing.UserSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		identityindexing.IndexTypeUsers,
	)
}

// AggregateMethodPermissions combines method permissions from all services into a single map.
// Each service provides its permissions via a typed map (e.g., SettingsMethodPermissions),
// which are then aggregated here for the auth interceptor.
func AggregateMethodPermissions(
	authPermissions authgrpc.AuthMethodPermissions,
	identityPermissions identitygrpc.IdentityMethodPermissions,
	issuereportsPermissions issuereportsgrpc.IssueReportsMethodPermissions,
	notificationsPermissions notificationsgrpc.NotificationsMethodPermissions,
	oauthPermissions oauthgrpc.OAuthMethodPermissions,
	playscriptsPermissions playscriptsgrpc.PlayScriptsMethodPermissions,
	settingsPermissions settingsgrpc.SettingsMethodPermissions,
	uploadedmediaPermissions uploadedmediagrpc.UploadedMediaMethodPermissions,
	waitlistsPermissions waitlistsgrpc.WaitlistsMethodPermissions,
	webhooksPermissions webhooksgrpc.WebhooksMethodPermissions,
) interceptors.MethodPermissionsMap {
	result := make(interceptors.MethodPermissionsMap)

	// Copy all service permissions into the aggregated map
	for method, perms := range authPermissions {
		result[method] = perms
	}
	for method, perms := range identityPermissions {
		result[method] = perms
	}
	for method, perms := range issuereportsPermissions {
		result[method] = perms
	}
	for method, perms := range notificationsPermissions {
		result[method] = perms
	}
	for method, perms := range oauthPermissions {
		result[method] = perms
	}
	for method, perms := range playscriptsPermissions {
		result[method] = perms
	}
	for method, perms := range settingsPermissions {
		result[method] = perms
	}
	for method, perms := range uploadedmediaPermissions {
		result[method] = perms
	}
	for method, perms := range waitlistsPermissions {
		result[method] = perms
	}
	for method, perms := range webhooksPermissions {
		result[method] = perms
	}

	return result
}
