package audit

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
)

func BuildDataChangeMessageFromContext(ctx context.Context, logger logging.Logger, eventType string, metadata map[string]any) *DataChangeMessage {
	sessionContext, err := sessions.FetchContextDataFromContext(ctx)
	if err != nil {
		logger.WithValue("event_type", eventType).Info("failed to extract session data from context")
	}

	x := &DataChangeMessage{
		EventType: eventType,
		Context:   metadata,
	}

	if sessionContext != nil {
		x.UserID = sessionContext.Requester.UserID
		x.AccountID = sessionContext.ActiveAccountID
	}

	return x
}
