package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	issuereports "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/issuereports"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

const (
	o11yName = "issue_reports_service"
)

var _ issuereportssvc.IssueReportsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		issuereportssvc.UnimplementedIssueReportsServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		issueReportRepository     issuereports.Repository
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	issueReportRepository issuereports.Repository,
) issuereportssvc.IssueReportsServiceServer {
	return &serviceImpl{
		logger:                    logging.EnsureLogger(logger).WithName(o11yName),
		tracer:                    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
		issueReportRepository:     issueReportRepository,
	}
}
