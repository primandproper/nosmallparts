package fakes

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeSessionContextData builds a faked ContextData.
func BuildFakeSessionContextData() *sessions.ContextData {
	return &sessions.ContextData{
		AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{},
		Requester: sessions.RequesterInfo{
			ServicePermissions:       nil,
			AccountStatus:            identity.GoodStandingUserAccountStatus.String(),
			AccountStatusExplanation: "fake",
			UserID:                   BuildFakeID(),
			EmailAddress:             fake.Email(),
			Username:                 buildUniqueString(),
		},
		ActiveAccountID: BuildFakeID(),
	}
}

// BuildFakeChangeActiveAccountInput builds a faked ChangeActiveAccountInput.
func BuildFakeChangeActiveAccountInput() *auth.ChangeActiveAccountInput {
	return &auth.ChangeActiveAccountInput{
		AccountID: fake.UUID(),
	}
}

// BuildFakeUserStatusResponse builds a faked UserStatusResponse.
func BuildFakeUserStatusResponse() *auth.UserStatusResponse {
	return &auth.UserStatusResponse{
		UserID:                   BuildFakeID(),
		AccountStatus:            identity.GoodStandingUserAccountStatus.String(),
		AccountStatusExplanation: "",
		ActiveAccount:            BuildFakeID(),
		UserIsAuthenticated:      true,
	}
}

// BuildFakeTokenResponse builds a faked TokenResponse.
func BuildFakeTokenResponse() *auth.TokenResponse {
	return &auth.TokenResponse{
		UserID:      BuildFakeID(),
		AccountID:   BuildFakeID(),
		AccessToken: fake.UUID(),
	}
}

func BuildFakeUserLoginInput() *auth.UserLoginInput {
	return &auth.UserLoginInput{
		Username:  BuildFakeID(),
		Password:  buildFakePassword(),
		TOTPToken: buildFakeTOTPToken(),
	}
}
