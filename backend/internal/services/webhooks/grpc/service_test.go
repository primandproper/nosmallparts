package grpc

import (
	"testing"

	webhookmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks/mock"
	webhookssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
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
		webhookRepo := &webhookmock.Repository{}

		service := NewService(logger, tracerProvider, webhookRepo)

		assert.NotNil(t, service)
		assert.Implements(t, (*webhookssvc.WebhooksServiceServer)(nil), service)
	})
}
