package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	auditsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	grpctypes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/observability/keys"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/audit/grpc/converters"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "audit_service"
)

var _ auditsvc.AuditServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		auditsvc.UnimplementedAuditServiceServer
		tracer          tracing.Tracer
		logger          logging.Logger
		auditRepository audit.Repository
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditRepository audit.Repository,
) auditsvc.AuditServiceServer {
	return &serviceImpl{
		logger:          logging.NewNamedLogger(logger, o11yName),
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		auditRepository: auditRepository,
	}
}

func (s *serviceImpl) GetAuditLogEntriesForAccount(ctx context.Context, request *auditsvc.GetAuditLogEntriesForAccountRequest) (*auditsvc.GetAuditLogEntriesForAccountResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("", "")
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	auditLogEntries, err := s.auditRepository.GetAuditLogEntriesForAccount(ctx, "TODO", filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "")
	}

	x := &auditsvc.GetAuditLogEntriesForAccountResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(auditLogEntries.Pagination, filter),
		Results:    nil,
	}

	for _, y := range auditLogEntries.Data {
		x.Results = append(x.Results, converters.ConvertAuditLogEntryToGRPCAuditLogEntry(y))
	}

	return x, nil
}

func (s *serviceImpl) GetAuditLogEntriesForUser(ctx context.Context, request *auditsvc.GetAuditLogEntriesForUserRequest) (*auditsvc.GetAuditLogEntriesForUserResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("", "")
	filter := grpcconverters.ConvertGRPCQueryFilterToQueryFilter(request.Filter)

	auditLogEntries, err := s.auditRepository.GetAuditLogEntriesForUser(ctx, "TODO", filter)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "")
	}

	x := &auditsvc.GetAuditLogEntriesForUserResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Pagination: grpcconverters.ConvertPaginationToGRPCPagination(auditLogEntries.Pagination, filter),
		Results:    nil,
	}

	for _, y := range auditLogEntries.Data {
		x.Results = append(x.Results, converters.ConvertAuditLogEntryToGRPCAuditLogEntry(y))
	}

	return x, nil
}

func (s *serviceImpl) GetAuditLogEntryByID(ctx context.Context, request *auditsvc.GetAuditLogEntryByIDRequest) (*auditsvc.GetAuditLogEntryByIDResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue(keys.AuditLogEntryIDKey, request.AuditLogEntryId)
	auditLogEntry, err := s.auditRepository.GetAuditLogEntry(ctx, request.AuditLogEntryId)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "")
	}

	returnValue := converters.ConvertAuditLogEntryToGRPCAuditLogEntry(auditLogEntry)

	x := &auditsvc.GetAuditLogEntryByIDResponse{
		ResponseDetails: &grpctypes.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Result: returnValue,
	}

	return x, nil
}
