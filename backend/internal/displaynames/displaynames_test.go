package displaynames

import (
	"strings"
	"testing"

	"github.com/shoenig/test"
	"github.com/shoenig/test/must"
)

// The cases here are ported from units.test.ts, under their original names.

func TestValidate(T *testing.T) {
	T.Parallel()

	T.Run("validates length, reserved tokens, profanity", func(t *testing.T) {
		t.Parallel()

		name, err := Validate("Alice")
		must.NoError(t, err)
		test.EqOp(t, "Alice", name)

		name, err = Validate("  spaced   out  ")
		must.NoError(t, err)
		test.EqOp(t, "spaced out", name)

		_, err = Validate(strings.Repeat("x", 32))
		test.NoError(t, err)

		_, err = Validate(strings.Repeat("x", 33))
		test.ErrorIs(t, err, ErrTooLong)

		_, err = Validate("")
		test.ErrorIs(t, err, ErrEmpty)

		for _, r := range []string{"host", "Admin", "DIRECTOR", "system", "Server"} {
			_, err = Validate(r)
			test.ErrorIs(t, err, ErrNotPermit, test.Sprintf("name %q", r))
		}

		_, err = Validate("fucknugget")
		test.ErrorIs(t, err, ErrNotPermit)
	})
}

func TestDedupe(T *testing.T) {
	T.Parallel()

	T.Run("dedupes with numeric suffixes within the length cap", func(t *testing.T) {
		t.Parallel()

		test.EqOp(t, "Sam", Dedupe("Sam", nil))
		test.EqOp(t, "Sam (2)", Dedupe("Sam", []string{"sam"}))
		test.EqOp(t, "Sam (3)", Dedupe("Sam", []string{"sam", "sam (2)"}))

		long := strings.Repeat("x", 32)
		deduped := Dedupe(long, []string{long})
		test.LessEq(t, 32, len(deduped))
		test.StrHasSuffix(t, "(2)", deduped)
	})
}
