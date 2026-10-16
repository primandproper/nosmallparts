// Package displaynames is the policy a name shown on a stage answers to: what
// a participant may call themselves, and how two people choosing the same name
// are told apart.
package displaynames

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

// MaxLength is the longest a display name may be, in characters.
const MaxLength = 32

// The refusals. Their words are what a participant is shown.
var (
	ErrEmpty     = platformerrors.New("Please enter a name.")
	ErrTooLong   = platformerrors.Newf("Name is too long (max %d characters).", MaxLength)
	ErrNotPermit = platformerrors.New("Please choose a different name.")
)

// reserved are names that would read as the system speaking.
var reserved = map[string]struct{}{
	"host":     {},
	"admin":    {},
	"director": {},
	"system":   {},
	"server":   {},
}

// embedded are strong words refused even inside a longer name ("fucknugget"),
// which the whole-word list cannot catch. It is curated to strings that are
// never part of an ordinary name; matching the whole list as substrings flags
// ordinary names ("spaced" contains "spac").
var embedded = []string{
	"fuck",
	"shit",
	"cunt",
	"nigger",
	"nigga",
	"faggot",
	"bitch",
	"asshole",
	"wanker",
	"dickhead",
	"cocksuck",
	"motherfuck",
}

//go:embed profanity.txt
var profanityList string

// wholeWords is the profanity list compiled once, on first use: each word
// literal, between word boundaries, without regard to case. It is the rule the
// TypeScript server's filter applied.
var wholeWords = sync.OnceValue(func() []*regexp.Regexp {
	var patterns []*regexp.Regexp

	for line := range strings.Lines(profanityList) {
		word := strings.TrimSpace(line)
		if word == "" || strings.HasPrefix(word, "#") {
			continue
		}

		patterns = append(patterns, regexp.MustCompile(`(?i)\b`+regexp.QuoteMeta(word)+`\b`))
	}

	return patterns
})

var whitespace = regexp.MustCompile(`\s+`)
var nonLetters = regexp.MustCompile(`[^a-z]`)

// Validate cleans a name, trimming it and collapsing its inner whitespace, and
// answers the cleaned name or the refusal to show the participant.
func Validate(raw string) (string, error) {
	name := whitespace.ReplaceAllString(strings.TrimSpace(raw), " ")

	switch {
	case name == "":
		return "", ErrEmpty
	case utf8.RuneCountInString(name) > MaxLength:
		return "", ErrTooLong
	case isReserved(name), isProfane(name):
		return "", ErrNotPermit
	}

	return name, nil
}

func isReserved(name string) bool {
	_, ok := reserved[strings.ToLower(name)]

	return ok
}

func isProfane(name string) bool {
	for _, pattern := range wholeWords() {
		if pattern.MatchString(name) {
			return true
		}
	}

	squashed := nonLetters.ReplaceAllString(strings.ToLower(name), "")

	for _, word := range embedded {
		if strings.Contains(squashed, word) {
			return true
		}
	}

	return false
}

// Dedupe answers name, or name with the first free " (n)" suffix if somebody on
// the stage already holds it. Names are compared without regard to case, and
// a suffixed name is cut to keep the whole within MaxLength.
func Dedupe(name string, taken []string) string {
	held := make(map[string]struct{}, len(taken))
	for _, t := range taken {
		held[strings.ToLower(strings.TrimSpace(t))] = struct{}{}
	}

	if _, clash := held[strings.ToLower(name)]; !clash {
		return name
	}

	for n := 2; ; n++ {
		suffix := fmt.Sprintf(" (%d)", n)
		base := strings.TrimRight(truncate(name, MaxLength-len(suffix)), " ")
		candidate := base + suffix

		if _, clash := held[strings.ToLower(candidate)]; !clash {
			return candidate
		}
	}
}

// truncate cuts s to at most n characters.
func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}

	return string([]rune(s)[:n])
}
