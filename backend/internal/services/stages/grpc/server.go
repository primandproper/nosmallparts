// Package grpc serves StagesService: creating a stage and resolving the code
// somebody typed. Everything after joining happens on the stage socket.
package grpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/clientip"
	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"
	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v15/callers"
	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/ratelimiting"
	ratelimitinggrpc "github.com/primandproper/primitives-go/v2/ratelimiting/grpc"

	"github.com/oklog/ulid/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serverName = "stages_grpc_server"

// The code lookup limit: 30 a minute per client address, so a join code
// cannot be guessed by trying them all.
const (
	ExistsPerMinute = 30
	ExistsBurst     = ExistsPerMinute
)

// codeAttempts bounds how many codes a create tries before giving up. With 31
// characters in six places a collision is rare enough that five in a row means
// something is wrong.
const codeAttempts = 5

// Server is StagesService.
type Server struct {
	stagepb.UnimplementedStagesServiceServer

	o11y      observability.Observer
	client    database.Client
	store     stages.Store
	scripts   *scripts.Service
	extractor callers.PrincipalExtractor
}

var _ stagepb.StagesServiceServer = (*Server)(nil)

// NewServer builds the server.
func NewServer(
	client database.Client,
	store stages.Store,
	scriptService *scripts.Service,
	extractor callers.PrincipalExtractor,
	pillars *observability.Pillars,
) (*Server, error) {
	if client == nil || store == nil || scriptService == nil || extractor == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "building the stages gRPC server")
	}

	o11y := observability.NewObserver(serverName, nil, nil)
	if pillars != nil {
		o11y = observability.NewObserver(serverName, pillars.Logger, pillars.TracerProvider)
	}

	return &Server{o11y: o11y, client: client, store: store, scripts: scriptService, extractor: extractor}, nil
}

// RegisterOn puts the service on a gRPC server.
func (s *Server) RegisterOn(srv *grpc.Server) {
	stagepb.RegisterStagesServiceServer(srv, s)
}

// RequireAuthentication declares what each method needs of the caller's
// credential: a user to create a stage, and nothing to look one up.
func RequireAuthentication(b *signingrpc.AuthenticationRequirementsBuilder) *signingrpc.AuthenticationRequirementsBuilder {
	return b.
		Declare(signingrpc.AuthenticationRequired, stagepb.StagesService_CreateStage_FullMethodName).
		Declare(signingrpc.AuthenticationAnonymous, stagepb.StagesService_StageExists_FullMethodName)
}

// Require declares every method public to the authorization enforcer: a stage
// is anybody's to create once they are signed in, which authentication decides.
func Require(b *authzgrpc.RequirementsBuilder) *authzgrpc.RequirementsBuilder {
	return b.
		Public(stagepb.StagesService_CreateStage_FullMethodName).
		Public(stagepb.StagesService_StageExists_FullMethodName)
}

// NewExistsLimiter is the in-memory limiter the code lookup spends from. The
// caller closes it.
func NewExistsLimiter() (*ratelimiting.InMemoryRateLimiter, error) {
	return ratelimiting.NewInMemoryRateLimiter(float64(ExistsPerMinute)/time.Minute.Seconds(), ExistsBurst)
}

// ExistsRateLimit is the interceptor that holds the code lookup to its limit,
// by client address, and exempts every other method. A refusal is
// RESOURCE_EXHAUSTED carrying a RetryInfo with when to come back.
func ExistsRateLimit(limiter ratelimiting.RateLimiter, pillars *observability.Pillars) (grpc.UnaryServerInterceptor, error) {
	opts := []ratelimitinggrpc.Option{}
	if pillars != nil {
		opts = append(opts,
			ratelimitinggrpc.WithLogger(pillars.Logger),
			ratelimitinggrpc.WithTracerProvider(pillars.TracerProvider),
			ratelimitinggrpc.WithMetricsProvider(pillars.MetricsProvider),
		)
	}

	return ratelimitinggrpc.NewUnaryServerInterceptor(limiter, func(ctx context.Context, info *grpc.UnaryServerInfo) (string, error) {
		if info.FullMethod != stagepb.StagesService_StageExists_FullMethodName {
			return "", nil
		}

		return "stage-exists:" + clientip.FromContext(ctx), nil
	}, opts...)
}

// CreateStage implements StagesServiceServer. A preselected play must be one
// the caller could see and that is switched on.
func (s *Server) CreateStage(ctx context.Context, req *stagepb.CreateStageRequest) (*stagepb.CreateStageResponse, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	principal, ok := s.extractor(ctx)
	if !ok || principal == nil {
		return nil, status.Error(codes.Unauthenticated, "sign in to create a stage")
	}

	playID := req.GetPlayScriptId()
	if playID != "" {
		if _, err := s.scripts.Get(ctx, &scripts.Viewer{UserID: principal.UserID()}, playID); err != nil {
			if errors.Is(err, scripts.ErrNotFound) {
				return nil, status.Error(codes.InvalidArgument, "Script unavailable.")
			}

			op.Logger().Error("checking a stage's play", err)

			return nil, status.Error(codes.Internal, "creating a stage failed")
		}
	}

	stage := &stages.Stage{ID: ulid.Make().String(), PlayScriptID: playID}

	for range codeAttempts {
		code := stages.NewCode()

		taken, err := s.store.CodeTaken(ctx, s.client.Reader(), code)
		if err != nil {
			op.Logger().Error("checking a stage code", err)

			return nil, status.Error(codes.Internal, "creating a stage failed")
		}

		if !taken {
			stage.Code = code

			break
		}
	}

	if stage.Code == "" {
		op.Logger().Error("minting a stage code", platformerrors.Newf("%d codes in a row were taken", codeAttempts))

		return nil, status.Error(codes.Unavailable, "no stage code is free right now; try again")
	}

	if err := s.client.WithTransaction(ctx, func(tx database.Tx) error {
		return s.store.CreateStage(ctx, tx, stage)
	}); err != nil {
		op.Logger().Error("creating a stage", err)

		return nil, status.Error(codes.Internal, "creating a stage failed")
	}

	return &stagepb.CreateStageResponse{Id: stage.ID, Code: stage.Code}, nil
}

// StageExists implements StagesServiceServer. The code is matched without
// regard to case.
func (s *Server) StageExists(ctx context.Context, req *stagepb.StageExistsRequest) (*stagepb.StageExistsResponse, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	stage, err := s.store.GetStageByCode(ctx, s.client.Reader(), strings.ToUpper(strings.TrimSpace(req.GetCode())))
	if err != nil {
		if errors.Is(err, stages.ErrStageNotFound) {
			return &stagepb.StageExistsResponse{Exists: false}, nil
		}

		op.Logger().Error("looking up a stage code", err)

		return nil, status.Error(codes.Internal, "looking up a stage failed")
	}

	return &stagepb.StageExistsResponse{Exists: true, Id: stage.ID, Code: stage.Code}, nil
}
