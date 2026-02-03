package datachangemessagehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	textsearch "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text"
	coreindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"
)

func (a *AsyncDataChangeMessageHandler) SearchIndexRequestsEventHandler(ctx context.Context, rawMsg []byte) error {
	ctx, span := a.tracer.StartSpan(ctx)
	defer span.End()

	start := time.Now()

	var searchIndexRequest textsearch.IndexRequest
	if err := json.NewDecoder(bytes.NewReader(rawMsg)).Decode(&searchIndexRequest); err != nil {
		return fmt.Errorf("decoding JSON body: %w", err)
	}

	switch searchIndexRequest.IndexType {
	case coreindexing.IndexTypeUsers:
		// we don't want to retry indexing perpetually in the event of a fundamental error, so we just log it and move on
		if err := a.userDataIndexer.HandleIndexRequest(ctx, &searchIndexRequest); err != nil {
			return fmt.Errorf("handling search indexing request: %w", err)
		}
	}

	a.searchIndexRequestsExecutionTimeHistogram.Record(ctx, float64(time.Since(start).Milliseconds()))

	return nil
}
