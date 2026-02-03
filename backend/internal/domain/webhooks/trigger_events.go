package webhooks

type TriggerEvent string

const (
	// TODO: make sure these match up with the database enums somehow?
	WebhookCreatedTriggerEvent  = "webhook_created"
	WebhookUpdatedTriggerEvent  = "webhook_updated"
	WebhookArchivedTriggerEvent = "webhook_archived"
)
