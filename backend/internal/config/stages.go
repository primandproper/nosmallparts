package config

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// The stage timings a config that names none takes.
const (
	DefaultHostGrace      = 15 * time.Second
	DefaultVacancyClaim   = 45 * time.Second
	DefaultCleanupDelay   = 5 * time.Second
	DefaultIdleAfter      = 10 * time.Minute
	DefaultReaperInterval = time.Minute
	DefaultCountdown      = 3 * time.Second
)

// StagesConfig is how long the things that happen on a stage take. Every one
// is a setting so a test can shorten it; a deployment has no reason to.
type StagesConfig struct {
	_ struct{} `json:"-"`

	// HostGrace is how long a host who disconnected has to come back before
	// the role passes to somebody else.
	HostGrace time.Duration `env:"HOST_GRACE" json:"hostGrace,omitempty"`
	// VacancyClaim is how long the audience has to claim a role its holder
	// left mid-performance before the play ends.
	VacancyClaim time.Duration `env:"VACANCY_CLAIM" json:"vacancyClaim,omitempty"`
	// CleanupDelay is how long an empty stage lives before it is deleted.
	CleanupDelay time.Duration `env:"CLEANUP_DELAY" json:"cleanupDelay,omitempty"`
	// IdleAfter is how long a stage with nobody connected and nothing
	// happening survives the reaper.
	IdleAfter time.Duration `env:"IDLE_AFTER" json:"idleAfter,omitempty"`
	// ReaperInterval is how often the reaper looks.
	ReaperInterval time.Duration `env:"REAPER_INTERVAL" json:"reaperInterval,omitempty"`
	// Countdown is the pause between starting a play and its first line.
	Countdown time.Duration `env:"COUNTDOWN" json:"countdown,omitempty"`
}

var _ validation.ValidatableWithContext = (*StagesConfig)(nil)

// EnsureDefaults fills every timing left at zero.
func (cfg *StagesConfig) EnsureDefaults() {
	for _, timing := range []struct {
		field    *time.Duration
		fallback time.Duration
	}{
		{&cfg.HostGrace, DefaultHostGrace},
		{&cfg.VacancyClaim, DefaultVacancyClaim},
		{&cfg.CleanupDelay, DefaultCleanupDelay},
		{&cfg.IdleAfter, DefaultIdleAfter},
		{&cfg.ReaperInterval, DefaultReaperInterval},
		{&cfg.Countdown, DefaultCountdown},
	} {
		if *timing.field <= 0 {
			*timing.field = timing.fallback
		}
	}
}

// ValidateWithContext applies the defaults; after that every timing is
// positive by construction.
func (cfg *StagesConfig) ValidateWithContext(context.Context) error {
	cfg.EnsureDefaults()

	return nil
}
