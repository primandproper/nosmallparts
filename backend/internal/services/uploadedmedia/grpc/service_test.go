package grpc

import (
	"testing"

	uploadedmediamock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/uploadedmedia/mock"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/uploaded_media"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	mockuploads "github.com/verygoodsoftwarenotvirus/platform/v4/uploads/mock"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		uploadedMediaRepo := &uploadedmediamock.Repository{}

		service := NewService(logger, tracerProvider, uploadedMediaRepo, &mockuploads.MockUploadManager{})

		assert.NotNil(t, service)
		assert.Implements(t, (*uploadedmediasvc.UploadedMediaServiceServer)(nil), service)
	})
}
