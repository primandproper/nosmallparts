// Package grpc serves LiveKitService: the token a stage's participant joins
// its video room with.
package grpc

import (
	"context"
	"errors"

	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/guests"
	"github.com/primandproper/nosmallparts/backend/internal/livekit"
	"github.com/primandproper/nosmallparts/backend/internal/proto/livekitpb"

	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v15/callers"
	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serverName = "livekit_grpc_server"

// Server is LiveKitService.
type Server struct {
	livekitpb.UnimplementedLiveKitServiceServer

	o11y      observability.Observer
	client    database.Client
	store     stages.Store
	minter    *livekit.Minter
	extractor callers.PrincipalExtractor
}

var _ livekitpb.LiveKitServiceServer = (*Server)(nil)

// NewServer builds the server.
func NewServer(
	client database.Client,
	store stages.Store,
	minter *livekit.Minter,
	extractor callers.PrincipalExtractor,
	pillars *observability.Pillars,
) (*Server, error) {
	if client == nil || store == nil || minter == nil || extractor == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "building the livekit gRPC server")
	}

	o11y := observability.NewObserver(serverName, nil, nil)
	if pillars != nil {
		o11y = observability.NewObserver(serverName, pillars.Logger, pillars.TracerProvider)
	}

	return &Server{o11y: o11y, client: client, store: store, minter: minter, extractor: extractor}, nil
}

// RegisterOn puts the service on a gRPC server.
func (s *Server) RegisterOn(srv *grpc.Server) {
	livekitpb.RegisterLiveKitServiceServer(srv, s)
}

// RequireAuthentication declares MintToken optional: a guest on a stage joins
// its video too, and is told apart by their cookie.
func RequireAuthentication(b *signingrpc.AuthenticationRequirementsBuilder) *signingrpc.AuthenticationRequirementsBuilder {
	return b.Declare(signingrpc.AuthenticationOptional, livekitpb.LiveKitService_MintToken_FullMethodName)
}

// Require declares MintToken public to the authorization enforcer; who may
// have a token is whoever the participant is, which the server checks.
func Require(b *authzgrpc.RequirementsBuilder) *authzgrpc.RequirementsBuilder {
	return b.Public(livekitpb.LiveKitService_MintToken_FullMethodName)
}

// MintToken implements LiveKitServiceServer. The participant must be on the
// stage now and must be the caller, so a stage and participant ID somebody
// overheard admit nobody else to the room. A performer may publish; the
// audience only watches.
func (s *Server) MintToken(ctx context.Context, req *livekitpb.MintTokenRequest) (*livekitpb.MintTokenResponse, error) {
	ctx, op := s.o11y.Begin(ctx,
		observability.WithValue("stage.id", req.GetStageId()),
		observability.WithValue("participant.id", req.GetParticipantId()),
	)
	defer op.End()

	participant, err := s.store.GetParticipant(ctx, s.client.Reader(), req.GetParticipantId())
	if errors.Is(err, stages.ErrParticipantNotFound) || (err == nil && (participant.StageID != req.GetStageId() || participant.LeftAt != nil)) {
		return nil, status.Error(codes.NotFound, "that participant is not on that stage")
	} else if err != nil {
		op.Logger().Error("reading a participant", err)

		return nil, status.Error(codes.Internal, "minting a video token failed")
	}

	if !participant.Identity.Matches(s.caller(ctx)) {
		return nil, status.Error(codes.PermissionDenied, "that participant is not you")
	}

	token, err := s.minter.Mint(&livekit.Grant{
		Room:     participant.StageID,
		Identity: participant.ID,
		Name:     participant.DisplayName,
		Publish:  !participant.IsAudience,
	})
	if errors.Is(err, livekit.ErrNotConfigured) {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	} else if err != nil {
		op.Logger().Error("minting a video token", err)

		return nil, status.Error(codes.Internal, "minting a video token failed")
	}

	return &livekitpb.MintTokenResponse{Token: token, Url: s.minter.URL(), Room: participant.StageID}, nil
}

// caller is who is asking: the signed-in user, or the guest their cookie
// names, or nobody.
func (s *Server) caller(ctx context.Context) stages.Identity {
	if principal, ok := s.extractor(ctx); ok && principal != nil {
		return stages.Identity{UserID: principal.UserID()}
	}

	if guestID, ok := guests.FromContext(ctx); ok {
		return stages.Identity{GuestID: guestID}
	}

	return stages.Identity{}
}
