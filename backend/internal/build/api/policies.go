package api

import (
	"context"
	"sync"
	"unicode/utf8"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	"github.com/primandproper/platform-go/v15/identity"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	grpcerrors "github.com/primandproper/primitives-go/v2/errors/grpc"
)

// The sign-in rules nosmallparts adds to platform sign-in, which ships a seam
// for each and a rule for neither. Both carry over what the service did before
// platform sign-in.

// MinPasswordLength is the shortest password anybody may choose.
const MinPasswordLength = 8

// ErrPasswordTooShort is the refusal a too-short password gets.
var ErrPasswordTooShort = platformerrors.Newf("use at least %d characters", MinPasswordLength)

// registerClientSafe makes ErrPasswordTooShort's words the ones a client is
// shown, rather than sign-in's generic refusal. The registry is process-wide,
// so it is filled once however many injectors a process builds.
var registerClientSafe = sync.OnceFunc(func() {
	grpcerrors.RegisterClientSafeSentinels(ErrPasswordTooShort)
})

// PasswordPolicy refuses a password shorter than MinPasswordLength characters.
func PasswordPolicy(_ context.Context, password string) error {
	if utf8.RuneCountInString(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}

	return nil
}

// RegistrationPolicy starts a registrant in good standing, so they can sign in
// straight away. Platform sign-in holds a registrant unverified until they
// prove their email address, and nothing here sends the mail that would let
// them; the service has never asked anybody to verify an address.
func RegistrationPolicy(_ context.Context, registration *signin.Registration) error {
	if registration.User != nil {
		registration.User.AccountStatus = identity.StatusGood
	}

	return nil
}
