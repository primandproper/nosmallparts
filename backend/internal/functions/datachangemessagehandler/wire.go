package datachangemessagehandler

import (
	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		NewAsyncDataChangeMessageHandler,
	)
)
