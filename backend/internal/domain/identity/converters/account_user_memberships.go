package converters

import (
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/platform/v4/identifiers"
)

// ConvertAccountUserMembershipToAccountUserMembershipDatabaseCreationInput builds a faked AccountUserMembershipCreationRequestInput.
func ConvertAccountUserMembershipToAccountUserMembershipDatabaseCreationInput(account *types.AccountUserMembership) *types.AccountUserMembershipDatabaseCreationInput {
	return &types.AccountUserMembershipDatabaseCreationInput{
		ID:          identifiers.New(),
		Reason:      "",
		UserID:      account.BelongsToUser,
		AccountID:   account.ID,
		AccountRole: account.AccountRole,
	}
}
