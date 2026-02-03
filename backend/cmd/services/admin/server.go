package main

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing"
	phttp "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/server/http"
)

const (
	o11yName                       = "admin_frontend"
	apiClientContextKey ContextKey = "api_client"
)

type ContextKey string

type AdminFrontendServer struct {
	tracer                          tracing.Tracer
	logger                          logging.Logger
	encoder                         encoding.ServerEncoderDecoder
	cookieManager                   cookies.Manager
	userIDRouteParamFetcher         func(req *http.Request) string
	accountIDRouteParamFetcher      func(req *http.Request) string
	settingIDRouteParamFetcher      func(req *http.Request) string
	oauth2ClientIDRouteParamFetcher func(req *http.Request) string
	waitlistIDRouteParamFetcher     func(req *http.Request) string
	issueReportIDRouteParamFetcher  func(req *http.Request) string
	config                          *config.AdminWebappConfig
	server                          phttp.Server
	componentRenderer               *components.ComponentRenderer
}

func NewAdminFrontendServer(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	encoder encoding.ServerEncoderDecoder,
	rpm routing.RouteParamManager,
	cfg *config.AdminWebappConfig,
) (*AdminFrontendServer, error) {
	cookieMan, err := cookies.NewCookieManager(&cfg.Cookies, tracerProvider)
	if err != nil {
		return nil, err
	}

	metricsProvider, err := cfg.Observability.Metrics.ProvideMetricsProvider(ctx, logger)
	if err != nil {
		return nil, err
	}

	router, err := cfg.Routing.ProvideRouter(logger, tracerProvider, metricsProvider)
	if err != nil {
		return nil, err
	}

	server, err := phttp.ProvideHTTPServer(cfg.HTTPServer, logger, router, tracerProvider)
	if err != nil {
		return nil, err
	}

	s := &AdminFrontendServer{
		tracer:                          tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(o11yName)),
		logger:                          logging.EnsureLogger(logger).WithName(o11yName),
		componentRenderer:               components.NewComponentRenderer(),
		userIDRouteParamFetcher:         rpm.BuildRouteParamStringIDFetcher(userIDURLParamKey),
		accountIDRouteParamFetcher:      rpm.BuildRouteParamStringIDFetcher(accountIDURLParamKey),
		settingIDRouteParamFetcher:      rpm.BuildRouteParamStringIDFetcher(settingIDURLParamKey),
		oauth2ClientIDRouteParamFetcher: rpm.BuildRouteParamStringIDFetcher(oauth2ClientIDURLParamKey),
		waitlistIDRouteParamFetcher:     rpm.BuildRouteParamStringIDFetcher(waitlistIDURLParamKey),
		issueReportIDRouteParamFetcher:  rpm.BuildRouteParamStringIDFetcher(issueReportIDURLParamKey),
		cookieManager:                   cookieMan,
		encoder:                         encoder,
		config:                          cfg,
		server:                          server,
	}

	s.setupRoutes(router)

	return s, nil
}

func (s *AdminFrontendServer) Serve() {
	s.server.Serve()
}
