package converters

import (
	"log"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
)

func ConvertStringToWebhookContentType(s string) webhookssvc.WebhookContentType {
	switch s {
	case encoding.ContentTypeToString(encoding.ContentTypeXML):
		return webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_XML
	case encoding.ContentTypeToString(encoding.ContentTypeJSON):
		return webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON
	default:
		log.Printf("unknown content type: %q", s)
		return webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON
	}
}

func ConvertWebhookContentTypeToString(s webhookssvc.WebhookContentType) string {
	switch s {
	case webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_XML:
		return encoding.ContentTypeToString(encoding.ContentTypeXML)
	case webhookssvc.WebhookContentType_WEBHOOK_CONTENT_TYPE_JSON:
		return encoding.ContentTypeToString(encoding.ContentTypeJSON)
	default:
		log.Printf("unknown content type: %q", s)
		return encoding.ContentTypeToString(encoding.ContentTypeJSON)
	}
}

func ConvertStringToWebhookMethod(s string) webhookssvc.WebhookMethod {
	switch s {
	case http.MethodGet:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_GET
	case http.MethodPut:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_PUT
	case http.MethodPatch:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_PATCH
	case http.MethodDelete:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_DELETE
	case http.MethodPost:
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST
	default:
		log.Printf("unknown webhook method: %q", s)
		return webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST
	}
}

func ConvertWebhookMethodToString(s webhookssvc.WebhookMethod) string {
	switch s {
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_GET:
		return http.MethodGet
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_PUT:
		return http.MethodPut
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_PATCH:
		return http.MethodPatch
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_DELETE:
		return http.MethodDelete
	case webhookssvc.WebhookMethod_WEBHOOK_METHOD_POST:
		return http.MethodPost
	default:
		log.Printf("unknown webhook method: %q", s)
		return http.MethodPost
	}
}

func ConvertWebhookToGRPCWebhook(webhook *webhooks.Webhook) *webhookssvc.Webhook {
	converted := &webhookssvc.Webhook{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(webhook.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(webhook.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(webhook.LastUpdatedAt),
		Name:             webhook.Name,
		Url:              webhook.URL,
		Method:           ConvertStringToWebhookMethod(webhook.Method),
		Id:               webhook.ID,
		BelongsToAccount: webhook.BelongsToAccount,
		ContentType:      ConvertStringToWebhookContentType(webhook.ContentType),
	}

	for _, event := range webhook.Events {
		converted.Events = append(converted.Events, ConvertWebhookTriggerEventToGRPCWebhookTriggerEvent(event))
	}

	return converted
}

func ConvertWebhookTriggerEventToGRPCWebhookTriggerEvent(z *webhooks.WebhookTriggerEvent) *webhookssvc.WebhookTriggerEvent {
	return &webhookssvc.WebhookTriggerEvent{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(z.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(z.ArchivedAt),
		Id:               z.ID,
		BelongsToWebhook: z.BelongsToWebhook,
		TriggerEvent:     z.TriggerEvent,
	}
}

func ConvertGRPCWebhookToWebhook(webhook *webhookssvc.Webhook) *webhooks.Webhook {
	converted := &webhooks.Webhook{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(webhook.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(webhook.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertPBTimestampToTimePointer(webhook.LastUpdatedAt),
		Name:             webhook.Name,
		URL:              webhook.Url,
		Method:           ConvertWebhookMethodToString(webhook.Method),
		ContentType:      ConvertWebhookContentTypeToString(webhook.ContentType),
		ID:               webhook.Id,
		BelongsToAccount: webhook.BelongsToAccount,
	}

	for _, event := range webhook.Events {
		converted.Events = append(converted.Events, ConvertGRPCWebhookTriggerEventToWebhookTriggerEvent(event))
	}

	return converted
}

func ConvertGRPCWebhookTriggerEventToWebhookTriggerEvent(z *webhookssvc.WebhookTriggerEvent) *webhooks.WebhookTriggerEvent {
	return &webhooks.WebhookTriggerEvent{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(z.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(z.ArchivedAt),
		ID:               z.Id,
		BelongsToWebhook: z.BelongsToWebhook,
		TriggerEvent:     z.TriggerEvent,
	}
}

func ConvertGRPCWebhookCreationRequestInputToWebhookDatabaseCreationInput(input *webhookssvc.WebhookCreationRequestInput, accountID string) *webhooks.WebhookDatabaseCreationInput {
	webhookID := identifiers.New()

	var events []*webhooks.WebhookTriggerEventDatabaseCreationInput
	for _, event := range input.Events {
		events = append(events, &webhooks.WebhookTriggerEventDatabaseCreationInput{
			ID:               identifiers.New(),
			BelongsToWebhook: webhookID,
			TriggerEvent:     event,
		})
	}

	x := &webhooks.WebhookDatabaseCreationInput{
		ID:               webhookID,
		Name:             input.Name,
		URL:              input.Url,
		Method:           ConvertWebhookMethodToString(input.Method),
		ContentType:      ConvertWebhookContentTypeToString(input.ContentType),
		BelongsToAccount: accountID,
		Events:           events,
	}

	return x
}

func ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(input *webhooks.WebhookCreationRequestInput) *webhookssvc.WebhookCreationRequestInput {
	return &webhookssvc.WebhookCreationRequestInput{
		Name:        input.Name,
		ContentType: ConvertStringToWebhookContentType(input.ContentType),
		Url:         input.URL,
		Method:      ConvertStringToWebhookMethod(input.Method),
		Events:      input.Events,
	}
}

func ConvertGRPCWebhookTriggerEventDatabaseCreationInputToWebhookTriggerEventDatabaseCreationInput(input *webhookssvc.WebhookTriggerEventCreationRequestInput) *webhooks.WebhookTriggerEventDatabaseCreationInput {
	return &webhooks.WebhookTriggerEventDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: input.BelongsToWebhook,
		TriggerEvent:     input.TriggerEvent,
	}
}

// ConvertUserDataCollectionToGRPCDataCollection converts a domain webhooks UserDataCollection to a proto DataCollection.
func ConvertUserDataCollectionToGRPCDataCollection(input *webhooks.UserDataCollection) *webhookssvc.DataCollection {
	result := &webhookssvc.DataCollection{
		Webhooks: make(map[string]*webhookssvc.WebhookList),
	}

	for accountID, webhookList := range input.Data {
		var grpcWebhooks []*webhookssvc.Webhook
		for i := range webhookList {
			grpcWebhooks = append(grpcWebhooks, ConvertWebhookToGRPCWebhook(&webhookList[i]))
		}
		result.Webhooks[accountID] = &webhookssvc.WebhookList{Webhooks: grpcWebhooks}
	}

	return result
}
