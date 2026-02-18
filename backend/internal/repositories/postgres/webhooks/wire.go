package webhooks

import "github.com/google/wire"

var (
	WebhookProviders = wire.NewSet(
		ProvideWebhooksRepository,
	)
)
