package waitlists

import (
	"context"
	"database/sql"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	generated "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/waitlists/generated"
)

var (
	_ types.Repository = (*repository)(nil)
)

// WaitlistIsNotExpired checks if a waitlist exists and is not expired.
func (r *repository) WaitlistIsNotExpired(ctx context.Context, waitlistID string) (bool, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistID == "" {
		return false, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, keys.WaitlistIDKey, waitlistID)

	exists, err := r.generatedQuerier.CheckWaitlistExistence(ctx, r.db, waitlistID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking waitlist existence")
	}

	if !exists {
		return false, sql.ErrNoRows
	}

	result, err := r.generatedQuerier.WaitlistIsNotExpired(ctx, r.db, waitlistID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking waitlist expiration status")
	}

	return result, nil
}

// GetWaitlist fetches a waitlist from the database.
func (r *repository) GetWaitlist(ctx context.Context, waitlistID string) (*types.Waitlist, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, keys.WaitlistIDKey, waitlistID)

	result, err := r.generatedQuerier.GetWaitlist(ctx, r.db, waitlistID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist")
	}

	waitlist := &types.Waitlist{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		Name:          result.Name,
		Description:   result.Description,
		ValidUntil:    result.ValidUntil,
	}

	return waitlist, nil
}

// GetWaitlists fetches waitlists with filtering.
func (r *repository) GetWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Waitlist], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetWaitlists(ctx, r.db, &generated.GetWaitlistsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlists from database")
	}

	var (
		data                      []*types.Waitlist
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.Waitlist{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Description:   result.Description,
			ValidUntil:    result.ValidUntil,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Waitlist) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// GetActiveWaitlists fetches non-expired waitlists with filtering.
func (r *repository) GetActiveWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Waitlist], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetActiveWaitlists(ctx, r.db, &generated.GetActiveWaitlistsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching active waitlists from database")
	}

	var (
		data                      []*types.Waitlist
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.Waitlist{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Description:   result.Description,
			ValidUntil:    result.ValidUntil,
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Waitlist) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateWaitlist creates a waitlist in the database.
func (r *repository) CreateWaitlist(ctx context.Context, input *types.WaitlistDatabaseCreationInput) (*types.Waitlist, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}

	logger := r.logger.WithValue(keys.WaitlistIDKey, input.ID)
	tracing.AttachToSpan(span, keys.WaitlistIDKey, input.ID)

	if err := r.generatedQuerier.CreateWaitlist(ctx, r.db, &generated.CreateWaitlistParams{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		ValidUntil:  input.ValidUntil,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing waitlist creation query")
	}

	x := &types.Waitlist{
		ID:          input.ID,
		CreatedAt:   r.CurrentTime(),
		Name:        input.Name,
		Description: input.Description,
		ValidUntil:  input.ValidUntil,
	}

	logger.Info("waitlist created")
	return x, nil
}

// UpdateWaitlist updates a waitlist.
func (r *repository) UpdateWaitlist(ctx context.Context, updated *types.Waitlist) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return database.ErrNilInputProvided
	}
	logger := r.logger.WithValue(keys.WaitlistIDKey, updated.ID)
	tracing.AttachToSpan(span, keys.WaitlistIDKey, updated.ID)

	if _, err := r.generatedQuerier.UpdateWaitlist(ctx, r.db, &generated.UpdateWaitlistParams{
		Name:        updated.Name,
		Description: updated.Description,
		ValidUntil:  updated.ValidUntil,
		ID:          updated.ID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating waitlist")
	}

	return nil
}

// ArchiveWaitlist archives a waitlist.
func (r *repository) ArchiveWaitlist(ctx context.Context, waitlistID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if waitlistID == "" {
		return database.ErrInvalidIDProvided
	}
	logger := r.logger.WithValue(keys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, keys.WaitlistIDKey, waitlistID)

	recordsChanged, err := r.generatedQuerier.ArchiveWaitlist(ctx, r.db, waitlistID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving waitlist")
	}

	if recordsChanged == 0 {
		return sql.ErrNoRows
	}

	logger.Info("waitlist archived")
	return nil
}

// GetWaitlistSignup fetches a waitlist signup from the database.
func (r *repository) GetWaitlistSignup(ctx context.Context, waitlistSignupID, waitlistID string) (*types.WaitlistSignup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistSignupID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.WaitlistSignupIDKey, waitlistSignupID)
	tracing.AttachToSpan(span, keys.WaitlistSignupIDKey, waitlistSignupID)

	if waitlistID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, keys.WaitlistIDKey, waitlistID)

	result, err := r.generatedQuerier.GetWaitlistSignup(ctx, r.db, &generated.GetWaitlistSignupParams{
		ID:                waitlistSignupID,
		BelongsToWaitlist: database.NullStringFromString(waitlistID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signup")
	}

	waitlistSignup := &types.WaitlistSignup{
		CreatedAt:         result.CreatedAt,
		LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
		ID:                result.ID,
		Notes:             result.Notes,
		BelongsToWaitlist: database.StringFromNullString(result.BelongsToWaitlist),
		BelongsToUser:     database.StringFromNullString(result.BelongsToUser),
		BelongsToAccount:  database.StringFromNullString(result.BelongsToAccount),
	}

	return waitlistSignup, nil
}

// GetWaitlistSignupsForWaitlist fetches waitlist signups for a waitlist with filtering.
func (r *repository) GetWaitlistSignupsForWaitlist(ctx context.Context, waitlistID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.WaitlistSignup], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if waitlistID == "" {
		return nil, database.ErrInvalidIDProvided
	}
	logger = logger.WithValue(keys.WaitlistIDKey, waitlistID)
	tracing.AttachToSpan(span, keys.WaitlistIDKey, waitlistID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetWaitlistSignupsForWaitlist(ctx, r.db, &generated.GetWaitlistSignupsForWaitlistParams{
		BelongsToWaitlist: database.NullStringFromString(waitlistID),
		CreatedAfter:      database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:     database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:     database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:      database.NullTimeFromTimePointer(filter.UpdatedAfter),
		IncludeArchived:   database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:            database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:       database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching waitlist signups from database")
	}

	var (
		data                      []*types.WaitlistSignup
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		data = append(data, &types.WaitlistSignup{
			CreatedAt:         result.CreatedAt,
			LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
			ID:                result.ID,
			Notes:             result.Notes,
			BelongsToWaitlist: database.StringFromNullString(result.BelongsToWaitlist),
			BelongsToUser:     database.StringFromNullString(result.BelongsToUser),
			BelongsToAccount:  database.StringFromNullString(result.BelongsToAccount),
		})

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.WaitlistSignup) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateWaitlistSignup creates a waitlist signup in the database.
func (r *repository) CreateWaitlistSignup(ctx context.Context, input *types.WaitlistSignupDatabaseCreationInput) (*types.WaitlistSignup, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, database.ErrNilInputProvided
	}

	logger := r.logger.WithValue(keys.WaitlistSignupIDKey, input.ID)
	tracing.AttachToSpan(span, keys.WaitlistSignupIDKey, input.ID)

	if err := r.generatedQuerier.CreateWaitlistSignup(ctx, r.db, &generated.CreateWaitlistSignupParams{
		ID:                input.ID,
		Notes:             input.Notes,
		BelongsToWaitlist: database.NullStringFromString(input.BelongsToWaitlist),
		BelongsToUser:     database.NullStringFromString(input.BelongsToUser),
		BelongsToAccount:  database.NullStringFromString(input.BelongsToAccount),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing waitlist signup creation query")
	}

	x := &types.WaitlistSignup{
		ID:                input.ID,
		CreatedAt:         r.CurrentTime(),
		Notes:             input.Notes,
		BelongsToWaitlist: input.BelongsToWaitlist,
		BelongsToUser:     input.BelongsToUser,
		BelongsToAccount:  input.BelongsToAccount,
	}

	logger.Info("waitlist signup created")
	return x, nil
}

// UpdateWaitlistSignup updates a waitlist signup.
func (r *repository) UpdateWaitlistSignup(ctx context.Context, updated *types.WaitlistSignup) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return database.ErrNilInputProvided
	}
	logger := r.logger.WithValue(keys.WaitlistSignupIDKey, updated.ID)
	tracing.AttachToSpan(span, keys.WaitlistSignupIDKey, updated.ID)

	if _, err := r.generatedQuerier.UpdateWaitlistSignup(ctx, r.db, &generated.UpdateWaitlistSignupParams{
		Notes: updated.Notes,
		ID:    updated.ID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating waitlist signup")
	}

	return nil
}

// ArchiveWaitlistSignup archives a waitlist signup.
func (r *repository) ArchiveWaitlistSignup(ctx context.Context, waitlistSignupID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if waitlistSignupID == "" {
		return database.ErrInvalidIDProvided
	}
	logger := r.logger.WithValue(keys.WaitlistSignupIDKey, waitlistSignupID)
	tracing.AttachToSpan(span, keys.WaitlistSignupIDKey, waitlistSignupID)

	recordsChanged, err := r.generatedQuerier.ArchiveWaitlistSignup(ctx, r.db, waitlistSignupID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving waitlist signup")
	}

	if recordsChanged == 0 {
		return sql.ErrNoRows
	}

	logger.Info("waitlist signup archived")
	return nil
}
