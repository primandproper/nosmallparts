package converters

import (
	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
)

// ConvertUserToUserCreationResponse builds a UserCreationResponse from a user.
func ConvertUserToUserCreationResponse(user *types.User) *types.UserCreationResponse {
	return &types.UserCreationResponse{
		CreatedUserID: user.ID,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		CreatedAt:     user.CreatedAt,
	}
}

// ConvertUserToUserDatabaseCreationInput builds a UserDatabaseCreationInput from a User.
func ConvertUserToUserDatabaseCreationInput(user *types.User) *types.UserDatabaseCreationInput {
	return &types.UserDatabaseCreationInput{
		ID:              user.ID,
		EmailAddress:    user.EmailAddress,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		HashedPassword:  user.HashedPassword,
		TwoFactorSecret: user.TwoFactorSecret,
		Birthday:        user.Birthday,
	}
}

func ConvertUserDetailsUpdateRequestInputToUserDetailsUpdateInput(x *types.UserDetailsUpdateRequestInput) *types.UserDetailsDatabaseUpdateInput {
	return &types.UserDetailsDatabaseUpdateInput{
		FirstName: x.FirstName,
		LastName:  x.LastName,
		Birthday:  x.Birthday,
	}
}
