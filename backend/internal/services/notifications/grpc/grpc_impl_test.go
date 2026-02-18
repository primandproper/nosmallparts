package grpc

import (
	"testing"

	notificationsmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications/mock"
	notificationssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		notificationsRepo := &notificationsmock.Repository{}

		service := NewService(logger, tracerProvider, notificationsRepo)

		assert.NotNil(t, service)
		assert.Implements(t, (*notificationssvc.UserNotificationsServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, notificationsRepo, impl.notificationsRepository)
		assert.NotNil(t, impl.sessionContextDataFetcher)
	})
}
