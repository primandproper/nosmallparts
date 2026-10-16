// Package api is the composition root of `nosmallparts serve`.
//
// It follows fijo's and DDB's shape: BuildInjector registers everything with a
// samber/do injector, and NewServer resolves what the process runs. The
// platform half is one call, service.Register, which walks the config and
// registers every subsystem it names; service.RegisterTransports mounts the
// sign-in surface and puts the script library beside it. What this file
// registers by hand is what no environment variable can express: the migrator,
// the password hasher and policy, the interceptors and the requirements they
// enforce, the script library, and the guest identity route.
package api

import (
	"context"
	"net/http"

	"github.com/primandproper/nosmallparts/backend/internal/admins"
	"github.com/primandproper/nosmallparts/backend/internal/config"
	"github.com/primandproper/nosmallparts/backend/internal/database/migrations"
	"github.com/primandproper/nosmallparts/backend/internal/domain/scripts"
	"github.com/primandproper/nosmallparts/backend/internal/domain/stages"
	"github.com/primandproper/nosmallparts/backend/internal/guests"
	"github.com/primandproper/nosmallparts/backend/internal/seed"
	scriptsgrpc "github.com/primandproper/nosmallparts/backend/internal/services/scripts/grpc"
	stagesgrpc "github.com/primandproper/nosmallparts/backend/internal/services/stages/grpc"
	"github.com/primandproper/nosmallparts/backend/internal/services/stages/ws"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	"github.com/primandproper/platform-go/v15/identity"
	identitycfg "github.com/primandproper/platform-go/v15/identity/config"
	"github.com/primandproper/platform-go/v15/service"
	"github.com/primandproper/primitives-go/v2/authentication"
	"github.com/primandproper/primitives-go/v2/authentication/argon2"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	grpcerrors "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/ratelimiting"
	"github.com/primandproper/primitives-go/v2/routing"
	grpcserver "github.com/primandproper/primitives-go/v2/server/grpc"

	"github.com/samber/do/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// IdentityRoute is where a visitor learns who they are.
const IdentityRoute = "/api/identity"

// BuildInjector registers everything `nosmallparts serve` is made of.
func BuildInjector(ctx context.Context, cfg *config.Config) (*do.RootScope, error) {
	if cfg == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil config")
	}

	if err := cfg.ValidateForServing(ctx); err != nil {
		return nil, err
	}

	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	// The migrator the database client runs on boot when RunMigrations is set.
	// Registered before the platform so the client's provider finds it.
	do.Provide(i, func(i do.Injector) (database.Migrator, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		return migrations.NewMigrator(pillars.Logger)
	})

	service.Register(i, &cfg.Config)
	registerClientSafe()

	registerAuth(i)
	registerScripts(i, cfg)
	registerStages(i, cfg)
	registerInterceptors(i)

	service.RegisterTransports(i, &service.Transports{
		Extractor: signingrpc.PrincipalFromContext,
		TenantOf:  service.DirectoryTenant,
		Registrations: []grpcserver.RegistrationFunc{
			func(srv *grpc.Server) { do.MustInvoke[*scriptsgrpc.Server](i).RegisterOn(srv) },
			func(srv *grpc.Server) { do.MustInvoke[*stagesgrpc.Server](i).RegisterOn(srv) },
		},
		// The identity surface is account and membership administration, which
		// nothing here offers; sign-in's Register is how people join.
		Skip: []service.Surface{service.SurfaceIdentity},
	})

	return i, nil
}

// NewServer builds the injector and the service over it.
func NewServer(ctx context.Context, cfg *config.Config) (*service.Service, error) {
	i, err := BuildInjector(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return New(i)
}

// New mounts the routes that are not gRPC on an injector BuildInjector
// returned, and builds the service over it. It is separate from NewServer so a
// caller holding the injector can reach what the service is built from, such
// as the address a server bound. Our own servers are resolved ahead of
// service.New so a misconfiguration in one is an error with a name rather than
// a panic inside a registration function.
func New(i do.Injector) (*service.Service, error) {
	if _, err := do.Invoke[*scriptsgrpc.Server](i); err != nil {
		return nil, platformerrors.Wrap(err, "building the scripts server")
	}

	if _, err := do.Invoke[*stagesgrpc.Server](i); err != nil {
		return nil, platformerrors.Wrap(err, "building the stages server")
	}

	hub, err := do.Invoke[*ws.Hub](i)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the stage socket")
	}

	limiter, err := do.Invoke[*ratelimiting.InMemoryRateLimiter](i)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the stage lookup limiter")
	}

	if err = mountHTTPRoutes(i, hub); err != nil {
		return nil, err
	}

	return service.New(i,
		service.WithRunners(hub),
		service.WithFlush("stage lookup limiter", func(context.Context) error { return limiter.Close() }),
	)
}

// Seed writes the sample plays into the configured database, migrating it
// first, and releases what it opened. It starts no server.
func Seed(ctx context.Context, cfg *config.Config) (err error) {
	i, err := BuildInjector(ctx, cfg)
	if err != nil {
		return err
	}

	client, err := do.Invoke[database.Client](i)
	if err != nil {
		return platformerrors.Wrap(err, "opening the database")
	}

	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			err = platformerrors.Join(err, platformerrors.Wrap(closeErr, "closing the database"))
		}
	}()

	seeder, err := do.Invoke[*seed.Seeder](i)
	if err != nil {
		return platformerrors.Wrap(err, "building the seeder")
	}

	return seeder.Seed(ctx)
}

// registerAuth registers the sign-in layer's application half: the directory
// service sign-in registers people through, the hasher and the password rule,
// and the extractor every request's caller is resolved by.
func registerAuth(i do.Injector) {
	identitycfg.RegisterService(i)

	do.Provide(i, func(_ do.Injector) (authentication.Authenticator, error) {
		return argon2.NewArgon2Authenticator(), nil
	})

	do.ProvideValue(i, signin.PasswordPolicy(PasswordPolicy))
	do.ProvideValue(i, signin.RegistrationPolicy(RegistrationPolicy))

	do.Provide(i, func(i do.Injector) (*signingrpc.PrincipalExtractor, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		issuer, err := do.Invoke[tokens.Issuer](i)
		if err != nil {
			return nil, err
		}

		client, err := do.Invoke[database.Client](i)
		if err != nil {
			return nil, err
		}

		directory, err := do.Invoke[identity.Store](i)
		if err != nil {
			return nil, err
		}

		signIns, err := do.Invoke[*signin.Service](i)
		if err != nil {
			return nil, err
		}

		// The sign-in check is what makes a sign-out stick: without it an
		// access token outlives the login it came from until it expires.
		return signingrpc.NewPrincipalExtractor(issuer, client, directory,
			signingrpc.WithSignInCheck(signIns),
			signingrpc.WithExtractorPillars(pillars),
		)
	})
}

// registerScripts registers the script library: its store, its rules, the
// admin list they consult, and its gRPC server.
func registerScripts(i do.Injector, cfg *config.Config) {
	do.ProvideValue(i, admins.New(cfg.AdminEmails))

	do.Provide(i, func(i do.Injector) (scripts.Store, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		client, err := do.Invoke[database.Client](i)
		if err != nil {
			return nil, err
		}

		store, err := scripts.NewSQLStore(client, scripts.WithPillars(pillars))
		if err != nil {
			return nil, err
		}

		return store, nil
	})

	do.Provide(i, func(i do.Injector) (*scripts.Service, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		client, err := do.Invoke[database.Client](i)
		if err != nil {
			return nil, err
		}

		store, err := do.Invoke[scripts.Store](i)
		if err != nil {
			return nil, err
		}

		return scripts.NewService(client, store, scripts.WithServicePillars(pillars))
	})

	do.Provide(i, func(i do.Injector) (*seed.Seeder, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		client, err := do.Invoke[database.Client](i)
		if err != nil {
			return nil, err
		}

		directory, err := do.Invoke[identity.Store](i)
		if err != nil {
			return nil, err
		}

		registrar, err := do.Invoke[*identity.Service](i)
		if err != nil {
			return nil, err
		}

		store, err := do.Invoke[scripts.Store](i)
		if err != nil {
			return nil, err
		}

		return seed.NewSeeder(client, directory, registrar, store, cfg.SignIn.DefaultOwnerRoles, pillars)
	})

	do.Provide(i, func(i do.Injector) (*scriptsgrpc.Server, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		svc, err := do.Invoke[*scripts.Service](i)
		if err != nil {
			return nil, err
		}

		return scriptsgrpc.NewServer(svc, do.MustInvoke[*admins.List](i), signingrpc.PrincipalFromContext, pillars)
	})
}

// registerStages registers the stage half: its store, StagesService, the
// limiter the code lookup spends from, and the stage socket.
func registerStages(i do.Injector, cfg *config.Config) {
	do.Provide(i, func(i do.Injector) (stages.Store, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		client, err := do.Invoke[database.Client](i)
		if err != nil {
			return nil, err
		}

		store, err := stages.NewSQLStore(client, stages.WithPillars(pillars))
		if err != nil {
			return nil, err
		}

		return store, nil
	})

	do.Provide(i, func(_ do.Injector) (*ratelimiting.InMemoryRateLimiter, error) {
		return stagesgrpc.NewExistsLimiter()
	})

	do.Provide(i, func(i do.Injector) (*stagesgrpc.Server, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		client, err := do.Invoke[database.Client](i)
		if err != nil {
			return nil, err
		}

		store, err := do.Invoke[stages.Store](i)
		if err != nil {
			return nil, err
		}

		library, err := do.Invoke[*scripts.Service](i)
		if err != nil {
			return nil, err
		}

		return stagesgrpc.NewServer(client, store, library, signingrpc.PrincipalFromContext, pillars)
	})

	do.Provide(i, func(i do.Injector) (*ws.Hub, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		client, err := do.Invoke[database.Client](i)
		if err != nil {
			return nil, err
		}

		store, err := do.Invoke[stages.Store](i)
		if err != nil {
			return nil, err
		}

		plays, err := do.Invoke[scripts.Store](i)
		if err != nil {
			return nil, err
		}

		library, err := do.Invoke[*scripts.Service](i)
		if err != nil {
			return nil, err
		}

		extractor, err := do.Invoke[*signingrpc.PrincipalExtractor](i)
		if err != nil {
			return nil, err
		}

		return ws.NewHub(client, store, plays, library, extractor, ws.Timings{
			HostGrace:      cfg.Stages.HostGrace,
			VacancyClaim:   cfg.Stages.VacancyClaim,
			CleanupDelay:   cfg.Stages.CleanupDelay,
			IdleAfter:      cfg.Stages.IdleAfter,
			ReaperInterval: cfg.Stages.ReaperInterval,
			Countdown:      cfg.Stages.Countdown,
		}, pillars)
	})
}

// healthMethods are the gRPC health service's, which a probe calls with no
// credential.
var healthMethods = []string{
	grpc_health_v1.Health_Check_FullMethodName,
	grpc_health_v1.Health_List_FullMethodName,
	grpc_health_v1.Health_Watch_FullMethodName,
}

// AuthenticationRequirements declares what every method on the server needs
// of the caller's credential. A method not declared is refused, which is how a
// new RPC that forgot its requirement fails on its first call rather than
// mounting open.
func AuthenticationRequirements() (*signingrpc.AuthenticationRequirements, error) {
	b := signingrpc.RequireAuthentication(signingrpc.NewAuthenticationRequirements())
	scriptsgrpc.RequireAuthentication(b)
	stagesgrpc.RequireAuthentication(b)
	b.Declare(signingrpc.AuthenticationAnonymous, healthMethods...)

	return b.Build()
}

// AuthorizationRequirements declares the permission every method needs. Only
// sign-in's operator methods need one, and nobody here holds it, so they are
// refused to everyone; everything else is public to the enforcer and decided
// by its own service.
func AuthorizationRequirements() (*authzgrpc.Requirements, error) {
	b := authzgrpc.NewRequirements()
	signingrpc.Require(b)
	scriptsgrpc.Require(b)
	stagesgrpc.Require(b)

	for _, method := range healthMethods {
		b.Public(method)
	}

	return b.Build()
}

// registerInterceptors registers the chains the gRPC server is built from,
// after the recovery it installs for itself: the stage lookup's rate limit,
// which needs nobody resolved and so refuses a flood before it costs a token
// check; authentication; authorization, which reads what authentication
// resolved; and the error encoder that turns a platform sentinel into a
// considered status code.
func registerInterceptors(i do.Injector) {
	do.Provide(i, func(_ do.Injector) (*signingrpc.AuthenticationRequirements, error) {
		return AuthenticationRequirements()
	})

	do.Provide(i, func(i do.Injector) (*authzgrpc.Enforcer, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		reqs, err := AuthorizationRequirements()
		if err != nil {
			return nil, err
		}

		extractor, err := do.Invoke[*signingrpc.PrincipalExtractor](i)
		if err != nil {
			return nil, err
		}

		return authzgrpc.NewEnforcer(reqs, extractor.Grants,
			authzgrpc.WithLogger(pillars.Logger),
			authzgrpc.WithMetricsProvider(pillars.MetricsProvider),
		)
	})

	do.Provide(i, func(i do.Injector) ([]grpc.UnaryServerInterceptor, error) {
		extractor, reqs, enforcer, err := invokeChain(i)
		if err != nil {
			return nil, err
		}

		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		limiter, err := do.Invoke[*ratelimiting.InMemoryRateLimiter](i)
		if err != nil {
			return nil, err
		}

		existsLimit, err := stagesgrpc.ExistsRateLimit(limiter, pillars)
		if err != nil {
			return nil, err
		}

		return []grpc.UnaryServerInterceptor{
			existsLimit,
			extractor.UnaryServerInterceptor(reqs),
			enforcer.UnaryServerInterceptor(),
			grpcerrors.UnaryErrorEncodingInterceptor(),
		}, nil
	})

	do.Provide(i, func(i do.Injector) ([]grpc.StreamServerInterceptor, error) {
		extractor, reqs, enforcer, err := invokeChain(i)
		if err != nil {
			return nil, err
		}

		return []grpc.StreamServerInterceptor{
			extractor.StreamServerInterceptor(reqs),
			enforcer.StreamServerInterceptor(),
			grpcerrors.StreamErrorEncodingInterceptor(),
		}, nil
	})
}

func invokeChain(i do.Injector) (*signingrpc.PrincipalExtractor, *signingrpc.AuthenticationRequirements, *authzgrpc.Enforcer, error) {
	extractor, err := do.Invoke[*signingrpc.PrincipalExtractor](i)
	if err != nil {
		return nil, nil, nil, err
	}

	reqs, err := do.Invoke[*signingrpc.AuthenticationRequirements](i)
	if err != nil {
		return nil, nil, nil, err
	}

	enforcer, err := do.Invoke[*authzgrpc.Enforcer](i)
	if err != nil {
		return nil, nil, nil, err
	}

	return extractor, reqs, enforcer, nil
}

// mountHTTPRoutes puts the routes that are not gRPC on the router: the guest
// identity route, which resolves the caller through the same extractor the
// gRPC server does so a signed-in visitor is answered as themselves, and the
// stage socket, which settles who is connecting itself.
func mountHTTPRoutes(i do.Injector, hub *ws.Hub) error {
	router, err := do.Invoke[*routing.Router](i)
	if err != nil {
		return platformerrors.Wrap(err, "resolving the router")
	}

	extractor, err := do.Invoke[*signingrpc.PrincipalExtractor](i)
	if err != nil {
		return platformerrors.Wrap(err, "resolving the principal extractor")
	}

	pillars, err := observability.InvokePillars(i)
	if err != nil {
		return err
	}

	handler := guests.NewHandler(signingrpc.PrincipalFromContext, pillars.Logger)

	router.Handle(http.MethodGet, IdentityRoute, extractor.HTTPMiddleware(handler))
	router.Handle(http.MethodGet, ws.Route, hub)

	if err = router.Err(); err != nil {
		return platformerrors.Wrap(err, "mounting the HTTP routes")
	}

	return nil
}
