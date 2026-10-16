// Package grpc serves the script library over gRPC as PlayScriptsService.
//
// Each method resolves who is calling, converts, calls the scripts service
// once, and converts back. The rules are the service's; what this package owns
// is the caller: their user ID, and whether their address is on the admin
// list.
package grpc

import (
	"context"
	"errors"

	"github.com/primandproper/nosmallparts/backend/internal/admins"
	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"
	"github.com/primandproper/nosmallparts/backend/internal/proto/scriptspb"

	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/platform-go/v15/identity"
	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const serverName = "scripts_grpc_server"

// Server is PlayScriptsService.
type Server struct {
	scriptspb.UnimplementedPlayScriptsServiceServer

	o11y      observability.Observer
	scripts   *scripts.Service
	admins    *admins.List
	extractor callers.PrincipalExtractor
}

var _ scriptspb.PlayScriptsServiceServer = (*Server)(nil)

// NewServer builds the server. The extractor is the one the authentication
// interceptor fills, read once per call.
func NewServer(
	service *scripts.Service,
	adminList *admins.List,
	extractor callers.PrincipalExtractor,
	pillars *observability.Pillars,
) (*Server, error) {
	if service == nil || adminList == nil || extractor == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "building the scripts gRPC server")
	}

	o11y := observability.NewObserver(serverName, nil, nil)
	if pillars != nil {
		o11y = observability.NewObserver(serverName, pillars.Logger, pillars.TracerProvider)
	}

	return &Server{o11y: o11y, scripts: service, admins: adminList, extractor: extractor}, nil
}

// RegisterOn puts the service on a gRPC server.
func (s *Server) RegisterOn(srv *grpc.Server) {
	scriptspb.RegisterPlayScriptsServiceServer(srv, s)
}

// RequireAuthentication declares what each method needs of the caller's
// credential. The reads answer anybody, and differently for somebody signed
// in; every write needs a user.
func RequireAuthentication(b *signingrpc.AuthenticationRequirementsBuilder) *signingrpc.AuthenticationRequirementsBuilder {
	return b.
		Declare(signingrpc.AuthenticationOptional,
			scriptspb.PlayScriptsService_ListPlayScripts_FullMethodName,
			scriptspb.PlayScriptsService_GetPlayScript_FullMethodName,
			scriptspb.PlayScriptsService_ListTags_FullMethodName,
		).
		Declare(signingrpc.AuthenticationRequired,
			scriptspb.PlayScriptsService_CreatePlayScript_FullMethodName,
			scriptspb.PlayScriptsService_UpdatePlayScript_FullMethodName,
			scriptspb.PlayScriptsService_DeletePlayScript_FullMethodName,
			scriptspb.PlayScriptsService_SetPlayScriptEnabled_FullMethodName,
			scriptspb.PlayScriptsService_RatePlayScript_FullMethodName,
		)
}

// Require declares every method public to the authorization enforcer. Who may
// do what here is decided by ownership and the admin list, which the scripts
// service checks per row, and not by a permission a role carries.
func Require(b *authzgrpc.RequirementsBuilder) *authzgrpc.RequirementsBuilder {
	for _, method := range scriptspb.PlayScriptsService_ServiceDesc.Methods {
		b.Public("/" + scriptspb.PlayScriptsService_ServiceDesc.ServiceName + "/" + method.MethodName)
	}

	return b
}

// ListPlayScripts implements PlayScriptsServiceServer.
func (s *Server) ListPlayScripts(ctx context.Context, req *scriptspb.ListPlayScriptsRequest) (*scriptspb.ListPlayScriptsResponse, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	plays, err := s.scripts.List(ctx, s.viewer(ctx), req.GetSearch())
	if err != nil {
		return nil, s.status(op, err, "listing scripts")
	}

	out := &scriptspb.ListPlayScriptsResponse{Plays: make([]*scriptspb.Play, 0, len(plays))}
	for _, play := range plays {
		out.Plays = append(out.Plays, convertPlay(play))
	}

	return out, nil
}

// GetPlayScript implements PlayScriptsServiceServer.
func (s *Server) GetPlayScript(ctx context.Context, req *scriptspb.GetPlayScriptRequest) (*scriptspb.GetPlayScriptResponse, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", req.GetId()))
	defer op.End()

	play, err := s.scripts.Get(ctx, s.viewer(ctx), req.GetId())
	if err != nil {
		return nil, s.status(op, err, "reading a script")
	}

	return &scriptspb.GetPlayScriptResponse{Play: convertPlay(play)}, nil
}

// CreatePlayScript implements PlayScriptsServiceServer.
func (s *Server) CreatePlayScript(ctx context.Context, req *scriptspb.CreatePlayScriptRequest) (*scriptspb.CreatePlayScriptResponse, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	play, err := s.scripts.Create(ctx, s.viewer(ctx), convertPlayInput(req.GetPlay()))
	if err != nil {
		return nil, s.status(op, err, "creating a script")
	}

	return &scriptspb.CreatePlayScriptResponse{Play: convertPlay(play)}, nil
}

// UpdatePlayScript implements PlayScriptsServiceServer.
func (s *Server) UpdatePlayScript(ctx context.Context, req *scriptspb.UpdatePlayScriptRequest) (*scriptspb.UpdatePlayScriptResponse, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", req.GetId()))
	defer op.End()

	play, err := s.scripts.Update(ctx, s.viewer(ctx), req.GetId(), convertPlayInput(req.GetPlay()))
	if err != nil {
		return nil, s.status(op, err, "updating a script")
	}

	return &scriptspb.UpdatePlayScriptResponse{Play: convertPlay(play)}, nil
}

// DeletePlayScript implements PlayScriptsServiceServer.
func (s *Server) DeletePlayScript(ctx context.Context, req *scriptspb.DeletePlayScriptRequest) (*scriptspb.DeletePlayScriptResponse, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", req.GetId()))
	defer op.End()

	if err := s.scripts.Delete(ctx, s.viewer(ctx), req.GetId()); err != nil {
		return nil, s.status(op, err, "deleting a script")
	}

	return &scriptspb.DeletePlayScriptResponse{}, nil
}

// SetPlayScriptEnabled implements PlayScriptsServiceServer.
func (s *Server) SetPlayScriptEnabled(ctx context.Context, req *scriptspb.SetPlayScriptEnabledRequest) (*scriptspb.SetPlayScriptEnabledResponse, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", req.GetId()))
	defer op.End()

	if err := s.scripts.SetEnabled(ctx, s.viewer(ctx), req.GetId(), req.GetEnabled()); err != nil {
		return nil, s.status(op, err, "switching a script")
	}

	return &scriptspb.SetPlayScriptEnabledResponse{Enabled: req.GetEnabled()}, nil
}

// RatePlayScript implements PlayScriptsServiceServer.
func (s *Server) RatePlayScript(ctx context.Context, req *scriptspb.RatePlayScriptRequest) (*scriptspb.RatePlayScriptResponse, error) {
	ctx, op := s.o11y.Begin(ctx, observability.WithValue("play_script.id", req.GetId()))
	defer op.End()

	stats, err := s.scripts.Rate(ctx, s.viewer(ctx), req.GetId(), req.GetRating())
	if err != nil {
		return nil, s.status(op, err, "rating a script")
	}

	return &scriptspb.RatePlayScriptResponse{RatingStats: convertRatingStats(stats)}, nil
}

// ListTags implements PlayScriptsServiceServer.
func (s *Server) ListTags(ctx context.Context, _ *scriptspb.ListTagsRequest) (*scriptspb.ListTagsResponse, error) {
	ctx, op := s.o11y.Begin(ctx)
	defer op.End()

	tags, err := s.scripts.Tags(ctx)
	if err != nil {
		return nil, s.status(op, err, "listing tags")
	}

	out := &scriptspb.ListTagsResponse{Tags: make([]*scriptspb.ScriptTag, 0, len(tags))}
	for _, tag := range tags {
		out.Tags = append(out.Tags, convertTag(tag))
	}

	return out, nil
}

// identified is a principal that carries its directory entry, which a
// sign-in token's caller does.
type identified interface {
	Identity() *identity.Principal
}

// viewer is who is calling: anonymous, a user, or a user whose address is on
// the admin list.
func (s *Server) viewer(ctx context.Context) *scripts.Viewer {
	principal, ok := s.extractor(ctx)
	if !ok || principal == nil {
		return &scripts.Viewer{}
	}

	viewer := &scripts.Viewer{UserID: principal.UserID()}

	if withIdentity, isIdentified := principal.(identified); isIdentified {
		if entry := withIdentity.Identity(); entry != nil && entry.User != nil {
			viewer.Admin = s.admins.Includes(entry.User.EmailAddress)
		}
	}

	return viewer
}

// status turns a refusal into the code its HTTP status mapped to before, and
// logs anything else as the internal failure it is.
func (s *Server) status(op observability.Operation, err error, description string) error {
	refusals := []struct {
		err  error
		code codes.Code
	}{
		{err: scripts.ErrNotFound, code: codes.NotFound},
		{err: scripts.ErrNotSignedIn, code: codes.Unauthenticated},
		{err: scripts.ErrNotAdmin, code: codes.PermissionDenied},
		{err: scripts.ErrSelfRating, code: codes.PermissionDenied},
		{err: scripts.ErrRatingOutOfRange, code: codes.InvalidArgument},
		{err: scripts.ErrInvalidInput, code: codes.InvalidArgument},
	}

	for _, refusal := range refusals {
		if errors.Is(err, refusal.err) {
			return status.Error(refusal.code, err.Error())
		}
	}

	op.Logger().Error(description, err)

	return status.Error(codes.Internal, description+" failed")
}
