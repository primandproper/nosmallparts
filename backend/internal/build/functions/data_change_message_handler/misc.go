package datachangemessagehandler

import (
	"github.com/verygoodsoftwarenotvirus/platform/v4/routing/chi"

	"github.com/google/wire"
)

var ProvidersMiscellaneous = wire.NewSet(
	chi.NewRouteParamManager,
)
