package main

import (
	"time"

	tokenscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	analyticscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	emailcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	featureflagscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/featureflags/config"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/pubsub"
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
	uploadscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/uploads/objectstorage"
	authservice "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/config"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/config"
)

const (
	gcpProjectID = "no-small-parts-dev"
)

func buildDevEnvironmentServerConfig() *config.APIServiceConfig {
	uploadsConfig := uploadscfg.Config{
		Debug: true,
		Storage: objectstorage.Config{
			UploadFilenameKey: "avatar",
			Provider:          objectstorage.GCPCloudStorageProvider,
			BucketName:        "media.nosmall.parts",
			BucketPrefix:      "avatars/",
			GCP: &objectstorage.GCPConfig{
				BucketName: "media.nosmall.parts",
			},
		},
	}

	cfg := &config.APIServiceConfig{
		Routing: routingcfg.Config{
			Provider: routingcfg.ProviderChi,
			Chi: &chi.Config{
				ServiceName:            otelServiceName,
				EnableCORSForLocalhost: true,
				SilenceRouteLogging:    false,
			},
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
				Provider: "", // msgconfig.ProviderPubSub,
				PubSub:   pubsub.Config{ProjectID: gcpProjectID},
			},
			Publisher: msgconfig.MessageQueueConfig{
				Provider: "", // msgconfig.ProviderPubSub,
				PubSub:   pubsub.Config{ProjectID: gcpProjectID},
			},
		},
		FeatureFlags: featureflagscfg.Config{
			// we're using a noop version of this in dev right now, but it still tries to instantiate a circuit breaker.
			CircuitBreaker: circuitbreaking.Config{
				Name:                   "feature_flagger",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
		},
		Email: emailcfg.Config{
			Provider: "",
			CircuitBreaker: circuitbreaking.Config{
				Name:                   "dev_emailer",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
			Sendgrid: &sendgrid.Config{},
		},
		Analytics: analyticscfg.Config{
			Provider: "",
			CircuitBreaker: circuitbreaking.Config{
				Name:                   "dev_analytics",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
			Segment: &segment.Config{APIToken: ""},
		},
		HTTPServer: http.Config{
			Debug:           true,
			Port:            defaultHTTPPort,
			StartupDeadline: time.Minute,
		},
		TextSearch: textsearchcfg.Config{
			Algolia: &algolia.Config{},
			CircuitBreaker: circuitbreaking.Config{
				Name:                   "dev_text_searcher",
				ErrorRate:              .5,
				MinimumSampleThreshold: 100,
			},
			Provider: "",
		},
		Database: databasecfg.Config{
			Provider:        databasecfg.ProviderPostgres,
			Debug:           true,
			LogQueries:      true,
			RunMigrations:   true,
			MaxPingAttempts: maxAttempts,
			PingWaitPeriod:  time.Second,
			ReadConnection: databasecfg.ConnectionDetails{
				Username:   "api_db_user",
				Password:   "",
				Database:   "no-small-parts",
				Host:       "",
				Port:       5432,
				DisableSSL: false,
			},
			WriteConnection: databasecfg.ConnectionDetails{
				Username:   "api_db_user",
				Password:   "",
				Database:   "no-small-parts",
				Host:       "",
				Port:       5432,
				DisableSSL: false,
			},
		},
		Observability: observability.Config{
			Logging: loggingcfg.Config{
				ServiceName: otelServiceName,
				Level:       logging.DebugLevel,
				Provider:    loggingcfg.ProviderOtelSlog,
				OtelSlog: &logotelgrpc.Config{
					CollectorEndpoint: internalKubernetesEndpoint("otel-collector-svc", "dev", 4317),
					Insecure:          true,
					Timeout:           2 * time.Second,
				},
			},
			Metrics: metricscfg.Config{
				ServiceName: otelServiceName,
				Provider:    tracingcfg.ProviderOtel,
				Otel: &otelgrpc.Config{
					CollectorEndpoint:  internalKubernetesEndpoint("otel-collector-svc", "dev", 4317),
					CollectionInterval: 30 * time.Second,
					Insecure:           true,
				},
			},
			Tracing: tracingcfg.Config{
				Provider:                  tracingcfg.ProviderOtel,
				ServiceName:               otelServiceName,
				SpanCollectionProbability: 1,
				Otel: &oteltrace.Config{
					CollectorEndpoint: internalKubernetesEndpoint("otel-collector-svc", "dev", 4317),
					Insecure:          true,
				},
			},
		},
		Services: config.ServicesConfig{
			Auth: authservice.Config{
				OAuth2: authservice.OAuth2Config{
					Domain:               "https://nosmall.parts",
					AccessTokenLifespan:  time.Hour,
					RefreshTokenLifespan: time.Hour,
					Debug:                false,
				},
				Debug:                 true,
				EnableUserSignup:      true,
				MinimumUsernameLength: 3,
				MinimumPasswordLength: 8,
				Tokens: tokenscfg.Config{
					Provider:                tokenscfg.ProviderPASETO,
					Audience:                "https://api.nosmall.parts",
					Base64EncodedSigningKey: "",
				},
				TokenLifetime: 5 * time.Minute,
			},
			DataPrivacy: dataprivacycfg.Config{
				Uploads: uploadscfg.Config{
					Storage: objectstorage.Config{
						GCP:        &objectstorage.GCPConfig{BucketName: "userdata.nosmall.parts"},
						BucketName: "userdata.nosmall.parts",
						Provider:   objectstorage.GCPCloudStorageProvider,
					},
					Debug: false,
				},
			},
			Users: identitycfg.Config{
				PublicMediaURLPrefix: "https://media.nosmall.parts/avatars",
				Uploads:              uploadsConfig,
			},
			UploadedMedia: uploadedmediacfg.Config{
				Uploads: uploadsConfig,
			},
		},
		// AppleAppSiteAssociation configures the apple-app-site-association endpoint for iOS Universal Links.
		// Set these values once you have an Apple Developer account.
		// TeamID: Your 10-character Apple Developer Team ID from https://developer.apple.com/account
		// BundleID: Your iOS app bundle identifier (e.g., "com.nosmallparts.ios")
		AppleAppSiteAssociation: config.AppleAppSiteAssociationConfig{
			TeamID:   "", // TODO: Set your Apple Developer Team ID
			BundleID: "com.nosmallparts.ios",
		},
	}

	return cfg
}
