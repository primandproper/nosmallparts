package grpc

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/playscripts"
)

// PlayScriptsMethodPermissions is a named type for Wire dependency injection.
type PlayScriptsMethodPermissions map[string][]authorization.Permission

// ProvidePlayScriptsMethodPermissions returns a Wire provider for the play scripts service's method permissions.
func ProvidePlayScriptsMethodPermissions() PlayScriptsMethodPermissions {
	return PlayScriptsMethodPermissions{
		playscriptssvc.PlayScriptsService_CreatePlayScript_FullMethodName: {
			authorization.CreatePlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_GetPlayScript_FullMethodName: {
			authorization.ReadPlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_GetPlayScripts_FullMethodName: {
			authorization.ReadPlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_GetPlayScriptsForAccount_FullMethodName: {
			authorization.ReadPlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_GetPublicPlayScripts_FullMethodName: {
			authorization.ReadPlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_UpdatePlayScript_FullMethodName: {
			authorization.UpdatePlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_ArchivePlayScript_FullMethodName: {
			authorization.ArchivePlayScriptsPermission,
		},
	}
}
