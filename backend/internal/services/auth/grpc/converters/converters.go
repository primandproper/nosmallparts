package converters

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	authsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
)

func ConvertGRPCUpdatePasswordRequestToPasswordUpdateInput(request *authsvc.UpdatePasswordRequest) *auth.PasswordUpdateInput {
	return &auth.PasswordUpdateInput{
		NewPassword:     request.NewPassword,
		CurrentPassword: request.CurrentPassword,
		TOTPToken:       request.TotpToken,
	}
}

func ConvertGRPCVerifyTOTPSecretRequestToTOTPSecretVerificationInput(request *authsvc.VerifyTOTPSecretRequest) *auth.TOTPSecretVerificationInput {
	return &auth.TOTPSecretVerificationInput{
		TOTPToken: request.TotpToken,
		UserID:    request.UserId,
	}
}

func ConvertGRPCVerifyEmailAddressRequestToEmailAddressVerificationRequestInput(request *authsvc.VerifyEmailAddressRequest) *auth.EmailAddressVerificationRequestInput {
	return &auth.EmailAddressVerificationRequestInput{
		Token: request.Token,
	}
}

func ConvertGRPCRequestUsernameReminderRequestToUsernameReminderRequestInput(request *authsvc.RequestUsernameReminderRequest) *auth.UsernameReminderRequestInput {
	return &auth.UsernameReminderRequestInput{
		EmailAddress: request.EmailAddress,
	}
}

func ConvertGRPCRequestPasswordResetTokenRequestToPasswordResetTokenCreationRequestInput(request *authsvc.RequestPasswordResetTokenRequest) *auth.PasswordResetTokenCreationRequestInput {
	return &auth.PasswordResetTokenCreationRequestInput{EmailAddress: request.EmailAddress}
}

func ConvertGRPCRefreshTOTPSecretRequestToTOTPSecretRefreshInput(request *authsvc.RefreshTOTPSecretRequest) *auth.TOTPSecretRefreshInput {
	return &auth.TOTPSecretRefreshInput{
		CurrentPassword: request.CurrentPassword,
		TOTPToken:       request.TotpToken,
	}
}

func ConvertTOTPSecretRefreshResponseToGRPCTOTPSecretRefreshResponse(input *auth.TOTPSecretRefreshResponse) *authsvc.TOTPSecretRefreshResponse {
	return &authsvc.TOTPSecretRefreshResponse{
		TwoFactorQrCode: input.TwoFactorQRCode,
		TwoFactorSecret: input.TwoFactorSecret,
	}
}

func ConvertGRPCRedeemPasswordResetTokenRequestToPasswordResetTokenRedemptionRequestInput(request *authsvc.RedeemPasswordResetTokenRequest) *auth.PasswordResetTokenRedemptionRequestInput {
	return &auth.PasswordResetTokenRedemptionRequestInput{
		Token:       request.Token,
		NewPassword: request.NewPassword,
	}
}

func ConvertGRPCCheckPermissionsRequestToUserPermissionsRequestInput(request *authsvc.UserPermissionsRequestInput) *auth.UserPermissionsRequestInput {
	return &auth.UserPermissionsRequestInput{
		Permissions: request.Permissions,
	}
}

func ConvertGRPCUserLoginInputToUserLoginInput(request *authsvc.UserLoginInput) *auth.UserLoginInput {
	return &auth.UserLoginInput{
		Username:  request.Username,
		Password:  request.Password,
		TOTPToken: request.TotpToken,
	}
}

func ConvertTokenResponseToGRPCTokenResponse(input *auth.TokenResponse) *authsvc.TokenResponse {
	return &authsvc.TokenResponse{
		UserId:       input.UserID,
		AccountId:    input.AccountID,
		AccessToken:  input.AccessToken,
		RefreshToken: input.RefreshToken,
		ExpiresUtc:   grpcconverters.ConvertTimeToPBTimestamp(input.ExpiresUTC),
	}
}
