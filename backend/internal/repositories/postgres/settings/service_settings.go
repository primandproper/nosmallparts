package settings

import (
	"context"
	"database/sql"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	platformkeys "github.com/verygoodsoftwarenotvirus/platform/v4/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/settings/generated"
)

const (
	resourceTypeServiceSettings  = "service_settings"
	serviceSettingsEnumDelimiter = "|"
)

var (
	_ types.ServiceSettingDataManager = (*repository)(nil)
)

// ServiceSettingExists fetches whether a service setting exists from the database.
func (q *repository) ServiceSettingExists(ctx context.Context, serviceSettingID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ServiceSettingIDKey, serviceSettingID)
	tracing.AttachToSpan(span, keys.ServiceSettingIDKey, serviceSettingID)

	result, err := q.generatedQuerier.CheckServiceSettingExistence(ctx, q.readDB, serviceSettingID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing service setting existence check")
	}

	return result, nil
}

// GetServiceSetting fetches a service setting from the database.
func (q *repository) GetServiceSetting(ctx context.Context, serviceSettingID string) (*types.ServiceSetting, error) {
	return q.getServiceSetting(ctx, q.readDB, serviceSettingID)
}

// getServiceSetting fetches a service setting from the database.
func (q *repository) getServiceSetting(ctx context.Context, db database.SQLQueryExecutor, serviceSettingID string) (*types.ServiceSetting, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ServiceSettingIDKey, serviceSettingID)
	tracing.AttachToSpan(span, keys.ServiceSettingIDKey, serviceSettingID)

	result, err := q.generatedQuerier.GetServiceSetting(ctx, db, serviceSettingID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing service setting fetch")
	}

	usableEnumeration := []string{}
	for _, x := range strings.Split(result.Enumeration, serviceSettingsEnumDelimiter) {
		if strings.TrimSpace(x) != "" {
			usableEnumeration = append(usableEnumeration, x)
		}
	}

	serviceSetting := &types.ServiceSetting{
		CreatedAt:     result.CreatedAt,
		DefaultValue:  database.StringPointerFromNullString(result.DefaultValue),
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		Name:          result.Name,
		Type:          string(result.Type),
		Description:   result.Description,
		Enumeration:   usableEnumeration,
		AdminsOnly:    result.AdminsOnly,
	}

	return serviceSetting, nil
}

// SearchForServiceSettings fetches a service setting from the database.
func (q *repository) SearchForServiceSettings(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSetting], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, keys.ServiceSettingIDKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.SearchForServiceSettings(ctx, q.readDB, &generated.SearchForServiceSettingsParams{
		NameQuery:       query,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing service settings list retrieval query")
	}

	var (
		data                      = []*types.ServiceSetting{}
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		rawEnumeration := strings.Split(result.Enumeration, serviceSettingsEnumDelimiter)
		usableEnumeration := []string{}
		for _, y := range rawEnumeration {
			if strings.TrimSpace(y) != "" {
				usableEnumeration = append(usableEnumeration, y)
			}
		}

		serviceSetting := &types.ServiceSetting{
			CreatedAt:     result.CreatedAt,
			DefaultValue:  database.StringPointerFromNullString(result.DefaultValue),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Type:          string(result.Type),
			Description:   result.Description,
			Enumeration:   usableEnumeration,
			AdminsOnly:    result.AdminsOnly,
		}

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
		data = append(data, serviceSetting)
	}

	result := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.ServiceSetting) string {
			return t.ID
		},
		filter,
	)

	return result, nil
}

// GetServiceSettings fetches a list of service settings from the database that meet a particular filter.
func (q *repository) GetServiceSettings(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ServiceSetting], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetServiceSettings(ctx, q.readDB, &generated.GetServiceSettingsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing service settings list retrieval query")
	}

	var (
		data                      = []*types.ServiceSetting{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		rawEnumeration := strings.Split(result.Enumeration, serviceSettingsEnumDelimiter)
		usableEnumeration := []string{}
		for _, y := range rawEnumeration {
			if strings.TrimSpace(y) != "" {
				usableEnumeration = append(usableEnumeration, y)
			}
		}

		data = append(data, &types.ServiceSetting{
			CreatedAt:     result.CreatedAt,
			DefaultValue:  database.StringPointerFromNullString(result.DefaultValue),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Type:          string(result.Type),
			Description:   result.Description,
			Enumeration:   usableEnumeration,
			AdminsOnly:    result.AdminsOnly,
		})
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.ServiceSetting) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateServiceSetting creates a service setting in the database.
func (q *repository) CreateServiceSetting(ctx context.Context, input *types.ServiceSettingDatabaseCreationInput) (*types.ServiceSetting, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, keys.ServiceSettingIDKey, input.ID)
	logger := q.logger.WithValue(keys.ServiceSettingIDKey, input.ID)

	tx, err := q.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	// create the service setting.
	if err = q.generatedQuerier.CreateServiceSetting(ctx, tx, &generated.CreateServiceSettingParams{
		ID:           input.ID,
		Name:         input.Name,
		Type:         generated.SettingType(input.Type),
		Description:  input.Description,
		Enumeration:  strings.Join(input.Enumeration, serviceSettingsEnumDelimiter),
		DefaultValue: database.NullStringFromStringPointer(input.DefaultValue),
		AdminsOnly:   input.AdminsOnly,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing service setting creation query")
	}

	x := &types.ServiceSetting{
		ID:           input.ID,
		Name:         input.Name,
		Type:         input.Type,
		Description:  input.Description,
		DefaultValue: input.DefaultValue,
		AdminsOnly:   input.AdminsOnly,
		Enumeration:  input.Enumeration,
		CreatedAt:    q.CurrentTime(),
	}

	if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeServiceSettings,
		RelevantID:   x.ID,
		EventType:    audit.AuditLogEventTypeCreated,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	logger.Info("service setting created")

	return x, nil
}

// ArchiveServiceSetting archives a service setting from the database by its ID.
func (q *repository) ArchiveServiceSetting(ctx context.Context, serviceSettingID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if serviceSettingID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.ServiceSettingIDKey, serviceSettingID)
	tracing.AttachToSpan(span, keys.ServiceSettingIDKey, serviceSettingID)

	tx, err := q.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	rowsAffected, err := q.generatedQuerier.ArchiveServiceSetting(ctx, tx, serviceSettingID)
	if err != nil {
		q.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "updating service setting")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeServiceSettings,
		RelevantID:   serviceSettingID,
		EventType:    audit.AuditLogEventTypeArchived,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	return nil
}
