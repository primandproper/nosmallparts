// Package admins says who the admins are: the holders of the email addresses
// the configuration lists. There is no table and no role; the list is read at
// boot, so a change to it takes effect on restart.
package admins

import (
	"strings"
)

// List is the configured admin addresses.
type List struct {
	emails map[string]struct{}
}

// New builds the list. Addresses are compared without regard to case or
// surrounding whitespace, and an empty one is skipped.
func New(emails []string) *List {
	l := &List{emails: make(map[string]struct{}, len(emails))}

	for _, email := range emails {
		if folded := fold(email); folded != "" {
			l.emails[folded] = struct{}{}
		}
	}

	return l
}

// Includes reports whether an address belongs to an admin. The empty address
// never does.
func (l *List) Includes(email string) bool {
	folded := fold(email)
	if folded == "" || l == nil {
		return false
	}

	_, ok := l.emails[folded]

	return ok
}

func fold(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
