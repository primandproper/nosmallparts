package grpc

import (
	"testing"

	waitlistmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists/mock"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		waitlistRepo := &waitlistmock.Repository{}

		service := NewService(logger, tracerProvider, waitlistRepo)

		assert.NotNil(t, service)
		assert.Implements(t, (*waitlistssvc.WaitlistsServiceServer)(nil), service)
	})
}
