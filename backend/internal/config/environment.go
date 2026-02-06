package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

// EnvironmentConfigSet contains a way of rendering a set of every config for a given environment to a given folder.
type EnvironmentConfigSet struct {
	RootConfig                               *APIServiceConfig
	APIServiceConfigPath                     string
	DBCleanerConfigPath                      string
	MealPlanFinalizerConfigPath              string
	MealPlanGroceryListInitializerConfigPath string
	MealPlanTaskCreatorConfigPath            string
	SearchDataIndexSchedulerConfigPath       string
	AsyncMessageHandlerConfigPath            string
	AdminWebappConfigPath                    string
}

func stringOrDefault(s, defaultStr string) string {
	if s != "" {
		return s
	}
	return defaultStr
}

func renderJSON(obj any, pretty bool) []byte {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(obj, "", "\t")
	} else {
		b, err = json.Marshal(obj)
	}

	if err != nil {
		panic(err)
	}

	return b
}

func writeFile(p string, content []byte) error {
	//nolint:gosec // I want this to be 644 I think
	return os.WriteFile(p, content, 0o0644)
}

const (
	dbcConfigObservabilityServiceName  = "db_cleaner"
	sdisConfigObservabilityServiceName = "search_data_index_scheduler"
	amhConfigObservabilityServiceName  = "async_message_handler"
	awaConfigObservabilityServiceName  = "admin_webapp"
)

func (s *EnvironmentConfigSet) Render(outputDir string, pretty, validate bool) error {
	if err := os.MkdirAll(outputDir, 0o0750); err != nil {
		return err
	}
	errs := &multierror.Error{}

	// write files
	if err := writeFile(
		path.Join(outputDir, stringOrDefault(s.APIServiceConfigPath, "api_service_config.json")),
		renderJSON(s.RootConfig, pretty),
	); err != nil {
		errs = multierror.Append(errs, err)
	}

	dbcConfig := &DBCleanerConfig{
		Observability: s.RootConfig.Observability,
		Database:      s.RootConfig.Database,
	}
	dbcConfig.Observability.Tracing.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Metrics.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Logging.ServiceName = dbcConfigObservabilityServiceName

	sdisConfig := &SearchDataIndexSchedulerConfig{
		Observability: s.RootConfig.Observability,
		Events:        s.RootConfig.Events,
		Database:      s.RootConfig.Database,
		Queues:        s.RootConfig.Queues,
	}
	sdisConfig.Observability.Tracing.ServiceName = sdisConfigObservabilityServiceName
	sdisConfig.Observability.Metrics.ServiceName = sdisConfigObservabilityServiceName
	sdisConfig.Observability.Logging.ServiceName = sdisConfigObservabilityServiceName

	amhConfig := &AsyncMessageHandlerConfig{
		Storage:       s.RootConfig.Services.DataPrivacy.Uploads.Storage,
		Queues:        s.RootConfig.Queues,
		Email:         s.RootConfig.Email,
		Analytics:     s.RootConfig.Analytics,
		Search:        s.RootConfig.TextSearch,
		Events:        s.RootConfig.Events,
		Observability: s.RootConfig.Observability,
		Database:      s.RootConfig.Database,
	}
	amhConfig.Observability.Tracing.ServiceName = amhConfigObservabilityServiceName
	amhConfig.Observability.Metrics.ServiceName = amhConfigObservabilityServiceName
	amhConfig.Observability.Logging.ServiceName = amhConfigObservabilityServiceName

	awaConfig := &AdminWebappConfig{
		// No equivalent for the following configs, I'll just have to rely on env vars to set them for now:
		//	- Cookies
		//	- APIServiceConnection
		//	- APIClientCache
		Cookies: cookies.Config{
			CookieName:            " ",
			Base64EncodedHashKey:  " ",
			Base64EncodedBlockKey: " ",
		},
		Encoding:      s.RootConfig.Encoding,
		Observability: s.RootConfig.Observability,
		Meta:          s.RootConfig.Meta,
		Routing:       s.RootConfig.Routing,
		HTTPServer:    s.RootConfig.HTTPServer,
	}
	awaConfig.Observability.Tracing.ServiceName = awaConfigObservabilityServiceName
	awaConfig.Observability.Metrics.ServiceName = awaConfigObservabilityServiceName
	awaConfig.Observability.Logging.ServiceName = awaConfigObservabilityServiceName

	if validate {
		allConfigs := []validation.ValidatableWithContext{
			s.RootConfig,
			dbcConfig,
			sdisConfig,
			amhConfig,
			awaConfig,
		}
		for i, cfg := range allConfigs {
			if err := cfg.ValidateWithContext(context.Background()); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("validating config %d: %w", i, err))
				continue
			}
		}
	}

	if err := errs.ErrorOrNil(); err != nil {
		return err
	}

	pathToConfigMap := map[string][]byte{
		path.Join(outputDir, stringOrDefault(s.DBCleanerConfigPath, "job_db_cleaner_config.json")):                                 renderJSON(dbcConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.SearchDataIndexSchedulerConfigPath, "job_search_data_index_scheduler_config.json")): renderJSON(sdisConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.AsyncMessageHandlerConfigPath, "async_message_handler_config.json")):                renderJSON(amhConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.AdminWebappConfigPath, "admin_webapp_config.json")):                                 renderJSON(awaConfig, pretty),
	}

	for p, b := range pathToConfigMap {
		if err := writeFile(p, b); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
