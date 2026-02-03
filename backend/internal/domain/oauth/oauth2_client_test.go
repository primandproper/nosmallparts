package oauth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2ClientCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &OAuth2ClientCreationRequestInput{
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
