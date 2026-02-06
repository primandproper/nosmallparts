module github.com/verygoodsoftwarenotvirus/nosmallparts/backend

go 1.25.5

require (
	cloud.google.com/go/pubsub/v2 v2.0.0
	github.com/BurntSushi/toml v1.4.0
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.26.0
	github.com/GuiaBolso/darwin v0.0.0-20191218124601-fd6d2aa3d244
	github.com/XSAM/otelsql v0.37.0
	github.com/alexedwards/argon2id v1.0.0
	github.com/algolia/algoliasearch-client-go/v3 v3.31.4
	github.com/aws/aws-sdk-go v1.55.6
	github.com/aws/aws-sdk-go-v2/service/s3 v1.77.1
	github.com/boombuler/barcode v1.0.2
	github.com/brianvoe/gofakeit/v5 v5.11.2
	github.com/brianvoe/gofakeit/v7 v7.2.1
	github.com/caarlos0/env/v11 v11.3.1
	github.com/codemodus/kace v0.5.1
	github.com/cristalhq/builq v0.15.0
	github.com/elastic/go-elasticsearch/v8 v8.17.1
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/cors v1.2.1
	github.com/go-faker/faker/v4 v4.6.0
	github.com/go-oauth2/oauth2/v4 v4.5.2
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/wire v0.6.0
	github.com/gorilla/securecookie v1.1.2
	github.com/gorilla/websocket v1.5.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/jackc/pgx/v5 v5.7.2
	github.com/keith-turner/ecoji/v2 v2.0.1
	github.com/klauspost/compress v1.18.0
	github.com/launchdarkly/go-sdk-common/v3 v3.2.0
	github.com/launchdarkly/go-server-sdk/v6 v6.2.1
	github.com/lib/pq v1.10.9
	github.com/livekit/protocol v1.44.0
	github.com/luna-duclos/instrumentedsql v1.1.3
	github.com/mailgun/mailgun-go/v4 v4.22.1
	github.com/mailjet/mailjet-apiv3-go/v4 v4.0.7
	github.com/markbates/goth v1.80.0
	github.com/matcornic/hermes/v2 v2.1.0
	github.com/mikespook/gorbac/v2 v2.3.3
	github.com/mitchellh/go-server-timing v1.0.1
	github.com/mssola/useragent v1.0.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/o1egl/paseto/v2 v2.1.1
	github.com/pinecone-io/go-pinecone/v2 v2.2.0
	github.com/posthog/posthog-go v1.3.1
	github.com/pquerna/otp v1.4.0
	github.com/qdrant/go-client v1.13.0
	github.com/riandyrn/otelchi v0.12.1
	github.com/rs/xid v1.6.0
	github.com/rs/zerolog v1.33.0
	github.com/rubyist/circuitbreaker v2.2.1+incompatible
	github.com/rudderlabs/analytics-go/v4 v4.2.1
	github.com/samber/slog-multi v1.4.0
	github.com/segmentio/analytics-go/v3 v3.3.0
	github.com/sendgrid/rest v2.6.9+incompatible
	github.com/sendgrid/sendgrid-go v3.16.0+incompatible
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	github.com/stripe/stripe-go/v75 v75.11.0
	github.com/testcontainers/testcontainers-go v0.34.0
	github.com/testcontainers/testcontainers-go/modules/elasticsearch v0.34.0
	github.com/testcontainers/testcontainers-go/modules/gcloud v0.34.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.34.0
	github.com/testcontainers/testcontainers-go/modules/redis v0.34.0
	github.com/wagslane/go-password-validator v0.3.0
	go.opentelemetry.io/contrib/bridges/otelslog v0.9.0
	go.opentelemetry.io/contrib/instrumentation/host v0.59.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.59.0
	go.opentelemetry.io/otel v1.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.10.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.34.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.39.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.34.0
	go.opentelemetry.io/otel/log v0.15.0
	go.opentelemetry.io/otel/metric v1.39.0
	go.opentelemetry.io/otel/sdk v1.39.0
	go.opentelemetry.io/otel/sdk/log v0.15.0
	go.opentelemetry.io/otel/sdk/metric v1.39.0
	go.opentelemetry.io/otel/trace v1.39.0
	go.uber.org/automaxprocs v1.6.0
	go.uber.org/zap v1.27.0
	gocloud.dev v0.40.0
	golang.org/x/crypto v0.45.0
	golang.org/x/net v0.47.0
	golang.org/x/oauth2 v0.32.0
	golang.org/x/text v0.31.0
	google.golang.org/api v0.252.0
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
	gopkg.in/matryer/try.v1 v1.0.0-20150601225556-312d2599e12e
	maragu.dev/gomponents v1.2.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.6-20250625184727-c923a0c2a132.1 // indirect
	buf.build/go/protovalidate v0.13.1 // indirect
	buf.build/go/protoyaml v0.6.0 // indirect
	cel.dev/expr v0.24.0 // indirect
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.17.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	cloud.google.com/go/longrunning v0.7.0 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	cloud.google.com/go/storage v1.56.1 // indirect
	cloud.google.com/go/trace v1.11.7 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20230124172434-306776ec8161 // indirect
	github.com/GoogleCloudPlatform/grpc-gcp-go/grpcgcp v1.5.2 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.30.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.53.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.53.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/PuerkitoBio/goquery v1.10.2 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/apache/arrow/go/v15 v15.0.2 // indirect
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.10 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.7 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.60 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.29 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.17.63 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.33 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.6.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.14 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.24.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.28.15 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.15 // indirect
	github.com/aws/smithy-go v1.22.3 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/cenk/backoff v2.2.1+incompatible // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cheekybits/is v0.0.0-20150225183255-68e9c0620927 // indirect
	github.com/cncf/xds/go v0.0.0-20251022180443-0feb69152e9f // indirect
	github.com/containerd/containerd v1.7.18 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/cznic/mathutil v0.0.0-20181122101859-297441e03548 // indirect
	github.com/cznic/ql v1.2.0 // indirect
	github.com/cznic/sortutil v0.0.0-20181122101858-f5f958428db8 // indirect
	github.com/cznic/strutil v0.0.0-20181122101858-275e90344537 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dennwc/iters v1.1.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/docker v27.1.1+incompatible // indirect
	github.com/docker/go-connections v0.5.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/ebitengine/purego v0.8.2 // indirect
	github.com/edsrzf/mmap-go v1.1.0 // indirect
	github.com/elastic/elastic-transport-go/v8 v8.6.1 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.35.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/frostbyte73/core v0.1.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/gammazero/deque v1.1.0 // indirect
	github.com/go-jose/go-jose/v3 v3.0.4 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/gddo v0.0.0-20210115222349-20d68f94ee1f // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/cel-go v0.25.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/gorilla/sessions v1.4.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jaytaylor/html2text v0.0.0-20230321000545-74c2419ad056 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/jxskiss/base62 v1.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/launchdarkly/ccache v1.1.0 // indirect
	github.com/launchdarkly/eventsource v1.8.0 // indirect
	github.com/launchdarkly/go-jsonstream/v3 v3.1.0 // indirect
	github.com/launchdarkly/go-sdk-events/v2 v2.0.2 // indirect
	github.com/launchdarkly/go-semver v1.0.3 // indirect
	github.com/launchdarkly/go-server-sdk-evaluation/v2 v2.0.2 // indirect
	github.com/lithammer/shortuuid/v4 v4.2.0 // indirect
	github.com/livekit/mageutil v0.0.0-20250511045019-0f1ff63f7731 // indirect
	github.com/livekit/psrpc v0.7.1 // indirect
	github.com/lufia/plan9stats v0.0.0-20240909124753-873cd0166683 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailgun/errors v0.4.0 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/matryer/try v0.0.0-20161228173917-9ac251b645a2 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mikespook/gorbac v2.3.0+incompatible // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/sys/user v0.1.0 // indirect
	github.com/moby/term v0.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nats-io/nats.go v1.43.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nxadm/tail v1.4.11 // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/onsi/gomega v1.29.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/peterbourgon/g2s v0.0.0-20170223122336-d4e7ad98afea // indirect
	github.com/pion/datachannel v1.5.10 // indirect
	github.com/pion/dtls/v3 v3.0.6 // indirect
	github.com/pion/ice/v4 v4.0.10 // indirect
	github.com/pion/interceptor v0.1.40 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.15 // indirect
	github.com/pion/rtp v1.8.19 // indirect
	github.com/pion/sctp v1.8.39 // indirect
	github.com/pion/sdp/v3 v3.0.14 // indirect
	github.com/pion/srtp/v3 v3.0.6 // indirect
	github.com/pion/stun/v3 v3.0.0 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/turn/v4 v4.0.2 // indirect
	github.com/pion/webrtc/v4 v4.1.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/puzpuzpuz/xsync/v3 v3.5.1 // indirect
	github.com/redis/go-redis/v9 v9.11.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/samber/lo v1.49.1 // indirect
	github.com/segmentio/backo-go v1.1.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shirou/gopsutil/v3 v3.23.12 // indirect
	github.com/shirou/gopsutil/v4 v4.25.1 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/spiffe/go-spiffe/v2 v2.6.0 // indirect
	github.com/ssor/bom v0.0.0-20170718123548-6386211fdfcf // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.9.0 // indirect
	github.com/twitchtv/twirp v8.1.3+incompatible // indirect
	github.com/valyala/fasthttp v1.49.0 // indirect
	github.com/vanng822/css v1.0.1 // indirect
	github.com/vanng822/go-premailer v1.23.0 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.einride.tech/aip v0.73.0 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.38.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/otel/sdk/log/logtest v0.15.0 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/telemetry v0.0.0-20251008203120-078029d740a8 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	google.golang.org/genproto v0.0.0-20251014184007-4626949a642f // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gotest.tools/v3 v3.5.2 // indirect
)
