package webhooks

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebhookCreationInput_Validate(T *testing.T) {
	T.Parallel()

	buildValidWebhookCreationInput := func() *WebhookCreationRequestInput {
		return &WebhookCreationRequestInput{
			Name:        "whatever",
			ContentType: "application/xml",
			URL:         "https://blah.verygoodsoftwarenotvirus.ru",
			Method:      http.MethodPatch,
			Events:      []string{"more_things"},
		}
	}

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, buildValidWebhookCreationInput().ValidateWithContext(t.Context()))
	})

	T.Run("bad name", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Name = ""

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("bad url", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		// much as we'd like to use testutils.InvalidRawURL here, it causes a cyclical import :'(
		exampleInput.URL = fmt.Sprintf(`%s://verygoodsoftwarenotvirus.ru`, string(byte(2<<6-1)))

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("bad method", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Method = "balogna"

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("bad content type", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.ContentType = "application/balogna"

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})

	T.Run("empty events", func(t *testing.T) {
		t.Parallel()
		exampleInput := buildValidWebhookCreationInput()
		exampleInput.Events = []string{}

		assert.Error(t, exampleInput.ValidateWithContext(t.Context()))
	})
}

func TestWebhookCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		name := t.Name()
		ctx := t.Context()
		x := &WebhookCreationRequestInput{
			Name:        name,
			ContentType: "application/json",
			URL:         "https://pkg.go.dev",
			Method:      http.MethodPatch,
			Events:      []string{name},
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestWebhookDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		name := t.Name()
		ctx := t.Context()
		x := &WebhookDatabaseCreationInput{
			ID:          name,
			Name:        name,
			ContentType: "application/json",
			URL:         "https://pkg.go.dev",
			Method:      http.MethodPatch,
			Events: []*WebhookTriggerEventDatabaseCreationInput{
				{},
			},
			BelongsToAccount: name,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
