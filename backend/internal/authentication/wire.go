package authentication

import "github.com/google/wire"

var (
	// AuthProviders are what we offer to dependency injection.
	AuthProviders = wire.NewSet(
		ProvideArgon2Authenticator,
		ProvideHasher,
		NewManager,
	)
)

func ProvideHasher(authenticator Authenticator) Hasher {
	return authenticator
}
