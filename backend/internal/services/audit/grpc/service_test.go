package grpc

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit"
	auditfakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit/fakes"
	auditmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/audit/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/filtering"
	auditsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/filtering"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/reflection"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *auditmock.Repository) {
	t.Helper()

	logger := logging.NewNoopLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	auditRepo := &auditmock.Repository{}

	service := &serviceImpl{
		tracer:          tracer,
		logger:          logger,
		auditRepository: auditRepo,
	}

	return service, auditRepo
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		tracerProvider := tracing.NewNoopTracerProvider()
		auditRepo := &auditmock.Repository{}

		service := NewService(logger, tracerProvider, auditRepo)

		assert.NotNil(t, service)
		assert.Implements(t, (*auditsvc.AuditServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, auditRepo, impl.auditRepository)
	})
}

func TestServiceImpl_GetAuditLogEntriesForAccount(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeAuditLogEntries := auditfakes.BuildFakeAuditLogEntriesList()
		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForAccount), testutils.ContextMatcher, "TODO", testutils.QueryFilterMatcher).Return(fakeAuditLogEntries, nil)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForAccountRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForAccount(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeAuditLogEntries.Data))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForAccount), testutils.ContextMatcher, "TODO", testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[audit.AuditLogEntry])(nil), errors.New("repository error"))

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForAccountRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForAccount(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetAuditLogEntriesForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeAuditLogEntries := auditfakes.BuildFakeAuditLogEntriesList()
		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForUser), testutils.ContextMatcher, "TODO", testutils.QueryFilterMatcher).Return(fakeAuditLogEntries, nil)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForUserRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForUser(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeAuditLogEntries.Data))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntriesForUser), testutils.ContextMatcher, "TODO", testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[audit.AuditLogEntry])(nil), errors.New("repository error"))

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &auditsvc.GetAuditLogEntriesForUserRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetAuditLogEntriesForUser(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetAuditLogEntryByID(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeAuditLogEntry := auditfakes.BuildFakeAuditLogEntry()
		entryID := fakeAuditLogEntry.ID

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntry), testutils.ContextMatcher, entryID).Return(fakeAuditLogEntry, nil)

		request := &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		}

		response, err := service.GetAuditLogEntryByID(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeAuditLogEntry.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		entryID := "nonexistent-entry"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetAuditLogEntry), testutils.ContextMatcher, entryID).Return((*audit.AuditLogEntry)(nil), errors.New("repository error"))

		request := &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		}

		response, err := service.GetAuditLogEntryByID(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}
