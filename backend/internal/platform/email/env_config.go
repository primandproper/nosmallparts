package email

import (
	"errors"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/matcornic/hermes/v2"
)

type (

	// EnvironmentConfig is the configuration for a given environment.
	EnvironmentConfig struct {
		baseURL template.URL
		outboundInvitesEmailAddress,
		passwordResetCreationEmailAddress,
		passwordResetRedemptionEmailAddress string
	}
)

var (
	ErrMissingEnvCfg = errors.New("missing environment configuration")

	envConfigsMapHat sync.Mutex
	envConfigsMap    = map[string]*EnvironmentConfig{
		"dev": {
			baseURL:                             "https://www.dinnerdonebetter.dev",
			outboundInvitesEmailAddress:         "noreply@dinnerdonebetter.dev",
			passwordResetCreationEmailAddress:   "noreply@dinnerdonebetter.dev",
			passwordResetRedemptionEmailAddress: "noreply@dinnerdonebetter.dev",
		},
		defaultEnv: {
			baseURL:                             "https://not.real.lol",
			outboundInvitesEmailAddress:         "not@real.lol",
			passwordResetCreationEmailAddress:   "not@real.lol",
			passwordResetRedemptionEmailAddress: "not@real.lol",
		},
	}
)

// BaseURL returns the BaseURL field.
func (c *EnvironmentConfig) BaseURL() template.URL {
	return c.baseURL
}

// OutboundInvitesEmailAddress returns the OutboundInvitesEmailAddress field.
func (c *EnvironmentConfig) OutboundInvitesEmailAddress() string {
	return c.outboundInvitesEmailAddress
}

// PasswordResetCreationEmailAddress returns the passwordResetCreationEmailAddress field.
func (c *EnvironmentConfig) PasswordResetCreationEmailAddress() string {
	return c.passwordResetCreationEmailAddress
}

// PasswordResetRedemptionEmailAddress returns the passwordResetRedemptionEmailAddress field.
func (c *EnvironmentConfig) PasswordResetRedemptionEmailAddress() string {
	return c.passwordResetRedemptionEmailAddress
}

func (c *EnvironmentConfig) BuildHermes() *hermes.Hermes {
	return &hermes.Hermes{
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name: companyName,
			Link: string(c.baseURL),
			// Optional product logo
			Logo:      logoURL,
			Copyright: fmt.Sprintf("Copyright © %d %s. All rights reserved.", time.Now().Year(), companyName),
		},
	}
}

func GetConfigForEnvironment(env string) *EnvironmentConfig {
	envConfigsMapHat.Lock()
	defer envConfigsMapHat.Unlock()

	return envConfigsMap[env]
}
