package interceptors

import "github.com/google/wire"

var (
	InterceptorProviders = wire.NewSet(
		ProvideAuthInterceptor,
	)
)
