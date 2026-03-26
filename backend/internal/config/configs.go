package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	authcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	analyticscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/config"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	emailcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	featureflagscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/featureflags/config"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	pushercfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pusher"
	routingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/config"
	textsearchcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/grpc"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/http"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/objectstorage"
	livekitsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/livekit"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	// DevelopmentRunMode is the run mode for a development environment.
	DevelopmentRunMode runMode = "development"
	// TestingRunMode is the run mode for a testing environment.
	TestingRunMode runMode = "testing"
	// ProductionRunMode is the run mode for a production environment.
	ProductionRunMode runMode = "production"

	EnvVarPrefix = "NO_SMALL_PARTS_"

	// ConfigurationFilePathEnvVarKey is the env var key we use to indicate where the config file is located.
	ConfigurationFilePathEnvVarKey = "CONFIGURATION_FILEPATH"
)

type (
	// runMode describes what method of operation the server is under.
	runMode string

	// CloserFunc calls all io.Closers in the service.
	CloserFunc func()

	configurations interface {
		APIServiceConfig |
			DBCleanerConfig |
			SearchDataIndexSchedulerConfig |
			AsyncMessageHandlerConfig |
			AdminWebappConfig
	}

	// APIServiceConfig configures an instance of the service. It is composed of all the other setting structs.
	APIServiceConfig struct {
		_                       struct{}                      `json:"-"`
		Queues                  msgconfig.QueuesConfig        `envPrefix:"QUEUES_"        json:"queues"`
		LiveKit                 livekitsvc.Config             `envPrefix:"LIVEKIT_"       json:"liveKit,omitempty"`
		Pusher                  pushercfg.Config              `envPrefix:"PUSHER_"        json:"pusher,omitempty"`
		AppleAppSiteAssociation AppleAppSiteAssociationConfig `envPrefix:"AASA_"          json:"appleAppSiteAssociation,omitempty"`
		Routing                 routingcfg.Config             `envPrefix:"ROUTING_"       json:"routing"`
		Encoding                encoding.Config               `envPrefix:"ENCODING_"      json:"encoding"`
		Events                  msgconfig.Config              `envPrefix:"EVENTS_"        json:"events"`
		Observability           observability.Config          `envPrefix:"OBSERVABILITY_" json:"observability"`
		GRPCServer              grpc.Config                   `envPrefix:"GRPC_"          json:"grpc"`
		Meta                    MetaSettings                  `envPrefix:"META_"          json:"meta"`
		Analytics               analyticscfg.Config           `envPrefix:"ANALYTICS_"     json:"analytics"`
		Email                   emailcfg.Config               `envPrefix:"EMAIL_"         json:"email"`
		TextSearch              textsearchcfg.Config          `envPrefix:"SEARCH_"        json:"search"`
		FeatureFlags            featureflagscfg.Config        `envPrefix:"FEATURE_FLAGS_" json:"featureFlags"`
		HTTPServer              http.Config                   `envPrefix:"HTTP_"          json:"http"`
		Auth                    authcfg.Config                `envPrefix:"AUTH_"          json:"auth"`
		Database                databasecfg.Config            `envPrefix:"DATABASE_"      json:"database"`
		Services                ServicesConfig                `envPrefix:"SERVICE_"       json:"services"`
		validateServices        bool
	}

	// DBCleanerConfig configures an instance of the database cleaner job.
	DBCleanerConfig struct {
		_ struct{} `json:"-"`

		Observability observability.Config `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config   `envPrefix:"DATABASE_"      json:"database"`
	}

	// SearchDataIndexSchedulerConfig configures an instance of the search data index scheduler job.
	SearchDataIndexSchedulerConfig struct {
		_ struct{} `json:"-"`

		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}

	// AsyncMessageHandlerConfig configures an instance of the search data index scheduler job.
	AsyncMessageHandlerConfig struct {
		_             struct{}               `json:"-"`
		Storage       objectstorage.Config   `envPrefix:"STORAGE_"       json:"storage"`
		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Encoding      encoding.Config        `envPrefix:"ENCODING_"      json:"encoding"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Email         emailcfg.Config        `envPrefix:"EMAIL_"         json:"email"`
		Analytics     analyticscfg.Config    `envPrefix:"ANALYTICS_"     json:"analytics"`
		Search        textsearchcfg.Config   `envPrefix:"SEARCH_"        json:"search"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}

	APIServiceOAuth2ConnectionConfig struct {
		_ struct{} `json:"-"`

		HTTPAPIServerURL      string `env:"HTTP_API_SERVER_URL"      json:"httpAPIServerURL"`
		GRPCAPIServerURL      string `env:"GRPC_API_SERVER_URL"      json:"grpcAPIServerURL"`
		OAuth2APIClientID     string `env:"OAUTH2_API_CLIENT_ID"     json:"oauth2APIClientID"`
		OAuth2APIClientSecret string `env:"OAUTH2_API_CLIENT_SECRET" json:"oauth2APIClientSecret"`
	}

	NamedCacheConfig struct {
		_ struct{} `json:"-"`

		CacheCapacity uint64        `env:"CACHE_CAPACITY" json:"cacheCapacity"`
		CacheTTL      time.Duration `env:"CACHE_TTL"      json:"cacheTTL"`
	}

	// AppleAppSiteAssociationConfig holds configuration for the apple-app-site-association file
	// used by iOS for Universal Links. See: https://developer.apple.com/documentation/xcode/supporting-associated-domains
	AppleAppSiteAssociationConfig struct {
		_ struct{} `json:"-"`

		// TeamID is the Apple Developer Team ID (e.g., "ABCD1234XY").
		// This can be found in the Apple Developer Portal under Membership.
		TeamID string `env:"TEAM_ID" json:"teamID,omitempty"`
		// BundleID is the iOS app bundle identifier (e.g., "com.nosmallparts.ios").
		BundleID string `env:"BUNDLE_ID" json:"bundleID,omitempty"`
	}

	// AdminWebappConfig configures an instance of the service. It is composed of all the other setting structs.
	AdminWebappConfig struct {
		_ struct{} `json:"-"`

		Cookies              cookies.Config                   `env:"init"                 envPrefix:"COOKIES_"    json:"cookies"`
		APIServiceConnection APIServiceOAuth2ConnectionConfig `envPrefix:"API_SERVICE_"   json:"apiServiceConfig"`
		Routing              routingcfg.Config                `envPrefix:"ROUTING_"       json:"routing"`
		Encoding             encoding.Config                  `envPrefix:"ENCODING_"      json:"encoding"`
		Observability        observability.Config             `envPrefix:"OBSERVABILITY_" json:"observability"`
		Meta                 MetaSettings                     `envPrefix:"META_"          json:"meta"`
		HTTPServer           http.Config                      `envPrefix:"SERVER_"        json:"server"`
	}
)

// EncodeToFile renders your config to a file given your favorite encoder.
func (cfg *APIServiceConfig) EncodeToFile(path string, marshaller func(v any) ([]byte, error)) error {
	if cfg == nil {
		return errors.New("nil config")
	}

	byteSlice, err := marshaller(*cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, byteSlice, 0o600)
}

func (cfg *APIServiceConfig) Commit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for i := range info.Settings {
			if info.Settings[i].Key == "vcs.revision" {
				return info.Settings[i].Value
			}
		}
	}

	return ""
}

var _ validation.ValidatableWithContext = (*APIServiceConfig)(nil)

// ValidateWithContext validates a APIServiceConfig struct.
func (cfg *APIServiceConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Routing":       cfg.Routing.ValidateWithContext,
		"Meta":          cfg.Meta.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
		"Encoding":      cfg.Encoding.ValidateWithContext,
		"Analytics":     cfg.Analytics.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"HTTPServer":    cfg.HTTPServer.ValidateWithContext,
		"Email":         cfg.Email.ValidateWithContext,
		"FeatureFlags":  cfg.FeatureFlags.ValidateWithContext,
		"TextSearch":    cfg.TextSearch.ValidateWithContext,
		// no "Events" here, that's a collection of publisher/subscriber configs that can each optionally be setup
	}

	if cfg.validateServices {
		validators["Services"] = cfg.Services.ValidateWithContext
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*DBCleanerConfig)(nil)

// ValidateWithContext validates a DBCleanerConfig struct.
func (cfg *DBCleanerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*SearchDataIndexSchedulerConfig)(nil)

// ValidateWithContext validates a SearchDataIndexSchedulerConfig struct.
func (cfg *SearchDataIndexSchedulerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*AsyncMessageHandlerConfig)(nil)

// ValidateWithContext validates a AsyncMessageHandlerConfig struct.
func (cfg *AsyncMessageHandlerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Queues":        cfg.Queues.ValidateWithContext,
		"Analytics":     cfg.Analytics.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Email":         cfg.Email.ValidateWithContext,
		"TextSearch":    cfg.Search.ValidateWithContext,
		"Storage":       cfg.Storage.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*AdminWebappConfig)(nil)

// ValidateWithContext validates a AdminWebappConfig struct.
func (cfg *AdminWebappConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Cookies":       cfg.Cookies.ValidateWithContext,
		"Encoding":      cfg.Encoding.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Meta":          cfg.Meta.ValidateWithContext,
		"Routing":       cfg.Routing.ValidateWithContext,
		"HTTPServer":    cfg.HTTPServer.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

func LoadConfigFromEnvironment[T configurations]() (*T, error) {
	configFilepath := os.Getenv(ConfigurationFilePathEnvVarKey)

	configBytes, err := os.ReadFile(configFilepath)
	if err != nil {
		return nil, fmt.Errorf("reading local config file: %w", err)
	}

	var cfg *T
	if err = json.NewDecoder(bytes.NewReader(configBytes)).Decode(&cfg); err != nil || cfg == nil {
		return nil, fmt.Errorf("decoding config file (%s) contents (%s): %w", configFilepath, string(configBytes), err)
	}

	if err = ApplyEnvironmentVariables(cfg); err != nil {
		return nil, fmt.Errorf("applying environment variables: %w", err)
	}

	return cfg, nil
}

func LoadConfigFromPath[T configurations](ctx context.Context, configurationFilepath string) (*T, error) {
	content, err := os.ReadFile(configurationFilepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read api configuration file: %w", err)
	}

	decoder := encoding.ProvideServerEncoderDecoder(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

	var x *T
	if err = decoder.DecodeBytes(ctx, content, &x); err != nil {
		return nil, fmt.Errorf("failed to decode api configuration file: %w", err)
	}

	if err = ApplyEnvironmentVariables(x); err != nil {
		return nil, fmt.Errorf("applying environment variables: %w", err)
	}

	return x, nil
}
