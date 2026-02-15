package grpc

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"
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
		playscriptssvc.PlayScriptsService_GetPlayScriptsWithTag_FullMethodName: {
			authorization.ReadPlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_CreateScriptTag_FullMethodName: {
			authorization.CreateScriptTagsPermission,
		},
		playscriptssvc.PlayScriptsService_ListScriptTags_FullMethodName: {
			authorization.ReadPlayScriptsPermission,
		},
		playscriptssvc.PlayScriptsService_AddTagToPlayScript_FullMethodName: {
			authorization.AddTagToPlayScriptPermission,
		},
		playscriptssvc.PlayScriptsService_RemoveTagFromPlayScript_FullMethodName: {
			authorization.RemoveTagFromPlayScriptPermission,
		},
		playscriptssvc.PlayScriptsService_UpdatePlayScriptLine_FullMethodName: {
			authorization.UpdatePlayScriptsPermission,
		},
	}
}
