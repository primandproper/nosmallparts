package config

import "github.com/google/wire"

var (
	UploadedMediaConfigProviders = wire.NewSet(
		wire.FieldsOf(new(*Config),
			"Uploads",
		),
	)
)
