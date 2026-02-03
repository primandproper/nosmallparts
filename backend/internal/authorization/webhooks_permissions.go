package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// CreateWebhooksPermission is an account admin permission.
	CreateWebhooksPermission Permission = "create.webhooks"
	// ReadWebhooksPermission is an account admin permission.
	ReadWebhooksPermission Permission = "read.webhooks"
	// UpdateWebhooksPermission is an account admin permission.
	UpdateWebhooksPermission Permission = "update.webhooks"
	// ArchiveWebhooksPermission is an account admin permission.
	ArchiveWebhooksPermission Permission = "archive.webhooks"
	// CreateWebhookTriggerEventsPermission is an account admin permission.
	CreateWebhookTriggerEventsPermission Permission = "create.webhook_trigger_events"
	// ArchiveWebhookTriggerEventsPermission is an account admin permission.
	ArchiveWebhookTriggerEventsPermission Permission = "archive.webhook_trigger_events"
)

var (
	// WebhooksPermissions contains all webhook-related permissions.
	WebhooksPermissions = []gorbac.Permission{
		CreateWebhooksPermission,
		ReadWebhooksPermission,
		UpdateWebhooksPermission,
		ArchiveWebhooksPermission,
		CreateWebhookTriggerEventsPermission,
		ArchiveWebhookTriggerEventsPermission,
	}
)
