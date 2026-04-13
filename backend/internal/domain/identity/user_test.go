package identity

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/platform/v4/pointer"

	"github.com/stretchr/testify/assert"
)

func TestUser_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		actual := User{
			Username:        "old_username",
			HashedPassword:  "hashed_pass",
			TwoFactorSecret: "two factor secret",
			Birthday:        pointer.To[time.Time](time.Now()),
		}

		exampleInput := User{
			Username:        "new_username",
			HashedPassword:  "updated_hashed_pass",
			TwoFactorSecret: "new fancy secret",
			FirstName:       "first",
			LastName:        "last",
			Birthday:        pointer.To[time.Time](time.Now()),
		}

		actual.Update(&exampleInput)
		assert.Equal(t, exampleInput, actual)
	})
}

func TestUser_IsBanned(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &User{AccountStatus: string(BannedUserAccountStatus)}

		assert.True(t, x.IsBanned())
	})
}

func TestUserRegistrationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserRegistrationInput{
			Username:     t.Name(),
			Password:     t.Name(),
			EmailAddress: "things@stuff.com",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("invalid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserRegistrationInput{
			Username:     "",
			EmailAddress: "",
			Password:     "",
		}

		err := x.ValidateWithContext(ctx)
		assert.Error(t, err)
	})
}

func TestUserDetailsUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserDetailsUpdateRequestInput{
			FirstName:       t.Name(),
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
