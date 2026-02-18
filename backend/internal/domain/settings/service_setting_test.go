package settings

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"

	"github.com/stretchr/testify/assert"
)

func TestServiceSettingCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingCreationRequestInput{
			Name:        t.Name(),
			Description: t.Name(),
			Type:        t.Name(),
			Enumeration: []string{
				t.Name(),
			},
			DefaultValue: pointer.To(t.Name()),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})

	T.Run("with invalid struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingCreationRequestInput{}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})

	T.Run("with invalid default value", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingCreationRequestInput{
			Name:        t.Name(),
			Description: t.Name(),
			Type:        t.Name(),
			Enumeration: []string{
				t.Name(),
			},
			DefaultValue: pointer.To("whatever"),
		}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})
}

func TestServiceSettingDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingDatabaseCreationInput{
			ID:          t.Name(),
			Name:        t.Name(),
			Description: t.Name(),
			Type:        t.Name(),
			Enumeration: []string{
				t.Name(),
			},
			DefaultValue: pointer.To(t.Name()),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})
}
