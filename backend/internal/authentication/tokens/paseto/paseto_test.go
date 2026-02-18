package paseto

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/mocks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/reflection"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/testutils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	exampleSigningKey = testutils.Example32ByteKey
	ed25519SigningKey = testutils.Example64ByteKey
	exampleToken      = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJUZXN0X2p3dFNpZ25lcl9Jc3N1ZUpXVC9zdGFuZGFyZCIsImV4cCI6MTcyNzU3MDU0OCwiaWF0IjoxNzI3NTY5OTQ4LCJpc3MiOiJkaW5uZXJkb25lYmV0dGVyIiwianRpIjoiY3JzYTA3NnRnM3FkdG1jY3E5MTAiLCJuYmYiOjE3Mjc1Njk4ODgsInN1YiI6ImNyc2EwNzZ0ZzNxZHRtY2NxOTBnIn0.tMASrQBoYAq4n1iwOElLqUQsYOARX5T1qxo8RKhvaAg"
	exampleExpiry     = time.Minute * 10
)

func Test_signer_IssueToken(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s, err := NewPASETOSigner(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), t.Name(), []byte(exampleSigningKey))
		require.NoError(t, err)

		ctx := t.Context()
		user := mocks.NewMockUser()
		user.On(reflection.GetMethodName(user.GetID)).Return("user_id").Times(2)

		actual, err := s.IssueToken(ctx, user, exampleExpiry)
		assert.NoError(t, err)

		parsed, err := s.ParseUserIDFromToken(ctx, actual)
		assert.NoError(t, err)
		assert.Equal(t, parsed, user.GetID())

		mock.AssertExpectationsForObjects(t, user)
	})
}

func Test_signer_ParseUserIDFromToken(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s, err := NewPASETOSigner(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), t.Name(), []byte(exampleSigningKey))
		require.NoError(t, err)

		ctx := t.Context()
		user := mocks.NewMockUser()
		user.On(reflection.GetMethodName(user.GetID)).Return("user_id").Times(2)

		issuedToken, err := s.IssueToken(ctx, user, exampleExpiry)
		assert.NoError(t, err)

		actual, err := s.ParseUserIDFromToken(ctx, issuedToken)
		assert.NoError(t, err)
		assert.Equal(t, actual, user.GetID())

		mock.AssertExpectationsForObjects(t, user)
	})

	T.Run("with invalid algo", func(t *testing.T) {
		t.Parallel()

		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.MapClaims{})

		cryptoSigner := ed25519.PrivateKey(ed25519SigningKey)
		tokenString, err := token.SignedString(cryptoSigner)
		require.NoError(t, err)

		ctx := t.Context()

		s, err := NewPASETOSigner(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), t.Name(), []byte(exampleSigningKey))
		require.NoError(t, err)

		actual, err := s.ParseUserIDFromToken(ctx, tokenString)
		assert.Error(t, err)
		assert.Empty(t, actual)
	})

	T.Run("with invalid key", func(t *testing.T) {
		t.Parallel()

		s, err := NewPASETOSigner(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), t.Name(), []byte(exampleSigningKey))
		require.NoError(t, err)

		s.(*signer).signingKey = nil

		ctx := t.Context()

		actual, err := s.ParseUserIDFromToken(ctx, exampleToken)
		assert.Error(t, err)
		assert.Empty(t, actual)
	})
}
