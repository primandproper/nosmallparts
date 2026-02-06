package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/types"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/dataprivacy/grpc/converters"

	"google.golang.org/grpc/codes"
)

const (
	o11yName = "data_privacy_service"
)

var _ dataprivacysvc.DataPrivacyServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		dataprivacysvc.UnimplementedDataPrivacyServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		dataPrivacyRepo           dataprivacy.Repository
		uploadManager             uploads.UploadManager
	}
)

// NewDataPrivacyService creates a new data privacy gRPC service.
func NewDataPrivacyService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error),
	dataPrivacyRepo dataprivacy.Repository,
	uploadManager uploads.UploadManager,
) dataprivacysvc.DataPrivacyServiceServer {
	return &serviceImpl{
		logger:                    logging.EnsureLogger(logger).WithName(o11yName),
		tracer:                    tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		sessionContextDataFetcher: sessionContextDataFetcher,
		dataPrivacyRepo:           dataPrivacyRepo,
		uploadManager:             uploadManager,
	}
}

// AggregateUserDataReport collects all user data and saves it to object storage for GDPR/CCPA disclosure.
func (s *serviceImpl) AggregateUserDataReport(ctx context.Context, _ *dataprivacysvc.AggregateUserDataReportRequest) (*dataprivacysvc.AggregateUserDataReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.Requester.UserID
	logger := s.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.UserIDKey, userID)

	// Generate a unique report ID
	reportID := identifiers.New()
	logger = logger.WithValue(keys.UserDataAggregationReportIDKey, reportID)
	tracing.AttachToSpan(span, keys.UserDataAggregationReportIDKey, reportID)

	logger.Info("aggregating user data")

	// Fetch all user data
	collection, err := s.dataPrivacyRepo.FetchUserDataCollection(ctx, userID)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "fetching user data collection")
	}

	// Marshal and save to object storage
	collectionBytes, err := json.Marshal(collection)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "marshaling user data collection")
	}

	if err = s.uploadManager.SaveFile(ctx, fmt.Sprintf("%s.json", reportID), collectionBytes); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "saving user data report")
	}

	logger.Info("user data aggregation complete")

	return &dataprivacysvc.AggregateUserDataReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		ReportId: reportID,
	}, nil
}

// DestroyAllUserData permanently deletes a user and all associated data.
func (s *serviceImpl) DestroyAllUserData(ctx context.Context, _ *dataprivacysvc.DestroyAllUserDataRequest) (*dataprivacysvc.DestroyAllUserDataResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	sessionContextData, err := s.sessionContextDataFetcher(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, s.logger, span, codes.Unauthenticated, "fetching session context data")
	}

	userID := sessionContextData.Requester.UserID
	logger := s.logger.WithValue(keys.UserIDKey, userID)
	tracing.AttachToSpan(span, keys.UserIDKey, userID)

	logger.Info("destroying all user data")

	if err = s.dataPrivacyRepo.DeleteUser(ctx, userID); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "deleting user")
	}

	logger.Info("user data destroyed successfully")

	return &dataprivacysvc.DestroyAllUserDataResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		Successful: true,
	}, nil
}

// FetchUserDataReport retrieves a previously generated user data report from object storage.
func (s *serviceImpl) FetchUserDataReport(ctx context.Context, request *dataprivacysvc.FetchUserDataReportRequest) (*dataprivacysvc.FetchUserDataReportResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	reportID := request.GetUserDataAggregationReportId()
	logger := s.logger.WithValue(keys.UserDataAggregationReportIDKey, reportID)
	tracing.AttachToSpan(span, keys.UserDataAggregationReportIDKey, reportID)

	logger.Info("fetching user data report")

	// Read the report from object storage
	reportBytes, err := s.uploadManager.ReadFile(ctx, fmt.Sprintf("%s.json", reportID))
	if err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.NotFound, "reading report from storage")
	}

	// Unmarshal the report
	var collection dataprivacy.UserDataCollection
	if err = json.Unmarshal(reportBytes, &collection); err != nil {
		return nil, observability.PrepareAndLogGRPCStatus(err, logger, span, codes.Internal, "unmarshaling report")
	}

	logger.Info("user data report fetched successfully")

	// Convert to proto type
	return &dataprivacysvc.FetchUserDataReportResponse{
		ResponseDetails: &types.ResponseDetails{
			TraceId: span.SpanContext().TraceID().String(),
		},
		UserDataCollection: converters.ConvertUserDataCollectionToGRPCUserDataCollection(&collection, reportID),
	}, nil
}
