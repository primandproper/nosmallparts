package grpc

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
)

// AuthMethodPermissions is a named type for Wire dependency injection.
type AuthMethodPermissions map[string][]authorization.Permission

var noPerms = []authorization.Permission{}

// ProvideMethodPermissions returns a Wire provider for the auth service's method permissions.
func ProvideMethodPermissions() AuthMethodPermissions {
	return AuthMethodPermissions{
		// Methods that don't require specific permissions (authenticated user only)
		authsvc.AuthService_CheckPermissions_FullMethodName:              noPerms,
		authsvc.AuthService_GetAuthStatus_FullMethodName:                 noPerms,
		authsvc.AuthService_GetActiveAccount_FullMethodName:              noPerms,
		authsvc.AuthService_UpdatePassword_FullMethodName:                noPerms,
		authsvc.AuthService_RefreshTOTPSecret_FullMethodName:             noPerms,
		authsvc.AuthService_VerifyTOTPSecret_FullMethodName:              noPerms,
		authsvc.AuthService_RequestPasswordResetToken_FullMethodName:     noPerms,
		authsvc.AuthService_RedeemPasswordResetToken_FullMethodName:      noPerms,
		authsvc.AuthService_ExchangeToken_FullMethodName:                 noPerms,
		authsvc.AuthService_AdminLoginForToken_FullMethodName:            noPerms,
		authsvc.AuthService_GetSelf_FullMethodName:                       noPerms,
		authsvc.AuthService_LoginForToken_FullMethodName:                 noPerms,
		authsvc.AuthService_RequestEmailVerificationEmail_FullMethodName: noPerms,
		authsvc.AuthService_RequestUsernameReminder_FullMethodName:       noPerms,
		authsvc.AuthService_VerifyEmailAddress_FullMethodName:            noPerms,
	}
}
