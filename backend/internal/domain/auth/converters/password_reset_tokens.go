package converters

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
)

// ConvertPasswordResetTokenToPasswordResetTokenDatabaseCreationInput builds a PasswordResetTokenDatabaseCreationInput from a PasswordResetToken.
func ConvertPasswordResetTokenToPasswordResetTokenDatabaseCreationInput(input *auth.PasswordResetToken) *auth.PasswordResetTokenDatabaseCreationInput {
	return &auth.PasswordResetTokenDatabaseCreationInput{
		ID:            input.ID,
		Token:         input.Token,
		BelongsToUser: input.BelongsToUser,
		ExpiresAt:     input.ExpiresAt,
	}
}
