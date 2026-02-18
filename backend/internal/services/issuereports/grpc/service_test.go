package grpc

import (
	"testing"

	issuereportmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/issuereports/mock"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"
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
		issueReportRepo := &issuereportmock.Repository{}

		service := NewService(logger, tracerProvider, issueReportRepo)

		assert.NotNil(t, service)
		assert.Implements(t, (*issuereportssvc.IssueReportsServiceServer)(nil), service)
	})
}
