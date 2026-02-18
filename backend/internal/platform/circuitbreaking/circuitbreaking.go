package circuitbreaking

import (
	"context"
	"fmt"
	"log"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"

	circuit "github.com/rubyist/circuitbreaker"
)

type CircuitBreaker interface {
	Failed()
	Succeeded()
	CanProceed() bool
	CannotProceed() bool
}

type baseImplementation struct {
	circuitBreaker *circuit.Breaker
}

func (b *baseImplementation) Failed() {
	b.circuitBreaker.Fail()
}

func (b *baseImplementation) Succeeded() {
	b.circuitBreaker.Success()
}

func (b *baseImplementation) CanProceed() bool {
	return b.circuitBreaker.Ready()
}

func (b *baseImplementation) CannotProceed() bool {
	return !b.circuitBreaker.Ready()
}

// EnsureCircuitBreaker ensures a valid CircuitBreaker is made available.
func EnsureCircuitBreaker(breaker CircuitBreaker) CircuitBreaker {
	if breaker == nil {
		log.Println("NOOP CircuitBreaker implementation in use.")
		return NewNoopCircuitBreaker()
	}

	return breaker
}

// ProvideCircuitBreaker provides a CircuitBreaker.
func (cfg *Config) ProvideCircuitBreaker(logger logging.Logger, metricsProvider metrics.Provider) (CircuitBreaker, error) {
	if cfg == nil {
		return nil, internalerrors.NilConfigError("circuit breaker")
	}

	logger = logging.EnsureLogger(logger).WithValue("circuit_breaker", cfg.Name)

	if err := cfg.ValidateWithContext(context.Background()); err != nil {
		logger.Error("invalid config passed, providing noop circuit breaker", err)
		return NewNoopCircuitBreaker(), nil
	}

	cfg.EnsureDefaults()

	brokenCounter, err := metricsProvider.NewInt64Counter(fmt.Sprintf("%s_circuit_breaker_tripped", cfg.Name))
	if err != nil {
		return nil, err
	}

	failureCounter, err := metricsProvider.NewInt64Counter(fmt.Sprintf("%s_circuit_breaker_failed", cfg.Name))
	if err != nil {
		return nil, err
	}

	resetCounter, err := metricsProvider.NewInt64Counter(fmt.Sprintf("%s_circuit_breaker_reset", cfg.Name))
	if err != nil {
		return nil, err
	}

	cb := circuit.NewBreakerWithOptions(&circuit.Options{
		ShouldTrip: func(cb *circuit.Breaker) bool {
			return uint64(cb.Failures()+cb.Successes()) >= cfg.MinimumSampleThreshold && cb.ErrorRate() >= cfg.ErrorRate
		},
		WindowTime:    circuit.DefaultWindowTime,
		WindowBuckets: circuit.DefaultWindowBuckets,
	})

	go handleCircuitBreakerEvents(logger, cb, failureCounter, resetCounter, brokenCounter)

	return &baseImplementation{
		circuitBreaker: cb,
	}, nil
}

func handleCircuitBreakerEvents(
	logger logging.Logger,
	cb *circuit.Breaker,
	failureCounter,
	resetCounter,
	brokenCounter metrics.Int64Counter,
) {
	ctx := context.Background()
	for be := range <-cb.Subscribe() {
		switch be {
		case circuit.BreakerTripped:
			brokenCounter.Add(ctx, 1)
		case circuit.BreakerReset:
			resetCounter.Add(ctx, 1)
		case circuit.BreakerFail:
			failureCounter.Add(ctx, 1)
		case circuit.BreakerReady:
			logger.Debug("circuit breaker is ready")
		}
	}
}
