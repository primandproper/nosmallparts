-- name: CreateWebhookTriggerEvent :exec
INSERT INTO webhook_trigger_events (
	id,
	trigger_event,
	belongs_to_webhook
) VALUES (
	sqlc.arg(id),
	sqlc.arg(trigger_event),
	sqlc.arg(belongs_to_webhook)
);

-- name: ArchiveWebhookTriggerEvent :execrows
UPDATE webhook_trigger_events SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_webhook = sqlc.arg(belongs_to_webhook);
