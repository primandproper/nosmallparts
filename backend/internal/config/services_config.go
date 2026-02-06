package config

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/config"
	livekitcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/livekit"
	oauthcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/oauth/config"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/config"

	"github.com/hashicorp/go-multierror"
)

type (
	// ServicesConfig collects the various service configurations.
	ServicesConfig struct {
		_ struct{} `json:"-"`

		LiveKit       livekitcfg.Config       `envPrefix:"LIVEKIT_"        json:"livekit"`
		Users         identitycfg.Config      `envPrefix:"USERS_"          json:"users"`
		DataPrivacy   dataprivacycfg.Config   `envPrefix:"DATA_PRIVACY_"   json:"dataPrivacy"`
		UploadedMedia uploadedmediacfg.Config `envPrefix:"UPLOADED_MEDIA_" json:"uploadedMedia"`
		Auth          authentication.Config   `envPrefix:"AUTH_"           json:"auth"`
		OAuth2Clients oauthcfg.Config         `envPrefix:"OAUTH2_CLIENTS_" json:"oauth2Clients"`
	}
)

// ValidateWithContext validates a APIServiceConfig struct.
func (cfg *ServicesConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validatorsToRun := map[string]func(context.Context) error{
		"Users":         cfg.Users.ValidateWithContext,
		"DataPrivacy":   cfg.DataPrivacy.ValidateWithContext,
		"OAuth2Clients": cfg.OAuth2Clients.ValidateWithContext,
	}

	for name, validator := range validatorsToRun {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}
