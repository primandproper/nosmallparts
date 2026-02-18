package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

// UserDataAggregationEventHandler handles user data aggregation requests for GDPR/CCPA compliance.
func (a *AsyncDataChangeMessageHandler) UserDataAggregationEventHandler(
	ctx context.Context,
	rawMsg []byte,
) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	start := time.Now()

	var userDataCollectionRequest dataprivacy.UserDataAggregationRequest
	if err := a.decoder.DecodeBytes(ctx, rawMsg, &userDataCollectionRequest); err != nil {
		return fmt.Errorf("decoding JSON body: %w", err)
	}

	logger := a.logger.WithValue(keys.UserDataAggregationReportIDKey, userDataCollectionRequest.ReportID).
		WithValue(keys.UserIDKey, userDataCollectionRequest.UserID)
	tracing.AttachToSpan(span, keys.UserDataAggregationReportIDKey, userDataCollectionRequest.ReportID)
	tracing.AttachToSpan(span, keys.UserIDKey, userDataCollectionRequest.UserID)
	logger.Info("loaded payload, aggregating user data")

	// Fetch the user's complete data collection
	collection, err := a.dataPrivacyRepo.FetchUserDataCollection(ctx, userDataCollectionRequest.UserID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching user data collection")
	}

	logger.Info("compiled user data payload, marshaling")

	// Marshal the collection to JSON
	collectionBytes, err := json.Marshal(collection)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marshaling collection")
	}

	logger.Info("saving file to object storage")

	// Save to object storage with report ID as filename
	if err = a.uploadManager.SaveFile(ctx, fmt.Sprintf("%s.json", userDataCollectionRequest.ReportID), collectionBytes); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "saving collection")
	}

	a.userDataAggregationExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()))

	logger.Info("user data aggregation complete")

	return nil
}
