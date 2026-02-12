package client

import (
	"context"

	auditgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	authgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	dataprivacygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	identitygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	internalopsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/internalops"
	issuereportsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"
	notificationsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	oauthgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/oauth"
	settingsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/settings"
	uploadedmediagrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/uploaded_media"
	waitlistsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
	webhooksgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
)

const (
	zuckModeUserHeader    = "X-Zuck-Mode-User"
	zuckModeAccountHeader = "X-Zuck-Mode-Account"
)

// contextKey type for impersonation context values (used by HTTP client to add headers).
type contextKey string

const (
	zuckModeUserContextKey    contextKey = "zuck_mode_user"
	zuckModeAccountContextKey contextKey = "zuck_mode_account"
)

type Client interface {
	authgrpc.AuthServiceClient
	identitygrpc.IdentityServiceClient
	auditgrpc.AuditServiceClient
	dataprivacygrpc.DataPrivacyServiceClient
	internalopsgrpc.InternalOperationsClient
	issuereportsgrpc.IssueReportsServiceClient
	notificationsgrpc.UserNotificationsServiceClient
	oauthgrpc.OAuthServiceClient
	settingsgrpc.SettingsServiceClient
	uploadedmediagrpc.UploadedMediaServiceClient
	waitlistsgrpc.WaitlistsServiceClient
	webhooksgrpc.WebhooksServiceClient
}

// ImpersonateUseAndAccountContext attaches user and account IDs to the context so the HTTP client can send X-Zuck-Mode-User and X-Zuck-Mode-Account headers.
func ImpersonateUseAndAccountContext(ctx context.Context, userID, accountID string) context.Context {
	ctx = context.WithValue(ctx, zuckModeUserContextKey, userID)
	ctx = context.WithValue(ctx, zuckModeAccountContextKey, accountID)
	return ctx
}
