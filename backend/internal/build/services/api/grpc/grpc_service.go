package grpcapi

import (
	auditsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	internalopssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/internalops"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"
	notificationssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	oauthsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/oauth"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"
	settingssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/settings"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/uploaded_media"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/grpc"
)

type GRPCService struct {
	auditsvc.AuditServiceServer
	authsvc.AuthServiceServer
	dataprivacysvc.DataPrivacyServiceServer
	identitysvc.IdentityServiceServer
	internalopssvc.InternalOperationsServer
	issuereportssvc.IssueReportsServiceServer
	notificationssvc.UserNotificationsServiceServer
	oauthsvc.OAuthServiceServer
	settingssvc.SettingsServiceServer
	uploadedmediasvc.UploadedMediaServiceServer
	waitlistssvc.WaitlistsServiceServer
	webhookssvc.WebhooksServiceServer
	playscriptssvc.PlayScriptsServiceServer
	*grpc.Server
}

func NewGRPCService(
	auditServiceServer auditsvc.AuditServiceServer,
	authServiceServer authsvc.AuthServiceServer,
	dataPrivacyServiceServer dataprivacysvc.DataPrivacyServiceServer,
	identityServiceServer identitysvc.IdentityServiceServer,
	internalOperationsServer internalopssvc.InternalOperationsServer,
	issueReportsServiceServer issuereportssvc.IssueReportsServiceServer,
	userNotificationsServiceServer notificationssvc.UserNotificationsServiceServer,
	oauthServiceServer oauthsvc.OAuthServiceServer,
	settingsServiceServer settingssvc.SettingsServiceServer,
	uploadedMediaServiceServer uploadedmediasvc.UploadedMediaServiceServer,
	webhooksServiceServer webhookssvc.WebhooksServiceServer,
	waitlistsServiceServer waitlistssvc.WaitlistsServiceServer,
	playScriptsServiceServer playscriptssvc.PlayScriptsServiceServer,
	server *grpc.Server,
) *GRPCService {
	return &GRPCService{
		Server:                         server,
		AuditServiceServer:             auditServiceServer,
		AuthServiceServer:              authServiceServer,
		DataPrivacyServiceServer:       dataPrivacyServiceServer,
		IdentityServiceServer:          identityServiceServer,
		InternalOperationsServer:       internalOperationsServer,
		IssueReportsServiceServer:      issueReportsServiceServer,
		UserNotificationsServiceServer: userNotificationsServiceServer,
		OAuthServiceServer:             oauthServiceServer,
		SettingsServiceServer:          settingsServiceServer,
		UploadedMediaServiceServer:     uploadedMediaServiceServer,
		WebhooksServiceServer:          webhooksServiceServer,
		WaitlistsServiceServer:         waitlistsServiceServer,
		PlayScriptsServiceServer:       playScriptsServiceServer,
	}
}
