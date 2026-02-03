package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/auth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/metrics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing"
	routingcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/config"
)

// appleAppSiteAssociation represents the structure of the apple-app-site-association file
// used by iOS for Universal Links.
type appleAppSiteAssociation struct {
	AppLinks appLinks `json:"applinks"`
}

type appLinks struct {
	Apps    []string    `json:"apps"`
	Details []appDetail `json:"details"`
}

type appDetail struct {
	AppID string   `json:"appID"`
	Paths []string `json:"paths"`
}

func ProvideAPIRouter(
	routingConfig routingcfg.Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	authService auth.AuthDataService,
	aasaConfig config.AppleAppSiteAssociationConfig,
) (routing.Router, error) {
	router, err := routingConfig.ProvideRouter(logger, tracerProvider, metricsProvider)
	if err != nil {
		return nil, err
	}

	router.Route("/_ops_", func(metaRouter routing.Router) {
		// Expose a readiness check on /ready
		metaRouter.Get("/live", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		})

		// Expose a readiness check on /ready
		metaRouter.Get("/ready", func(res http.ResponseWriter, req *http.Request) {
			// TODO: check readiness here lol
			res.WriteHeader(http.StatusOK)
		})
	})

	router.Route("/oauth2", func(userRouter routing.Router) {
		userRouter.Get("/authorize", authService.AuthorizeHandler)
		userRouter.Post("/token", authService.TokenHandler)
	})

	// Apple App Site Association for iOS Universal Links
	// This endpoint is required for iOS to recognize the app as a handler for URLs on this domain.
	// See: https://developer.apple.com/documentation/xcode/supporting-associated-domains
	if aasaConfig.TeamID != "" && aasaConfig.BundleID != "" {
		appID := fmt.Sprintf("%s.%s", aasaConfig.TeamID, aasaConfig.BundleID)
		router.Route("/.well-known", func(wellKnownRouter routing.Router) {
			wellKnownRouter.Get("/apple-app-site-association", func(res http.ResponseWriter, req *http.Request) {
				aasa := appleAppSiteAssociation{
					AppLinks: appLinks{
						Apps: []string{},
						Details: []appDetail{
							{
								// App ID format: <TeamID>.<BundleID>
								AppID: appID,
								Paths: []string{
									"/accept_invitation",
									"/accept_invitation/*",
								},
							},
						},
					},
				}

				res.Header().Set("Content-Type", "application/json")
				if err = json.NewEncoder(res).Encode(aasa); err != nil {
					logger.Error("encoding apple-app-site-association", err)
					res.WriteHeader(http.StatusInternalServerError)
				}
			})
		})
	}

	return router, nil
}
