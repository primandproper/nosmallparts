package api_test

// The cases in this file are ported from the TypeScript server's
// http-scripts.test.ts, each subtest under its original name. A guarantee the
// old suite asserted through Better Auth or REST is asserted the same way
// through platform sign-in or gRPC, and the status codes map as 401 to
// Unauthenticated, 403 to PermissionDenied, 404 to NotFound and 400 to
// InvalidArgument.

import (
	"encoding/json"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/primandproper/nosmallparts/backend/internal/build/api"
	"github.com/primandproper/nosmallparts/backend/internal/guests"
	"github.com/primandproper/nosmallparts/backend/internal/proto/scriptspb"

	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ulidPattern = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)

type identityAnswer struct {
	UserID  *string `json:"userId"`
	GuestID *string `json:"guestId"`
}

// getIdentity calls GET /api/identity with the given cookie and bearer token,
// either of which may be empty.
func (h *harness) getIdentity(t *testing.T, cookie *http.Cookie, token string) (*identityAnswer, *http.Response) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, h.httpURL+api.IdentityRoute, http.NoBody)
	must.NoError(t, err)

	if cookie != nil {
		req.AddCookie(cookie)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := h.http.Do(req)
	must.NoError(t, err)

	t.Cleanup(func() { must.NoError(t, res.Body.Close()) })

	must.EqOp(t, http.StatusOK, res.StatusCode)

	var answer identityAnswer
	must.NoError(t, json.NewDecoder(res.Body).Decode(&answer))

	return &answer, res
}

func guestCookie(res *http.Response) *http.Cookie {
	for _, cookie := range res.Cookies() {
		if cookie.Name == guests.CookieName {
			return cookie
		}
	}

	return nil
}

func TestIdentity(T *testing.T) {
	T.Parallel()

	T.Run("mints a guest cookie for a fresh visitor", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		answer, res := h.getIdentity(t, nil, "")
		test.Nil(t, answer.UserID)
		must.NotNil(t, answer.GuestID)
		test.True(t, ulidPattern.MatchString(*answer.GuestID))

		cookie := guestCookie(res)
		must.NotNil(t, cookie)
		test.EqOp(t, *answer.GuestID, cookie.Value)
		test.True(t, cookie.HttpOnly)
	})

	T.Run("keeps the same guest id across requests with the cookie", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		first, res := h.getIdentity(t, nil, "")
		cookie := guestCookie(res)
		must.NotNil(t, cookie)

		second, _ := h.getIdentity(t, &http.Cookie{Name: cookie.Name, Value: cookie.Value}, "")
		must.NotNil(t, second.GuestID)
		test.EqOp(t, *first.GuestID, *second.GuestID)
	})

	T.Run("treats a tampered non-ULID cookie as no identity and mints a new one", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		answer, res := h.getIdentity(t, &http.Cookie{Name: guests.CookieName, Value: "not-a-ulid"}, "")
		must.NotNil(t, answer.GuestID)
		test.True(t, ulidPattern.MatchString(*answer.GuestID))
		test.NotEq(t, "not-a-ulid", *answer.GuestID)
		test.NotNil(t, guestCookie(res))
	})

	T.Run("reports userId for a signed-in user", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		answer, _ := h.getIdentity(t, nil, owner.Token)
		must.NotNil(t, answer.UserID)
		test.EqOp(t, owner.ID, *answer.UserID)
		test.Nil(t, answer.GuestID)
	})
}

func TestAuth(T *testing.T) {
	T.Parallel()

	T.Run("register with an existing email fails without creating a partial account", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		_, err := h.signIn.Register(t.Context(), registerRequest(owner.Email, "another-password-1", "Dup"))
		must.Error(t, err)
		test.EqOp(t, codes.AlreadyExists, status.Code(err))

		// Nothing of the second registration stands: its password opens
		// nothing, and the first account's still does.
		_, err = h.signIn.LoginForToken(t.Context(), &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{EmailAddress: owner.Email, Password: "another-password-1"},
		})
		test.Error(t, err)

		_, err = h.signIn.LoginForToken(t.Context(), &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{EmailAddress: owner.Email, Password: owner.Password},
		})
		test.NoError(t, err)
	})

	T.Run("register with a weak password fails validation", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		_, err := h.signIn.Register(t.Context(), registerRequest("weak@test.example", "short", "Weak"))
		must.Error(t, err)
		test.EqOp(t, codes.InvalidArgument, status.Code(err))
		test.StrContains(t, status.Convert(err).Message(), api.ErrPasswordTooShort.Error())
	})

	T.Run("login with wrong creds yields no session", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		res, err := h.signIn.LoginForToken(t.Context(), &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{EmailAddress: owner.Email, Password: "wrong-password-1"},
		})
		must.Error(t, err)
		test.EqOp(t, codes.Unauthenticated, status.Code(err))
		test.Nil(t, res.GetToken())
	})

	T.Run("login with correct creds sets a session cookie", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		// On the native transport the refresh token is in the body; on the
		// gRPC-web route platform sign-in seals that same field into the
		// cookie, which is its half of this guarantee and its own tests'.
		res, err := h.signIn.LoginForToken(t.Context(), &signinpb.LoginForTokenRequest{
			Credentials: &signinpb.Credentials{EmailAddress: owner.Email, Password: owner.Password},
		})
		must.NoError(t, err)
		test.NotEq(t, "", res.GetToken().GetRefreshToken())

		self, err := h.signIn.GetSelf(as(t.Context(), &testUser{Token: res.GetToken().GetToken()}), &signinpb.GetSelfRequest{})
		must.NoError(t, err)
		test.EqOp(t, owner.ID, self.GetUser().GetId())
	})

	T.Run("sign out clears the session; protected routes 401", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		user := h.signUpUser(t, "")

		_, err := h.signIn.SignOut(as(t.Context(), user), &signinpb.SignOutRequest{RefreshToken: user.RefreshToken})
		must.NoError(t, err)

		_, err = h.scripts.CreatePlayScript(as(t.Context(), user), &scriptspb.CreatePlayScriptRequest{Play: sampleScript()})
		must.Error(t, err)
		test.EqOp(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestAuthGates(T *testing.T) {
	T.Parallel()

	T.Run("POST /api/play-scripts requires user", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		_, err := h.scripts.CreatePlayScript(t.Context(), &scriptspb.CreatePlayScriptRequest{Play: sampleScript()})
		must.Error(t, err)
		test.EqOp(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("POST /api/play-scripts/:id/rate requires user", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		play := h.createScript(t, owner, nil)

		_, err := h.scripts.RatePlayScript(t.Context(), &scriptspb.RatePlayScriptRequest{Id: play.GetId(), Rating: 5})
		must.Error(t, err)
		test.EqOp(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("PUT of a script you do not own returns 404", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		rater := h.signUpUser(t, "")
		play := h.createScript(t, owner, nil)

		hijacked := sampleScript()
		hijacked.Title = "Hijacked"

		_, err := h.scripts.UpdatePlayScript(as(t.Context(), rater), &scriptspb.UpdatePlayScriptRequest{Id: play.GetId(), Play: hijacked})
		must.Error(t, err)
		test.EqOp(t, codes.NotFound, status.Code(err))
	})
}

func TestScriptsCRUD(T *testing.T) {
	T.Parallel()

	T.Run("creates a script with cast, lines and tags; defaults to enabled", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		play := h.createScript(t, owner, func(in *scriptspb.PlayInput) { in.TagNames = []string{"Comedy", "Short"} })
		test.True(t, play.GetEnabled())

		res, err := h.scripts.GetPlayScript(as(t.Context(), owner), &scriptspb.GetPlayScriptRequest{Id: play.GetId()})
		must.NoError(t, err)

		test.SliceLen(t, 2, res.GetPlay().GetCast())
		test.SliceLen(t, 3, res.GetPlay().GetLines())

		var names []string
		for _, tag := range res.GetPlay().GetTags() {
			names = append(names, tag.GetName())
		}

		slices.Sort(names)
		test.Eq(t, []string{"Comedy", "Short"}, names)
	})

	T.Run("edit own script persists title/cast/lines and preserves ordering", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		play := h.createScript(t, owner, nil)

		edited := sampleScript()
		edited.Title = "Renamed"
		edited.Lines = []*scriptspb.PlayLine{
			{CastMemberIds: []string{"b"}, Text: "New first", Index: 0},
			{CastMemberIds: []string{"a"}, Text: "New second", Index: 1},
		}

		res, err := h.scripts.UpdatePlayScript(as(t.Context(), owner), &scriptspb.UpdatePlayScriptRequest{Id: play.GetId(), Play: edited})
		must.NoError(t, err)

		updated := res.GetPlay()
		test.EqOp(t, "Renamed", updated.GetTitle())

		var (
			texts   []string
			indexes []int32
		)

		for _, line := range updated.GetLines() {
			texts = append(texts, line.GetText())
			indexes = append(indexes, line.GetIndex())
		}

		test.Eq(t, []string{"New first", "New second"}, texts)
		test.Eq(t, []int32{0, 1}, indexes)
	})

	T.Run("anonymous visitors see only public enabled scripts", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		public := h.createScript(t, owner, func(in *scriptspb.PlayInput) { in.Title, in.IsPublic = "PublicOne", true })
		private := h.createScript(t, owner, func(in *scriptspb.PlayInput) { in.Title, in.IsPublic = "PrivateOne", false })

		res, err := h.scripts.ListPlayScripts(t.Context(), &scriptspb.ListPlayScriptsRequest{})
		must.NoError(t, err)

		ids := playIDs(res.GetPlays())
		test.SliceContains(t, ids, public.GetId())
		test.SliceNotContains(t, ids, private.GetId())
	})

	T.Run("?search= filters case-insensitively across title/description/author", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		h.createScript(t, owner, func(in *scriptspb.PlayInput) { in.Title = "ZebraQuest Saga" })
		h.createScript(t, owner, nil)

		res, err := h.scripts.ListPlayScripts(as(t.Context(), owner), &scriptspb.ListPlayScriptsRequest{Search: pointer.To("zebraquest")})
		must.NoError(t, err)

		must.SliceNotEmpty(t, res.GetPlays())

		for _, play := range res.GetPlays() {
			test.StrContains(t, strings.ToLower(searchable(play)), "zebraquest")
		}
	})
}

// searchable is every field a search covers, run together.
func searchable(play *scriptspb.Play) string {
	fields := []string{play.GetTitle(), play.GetDescription(), play.GetAuthor()}

	for _, member := range play.GetCast() {
		fields = append(fields, member.GetName())
	}

	for _, tag := range play.GetTags() {
		fields = append(fields, tag.GetName())
	}

	return strings.Join(fields, " ")
}

func playIDs(plays []*scriptspb.Play) []string {
	ids := make([]string, 0, len(plays))
	for _, play := range plays {
		ids = append(ids, play.GetId())
	}

	return ids
}

func TestRatings(T *testing.T) {
	T.Parallel()

	T.Run("rates 1–5, averages, re-rates without double counting, blocks self-rating", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		rater := h.signUpUser(t, "")
		play := h.createScript(t, owner, nil)

		// Owner cannot rate their own script.
		_, err := h.scripts.RatePlayScript(as(t.Context(), owner), &scriptspb.RatePlayScriptRequest{Id: play.GetId(), Rating: 5})
		must.Error(t, err)
		test.EqOp(t, codes.PermissionDenied, status.Code(err))

		// Somebody else rates.
		first, err := h.scripts.RatePlayScript(as(t.Context(), rater), &scriptspb.RatePlayScriptRequest{Id: play.GetId(), Rating: 4})
		must.NoError(t, err)
		test.EqOp(t, int32(1), first.GetRatingStats().GetRatingCount())
		test.EqOp(t, 4.0, first.GetRatingStats().GetAverageRating())

		// Re-rating replaces rather than double counting.
		second, err := h.scripts.RatePlayScript(as(t.Context(), rater), &scriptspb.RatePlayScriptRequest{Id: play.GetId(), Rating: 2})
		must.NoError(t, err)
		test.EqOp(t, int32(1), second.GetRatingStats().GetRatingCount())
		test.EqOp(t, 2.0, second.GetRatingStats().GetAverageRating())

		// The rater's own rating is on the listing they see.
		list, err := h.scripts.ListPlayScripts(as(t.Context(), rater), &scriptspb.ListPlayScriptsRequest{})
		must.NoError(t, err)

		idx := slices.IndexFunc(list.GetPlays(), func(p *scriptspb.Play) bool { return p.GetId() == play.GetId() })
		must.NotEq(t, -1, idx)
		must.NotNil(t, list.GetPlays()[idx].UserRating)
		test.EqOp(t, int32(2), list.GetPlays()[idx].GetUserRating())
	})

	// The old case also sent 3.5. rating is an int32 on the wire now, so a
	// fractional rating cannot be sent at all.
	T.Run("rejects out-of-range ratings with 400", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		rater := h.signUpUser(t, "")
		play := h.createScript(t, owner, nil)

		for _, rating := range []int32{0, 6} {
			_, err := h.scripts.RatePlayScript(as(t.Context(), rater), &scriptspb.RatePlayScriptRequest{Id: play.GetId(), Rating: rating})
			must.Error(t, err)
			test.EqOp(t, codes.InvalidArgument, status.Code(err), test.Sprintf("rating %d", rating))
		}
	})
}

func TestAdminAndEnabledFlag(T *testing.T) {
	T.Parallel()

	T.Run("non-admin DELETE (even the owner) gets 403; admin DELETE succeeds", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		admin := h.signUpUser(t, "admin@example.com")
		play := h.createScript(t, owner, nil)

		_, err := h.scripts.DeletePlayScript(as(t.Context(), owner), &scriptspb.DeletePlayScriptRequest{Id: play.GetId()})
		must.Error(t, err)
		test.EqOp(t, codes.PermissionDenied, status.Code(err))

		_, err = h.scripts.DeletePlayScript(as(t.Context(), admin), &scriptspb.DeletePlayScriptRequest{Id: play.GetId()})
		must.NoError(t, err)

		_, err = h.scripts.GetPlayScript(as(t.Context(), admin), &scriptspb.GetPlayScriptRequest{Id: play.GetId()})
		must.Error(t, err)
		test.EqOp(t, codes.NotFound, status.Code(err))
	})

	T.Run("disable hides the script from non-admins (owner included); re-enable restores", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		admin := h.signUpUser(t, "admin@example.com")
		play := h.createScript(t, owner, func(in *scriptspb.PlayInput) { in.Title, in.IsPublic = "ToDisable", true })

		_, err := h.scripts.SetPlayScriptEnabled(as(t.Context(), admin), &scriptspb.SetPlayScriptEnabledRequest{Id: play.GetId(), Enabled: false})
		must.NoError(t, err)

		// Hidden from the anonymous list and from the owner's own.
		for _, viewer := range []*testUser{nil, owner} {
			res, listErr := h.scripts.ListPlayScripts(as(t.Context(), viewer), &scriptspb.ListPlayScriptsRequest{})
			must.NoError(t, listErr)
			test.SliceNotContains(t, playIDs(res.GetPlays()), play.GetId())
		}

		// The admin still sees it, flagged disabled.
		adminList, err := h.scripts.ListPlayScripts(as(t.Context(), admin), &scriptspb.ListPlayScriptsRequest{})
		must.NoError(t, err)

		idx := slices.IndexFunc(adminList.GetPlays(), func(p *scriptspb.Play) bool { return p.GetId() == play.GetId() })
		must.NotEq(t, -1, idx)
		test.False(t, adminList.GetPlays()[idx].GetEnabled())

		// A direct read is NotFound for anybody else.
		_, err = h.scripts.GetPlayScript(as(t.Context(), owner), &scriptspb.GetPlayScriptRequest{Id: play.GetId()})
		must.Error(t, err)
		test.EqOp(t, codes.NotFound, status.Code(err))

		// The old case also refused a new stage against the disabled script.
		// Stage creation arrives with the stages service (#5), which carries
		// that assertion in its own port of this case.

		// Re-enabling restores the listing.
		_, err = h.scripts.SetPlayScriptEnabled(as(t.Context(), admin), &scriptspb.SetPlayScriptEnabledRequest{Id: play.GetId(), Enabled: true})
		must.NoError(t, err)

		restored, err := h.scripts.ListPlayScripts(t.Context(), &scriptspb.ListPlayScriptsRequest{})
		must.NoError(t, err)
		test.SliceContains(t, playIDs(restored.GetPlays()), play.GetId())
	})

	T.Run("PATCH enabled is admin-only, case-insensitive on ADMIN_EMAILS", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")
		play := h.createScript(t, owner, nil)

		_, err := h.scripts.SetPlayScriptEnabled(as(t.Context(), owner), &scriptspb.SetPlayScriptEnabledRequest{Id: play.GetId(), Enabled: false})
		must.Error(t, err)
		test.EqOp(t, codes.PermissionDenied, status.Code(err))

		// "Boss@Example.COM" in the admin list matches boss@example.com.
		boss := h.signUpUser(t, "boss@example.com")

		res, err := h.scripts.SetPlayScriptEnabled(as(t.Context(), boss), &scriptspb.SetPlayScriptEnabledRequest{Id: play.GetId(), Enabled: true})
		must.NoError(t, err)
		test.True(t, res.GetEnabled())
	})
}

func TestTags(T *testing.T) {
	T.Parallel()

	T.Run("lists tags and merges duplicates case-insensitively", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)
		owner := h.signUpUser(t, "")

		h.createScript(t, owner, func(in *scriptspb.PlayInput) { in.TagNames = []string{"Drama"} })
		h.createScript(t, owner, func(in *scriptspb.PlayInput) { in.TagNames = []string{" drama ", "DRAMA"} })

		res, err := h.scripts.ListTags(t.Context(), &scriptspb.ListTagsRequest{})
		must.NoError(t, err)

		dramas := 0

		for _, tag := range res.GetTags() {
			if strings.EqualFold(tag.GetName(), "drama") {
				dramas++
			}
		}

		test.EqOp(t, 1, dramas)
	})
}
