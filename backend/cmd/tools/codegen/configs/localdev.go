package main

import (
	"encoding/base64"
	"time"

	tokenscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	analyticscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	featureflagscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/featureflags/config"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/redis"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	loggingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/config"
	logotelgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging/otelgrpc"
	metricscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics/otelgrpc"
	tracingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing/oteltrace"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/chi"
	routingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/algolia"
	textsearchcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/http"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/testutils"
	uploadscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/objectstorage"
	authservice "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/config"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/config"
)

const (
	dockerComposeWorkerQueueAddress = "worker_queue:6379"
	localOAuth2TokenEncryptionKey   = debugCookieHashKey
)

var (
	localdevPostgresDBConnectionDetails = databasecfg.ConnectionDetails{
		Username:   "dbuser",
		Password:   "hunter2",
		Database:   "no-small-parts",
		Host:       "pgdatabase",
		Port:       5432,
		DisableSSL: true,
	}

	localObservabilityConfig = observability.Config{
		Logging: loggingcfg.Config{
			ServiceName: otelServiceName,
			Level:       logging.DebugLevel,
			Provider:    loggingcfg.ProviderOtelSlog,
			OtelSlog: &logotelgrpc.Config{
				CollectorEndpoint: "otel_collector:4317",
				Insecure:          true,
				Timeout:           time.Second * 3,
			},
		},
		Metrics: metricscfg.Config{
			ServiceName: otelServiceName,
			Otel: &otelgrpc.Config{
				Insecure:           true,
				CollectorEndpoint:  "otel_collector:4317",
				CollectionInterval: time.Second,
			},
			Provider: metricscfg.ProviderOtel,
		},
		Tracing: tracingcfg.Config{
			Provider:                  tracingcfg.ProviderOtel,
			ServiceName:               otelServiceName,
			SpanCollectionProbability: 1,
			Otel: &oteltrace.Config{
				Insecure:          true,
				CollectorEndpoint: "otel_collector:4317",
			},
		},
	}

	localRoutingConfig = routingcfg.Config{
		Provider: routingcfg.ProviderChi,
		Chi: &chi.Config{
			ServiceName:            otelServiceName,
			EnableCORSForLocalhost: true,
			SilenceRouteLogging:    false,
		},
	}
)

func buildLocalDevConfig() *config.APIServiceConfig {
	uploadsConfig := uploadscfg.Config{
		Debug: true,
		Storage: objectstorage.Config{
			UploadFilenameKey: "avatar",
			Provider:          objectstorage.FilesystemProvider,
			BucketName:        "avatars",
			FilesystemConfig: &objectstorage.FilesystemConfig{
				RootDirectory: "/uploads",
			},
		},
	}

	return &config.APIServiceConfig{
		Routing: localRoutingConfig,
		Queues: msgconfig.QueuesConfig{
			DataChangesTopicName:              dataChangesTopicName,
			OutboundEmailsTopicName:           outboundEmailsTopicName,
			SearchIndexRequestsTopicName:      searchIndexRequestsTopicName,
			UserDataAggregationTopicName:      userDataAggregationTopicName,
			WebhookExecutionRequestsTopicName: webhookExecutionRequestsTopicName,
		},
		Meta: config.MetaSettings{
			Debug:   true,
			RunMode: developmentEnv,
		},
		Encoding: encoding.Config{
			ContentType: contentTypeJSON,
		},
		Events: msgconfig.Config{
			Consumer: msgconfig.MessageQueueConfig{
				Provider: msgconfig.ProviderRedis,
				Redis: redis.Config{
					QueueAddresses: []string{dockerComposeWorkerQueueAddress},
				},
			},
			Publisher: msgconfig.MessageQueueConfig{
				Provider: msgconfig.ProviderRedis,
				Redis: redis.Config{
					QueueAddresses: []string{dockerComposeWorkerQueueAddress},
				},
			},
		},
		FeatureFlags: featureflagscfg.Config{
			// we're using a noop version of this in localdev right now, but it still tries to instantiate a circuit breaker.
			CircuitBreaker: circuitbreaking.Config{
				Name:                   "feature_flagger",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Analytics: analyticscfg.Config{
			// we're using a noop version of this in localdev right now, but it still tries to instantiate a circuit breaker.
			CircuitBreaker: circuitbreaking.Config{
				Name:                   "feature_flagger",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		TextSearch: textsearchcfg.Config{
			Algolia:  &algolia.Config{},
			Provider: textsearchcfg.AlgoliaProvider,
			CircuitBreaker: circuitbreaking.Config{
				Name:                   "dev_text_searcher",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		HTTPServer: http.Config{
			Debug:           true,
			Port:            defaultHTTPPort,
			StartupDeadline: time.Minute,
		},
		Database: databasecfg.Config{
			Provider:                 databasecfg.ProviderPostgres,
			OAuth2TokenEncryptionKey: localOAuth2TokenEncryptionKey,
			Debug:                    true,
			RunMigrations:            true,
			LogQueries:               true,
			MaxPingAttempts:          maxAttempts,
			PingWaitPeriod:           time.Second,
			ReadConnection:           localdevPostgresDBConnectionDetails,
			WriteConnection:          localdevPostgresDBConnectionDetails,
		},
		Observability: localObservabilityConfig,
		Services: config.ServicesConfig{
			Auth: authservice.Config{
				OAuth2: authservice.OAuth2Config{
					Domain:               "http://localhost:9000",
					AccessTokenLifespan:  time.Hour,
					RefreshTokenLifespan: time.Hour,
					Debug:                false,
				},
				SSO: authservice.SSOConfigs{
					Google: authservice.GoogleSSOConfig{
						CallbackURL: "https://app.nosmall.parts/auth/google/callback",
					},
				},
				Debug:                 true,
				EnableUserSignup:      true,
				MinimumUsernameLength: 3,
				MinimumPasswordLength: 8,
				TokenLifetime:         5 * time.Minute,
				Tokens: tokenscfg.Config{
					Provider:                tokenscfg.ProviderPASETO,
					Audience:                "https://api.nosmall.parts",
					Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
				},
			},
			DataPrivacy: dataprivacycfg.Config{
				Uploads: uploadscfg.Config{
					Storage: objectstorage.Config{
						FilesystemConfig: &objectstorage.FilesystemConfig{RootDirectory: "/tmp"},
						BucketName:       "userdata",
						Provider:         objectstorage.FilesystemProvider,
					},
					Debug: false,
				},
			},
			Users: identitycfg.Config{
				Uploads: uploadsConfig,
			},
			UploadedMedia: uploadedmediacfg.Config{
				Uploads: uploadsConfig,
			},
		},
		// AppleAppSiteAssociation is not needed for local development since
		// Universal Links require HTTPS and a publicly accessible domain.
		// These values are left empty for localdev.
		AppleAppSiteAssociation: config.AppleAppSiteAssociationConfig{
			TeamID:   "",
			BundleID: "",
		},
	}
}
