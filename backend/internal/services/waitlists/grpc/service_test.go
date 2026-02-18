package grpc

import (
	"testing"

	waitlistmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists/mock"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
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
		waitlistRepo := &waitlistmock.Repository{}

		service := NewService(logger, tracerProvider, waitlistRepo)

		assert.NotNil(t, service)
		assert.Implements(t, (*waitlistssvc.WaitlistsServiceServer)(nil), service)
	})
}
