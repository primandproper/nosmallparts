package api_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/primandproper/nosmallparts/backend/internal/build/api"
	"github.com/primandproper/nosmallparts/backend/internal/config"
	"github.com/primandproper/nosmallparts/backend/internal/proto/scriptspb"
	"github.com/primandproper/nosmallparts/backend/internal/proto/stagepb"

	signincfg "github.com/primandproper/platform-go/v15/authentication/signin/config"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	identitycfg "github.com/primandproper/platform-go/v15/identity/config"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	"github.com/primandproper/platform-go/v15/service"
	tokenscfg "github.com/primandproper/primitives-go/v2/authentication/tokens/config"
	"github.com/primandproper/primitives-go/v2/cookies"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/encoding"
	"github.com/primandproper/primitives-go/v2/observability"
	loggingcfg "github.com/primandproper/primitives-go/v2/observability/logging/config"
	"github.com/primandproper/primitives-go/v2/pointer"
	"github.com/primandproper/primitives-go/v2/routing/backends/stdlib"
	routingcfg "github.com/primandproper/primitives-go/v2/routing/config"
	grpcserver "github.com/primandproper/primitives-go/v2/server/grpc"
	httpserver "github.com/primandproper/primitives-go/v2/server/http"

	"github.com/samber/do/v2"
	"github.com/shoenig/test/must"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// testAdminEmails is what the TypeScript suite's setup put in ADMIN_EMAILS,
// mixed case included, so the case-insensitive match is exercised end to end.
var testAdminEmails = []string{"admin@example.com", "Boss@Example.COM"}

// testPassword is long enough for the password policy.
const testPassword = "password-123"

// harness is one running server over a fresh database: the gRPC services on
// one loopback port and the HTTP routes on another.
type harness struct {
	injector do.Injector
	signIn   signinpb.SignInServiceClient
	scripts  scriptspb.PlayScriptsServiceClient
	stages   stagepb.StagesServiceClient
	http     *http.Client
	httpURL  string
	users    atomic.Int64
}

// testConfig is a whole serving config over a database in dir, with every
// listener on an ephemeral port and every sweeper off.
func testConfig(dir string) *config.Config {
	path := filepath.Join(dir, "test.db")

	return &config.Config{
		Name: config.DefaultServiceName,
		Observability: observability.Config{
			// Logging is left unconfigured, which is a noop: a parallel suite
			// writing every request's log line is noise nobody reads.
			Logging: loggingcfg.Config{},
		},
		Database: &databasecfg.Config{
			Provider:        databasecfg.ProviderSQLite,
			ReadConnection:  databasecfg.ConnectionDetails{Database: path},
			WriteConnection: databasecfg.ConnectionDetails{Database: path},
			RunMigrations:   true,
			MaxPingAttempts: 1,
		},
		Identity: &identitycfg.Config{InvitationTTL: time.Hour},
		SignIn: &signincfg.Config{
			DefaultOwnerRoles: []string{"owner"},
			RefreshTokens:     signincfg.RefreshTokensConfig{SweepInterval: pointer.To(time.Duration(0))},
		},
		Tokens: &tokenscfg.Config{
			Provider:                tokenscfg.ProviderJWT,
			Issuer:                  "nosmallparts-test",
			Audience:                "nosmallparts-test",
			Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte("nosmallparts-test-signing-key-00")),
		},
		Cookies: &cookies.Config{
			Base64EncodedHashKey:  base64.StdEncoding.EncodeToString([]byte("nosmallparts-test-cookie-hash-key-0000000")),
			Base64EncodedBlockKey: base64.StdEncoding.EncodeToString([]byte("nosmallparts-test-block-key-0000")),
			Lifetime:              time.Hour,
		},
		Encoding: &encoding.Config{ContentType: "application/json"},
		Routing: &routingcfg.Config{
			Provider: routingcfg.ProviderStdlib,
			Stdlib:   &stdlib.Config{ServiceName: config.DefaultServiceName, SilenceRouteLogging: true},
		},
		// Port 0 is an ephemeral port, but a block holding nothing beyond its
		// defaults is released as unconfigured, so each names one setting.
		HTTPServer:  &httpserver.Config{StartupDeadline: 10 * time.Second},
		GRPCServer:  &grpcserver.Config{MaxReceiveMessageSize: grpcserver.DefaultMaxMessageSize},
		GRPCWeb:     &service.GRPCWebConfig{PathPrefix: "/grpc"},
		AdminEmails: testAdminEmails,
		// The TypeScript suite's shortened timers, so the timed cases run fast.
		// The cleanup delay and the vacancy claim window are longer than its
		// 100ms and 150ms: a real socket reconnecting or claiming under the
		// race detector can take longer than that, where the old suite's fake
		// one was instant, and those cases are about acting inside the window.
		Stages: config.StagesConfig{
			HostGrace:      80 * time.Millisecond,
			VacancyClaim:   600 * time.Millisecond,
			CleanupDelay:   500 * time.Millisecond,
			IdleAfter:      5 * time.Minute,
			ReaperInterval: time.Minute,
			Countdown:      40 * time.Millisecond,
		},
	}
}

// newHarness starts a server and stops it when the test ends.
func newHarness(t *testing.T) *harness {
	t.Helper()

	ctx, cancel := context.WithCancel(context.WithoutCancel(t.Context()))

	cfg := testConfig(t.TempDir())
	must.NoError(t, cfg.ValidateWithContext(ctx))

	i, err := api.BuildInjector(ctx, cfg)
	must.NoError(t, err)

	svc, err := api.New(i)
	must.NoError(t, err)

	done := make(chan error, 1)

	go func() { done <- svc.Run(ctx) }()

	t.Cleanup(func() {
		cancel()

		select {
		case runErr := <-done:
			must.NoError(t, runErr)
		case <-time.After(10 * time.Second):
			t.Error("the server did not stop within 10s")
		}
	})

	grpcAddr, err := do.MustInvoke[*grpcserver.Server](i).Addr(ctx)
	must.NoError(t, err)

	httpAddr, err := do.MustInvoke[*httpserver.APIServer](i).Addr(ctx)
	must.NoError(t, err)

	conn, err := grpc.NewClient(grpcAddr.String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	must.NoError(t, err)

	t.Cleanup(func() { must.NoError(t, conn.Close()) })

	return &harness{
		injector: i,
		signIn:   signinpb.NewSignInServiceClient(conn),
		scripts:  scriptspb.NewPlayScriptsServiceClient(conn),
		stages:   stagepb.NewStagesServiceClient(conn),
		http:     &http.Client{Timeout: 10 * time.Second},
		httpURL:  "http://" + httpAddr.String(),
	}
}

// testUser is somebody registered and signed in.
type testUser struct {
	ID           string
	Email        string
	Password     string
	Token        string
	RefreshToken string
}

// signUpUser registers a user the way the web app will: the email is the
// username and the name is the first name. Then it signs them in.
func (h *harness) signUpUser(t *testing.T, email string) *testUser {
	t.Helper()

	n := h.users.Add(1)
	if email == "" {
		email = fmt.Sprintf("user%d@test.example", n)
	}

	registered, err := h.signIn.Register(t.Context(), registerRequest(email, testPassword, fmt.Sprintf("User %d", n)))
	must.NoError(t, err)

	user := &testUser{ID: registered.GetRegistration().GetUser().GetId(), Email: email, Password: testPassword}

	issued, err := h.signIn.LoginForToken(t.Context(), &signinpb.LoginForTokenRequest{
		Credentials: &signinpb.Credentials{EmailAddress: email, Password: testPassword},
	})
	must.NoError(t, err)

	user.Token = issued.GetToken().GetToken()
	user.RefreshToken = issued.GetToken().GetRefreshToken()

	return user
}

func registerRequest(email, password, name string) *signinpb.RegisterRequest {
	return &signinpb.RegisterRequest{
		User: &identitypb.UserRegistrationInput{
			Username:     email,
			EmailAddress: email,
			FirstName:    name,
		},
		Account:    &identitypb.AccountCreationInput{Name: name},
		Credential: &signinpb.RegisterRequest_Password{Password: password},
	}
}

// as is a context carrying a user's access token, or none for nil.
func as(ctx context.Context, user *testUser) context.Context {
	if user == nil {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+user.Token)
}

// sampleScript is the TypeScript suite's SAMPLE_SCRIPT_INPUT.
func sampleScript() *scriptspb.PlayInput {
	return &scriptspb.PlayInput{
		Title:       "Two Hander",
		Description: "A short scene",
		Author:      pointer.To("Test Author"),
		IsPublic:    true,
		Cast: []*scriptspb.CastMember{
			{Id: "a", Name: "Alice"},
			{Id: "b", Name: "Bob"},
		},
		Lines: []*scriptspb.PlayLine{
			{CastMemberIds: []string{"a"}, Text: "Line one", Index: 0},
			{CastMemberIds: []string{"b"}, Text: "Line two", Index: 1},
			{CastMemberIds: []string{"a"}, Text: "Line three", Index: 2},
		},
	}
}

// createScript creates a script as a user, with the sample's fields changed by
// edit, and answers what came back.
func (h *harness) createScript(t *testing.T, user *testUser, edit func(*scriptspb.PlayInput)) *scriptspb.Play {
	t.Helper()

	input := sampleScript()
	if edit != nil {
		edit(input)
	}

	res, err := h.scripts.CreatePlayScript(as(t.Context(), user), &scriptspb.CreatePlayScriptRequest{Play: input})
	must.NoError(t, err)

	return res.GetPlay()
}
