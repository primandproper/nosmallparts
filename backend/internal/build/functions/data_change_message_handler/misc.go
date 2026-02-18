package datachangemessagehandler

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/chi"

	"github.com/google/wire"
)

var ProvidersMiscellaneous = wire.NewSet(
	chi.NewRouteParamManager,
)
