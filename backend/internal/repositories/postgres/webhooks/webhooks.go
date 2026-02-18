package webhooks

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/webhooks/generated"
)

const (
	resourceTypeWebhooks             = "webhooks"
	resourceTypeWebhookTriggerEvents = "webhook_trigger_events"
)

var (
	_ types.WebhookDataManager = (*repository)(nil)
)

// WebhookExists fetches whether a webhook exists from the database.
func (r *repository) WebhookExists(ctx context.Context, webhookID, accountID string) (exists bool, err error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if webhookID == "" {
		return false, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, keys.WebhookIDKey, webhookID)

	if accountID == "" {
		return false, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	result, err := r.generatedQuerier.CheckWebhookExistence(ctx, r.readDB, &generated.CheckWebhookExistenceParams{
		BelongsToAccount: accountID,
		ID:               webhookID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing webhook existence check")
	}

	return result, nil
}

// GetWebhook fetches a webhook from the database.
func (r *repository) GetWebhook(ctx context.Context, webhookID, accountID string) (*types.Webhook, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if webhookID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, keys.WebhookIDKey, webhookID)

	if accountID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	results, err := r.generatedQuerier.GetWebhook(ctx, r.readDB, &generated.GetWebhookParams{
		BelongsToAccount: accountID,
		ID:               webhookID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhook")
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	var webhook *types.Webhook
	for _, result := range results {
		if webhook == nil {
			webhook = &types.Webhook{
				CreatedAt:        result.WebhookCreatedAt,
				ArchivedAt:       database.TimePointerFromNullTime(result.WebhookArchivedAt),
				LastUpdatedAt:    database.TimePointerFromNullTime(result.WebhookLastUpdatedAt),
				Name:             result.WebhookName,
				URL:              result.WebhookUrl,
				Method:           result.WebhookMethod,
				ID:               result.WebhookID,
				BelongsToAccount: result.WebhookBelongsToAccount,
				ContentType:      result.WebhookContentType,
				Events:           []*types.WebhookTriggerEvent{},
			}
		}

		if result.WebhookTriggerEventID.Valid {
			// we can safely assume most of these are valid
			webhook.Events = append(webhook.Events, &types.WebhookTriggerEvent{
				CreatedAt:        database.TimeFromNullTime(result.WebhookTriggerEventCreatedAt),
				ArchivedAt:       database.TimePointerFromNullTime(result.WebhookTriggerEventArchivedAt),
				ID:               database.StringFromNullString(result.WebhookTriggerEventID),
				BelongsToWebhook: database.StringFromNullString(result.WebhookTriggerEventBelongsToWebhook),
				TriggerEvent:     string(result.WebhookTriggerEventTriggerEvent.WebhookEvent),
			})
		}
	}

	return webhook, nil
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (r *repository) GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Webhook], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetWebhooksForAccount(ctx, r.readDB, &generated.GetWebhooksForAccountParams{
		BelongsToAccount: accountID,
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhooks from database")
	}

	var (
		data                      []*types.Webhook
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.Webhook{
			CreatedAt:        database.TimeFromNullTime(result.CreatedAt),
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			Name:             result.Name,
			URL:              result.URL,
			Method:           result.Method,
			ID:               result.ID,
			BelongsToAccount: result.BelongsToAccount,
			ContentType:      result.ContentType,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Webhook) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetWebhooksForAccountAndEvent fetches a list of webhooks from the database that meet a particular filter.
func (r *repository) GetWebhooksForAccountAndEvent(ctx context.Context, accountID, eventType string) ([]*types.Webhook, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	databaseResults, err := r.generatedQuerier.GetWebhooksForAccountAndEvent(ctx, r.readDB, &generated.GetWebhooksForAccountAndEventParams{
		BelongsToAccount: accountID,
		TriggerEvent:     generated.WebhookEvent(eventType),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhooks from database")
	}

	results := []*types.Webhook{}
	for _, result := range databaseResults {
		results = append(results, &types.Webhook{
			CreatedAt:        result.CreatedAt,
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			Name:             result.Name,
			URL:              result.URL,
			Method:           result.Method,
			ID:               result.ID,
			BelongsToAccount: result.BelongsToAccount,
			ContentType:      result.ContentType,
		})
	}

	return results, nil
}

// CreateWebhook creates a webhook in a database.
func (r *repository) CreateWebhook(ctx context.Context, input *types.WebhookDatabaseCreationInput) (*types.Webhook, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, keys.AccountIDKey, input.BelongsToAccount)
	logger = logger.WithValue(keys.AccountIDKey, input.BelongsToAccount)

	logger.Debug("CreateWebhook invoked")

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	if err = r.generatedQuerier.CreateWebhook(ctx, tx, &generated.CreateWebhookParams{
		ID:               input.ID,
		Name:             input.Name,
		ContentType:      input.ContentType,
		URL:              input.URL,
		Method:           input.Method,
		BelongsToAccount: input.BelongsToAccount,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing webhook creation query")
	}

	x := &types.Webhook{
		ID:               input.ID,
		Name:             input.Name,
		ContentType:      input.ContentType,
		URL:              input.URL,
		Method:           input.Method,
		BelongsToAccount: input.BelongsToAccount,
		CreatedAt:        r.CurrentTime(),
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &x.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeWebhooks,
		RelevantID:       x.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	for i := range input.Events {
		evt := input.Events[i]
		evt.BelongsToWebhook = input.ID

		e, webhookTriggerEventCreationErr := r.createWebhookTriggerEvent(ctx, tx, x.BelongsToAccount, evt)
		if webhookTriggerEventCreationErr != nil {
			r.RollbackTransaction(ctx, tx)
			return nil, observability.PrepareAndLogError(webhookTriggerEventCreationErr, logger, span, "performing webhook creation query")
		}

		x.Events = append(x.Events, e)
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	tracing.AttachToSpan(span, keys.WebhookIDKey, x.ID)

	return x, nil
}

// createWebhookTriggerEvent creates a webhook trigger event in a database.
func (r *repository) createWebhookTriggerEvent(ctx context.Context, querier database.SQLQueryExecutor, accountID string, input *types.WebhookTriggerEventDatabaseCreationInput) (*types.WebhookTriggerEvent, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, keys.WebhookIDKey, input.BelongsToWebhook)

	if accountID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)
	logger = logger.WithValue(keys.AccountIDKey, accountID)

	if err := r.generatedQuerier.CreateWebhookTriggerEvent(ctx, querier, &generated.CreateWebhookTriggerEventParams{
		ID:               input.ID,
		TriggerEvent:     generated.WebhookEvent(input.TriggerEvent),
		BelongsToWebhook: input.BelongsToWebhook,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing webhook trigger event creation query")
	}

	if _, err := r.auditLogEntryRepo.CreateAuditLogEntry(ctx, querier, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeWebhookTriggerEvents,
		RelevantID:       input.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	x := &types.WebhookTriggerEvent{
		ID:               input.ID,
		TriggerEvent:     input.TriggerEvent,
		BelongsToWebhook: input.BelongsToWebhook,
		CreatedAt:        r.CurrentTime(),
		ArchivedAt:       nil,
	}

	return x, nil
}

// ArchiveWebhook archives a webhook from the database.
func (r *repository) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.WebhookIDKey, webhookID)

	if accountID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	logger := r.logger.WithValues(map[string]any{
		keys.WebhookIDKey: webhookID,
		keys.AccountIDKey: accountID,
	})

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	rowsAffected, err := r.generatedQuerier.ArchiveWebhook(ctx, tx, &generated.ArchiveWebhookParams{
		BelongsToAccount: accountID,
		ID:               webhookID,
	})
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "archiving webhook")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeWebhooks,
		RelevantID:       webhookID,
		EventType:        audit.AuditLogEventTypeArchived,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	return nil
}

// AddWebhookTriggerEvent adds a webhook trigger event from the database.
func (r *repository) AddWebhookTriggerEvent(ctx context.Context, accountID string, input *types.WebhookTriggerEventDatabaseCreationInput) (*types.WebhookTriggerEvent, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.AccountIDKey, accountID)

	if input == nil {
		return nil, database.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, keys.WebhookIDKey, input.BelongsToWebhook)

	logger := r.logger.WithValues(map[string]any{
		keys.WebhookIDKey:             input.BelongsToWebhook,
		keys.WebhookTriggerEventIDKey: input.ID,
		keys.AccountIDKey:             accountID,
	})

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	e, err := r.createWebhookTriggerEvent(ctx, tx, accountID, input)
	if err != nil {
		r.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing webhook creation query")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	return e, nil
}

// ArchiveWebhookTriggerEvent archives a webhook trigger event from the database.
func (r *repository) ArchiveWebhookTriggerEvent(ctx context.Context, webhookID, webhookTriggerEventID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.WebhookIDKey, webhookID)

	if webhookTriggerEventID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.WebhookTriggerEventIDKey, webhookTriggerEventID)

	logger := r.logger.WithValues(map[string]any{
		keys.WebhookIDKey:             webhookID,
		keys.WebhookTriggerEventIDKey: webhookTriggerEventID,
	})

	tx, err := r.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	if _, err = r.generatedQuerier.ArchiveWebhookTriggerEvent(ctx, tx, &generated.ArchiveWebhookTriggerEventParams{
		BelongsToWebhook: webhookID,
		ID:               webhookTriggerEventID,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "archiving webhook trigger event")
	}

	if _, err = r.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeWebhookTriggerEvents,
		RelevantID:   webhookID,
		EventType:    audit.AuditLogEventTypeArchived,
	}); err != nil {
		r.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing database transaction")
	}

	return nil
}
