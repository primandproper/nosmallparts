package settings

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestServiceSettingConfiguration_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ServiceSettingConfiguration{}
		input := &ServiceSettingConfigurationUpdateRequestInput{}

		assert.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestServiceSettingConfigurationCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationCreationRequestInput{
			Value:            t.Name(),
			ServiceSettingID: t.Name(),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})

	T.Run("with invalid struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationCreationRequestInput{}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})
}

func TestServiceSettingConfigurationDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationDatabaseCreationInput{
			ID:            t.Name(),
			Value:         t.Name(),
			BelongsToUser: t.Name(),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})

	T.Run("with invalid struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationDatabaseCreationInput{}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})
}

func TestServiceSettingConfigurationUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationUpdateRequestInput{
			Value:            pointer.To(t.Name()),
			ServiceSettingID: pointer.To(t.Name()),
		}

		actual := x.ValidateWithContext(ctx)

		assert.NoError(t, actual)
	})

	T.Run("with invalid struct", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ServiceSettingConfigurationUpdateRequestInput{}

		actual := x.ValidateWithContext(ctx)

		assert.Error(t, actual)
	})
}
