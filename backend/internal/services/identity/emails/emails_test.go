package emails

import (
	"testing"
	"time"

	authfakes "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth/fakes"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"

	"github.com/stretchr/testify/assert"
)

func TestBuildGeneratedPasswordResetTokenEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = pointer.To(time.Now())
		token := authfakes.BuildFakePasswordResetToken()

		actual, err := BuildGeneratedPasswordResetTokenEmail(user, token, &email.EnvironmentConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Contains(t, actual.HTMLContent, logoURL)
	})
}

func TestBuildInviteMemberEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		invitation := fakes.BuildFakeAccountInvitation()

		actual, err := BuildInviteMemberEmail(user, invitation, &email.EnvironmentConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func TestBuildPasswordResetTokenRedeemedEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = pointer.To(time.Now())

		actual, err := BuildPasswordResetTokenRedeemedEmail(user, &email.EnvironmentConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func TestBuildUsernameReminderEmail(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := fakes.BuildFakeUser()
		user.EmailAddressVerifiedAt = pointer.To(time.Now())

		actual, err := BuildUsernameReminderEmail(user, &email.EnvironmentConfig{})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}
