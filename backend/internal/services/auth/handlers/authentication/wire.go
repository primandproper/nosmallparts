package authentication

import (
	"github.com/google/wire"
)

// AuthHTTPServiceProviders are our collection of what we provide to other services.
var AuthHTTPServiceProviders = wire.NewSet(
	ProvideService,
	ProvideOAuth2ClientManager,
	ProvideOAuth2ServerImplementation,
)
