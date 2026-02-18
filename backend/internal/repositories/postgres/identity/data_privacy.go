package identity

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

// DeleteUser hard-deletes a user and all associated data via ON DELETE CASCADE.
func (r *repository) DeleteUser(ctx context.Context, userID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return database.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, keys.UserIDKey, userID)
	logger := r.logger.WithValue(keys.UserIDKey, userID)

	changed, err := r.generatedQuerier.DeleteUser(ctx, r.writeDB, userID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "deleting user")
	}

	if changed == 0 {
		return sql.ErrNoRows
	}

	logger.Info("user deleted")

	return nil
}
