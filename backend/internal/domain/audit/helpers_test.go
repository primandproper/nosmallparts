package audit

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"

	"github.com/stretchr/testify/assert"
)

func Test_buildDataChangeMessageFromContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		sessionContextData := &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: identifiers.New()},
			ActiveAccountID: identifiers.New(),
		}
		ctx = context.WithValue(ctx, sessions.SessionContextDataKey, sessionContextData)

		expected := &DataChangeMessage{
			EventType: AuditLogEventTypeArchived,
			Context: map[string]any{
				"things": "stuff",
			},
			UserID:    sessionContextData.Requester.UserID,
			AccountID: sessionContextData.ActiveAccountID,
		}

		actual := BuildDataChangeMessageFromContext(ctx, loggingnoop.NewLogger(), expected.EventType, expected.Context)

		assert.Equal(t, expected, actual)
	})
}
