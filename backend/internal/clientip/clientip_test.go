package clientip

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shoenig/test"
	"google.golang.org/grpc/metadata"
)

func TestFromRequest(T *testing.T) {
	T.Parallel()

	T.Run("takes the last forwarded entry", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
		r.Header.Add("X-Forwarded-For", "203.0.113.9, 198.51.100.4")

		test.EqOp(t, "198.51.100.4", FromRequest(r))
	})

	T.Run("falls back to the socket address", func(t *testing.T) {
		t.Parallel()

		r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
		r.RemoteAddr = "192.0.2.7:51234"

		test.EqOp(t, "192.0.2.7", FromRequest(r))
	})
}

func TestFromContext(T *testing.T) {
	T.Parallel()

	T.Run("reads the forwarded metadata", func(t *testing.T) {
		t.Parallel()

		ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("x-forwarded-for", "203.0.113.9, 198.51.100.4"))

		test.EqOp(t, "198.51.100.4", FromContext(ctx))
	})

	T.Run("is empty with nothing to read", func(t *testing.T) {
		t.Parallel()

		test.EqOp(t, "", FromContext(t.Context()))
	})
}
