package launchdarkly

import (
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
)

type (
	Config struct {
		SDKKey               string                 `env:"SDK_KEY"                 json:"sdkKey"`
		CircuitBreakerConfig circuitbreaking.Config `envPrefix:"CIRCUIT_BREAKING_" json:"circuitBreakerConfig"`
		InitTimeout          time.Duration          `env:"INIT_TIMEOUT"            json:"initTimeout"`
	}
)
