package authcfg

import "github.com/google/wire"

var (
	Providers = wire.NewSet(
		wire.FieldsOf(
			new(*Config),
			"Tokens",
			"SSO",
		),
	)
)
