package main

import (
	"encoding/base64"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/config"

	signincfg "github.com/primandproper/platform-go/v15/authentication/signin/config"
	identitycfg "github.com/primandproper/platform-go/v15/identity/config"
	"github.com/primandproper/platform-go/v15/service"
	tokenscfg "github.com/primandproper/primitives-go/v2/authentication/tokens/config"
	"github.com/primandproper/primitives-go/v2/cookies"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/encoding"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	loggingcfg "github.com/primandproper/primitives-go/v2/observability/logging/config"
	"github.com/primandproper/primitives-go/v2/routing/backends/stdlib"
	routingcfg "github.com/primandproper/primitives-go/v2/routing/config"
	grpcserver "github.com/primandproper/primitives-go/v2/server/grpc"
	httpserver "github.com/primandproper/primitives-go/v2/server/http"
)

// Each builder returns a fully-formed *config.Config for one environment. The
// configs are constructed here as real, typed Go objects (the compiler and the
// config's own ValidateWithContext are the guardrails) and then Render projects
// them to the JSON files under config/.
//
// A platform subsystem is composed into the process by being present here and
// by nothing else, so this file is also the shortest statement of what a
// nosmallparts process is made of: a SQLite database, the identity directory,
// sign-in over it, a token issuer, the cookie keys a browser's refresh token is
// sealed with, and an HTTP server that carries the gRPC server's services as
// gRPC-web.

const (
	// replaceAtDeploy marks a secret the deployment supplies through the
	// environment. Every one has a NOSMALLPARTS_ variable in the envvars
	// package, so a deployment that forgot one fails loudly at the provider
	// rather than quietly with a placeholder.
	replaceAtDeploy = "REPLACE_AT_DEPLOY"

	// httpPort is the one port a browser reaches: the web app, gRPC-web and
	// the stage WebSocket. The native gRPC listener stays internal.
	httpPort = 3001
	grpcPort = 50051

	// grpcWebPrefix is where the gRPC services are reached over HTTP.
	grpcWebPrefix = "/grpc"

	tokenIssuer = "nosmallparts"

	// ownerRole is what a registrant holds in the account sign-in mints for
	// them. Nothing reads it; sign-in refuses to register without one.
	ownerRole = "owner"

	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour

	// invitationTTL is set only because an identity block holding nothing but
	// defaults is released as unconfigured. Nothing here invites anybody.
	invitationTTL = 7 * 24 * time.Hour
)

// The local keys are fixed so a local sign-in survives a restart. They sign and
// seal nothing outside a laptop.
const (
	localDevSigningKey = "nosmallparts-localdev-signing-00"
	localDevHashKey    = "nosmallparts-localdev-cookie-hash-key-000"
	localDevBlockKey   = "nosmallparts-localdev-blockkey00"
)

// The production markers decode cleanly at the right lengths so the rendered
// file validates; the environment replaces each at deploy.
const (
	productionSigningKeyMarker = "REPLACE_AT_DEPLOY_signing_key_00"
	productionBlockKeyMarker   = "REPLACE_AT_DEPLOY_block_key_0000"
)

func base64URLKey(raw string) string {
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func base64StdKey(raw string) string {
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// platform is what both environments share: every subsystem present, the
// table prefixes left at their defaults because nosmallparts has the database
// to itself, and migrations run on boot because one machine owns the file.
func platform(level logging.Level, databasePath, signingKey string, cookieKeys *cookies.Config) service.Config {
	return service.Config{
		Name: config.DefaultServiceName,
		Observability: observability.Config{
			Logging: loggingcfg.Config{
				Provider:    loggingcfg.ProviderSlog,
				ServiceName: config.DefaultServiceName,
				Level:       level,
			},
		},
		Database: &databasecfg.Config{
			Provider:        databasecfg.ProviderSQLite,
			ReadConnection:  databasecfg.ConnectionDetails{Database: databasePath},
			WriteConnection: databasecfg.ConnectionDetails{Database: databasePath},
			RunMigrations:   true,
		},
		Identity: &identitycfg.Config{InvitationTTL: invitationTTL},
		SignIn: &signincfg.Config{
			DefaultOwnerRoles: []string{ownerRole},
			TokenTTL:          accessTokenTTL,
			RefreshTokens:     signincfg.RefreshTokensConfig{TTL: refreshTokenTTL},
		},
		Tokens: &tokenscfg.Config{
			Provider:                tokenscfg.ProviderJWT,
			Issuer:                  tokenIssuer,
			Audience:                tokenIssuer,
			Base64EncodedSigningKey: signingKey,
		},
		Cookies:  cookieKeys,
		Encoding: &encoding.Config{ContentType: "application/json"},
		Routing: &routingcfg.Config{
			Provider: routingcfg.ProviderStdlib,
			Stdlib:   &stdlib.Config{ServiceName: config.DefaultServiceName},
		},
		HTTPServer: &httpserver.Config{Port: httpPort},
		GRPCServer: &grpcserver.Config{Port: grpcPort},
		GRPCWeb:    &service.GRPCWebConfig{PathPrefix: grpcWebPrefix},
	}
}

// buildLocalDevConfig is the config a developer runs against locally: debug
// logging, a database under ./data, and cookies a plain-HTTP localhost accepts.
func buildLocalDevConfig() *config.Config {
	return &config.Config{
		Config: platform(
			logging.DebugLevel,
			"data/nosmallparts.db",
			base64URLKey(localDevSigningKey),
			&cookies.Config{
				Base64EncodedHashKey:  base64StdKey(localDevHashKey),
				Base64EncodedBlockKey: base64StdKey(localDevBlockKey),
				Lifetime:              refreshTokenTTL,
			},
		),
	}
}

// buildProductionConfig is the config for the Fly deployment: info logging,
// the database on the /data volume, and secure cookies keyed from the
// environment.
func buildProductionConfig() *config.Config {
	return &config.Config{
		Config: platform(
			logging.InfoLevel,
			"/data/nosmallparts.db",
			base64URLKey(productionSigningKeyMarker),
			&cookies.Config{
				Base64EncodedHashKey:  base64StdKey(replaceAtDeploy),
				Base64EncodedBlockKey: base64StdKey(productionBlockKeyMarker),
				Lifetime:              refreshTokenTTL,
				SecureOnly:            true,
			},
		),
	}
}
