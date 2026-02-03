package grpc

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
)

// WebhooksMethodPermissions is a named type for Wire dependency injection.
type WebhooksMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the webhooks service's method permissions.
func ProvideMethodPermissions() WebhooksMethodPermissions {
	return WebhooksMethodPermissions{
		webhookssvc.WebhooksService_GetWebhook_FullMethodName: {
			authorization.ReadWebhooksPermission,
		},
		webhookssvc.WebhooksService_GetWebhooks_FullMethodName: {
			authorization.ReadWebhooksPermission,
		},
		webhookssvc.WebhooksService_CreateWebhook_FullMethodName: {
			authorization.CreateWebhooksPermission,
		},
		webhookssvc.WebhooksService_ArchiveWebhook_FullMethodName: {
			authorization.ArchiveWebhooksPermission,
		},
		webhookssvc.WebhooksService_AddWebhookTriggerEvent_FullMethodName: {
			authorization.CreateWebhookTriggerEventsPermission,
		},
		webhookssvc.WebhooksService_ArchiveWebhookTriggerEvent_FullMethodName: {
			authorization.ArchiveWebhookTriggerEventsPermission,
		},
	}
}
