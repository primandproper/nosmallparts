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

	logger := a.logger.WithValue(keys.UserDataAggregationReportIDKey, userDataCollectionRequest.ReportID)
	tracing.AttachToSpan(span, keys.UserDataAggregationReportIDKey, userDataCollectionRequest.ReportID)
	logger.Info("loaded payload, aggregating data")

	logger.Info("compiled payload, saving")

	logger.Info("establishing upload manager")

	collectionBytes, err := json.Marshal(struct{}{})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marshaling collection")
	}

	logger.Info("saving file")

	if err = a.uploadManager.SaveFile(ctx, fmt.Sprintf("%s.json", userDataCollectionRequest.ReportID), collectionBytes); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "saving collection")
	}

	a.userDataAggregationExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()))

	return nil
}
