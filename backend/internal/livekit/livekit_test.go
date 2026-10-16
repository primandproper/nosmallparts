package livekit

import (
	"testing"

	"github.com/livekit/protocol/auth"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

func testConfig() *Config {
	return &Config{APIKey: "test-key", APISecret: "test-secret-that-is-long-enough-000", URL: "wss://video.example"}
}

// grantsOf verifies a token against the test secret and answers its video
// grant.
func grantsOf(t *testing.T, token string) *auth.ClaimGrants {
	t.Helper()

	verifier, err := auth.ParseAPIToken(token)
	must.NoError(t, err)

	_, grants, err := verifier.Verify(testConfig().APISecret)
	must.NoError(t, err)

	return grants
}

func TestMinter_Mint(T *testing.T) {
	T.Parallel()

	T.Run("a performer may publish and subscribe in the stage's room", func(t *testing.T) {
		t.Parallel()

		token, err := NewMinter(testConfig()).Mint(&Grant{Room: "stage-1", Identity: "participant-1", Name: "Alice", Publish: true})
		must.NoError(t, err)

		grants := grantsOf(t, token)
		test.EqOp(t, "participant-1", grants.Identity)
		test.EqOp(t, "Alice", grants.Name)
		must.NotNil(t, grants.Video)
		test.True(t, grants.Video.RoomJoin)
		test.EqOp(t, "stage-1", grants.Video.Room)
		test.True(t, grants.Video.GetCanPublish())
		test.True(t, grants.Video.GetCanPublishData())
		test.True(t, grants.Video.GetCanSubscribe())
	})

	T.Run("the audience only subscribes", func(t *testing.T) {
		t.Parallel()

		token, err := NewMinter(testConfig()).Mint(&Grant{Room: "stage-1", Identity: "participant-2", Name: "Watcher"})
		must.NoError(t, err)

		grants := grantsOf(t, token)
		test.False(t, grants.Video.GetCanPublish())
		test.False(t, grants.Video.GetCanPublishData())
		test.True(t, grants.Video.GetCanSubscribe())
	})

	T.Run("refuses with no server configured", func(t *testing.T) {
		t.Parallel()

		_, err := NewMinter(&Config{}).Mint(&Grant{Room: "stage-1", Identity: "participant-1", Publish: true})
		test.ErrorIs(t, err, ErrNotConfigured)
	})
}

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("accepts all three or none", func(t *testing.T) {
		t.Parallel()

		test.NoError(t, testConfig().ValidateWithContext(t.Context()))
		test.NoError(t, (&Config{}).ValidateWithContext(t.Context()))
	})

	T.Run("refuses a half-configured server", func(t *testing.T) {
		t.Parallel()

		test.Error(t, (&Config{APIKey: "key"}).ValidateWithContext(t.Context()))
		test.Error(t, (&Config{APIKey: "key", APISecret: "secret"}).ValidateWithContext(t.Context()))
	})
}

func TestNewAdmin(T *testing.T) {
	T.Parallel()

	T.Run("is a noop with no server configured", func(t *testing.T) {
		t.Parallel()

		_, isNoop := NewAdmin(&Config{}, nil).(NoopAdmin)
		test.True(t, isNoop)
	})

	T.Run("talks to the room service over HTTP at the server's host", func(t *testing.T) {
		t.Parallel()

		test.EqOp(t, "https://video.example", httpURL("wss://video.example"))
		test.EqOp(t, "http://localhost:7880", httpURL("ws://localhost:7880"))

		_, isRoom := NewAdmin(testConfig(), nil).(*RoomAdmin)
		test.True(t, isRoom)
	})
}
