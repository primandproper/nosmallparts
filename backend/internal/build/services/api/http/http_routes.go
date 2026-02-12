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
	livekitsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/livekit"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/playscripts"
	stagessvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/stages"
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
	stagesHandler *stagessvc.Handler,
	playScriptsHandler *playscriptssvc.HTTPHandler,
	livekitHandler *livekitsvc.Handler,
	sessionMiddleware routing.Middleware,
	sessionCookieHandler *SessionCookieHandler,
	registerHandler *RegisterHandler,
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

	// WebSocket endpoint for stages
	router.Route("/ws", func(wsRouter routing.Router) {
		wsRouter.Get("/stages", stagesHandler.ServeHTTP)
	})

	// /api: session-issuing endpoint (Bearer token only, no session cookie required) and session-protected routes
	router.Route("/api", func(apiRouter routing.Router) {
		// Login with username/password; sets session cookie for later requests
		apiRouter.Post("/login", authService.LoginHandler)
		// User registration (unauthenticated)
		apiRouter.Post("/register", registerHandler.ServeHTTP)
		// Login for token: returns JSON with accessToken (for API clients, e.g. integration tests)
		apiRouter.Post("/login-for-token", authService.LoginForTokenHandler)
		// Verify TOTP secret for a user (e.g. after registration)
		apiRouter.Post("/verify-totp", authService.VerifyTOTPHandler)
		// Issue session cookie given Bearer token; frontend stores and sends it on subsequent requests
		apiRouter.Get("/session", sessionCookieHandler.ServeHTTP)
		apiRouter.Post("/session", sessionCookieHandler.ServeHTTP)

		// Routes that use session-from-cookie middleware
		apiWithSession := apiRouter.WithMiddleware(sessionMiddleware)
		apiWithSession.Route("/stages", func(stagesRouter routing.Router) {
			stagesRouter.Post("/", stagesHandler.CreateStageHTTP)
			stagesRouter.Get("/{code}/exists", stagesHandler.CheckStageExists)
		})
		apiWithSession.Route("/play-scripts", func(psRouter routing.Router) {
			psRouter.Get("/public", playScriptsHandler.GetPublicPlayScripts)
			psRouter.Get("/account", playScriptsHandler.GetPlayScriptsForAccount)
			psRouter.Get("/{id}", playScriptsHandler.GetPlayScript)
			psRouter.Patch("/{id}", playScriptsHandler.UpdatePlayScript)
			psRouter.Delete("/{id}", playScriptsHandler.ArchivePlayScript)
			psRouter.Post("/", playScriptsHandler.CreatePlayScript)
		})
		if livekitHandler.Enabled() {
			apiWithSession.Route("/livekit", func(livekitRouter routing.Router) {
				livekitRouter.Post("/token", livekitHandler.GenerateToken)
			})
		}
	})

	return router, nil
}
