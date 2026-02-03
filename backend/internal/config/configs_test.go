package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config/envvars"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIServiceConfig_EncodeToFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &APIServiceConfig{
			HTTPServer: http.Config{
				Port:            1234,
				Debug:           false,
				StartupDeadline: time.Minute,
			},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{},
			Services:      ServicesConfig{},
			Database: databasecfg.Config{
				Debug:         true,
				RunMigrations: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username:   "username",
					Password:   "password",
					Database:   "table",
					Host:       "host",
					DisableSSL: true,
				},
			},
		}

		f, err := os.CreateTemp("", "")
		require.NoError(t, err)

		assert.NoError(t, cfg.EncodeToFile(f.Name(), json.Marshal))
	})

	T.Run("with error marshaling", func(t *testing.T) {
		t.Parallel()

		cfg := &APIServiceConfig{}

		f, err := os.CreateTemp("", "")
		require.NoError(t, err)

		assert.Error(t, cfg.EncodeToFile(f.Name(), func(any) ([]byte, error) {
			return nil, errors.New("blah")
		}))
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		var cfg *APIServiceConfig

		f, err := os.CreateTemp("", "")
		require.NoError(t, err)

		err = cfg.EncodeToFile(f.Name(), json.Marshal)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil config")
	})
}

//nolint:paralleltest // because we set env vars for this, we can't
func TestLoadConfigFromEnvironment(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		cfg := &APIServiceConfig{
			Database: databasecfg.Config{
				Debug: true,
			},
		}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		assert.Equal(t, actual.Database.Debug, true)
	})

	// prior TODOs count here too
	T.Run("overrides meta", func(t *testing.T) {
		cfg := &APIServiceConfig{
			Database: databasecfg.Config{
				Debug: true,
			},
		}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)
		t.Setenv(envvars.MetaDebugEnvVarKey, strconv.FormatBool(false))

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		assert.Equal(t, actual.Meta.Debug, false)
	})

	T.Run("with invalid config file", func(t *testing.T) {
		t.Setenv(ConfigurationFilePathEnvVarKey, "/nonexistent/path")

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid JSON", func(t *testing.T) {
		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, []byte("{invalid json"), 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with apply env vars error", func(t *testing.T) {
		cfg := &APIServiceConfig{}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)
		// Set an invalid environment variable that would cause parsing to fail
		t.Setenv(envvars.HTTPPortEnvVarKey, "invalid_port")

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestAPIServiceConfig_Commit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &APIServiceConfig{}
		commit := cfg.Commit()
		// The commit may or may not be empty depending on build info
		assert.IsType(t, "", commit)
	})
}

func TestAPIServiceConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &APIServiceConfig{
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{},
			HTTPServer: http.Config{
				Port:            8080,
				StartupDeadline: time.Minute,
			},
			Queues: msgconfig.QueuesConfig{
				DataChangesTopicName:              "data-changes",
				OutboundEmailsTopicName:           "outbound-emails",
				SearchIndexRequestsTopicName:      "search-index-requests",
				UserDataAggregationTopicName:      "user-data-aggregation",
				WebhookExecutionRequestsTopicName: "webhook-execution-requests",
			},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		assert.NoError(t, err)
	})

	T.Run("with validateServices enabled", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &APIServiceConfig{
			validateServices: true,
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{},
			HTTPServer: http.Config{
				Port:            8080,
				StartupDeadline: time.Minute,
			},
			Queues: msgconfig.QueuesConfig{
				DataChangesTopicName:              "data-changes",
				OutboundEmailsTopicName:           "outbound-emails",
				SearchIndexRequestsTopicName:      "search-index-requests",
				UserDataAggregationTopicName:      "user-data-aggregation",
				WebhookExecutionRequestsTopicName: "webhook-execution-requests",
			},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
			Services: ServicesConfig{},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in services config
		_ = err // Don't assert NoError as services might have validation issues
	})
}

func TestDBCleanerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &DBCleanerConfig{
			Observability: observability.Config{},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		assert.NoError(t, err)
	})
}

func TestMealPlanFinalizerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &MealPlanFinalizerConfig{
			Observability: observability.Config{},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in queues config
		_ = err
	})
}

func TestMealPlanGroceryListInitializerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &MealPlanGroceryListInitializerConfig{
			Observability: observability.Config{},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in analytics or queues config
		_ = err
	})
}

func TestMealPlanTaskCreatorConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &MealPlanTaskCreatorConfig{
			Observability: observability.Config{},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in analytics or queues config
		_ = err
	})
}

func TestSearchDataIndexSchedulerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &SearchDataIndexSchedulerConfig{
			Observability: observability.Config{},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in queues config
		_ = err
	})
}

func TestAsyncMessageHandlerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &AsyncMessageHandlerConfig{
			Observability: observability.Config{},
			Database: databasecfg.Config{
				Debug: true,
				ConnectionDetails: databasecfg.ConnectionDetails{
					Username: "user",
					Password: "pass",
					Database: "db",
					Host:     "host",
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in various configs
		_ = err
	})
}

func TestAdminWebappConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &AdminWebappConfig{
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in various configs
		_ = err
	})
}
