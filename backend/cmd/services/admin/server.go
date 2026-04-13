package main

import (
	"context"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/cookies"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/platform/v4/encoding"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/routing"
	phttp "github.com/verygoodsoftwarenotvirus/platform/v4/server/http"
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
	playScriptIDRouteParamFetcher   func(req *http.Request) string
	lineIDRouteParamFetcher         func(req *http.Request) string
	tagIDRouteParamFetcher          func(req *http.Request) string
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

	server, err := phttp.ProvideHTTPServer(cfg.HTTPServer, logger, router, tracerProvider, "admin")
	if err != nil {
		return nil, err
	}

	s := &AdminFrontendServer{
		tracer:                          tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:                          logging.NewNamedLogger(logger, o11yName),
		componentRenderer:               components.NewComponentRenderer(),
		userIDRouteParamFetcher:         rpm.BuildRouteParamStringIDFetcher(userIDURLParamKey),
		accountIDRouteParamFetcher:      rpm.BuildRouteParamStringIDFetcher(accountIDURLParamKey),
		settingIDRouteParamFetcher:      rpm.BuildRouteParamStringIDFetcher(settingIDURLParamKey),
		oauth2ClientIDRouteParamFetcher: rpm.BuildRouteParamStringIDFetcher(oauth2ClientIDURLParamKey),
		waitlistIDRouteParamFetcher:     rpm.BuildRouteParamStringIDFetcher(waitlistIDURLParamKey),
		issueReportIDRouteParamFetcher:  rpm.BuildRouteParamStringIDFetcher(issueReportIDURLParamKey),
		playScriptIDRouteParamFetcher:   rpm.BuildRouteParamStringIDFetcher(playScriptIDURLParamKey),
		lineIDRouteParamFetcher:         rpm.BuildRouteParamStringIDFetcher(lineIDURLParamKey),
		tagIDRouteParamFetcher:          rpm.BuildRouteParamStringIDFetcher(tagIDURLParamKey),
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
