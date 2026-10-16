package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/primandproper/primitives-go/v2/observability/logging/noop"

	"github.com/shoenig/test/must"
)

const indexBody = "<!doctype html><title>app</title>"

func builtFiles() fstest.MapFS {
	return fstest.MapFS{
		"index.html":           {Data: []byte(indexBody)},
		"assets/index-abc.js":  {Data: []byte("console.log(1)")},
		"assets/index-abc.css": {Data: []byte("body{}")},
		"favicon.svg":          {Data: []byte("<svg/>")},
	}
}

func get(t *testing.T, h http.Handler, target string) (res *http.Response, body string) {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody))

	res = rec.Result()
	raw, err := io.ReadAll(res.Body)
	must.NoError(t, err)
	must.NoError(t, res.Body.Close())

	return res, string(raw)
}

func TestHandler_ServeHTTP(T *testing.T) {
	T.Parallel()

	T.Run("serves index.html at the root, revalidated", func(t *testing.T) {
		t.Parallel()

		h := newHandler(builtFiles(), noop.NewLogger())

		res, body := get(t, h, "/")
		must.EqOp(t, http.StatusOK, res.StatusCode)
		must.EqOp(t, indexBody, body)
		must.EqOp(t, cacheRevalidate, res.Header.Get("Cache-Control"))
		must.StrContains(t, res.Header.Get("Content-Type"), "text/html")
	})

	T.Run("serves index.html for a client-side route", func(t *testing.T) {
		t.Parallel()

		h := newHandler(builtFiles(), noop.NewLogger())

		res, body := get(t, h, "/stage/ABCDEF")
		must.EqOp(t, http.StatusOK, res.StatusCode)
		must.EqOp(t, indexBody, body)
	})

	T.Run("serves /index.html itself without a redirect", func(t *testing.T) {
		t.Parallel()

		h := newHandler(builtFiles(), noop.NewLogger())

		res, body := get(t, h, "/index.html")
		must.EqOp(t, http.StatusOK, res.StatusCode)
		must.EqOp(t, indexBody, body)
	})

	T.Run("caches a hashed asset forever", func(t *testing.T) {
		t.Parallel()

		h := newHandler(builtFiles(), noop.NewLogger())

		res, body := get(t, h, "/assets/index-abc.js")
		must.EqOp(t, http.StatusOK, res.StatusCode)
		must.EqOp(t, "console.log(1)", body)
		must.EqOp(t, cacheImmutable, res.Header.Get("Cache-Control"))
	})

	T.Run("revalidates an unhashed file", func(t *testing.T) {
		t.Parallel()

		h := newHandler(builtFiles(), noop.NewLogger())

		res, _ := get(t, h, "/favicon.svg")
		must.EqOp(t, http.StatusOK, res.StatusCode)
		must.EqOp(t, cacheRevalidate, res.Header.Get("Cache-Control"))
	})

	T.Run("answers a missing asset with 404, not the app", func(t *testing.T) {
		t.Parallel()

		h := newHandler(builtFiles(), noop.NewLogger())

		res, body := get(t, h, "/assets/index-gone.js")
		must.EqOp(t, http.StatusNotFound, res.StatusCode)
		must.StrNotContains(t, body, indexBody)
	})

	T.Run("keeps a traversal inside the build", func(t *testing.T) {
		t.Parallel()

		h := newHandler(builtFiles(), noop.NewLogger())

		res, body := get(t, h, "/../../etc/passwd")
		must.EqOp(t, http.StatusOK, res.StatusCode)
		must.EqOp(t, indexBody, body)
	})

	T.Run("says so when the binary carries no build", func(t *testing.T) {
		t.Parallel()

		h := newHandler(fstest.MapFS{"README.md": {Data: []byte("placeholder")}}, noop.NewLogger())

		res, body := get(t, h, "/stage/ABCDEF")
		must.EqOp(t, http.StatusServiceUnavailable, res.StatusCode)
		must.EqOp(t, notBuilt, body)
	})
}

func TestNewHandler(T *testing.T) {
	T.Parallel()

	T.Run("builds over the embedded files", func(t *testing.T) {
		t.Parallel()

		h, err := NewHandler(noop.NewLogger())
		must.NoError(t, err)
		must.NotNil(t, h)
	})
}
