package main

import (
	"time"

	tokenscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	analyticscfg "github.com/verygoodsoftwarenotvirus/platform/v4/analytics/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/analytics/segment"
	"github.com/verygoodsoftwarenotvirus/platform/v4/circuitbreaking"
	databasecfg "github.com/verygoodsoftwarenotvirus/platform/v4/database/config"
	emailcfg "github.com/verygoodsoftwarenotvirus/platform/v4/email/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	featureflagscfg "github.com/verygoodsoftwarenotvirus/platform/v4/featureflags/config"
	msgconfig "github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/messagequeue/pubsub"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	loggingcfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging/config"
	logotelgrpc "github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging/otelgrpc"
	metricscfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/metrics/otelgrpc"
	tracingcfg "github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing/oteltrace"
	"github.com/verygoodsoftwarenotvirus/platform/v4/routing/chi"
	routingcfg "github.com/verygoodsoftwarenotvirus/platform/v4/routing/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/search/text/algolia"
	textsearchcfg "github.com/verygoodsoftwarenotvirus/platform/v4/search/text/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/server/grpc"
	"github.com/verygoodsoftwarenotvirus/platform/v4/server/http"
	uploadscfg "github.com/verygoodsoftwarenotvirus/platform/v4/uploads/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/uploads/objectstorage"
	authservice "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/config"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/config"
)

const (
	prodGCPProjectID   = "no-small-parts"
	prodMediaBucket    = "no-small-parts-media"
	prodUserDataBucket = "no-small-parts-userdata"
)

func buildProdEnvironmentServerConfig() *config.APIServiceConfig {
	uploadsConfig := uploadscfg.Config{
		Debug: true,
		Storage: objectstorage.Config{
			UploadFilenameKey: "avatar",
			Provider:          objectstorage.GCPCloudStorageProvider,
			BucketName:        prodMediaBucket,
			BucketPrefix:      "avatars/",
			GCP: &objectstorage.GCPConfig{
				BucketName: prodMediaBucket,
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
				Provider: "",
				PubSub:   pubsub.Config{ProjectID: prodGCPProjectID},
			},
			Publisher: msgconfig.MessageQueueConfig{
				Provider: "",
				PubSub:   pubsub.Config{ProjectID: prodGCPProjectID},
			},
		},
		FeatureFlags: featureflagscfg.Config{
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
			SourceConfig: analyticscfg.SourceConfig{
				Provider: "",
				CircuitBreaker: circuitbreaking.Config{
					Name:                   "dev_analytics",
					ErrorRate:              .5,
					MinimumSampleThreshold: 100,
				},
				Segment: &segment.Config{APIToken: ""},
			},
		},
		HTTPServer: http.Config{
			Debug:           true,
			Port:            defaultHTTPPort,
			StartupDeadline: time.Minute,
		},
		GRPCServer: grpc.Config{
			Port: defaultGRPCPort,
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
					CollectorEndpoint: internalKubernetesEndpoint("otel-collector-svc", "prod", 4317),
					Insecure:          true,
					Timeout:           2 * time.Second,
				},
			},
			Metrics: metricscfg.Config{
				ServiceName: otelServiceName,
				Provider:    tracingcfg.ProviderOtel,
				Otel: &otelgrpc.Config{
					CollectorEndpoint:  internalKubernetesEndpoint("otel-collector-svc", "prod", 4317),
					CollectionInterval: 30 * time.Second,
					Insecure:           true,
				},
			},
			Tracing: tracingcfg.Config{
				Provider:                  tracingcfg.ProviderOtel,
				ServiceName:               otelServiceName,
				SpanCollectionProbability: 1,
				Otel: &oteltrace.Config{
					CollectorEndpoint: internalKubernetesEndpoint("otel-collector-svc", "prod", 4317),
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
					MaxAccessTokenLifetime:  7 * 24 * time.Hour,
					MaxRefreshTokenLifetime: 30 * 24 * time.Hour,
				},
				TokenLifetime: 5 * time.Minute,
			},
			DataPrivacy: dataprivacycfg.Config{
				Uploads: uploadscfg.Config{
					Storage: objectstorage.Config{
						GCP:        &objectstorage.GCPConfig{BucketName: prodUserDataBucket},
						BucketName: prodUserDataBucket,
						Provider:   objectstorage.GCPCloudStorageProvider,
					},
					Debug: false,
				},
			},
			Users: identitycfg.Config{
				PublicMediaURLPrefix: "https://storage.googleapis.com/no-small-parts-media/avatars",
				Uploads:              uploadsConfig,
			},
			UploadedMedia: uploadedmediacfg.Config{
				Uploads: uploadsConfig,
			},
		},
		AppleAppSiteAssociation: config.AppleAppSiteAssociationConfig{
			TeamID:   "",
			BundleID: "com.nosmallparts.ios",
		},
	}

	return cfg
}
