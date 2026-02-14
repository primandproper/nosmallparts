package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	auditgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	authgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	dataprivacygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	identitygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	internalopsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/internalops"
	issuereportsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"
	notificationsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	oauthgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/oauth"
	settingsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/settings"
	uploadedmediagrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/uploaded_media"
	waitlistsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
	webhooksgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/random"

	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
	"google.golang.org/grpc/metadata"
)

const (
	zuckModeUserHeader    = "X-Zuck-Mode-User"
	zuckModeAccountHeader = "X-Zuck-Mode-Account"
)

type Client interface {
	authgrpc.AuthServiceClient
	identitygrpc.IdentityServiceClient
	auditgrpc.AuditServiceClient
	dataprivacygrpc.DataPrivacyServiceClient
	internalopsgrpc.InternalOperationsClient
	issuereportsgrpc.IssueReportsServiceClient
	notificationsgrpc.UserNotificationsServiceClient
	oauthgrpc.OAuthServiceClient
	settingsgrpc.SettingsServiceClient
	uploadedmediagrpc.UploadedMediaServiceClient
	waitlistsgrpc.WaitlistsServiceClient
	webhooksgrpc.WebhooksServiceClient
}

type client struct {
	authgrpc.AuthServiceClient
	identitygrpc.IdentityServiceClient
	auditgrpc.AuditServiceClient
	dataprivacygrpc.DataPrivacyServiceClient
	internalopsgrpc.InternalOperationsClient
	issuereportsgrpc.IssueReportsServiceClient
	notificationsgrpc.UserNotificationsServiceClient
	oauthgrpc.OAuthServiceClient
	settingsgrpc.SettingsServiceClient
	uploadedmediagrpc.UploadedMediaServiceClient
	waitlistsgrpc.WaitlistsServiceClient
	webhooksgrpc.WebhooksServiceClient
}

// BuildClient builds a new Client.
func BuildClient(grpcServerAddress string, opts ...grpc.DialOption) (Client, error) {
	conn, err := grpc.NewClient(grpcServerAddress, opts...)
	if err != nil {
		return nil, fmt.Errorf("building grpc client: %w", err)
	}

	c := &client{
		AuthServiceClient:              authgrpc.NewAuthServiceClient(conn),
		IdentityServiceClient:          identitygrpc.NewIdentityServiceClient(conn),
		AuditServiceClient:             auditgrpc.NewAuditServiceClient(conn),
		DataPrivacyServiceClient:       dataprivacygrpc.NewDataPrivacyServiceClient(conn),
		InternalOperationsClient:       internalopsgrpc.NewInternalOperationsClient(conn),
		IssueReportsServiceClient:      issuereportsgrpc.NewIssueReportsServiceClient(conn),
		UserNotificationsServiceClient: notificationsgrpc.NewUserNotificationsServiceClient(conn),
		OAuthServiceClient:             oauthgrpc.NewOAuthServiceClient(conn),
		SettingsServiceClient:          settingsgrpc.NewSettingsServiceClient(conn),
		UploadedMediaServiceClient:     uploadedmediagrpc.NewUploadedMediaServiceClient(conn),
		WaitlistsServiceClient:         waitlistsgrpc.NewWaitlistsServiceClient(conn),
		WebhooksServiceClient:          webhooksgrpc.NewWebhooksServiceClient(conn),
	}

	return c, nil
}

func BuildUnauthenticatedGRPCClient(grpcServerAddr string) (Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	return BuildClient(grpcServerAddr, opts...)
}

// staticBearerToken implements credentials.PerRPCCredentials for a fixed Bearer token.
type staticBearerToken struct{ token string }

func (s *staticBearerToken) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + s.token}, nil
}

func (s *staticBearerToken) RequireTransportSecurity() bool { return false }

// ExchangeJWTForOAuth2Token performs the OAuth2 authorization code flow using the given JWT for
// user authorization. The JWT is sent to /oauth2/authorize; the auth server uses it to identify
// the user and redirects with a code; we exchange that code for an OAuth2 access token. Returns
// the OAuth2 token (use .AccessToken for cookie storage and BuildClientWithBearerToken).
func ExchangeJWTForOAuth2Token(
	ctx context.Context,
	authServerHTTPBase,
	clientID,
	clientSecret,
	jwt string,
) (*oauth2.Token, error) {
	state, err := random.GenerateBase64EncodedString(ctx, 32)
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  authServerHTTPBase,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   authServerHTTPBase + "/oauth2/authorize",
			TokenURL:  authServerHTTPBase + "/oauth2/token",
		},
	}

	authCodeURL := oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge_method", "plain"),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authCodeURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("building oauth2 authorize request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwt)

	c := tracing.BuildTracedHTTPClient()
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getting oauth2 code: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Println("failed to close oauth2 response body", closeErr)
		}
	}()

	location, err := res.Location()
	if err != nil {
		return nil, err
	}
	code := location.Query().Get("code")
	if code == "" {
		return nil, errors.New("code not returned from oauth2 redirect")
	}

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging oauth2 code for token: %w", err)
	}
	return token, nil
}

// BuildClientWithBearerToken builds a gRPC Client that sends the given Bearer token on every call.
// Use with the OAuth2 access token from ExchangeJWTForOAuth2Token (stored in cookie after login).
func BuildClientWithBearerToken(grpcServerAddr, bearerToken string) (Client, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(&staticBearerToken{token: bearerToken}),
	}

	return BuildClient(grpcServerAddr, opts...)
}

func WithOAuth2Credentials(
	ctx context.Context,
	authServerAddress,
	clientID,
	clientSecret,
	authToken string,
) (grpc.DialOption, error) {
	oauth2Token, err := ExchangeJWTForOAuth2Token(ctx, authServerAddress, clientID, clientSecret, authToken)
	if err != nil {
		return nil, err
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  authServerAddress,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   authServerAddress + "/oauth2/authorize",
			TokenURL:  authServerAddress + "/oauth2/token",
		},
	}

	ts := oauth2.ReuseTokenSource(oauth2Token, oauth2Config.TokenSource(ctx, oauth2Token))

	return grpc.WithPerRPCCredentials(oauth.TokenSource{
		TokenSource: ts,
	}), nil
}

func ImpersonateUseAndAccountContext(ctx context.Context, userID, accountID string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		zuckModeUserHeader:    userID,
		zuckModeAccountHeader: accountID,
	}))
}
