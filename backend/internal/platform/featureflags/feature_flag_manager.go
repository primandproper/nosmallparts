package featureflags

import (
	"context"
)

type (
	User interface {
		GetID() string
		GetEmail() string
		GetUsername() string
		GetFirstName() string
		GetLastName() string
	}

	// FeatureFlagManager manages feature flags.
	FeatureFlagManager interface {
		Identify(ctx context.Context, user User) error
		CanUseFeature(ctx context.Context, userID, feature string) (bool, error)
		Close() error
	}
)

func NewNoopFeatureFlagManager() FeatureFlagManager {
	return &NoopFeatureFlagManager{}
}

// NoopFeatureFlagManager is a no-op FeatureFlagManager.
type NoopFeatureFlagManager struct{}

// Identify implements the FeatureFlagManager interface.
func (m *NoopFeatureFlagManager) Identify(context.Context, User) error {
	return nil
}

// CanUseFeature implements the FeatureFlagManager interface.
func (*NoopFeatureFlagManager) CanUseFeature(context.Context, string, string) (bool, error) {
	return false, nil
}

// Close implements the FeatureFlagManager interface.
func (*NoopFeatureFlagManager) Close() error {
	return nil
}
