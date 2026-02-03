package webhooks

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/webhooks/generated"
)

func (r *repository) CollectUserData(ctx context.Context, accountIDs []string) (*webhooks.UserDataCollection, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span)
	filter := filtering.DefaultQueryFilter()

	x := &webhooks.UserDataCollection{}

	for _, accountID := range accountIDs {
		accountWebhooks, err := r.generatedQuerier.GetWebhooksForAccount(ctx, r.db, &generated.GetWebhooksForAccountParams{
			CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
			CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
			UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
			UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
			Cursor:           database.NullStringFromStringPointer(filter.Cursor),
			ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
			IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
			BelongsToAccount: accountID,
		})
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "retrieving webhooks for account")
		}

		for _, webhook := range accountWebhooks {
			x.Data[accountID] = append(x.Data[accountID], webhooks.Webhook{
				CreatedAt:        database.TimeFromNullTime(webhook.CreatedAt),
				ArchivedAt:       database.TimePointerFromNullTime(webhook.ArchivedAt),
				LastUpdatedAt:    database.TimePointerFromNullTime(webhook.LastUpdatedAt),
				Name:             webhook.Name,
				URL:              webhook.URL,
				Method:           webhook.Method,
				ID:               webhook.ID,
				BelongsToAccount: webhook.BelongsToAccount,
				ContentType:      webhook.ContentType,
				Events:           nil,
			})
		}
	}

	return x, nil
}
