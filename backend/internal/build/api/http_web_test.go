package api_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/primandproper/nosmallparts/backend/internal/build/api"

	"github.com/shoenig/test/must"
)

// get calls GET on the server's HTTP listener and returns the status and body.
func (h *harness) get(t *testing.T, path string) (status int, contentType, body string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, h.httpURL+path, http.NoBody)
	must.NoError(t, err)

	res, err := h.http.Do(req)
	must.NoError(t, err)

	t.Cleanup(func() { must.NoError(t, res.Body.Close()) })

	raw, err := io.ReadAll(res.Body)
	must.NoError(t, err)

	return res.StatusCode, res.Header.Get("Content-Type"), string(raw)
}

// isWebApp reports whether a body is the web app's answer: index.html when the
// binary carries a build, or the note saying it does not.
func isWebApp(body string) bool {
	return strings.Contains(strings.ToLower(body), "<!doctype html") || strings.Contains(body, "make web")
}

func TestWebApp(T *testing.T) {
	T.Parallel()

	T.Run("answers a deep link with the app", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		status, _, body := h.get(t, "/stage/ABCDEF")
		must.NotEq(t, http.StatusNotFound, status)
		must.True(t, isWebApp(body))
	})

	T.Run("leaves the routes it does not own alone", func(t *testing.T) {
		t.Parallel()

		h := newHarness(t)

		status, contentType, body := h.get(t, api.IdentityRoute)
		must.EqOp(t, http.StatusOK, status)
		must.StrContains(t, contentType, "application/json")
		must.False(t, isWebApp(body))

		status, _, body = h.get(t, "/readyz")
		must.EqOp(t, http.StatusOK, status)
		must.False(t, isWebApp(body))
	})
}
