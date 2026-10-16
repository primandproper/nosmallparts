package api_test

// The TypeScript suite had no LiveKit cases; these are #6's acceptance: the
// token's scope for each role, and who may have one.

import (
	"context"
	"testing"

	"github.com/primandproper/nosmallparts/backend/internal/config"
	"github.com/primandproper/nosmallparts/backend/internal/guests"
	"github.com/primandproper/nosmallparts/backend/internal/livekit"
	"github.com/primandproper/nosmallparts/backend/internal/proto/livekitpb"

	lkauth "github.com/livekit/protocol/auth"
	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const testLiveKitSecret = "test-secret-that-is-long-enough-000"

func withLiveKit(cfg *config.Config) {
	cfg.LiveKit = livekit.Config{APIKey: "test-key", APISecret: testLiveKitSecret, URL: "wss://video.example"}
}

// asGuest is a context carrying a guest's cookie, as a browser's gRPC-web
// request does.
func asGuest(ctx context.Context, guestID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "cookie", guests.CookieName+"="+guestID)
}

func videoGrant(t *testing.T, token string) *lkauth.VideoGrant {
	t.Helper()

	verifier, err := lkauth.ParseAPIToken(token)
	must.NoError(t, err)

	_, grants, err := verifier.Verify(testLiveKitSecret)
	must.NoError(t, err)
	must.NotNil(t, grants.Video)

	return grants.Video
}

func TestLiveKitTokens(T *testing.T) {
	T.Parallel()

	T.Run("a performer may publish; the audience only subscribes", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t, withLiveKit)
		stage := h.makeStage(t, "")

		host := h.userIdentity(t)
		hostSocket := h.joinClient(t, host, stage.ID, "Hosty")

		watcher := guestIdentity()
		watcherSocket := h.joinClient(t, watcher, stage.ID, "Watcher")
		watcherSocket.send(setAudience(true))
		watcherSocket.waitFor("the audience", func() bool { return watcherSocket.me().GetIsAudience() })

		performer, err := h.livekit.MintToken(as(t.Context(), host.user), &livekitpb.MintTokenRequest{
			StageId:       stage.ID,
			ParticipantId: hostSocket.participantID(),
		})
		must.NoError(t, err)
		test.EqOp(t, stage.ID, performer.GetRoom())
		test.EqOp(t, "wss://video.example", performer.GetUrl())

		grant := videoGrant(t, performer.GetToken())
		test.EqOp(t, stage.ID, grant.Room)
		test.True(t, grant.GetCanPublish())
		test.True(t, grant.GetCanSubscribe())

		audience, err := h.livekit.MintToken(asGuest(t.Context(), watcher.guestID), &livekitpb.MintTokenRequest{
			StageId:       stage.ID,
			ParticipantId: watcherSocket.participantID(),
		})
		must.NoError(t, err)

		grant = videoGrant(t, audience.GetToken())
		test.False(t, grant.GetCanPublish())
		test.True(t, grant.GetCanSubscribe())
	})

	T.Run("refuses a participant who is not the caller", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t, withLiveKit)
		stage := h.makeStage(t, "")
		socket := h.joinClient(t, guestIdentity(), stage.ID, "Somebody")

		for _, ctx := range []context.Context{
			t.Context(),
			asGuest(t.Context(), guestIdentity().guestID),
			as(t.Context(), h.signUpUser(t, "")),
		} {
			_, err := h.livekit.MintToken(ctx, &livekitpb.MintTokenRequest{StageId: stage.ID, ParticipantId: socket.participantID()})
			must.Error(t, err)
			test.EqOp(t, codes.PermissionDenied, status.Code(err))
		}
	})

	T.Run("refuses a participant who is not on the stage", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t, withLiveKit)
		stage := h.makeStage(t, "")
		other := h.makeStage(t, "")

		who := guestIdentity()
		socket := h.joinClient(t, who, stage.ID, "Somebody")

		_, err := h.livekit.MintToken(asGuest(t.Context(), who.guestID), &livekitpb.MintTokenRequest{
			StageId:       other.ID,
			ParticipantId: socket.participantID(),
		})
		must.Error(t, err)
		test.EqOp(t, codes.NotFound, status.Code(err))
	})

	T.Run("answers with the fallback message when video is not configured", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		stage := h.makeStage(t, "")

		who := guestIdentity()
		socket := h.joinClient(t, who, stage.ID, "Somebody")

		_, err := h.livekit.MintToken(asGuest(t.Context(), who.guestID), &livekitpb.MintTokenRequest{
			StageId:       stage.ID,
			ParticipantId: socket.participantID(),
		})
		must.Error(t, err)
		test.EqOp(t, codes.FailedPrecondition, status.Code(err))
		test.EqOp(t, livekit.ErrNotConfigured.Error(), status.Convert(err).Message())
	})
}
