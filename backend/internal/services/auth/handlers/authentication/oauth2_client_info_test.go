package authentication

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth/fakes"

	"github.com/stretchr/testify/assert"
)

func TestOAuth2ClientInfoImpl_GetID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.GetID()
		assert.Equal(t, client.ID, result)
	})
}

func TestOAuth2ClientInfoImpl_GetSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.GetSecret()
		assert.Equal(t, client.ClientSecret, result)
	})
}

func TestOAuth2ClientInfoImpl_GetDomain(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		domain := "example.com"
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: domain,
		}

		result := impl.GetDomain()
		assert.Equal(t, domain, result)
	})
}

func TestOAuth2ClientInfoImpl_IsPublic(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.IsPublic()
		assert.False(t, result)
	})
}

func TestOAuth2ClientInfoImpl_GetUserID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		client := fakes.BuildFakeOAuth2Client()
		impl := &oauth2ClientInfoImpl{
			client: client,
			domain: "example.com",
		}

		result := impl.GetUserID()
		assert.Empty(t, result)
	})
}
