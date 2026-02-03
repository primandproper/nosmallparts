package fakes

import (
	"net/http"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeWebhook builds a faked Webhook.
func BuildFakeWebhook() *types.Webhook {
	webhookID := BuildFakeID()

	fakeEvent := BuildFakeWebhookTriggerEvent()
	fakeEvent.BelongsToWebhook = webhookID
	events := []*types.WebhookTriggerEvent{fakeEvent}

	return &types.Webhook{
		ID:               webhookID,
		Name:             fake.UUID(),
		ContentType:      "application/json",
		URL:              fake.URL(),
		Method:           http.MethodPost,
		Events:           events,
		CreatedAt:        BuildFakeTime(),
		ArchivedAt:       nil,
		BelongsToAccount: fake.UUID(),
	}
}

// BuildFakeWebhooksList builds a faked WebhookList.
func BuildFakeWebhooksList() *filtering.QueryFilteredResult[types.Webhook] {
	var examples []*types.Webhook
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeWebhook())
	}

	return &filtering.QueryFilteredResult[types.Webhook]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeWebhookTriggerEvent builds a faked WebhookTriggerEvent.
func BuildFakeWebhookTriggerEvent() *types.WebhookTriggerEvent {
	return &types.WebhookTriggerEvent{
		ID:               BuildFakeID(),
		TriggerEvent:     string(types.WebhookCreatedServiceEventType),
		BelongsToWebhook: BuildFakeID(),
		CreatedAt:        BuildFakeTime(),
		ArchivedAt:       nil,
	}
}

// BuildFakeWebhookTriggerEventList builds a faked WebhookList.
func BuildFakeWebhookTriggerEventList() *filtering.QueryFilteredResult[types.WebhookTriggerEvent] {
	var examples []*types.WebhookTriggerEvent
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeWebhookTriggerEvent())
	}

	return &filtering.QueryFilteredResult[types.WebhookTriggerEvent]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

func BuildFakeWebhookTriggerEventCreationRequestInput() *types.WebhookTriggerEventCreationRequestInput {
	triggerEvent := BuildFakeWebhookTriggerEvent()
	return converters.ConvertWebhookTriggerEventToWebhookTriggerEventCreationRequestInput(triggerEvent)
}

// BuildFakeWebhookCreationRequestInput builds a faked WebhookCreationRequestInput from a webhook.
func BuildFakeWebhookCreationRequestInput() *types.WebhookCreationRequestInput {
	webhook := BuildFakeWebhook()
	return converters.ConvertWebhookToWebhookCreationRequestInput(webhook)
}
