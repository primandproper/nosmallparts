package indexing

import (
	"context"
	"errors"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	textsearch "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text"
)

const (
	o11yName = "core_search_indexer"
)

var (
	ErrNilIndexRequest = errors.New("nil index request")
)

type UserDataIndexer struct {
	logger          logging.Logger
	tracer          tracing.Tracer
	identityRepo    identity.Repository
	userSearchIndex UserTextSearcher
}

func NewCoreDataIndexer(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	identityRepo identity.Repository,
	userSearchIndex UserTextSearcher,
) *UserDataIndexer {
	return &UserDataIndexer{
		logger:          logging.EnsureLogger(logger).WithName(o11yName),
		tracer:          tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		identityRepo:    identityRepo,
		userSearchIndex: userSearchIndex,
	}
}

func (i *UserDataIndexer) HandleIndexRequest(
	ctx context.Context,
	indexReq *textsearch.IndexRequest,
) error {
	ctx, span := i.tracer.StartSpan(ctx)
	defer span.End()

	if indexReq == nil {
		return observability.PrepareAndLogError(ErrNilIndexRequest, i.logger, span, "handling index requests")
	}

	logger := i.logger.WithValue("index_type_requested", indexReq.IndexType)

	var (
		toBeIndexed       any
		err               error
		markAsIndexedFunc func() error
	)

	switch indexReq.IndexType {
	case IndexTypeUsers:
		var user *identity.User
		user, err = i.identityRepo.GetUser(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting user")
		}

		toBeIndexed = ConvertUserToUserSearchSubset(user)
		markAsIndexedFunc = func() error { return i.identityRepo.MarkUserAsIndexed(ctx, indexReq.RowID) }
	default:
		logger.Info("invalid index type specified, exiting")
		return nil
	}

	if indexReq.Delete {
		if err = i.userSearchIndex.Delete(ctx, indexReq.RowID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "deleting data")
		}

		return nil
	} else {
		if err = i.userSearchIndex.Index(ctx, indexReq.RowID, toBeIndexed); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "indexing data")
		}

		if err = markAsIndexedFunc(); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "marking data as indexed")
		}
	}

	return nil
}
