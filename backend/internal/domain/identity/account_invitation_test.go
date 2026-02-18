package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountInvitationCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &AccountInvitationCreationRequestInput{
			ToName: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestAccountInvitationUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &AccountInvitationUpdateRequestInput{
			Token: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
