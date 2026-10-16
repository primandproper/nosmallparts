package admins

import (
	"testing"

	"github.com/shoenig/test"
)

func TestList_Includes(T *testing.T) {
	T.Parallel()

	// Ported from units.test.ts, whose setup set ADMIN_EMAILS to
	// "admin@example.com, Boss@Example.COM". A null or undefined address there
	// is the empty one here.
	T.Run("matches case-insensitively against ADMIN_EMAILS", func(t *testing.T) {
		t.Parallel()

		list := New([]string{"admin@example.com", " Boss@Example.COM"})

		test.True(t, list.Includes("admin@example.com"))
		test.True(t, list.Includes("ADMIN@EXAMPLE.COM"))
		test.True(t, list.Includes("boss@example.com"))
		test.False(t, list.Includes("someone@example.com"))
		test.False(t, list.Includes(""))
	})
}
