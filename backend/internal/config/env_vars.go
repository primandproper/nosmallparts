package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

const (
	// CeaseOperationEnvVarKey is the env var key used to indicate a function or job should just quit early.
	CeaseOperationEnvVarKey = "CEASE_OPERATION"
	// RunningInKubernetesEnvVarKey is the env var key we use to indicate we're running in Kubernetes.
	RunningInKubernetesEnvVarKey = "RUNNING_IN_KUBERNETES"
)

func ConditionallyCease() {
	if ShouldCeaseOperation() {
		slog.Info(fmt.Sprintf("%s is set to true, exiting", CeaseOperationEnvVarKey))
		os.Exit(0)
	}
}

// ShouldCeaseOperation returns whether a job should just quit without trying.
func ShouldCeaseOperation() bool {
	return strings.TrimSpace(strings.ToLower(os.Getenv(CeaseOperationEnvVarKey))) == "true"
}

// RunningInKubernetes returns whether the service is running in a Kubernetes cluster.
func RunningInKubernetes() bool {
	return os.Getenv(RunningInKubernetesEnvVarKey) != ""
}

func ApplyEnvironmentVariables(cfg any) error {
	return env.ParseWithOptions(cfg, env.Options{
		Prefix: EnvVarPrefix,
		OnSet: func(tag string, value any, isDefault bool) {
			slog.Debug("env var set",
				slog.String("tag", tag),
				slog.String("value", fmt.Sprintf("%+v", value)),
				slog.Bool("isDefault", isDefault),
			)
		},
	})
}
