// Package guests is the identity a visitor has before they sign in: a ULID the
// server mints, kept in an httpOnly cookie for six months.
//
// The server is the only one who mints it. A cookie whose value is not a ULID
// is treated as no identity at all, and its holder is minted a fresh one, so a
// tampered value cannot be used to pose as somebody else's guest identity
// beyond guessing a 128-bit value.
package guests

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"

	"github.com/oklog/ulid/v2"
)

// CookieName is the cookie a guest identity travels in.
const CookieName = "nsp_guest_id"

// cookieMaxAge is how long a browser keeps a guest identity.
const cookieMaxAge = 180 * 24 * time.Hour

const handlerName = "guest_identity_handler"

// validID is a ULID in its canonical form: 26 characters of Crockford's
// base32, upper case. It is stricter than ulid.Parse, which accepts lower case
// and so would let one identity be spelled two ways.
var validID = regexp.MustCompile(`^[0-9A-HJKMNP-TV-Z]{26}$`)

// FromRequest is the guest identity a request's cookie carries, if it carries a
// valid one.
func FromRequest(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", false
	}

	if !validID.MatchString(cookie.Value) {
		return "", false
	}

	return cookie.Value, true
}

// Mint gives the response a new guest identity and returns it.
//
// The cookie is Secure everywhere, local development included: browsers treat
// http://localhost as a secure context and keep a Secure cookie from it, which
// sign-in's refresh token cookie already relies on.
func Mint(w http.ResponseWriter) string {
	id := ulid.Make().String()

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    id,
		Path:     "/",
		MaxAge:   int(cookieMaxAge.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	return id
}

// identityResponse is the answer's shape: exactly one of the two is set.
type identityResponse struct {
	UserID  *string `json:"userId"`
	GuestID *string `json:"guestId"`
}

// Handler answers GET /api/identity: who this visitor is.
type Handler struct {
	o11y      observability.Observer
	extractor callers.PrincipalExtractor
}

// NewHandler builds the handler. The extractor reads the caller the
// authentication middleware in front of it resolved.
func NewHandler(extractor callers.PrincipalExtractor, logger logging.Logger) *Handler {
	return &Handler{
		o11y:      observability.NewObserver(handlerName, logger, nil),
		extractor: extractor,
	}
}

// ServeHTTP answers a signed-in user with their user ID, and anybody else with
// their guest identity, minted when they arrive without a valid one.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, op := h.o11y.Begin(r.Context())
	defer op.End()

	var answer identityResponse

	if principal, ok := h.extractor(ctx); ok && principal != nil {
		answer.UserID = new(principal.UserID())
	} else {
		guestID, valid := FromRequest(r)
		if !valid {
			guestID = Mint(w)
		}

		answer.GuestID = new(guestID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	if err := json.NewEncoder(w).Encode(answer); err != nil {
		op.Acknowledge(err, "writing the identity answer")
	}
}
