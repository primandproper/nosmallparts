package fakes

import (
	"encoding/base32"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	"github.com/verygoodsoftwarenotvirus/platform/v4/pointer"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUser builds a faked User.
func BuildFakeUser() *identity.User {
	fakeDate := BuildFakeTime()

	return &identity.User{
		ID:                        BuildFakeID(),
		FirstName:                 fake.FirstName(),
		LastName:                  fake.LastName(),
		EmailAddress:              fake.Email(),
		Username:                  fmt.Sprintf("%s_%d_%s", fake.Username(), fake.Uint8(), fake.Username()),
		Birthday:                  pointer.To(BuildFakeTime()),
		AccountStatus:             string(identity.UnverifiedAccountStatus),
		TwoFactorSecret:           base32.StdEncoding.EncodeToString([]byte(fake.Password(false, true, true, false, false, 32))),
		TwoFactorSecretVerifiedAt: &fakeDate,
		ServiceRole:               authorization.ServiceUserRole.String(),
		CreatedAt:                 BuildFakeTime(),
	}
}

// BuildFakeUsersList builds a faked UserList.
func BuildFakeUsersList() *filtering.QueryFilteredResult[identity.User] {
	var examples []*identity.User
	for i := 0; i < exampleQuantity; i++ {
		examples = append(examples, BuildFakeUser())
	}

	return &filtering.QueryFilteredResult[identity.User]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeUserCreationInput builds a faked UserRegistrationInput.
func BuildFakeUserCreationInput() *identity.UserRegistrationInput {
	exampleUser := BuildFakeUser()

	return &identity.UserRegistrationInput{
		Username:     exampleUser.Username,
		EmailAddress: fake.Email(),
		FirstName:    exampleUser.FirstName,
		LastName:     exampleUser.LastName,
		Password:     buildFakePassword(),
		Birthday:     exampleUser.Birthday,
	}
}

// BuildFakeUserRegistrationInput builds a faked UserRegistrationInput.
func BuildFakeUserRegistrationInput() *identity.UserRegistrationInput {
	user := BuildFakeUser()
	return &identity.UserRegistrationInput{
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		EmailAddress: user.EmailAddress,
		Password:     buildFakePassword(),
		Birthday:     user.Birthday,
	}
}

// BuildFakeUserRegistrationInputFromUser builds a faked UserRegistrationInput.
func BuildFakeUserRegistrationInputFromUser(user *identity.User) *identity.UserRegistrationInput {
	return &identity.UserRegistrationInput{
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		EmailAddress: user.EmailAddress,
		Password:     buildFakePassword(),
		Birthday:     user.Birthday,
	}
}

// BuildFakeUserRegistrationInputWithInviteFromUser builds a faked UserRegistrationInput.
func BuildFakeUserRegistrationInputWithInviteFromUser(user *identity.User) *identity.UserRegistrationInput {
	return &identity.UserRegistrationInput{
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		EmailAddress:    user.EmailAddress,
		Password:        buildFakePassword(),
		Birthday:        user.Birthday,
		InvitationToken: fake.UUID(),
		InvitationID:    BuildFakeID(),
	}
}

// BuildFakeUserCreationResponse builds a faked UserAccountStatusUpdateInput.
func BuildFakeUserCreationResponse() *identity.UserCreationResponse {
	user := BuildFakeUser()
	return &identity.UserCreationResponse{
		CreatedAt:       user.CreatedAt,
		Birthday:        user.Birthday,
		Username:        user.Username,
		EmailAddress:    user.EmailAddress,
		TwoFactorQRCode: fake.UUID(),
		CreatedUserID:   user.ID,
		AccountStatus:   user.AccountStatus,
		TwoFactorSecret: user.TwoFactorSecret,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
	}
}

// BuildFakeAvatarUpdateInput builds a faked AvatarUpdateInput.
func BuildFakeAvatarUpdateInput() *identity.AvatarUpdateInput {
	return &identity.AvatarUpdateInput{
		Base64EncodedData: buildUniqueString(),
	}
}

func BuildFakeUserDetailsUpdateRequestInput() *identity.UserDetailsUpdateRequestInput {
	return &identity.UserDetailsUpdateRequestInput{
		FirstName:       buildUniqueString(),
		LastName:        buildUniqueString(),
		Birthday:        BuildFakeTime(),
		CurrentPassword: fake.Password(true, true, true, false, false, 32),
		TOTPToken:       "123456",
	}
}

func BuildFakeUserDetailsDatabaseUpdateInput() *identity.UserDetailsDatabaseUpdateInput {
	return &identity.UserDetailsDatabaseUpdateInput{
		FirstName: buildUniqueString(),
		LastName:  buildUniqueString(),
		Birthday:  BuildFakeTime(),
	}
}

// BuildFakeUserPermissionModificationInput builds a faked ModifyUserPermissionsInput.
func BuildFakeUserPermissionModificationInput() *identity.ModifyUserPermissionsInput {
	return &identity.ModifyUserPermissionsInput{
		Reason:  fake.Sentence(10),
		NewRole: authorization.AccountMemberRole.String(),
	}
}
