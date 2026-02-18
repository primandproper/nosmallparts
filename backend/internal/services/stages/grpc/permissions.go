package grpc

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	stagessvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/stages"
)

// StagesMethodPermissions is a named type for Wire dependency injection.
// Stages service methods are unauthenticated (anonymous participants join by name).
// We provide empty permission slices so the auth interceptor does not reject.
// The actual unauthenticated bypass is configured in the auth interceptor's unauthenticatedRoutes.
type StagesMethodPermissions map[string][]authorization.Permission

// ProvideStagesMethodPermissions returns a Wire provider.
func ProvideStagesMethodPermissions() StagesMethodPermissions {
	return StagesMethodPermissions{
		stagessvc.StagesService_JoinStage_FullMethodName:         {},
		stagessvc.StagesService_GetStageState_FullMethodName:     {},
		stagessvc.StagesService_UpdateParticipant_FullMethodName: {},
		stagessvc.StagesService_ToggleRole_FullMethodName:        {},
		stagessvc.StagesService_UpdateStage_FullMethodName:       {},
		stagessvc.StagesService_SendChat_FullMethodName:          {},
		stagessvc.StagesService_LeaveStage_FullMethodName:        {},
		stagessvc.StagesService_StartPlay_FullMethodName:         {},
		stagessvc.StagesService_NextLine_FullMethodName:          {},
		stagessvc.StagesService_ResetPlay_FullMethodName:         {},
	}
}
