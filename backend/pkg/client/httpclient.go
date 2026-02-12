package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	authgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	identitygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	identityconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"google.golang.org/grpc"
)

var errUnimplemented = errors.New("not implemented: HTTP client only supports CreateUser, LoginForToken, AdminLoginForToken, VerifyTOTPSecret")

// httpClient implements Client by calling the backend HTTP API for auth/identity and returning errUnimplemented for all other methods.
// It embeds stub types for every service; the few implemented methods are defined on *httpClient and override the stubs.
type httpClient struct {
	stubAuthClient
	stubIdentityClient
	stubAuditClient
	stubDataPrivacyClient
	stubInternalOpsClient
	stubIssueReportsClient
	stubNotificationsClient
	stubOAuthClient
	stubSettingsClient
	stubUploadedMediaClient
	stubWaitlistsClient
	stubWebhooksClient
	baseURL    string
	httpClient *http.Client
	token      string // optional Bearer token for authenticated requests
}

// httpClient must implement Client; unimplemented methods are provided by embedded stub types in client_stubs.go.

// BuildUnauthenticatedHTTPClient returns a Client that talks to the backend HTTP API at baseURL (e.g. "http://localhost:8000").
// Only CreateUser, LoginForToken, AdminLoginForToken, and VerifyTOTPSecret are implemented; other methods return errUnimplemented.
func BuildUnauthenticatedHTTPClient(baseURL string) (Client, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &httpClient{
		baseURL:    baseURL,
		httpClient: tracing.BuildTracedHTTPClient(),
	}, nil
}

// BuildAuthenticatedHTTPClient returns a Client that uses the given Bearer token on API requests.
func BuildAuthenticatedHTTPClient(baseURL, bearerToken string) (Client, error) {
	c, err := BuildUnauthenticatedHTTPClient(baseURL)
	if err != nil {
		return nil, err
	}
	h := c.(*httpClient)
	h.token = bearerToken
	return h, nil
}

func (h *httpClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, h.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if h.token != "" {
		req.Header.Set("Authorization", "Bearer "+h.token)
	}
	if userID, ok := ctx.Value(zuckModeUserContextKey).(string); ok && userID != "" {
		req.Header.Set(zuckModeUserHeader, userID)
	}
	if accountID, ok := ctx.Value(zuckModeAccountContextKey).(string); ok && accountID != "" {
		req.Header.Set(zuckModeAccountHeader, accountID)
	}
	return h.httpClient.Do(req)
}

// registerRequest is the JSON body for POST /api/register.
type registerRequest struct {
	Username              string  `json:"username"`
	Password              string  `json:"password"`
	EmailAddress          string  `json:"emailAddress"`
	FirstName             string  `json:"firstName"`
	LastName              string  `json:"lastName"`
	Birthday              *string `json:"birthday,omitempty"`
	AccountName           string  `json:"accountName"`
	InvitationToken       string  `json:"invitationToken,omitempty"`
	InvitationID          string  `json:"invitationId,omitempty"`
	AcceptedTOS           bool    `json:"acceptedTos"`
	AcceptedPrivacyPolicy bool    `json:"acceptedPrivacyPolicy"`
}

// registerResponse is the JSON response for POST /api/register.
type registerResponse struct {
	CreatedUserID   string `json:"createdUserId"`
	TwoFactorQRCode string `json:"twoFactorQrCode"`
	TwoFactorSecret string `json:"twoFactorSecret"`
}

func (h *httpClient) CreateUser(ctx context.Context, in *identitygrpc.CreateUserRequest, opts ...grpc.CallOption) (*identitygrpc.CreateUserResponse, error) {
	if in == nil || in.Input == nil {
		return nil, errors.New("CreateUserRequest.Input is required")
	}
	dom := identityconverters.ConvertGRPCUserRegistrationInputToUserRegistrationInput(in.Input)
	var birthday *string
	if dom.Birthday != nil {
		s := dom.Birthday.Format("2006-01-02")
		birthday = &s
	}
	body := registerRequest{
		Username:              dom.Username,
		Password:              dom.Password,
		EmailAddress:          dom.EmailAddress,
		FirstName:             dom.FirstName,
		LastName:              dom.LastName,
		Birthday:              birthday,
		AccountName:           dom.AccountName,
		InvitationToken:       dom.InvitationToken,
		InvitationID:          dom.InvitationID,
		AcceptedTOS:           dom.AcceptedTOS,
		AcceptedPrivacyPolicy: dom.AcceptedPrivacyPolicy,
	}
	res, err := h.do(ctx, http.MethodPost, "/api/register", body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("register returned %d: %s", res.StatusCode, string(bb))
	}
	var rr registerResponse
	if err := json.NewDecoder(res.Body).Decode(&rr); err != nil {
		return nil, err
	}
	created := &identitygrpc.UserCreationResponse{
		CreatedUserId:   rr.CreatedUserID,
		TwoFactorQrCode: rr.TwoFactorQRCode,
		TwoFactorSecret: rr.TwoFactorSecret,
	}
	return &identitygrpc.CreateUserResponse{Created: created}, nil
}

// loginForTokenBody is the JSON body for POST /api/login-for-token.
type loginForTokenBody struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TOTPToken  string `json:"totpToken,omitempty"`
}

// loginForTokenData is the "data" part of the API response.
type loginForTokenData struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	UserID       string `json:"userID"`
	AccountID    string `json:"accountID"`
	Expires      string `json:"expires"`
}

// apiResponse wraps the standard API response shape.
type apiResponse struct {
	Data *loginForTokenData `json:"data"`
}

func (h *httpClient) LoginForToken(ctx context.Context, in *authgrpc.LoginForTokenRequest, opts ...grpc.CallOption) (*authgrpc.LoginForTokenResponse, error) {
	if in == nil || in.Input == nil {
		return nil, errors.New("LoginForTokenRequest.Input is required")
	}
	body := loginForTokenBody{
		Username:  in.Input.Username,
		Password:  in.Input.Password,
		TOTPToken: in.Input.TotpToken,
	}
	res, err := h.do(ctx, http.MethodPost, "/api/login-for-token", body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("login-for-token returned %d: %s", res.StatusCode, string(bb))
	}
	var wrap apiResponse
	if err := json.NewDecoder(res.Body).Decode(&wrap); err != nil {
		return nil, err
	}
	if wrap.Data == nil {
		return nil, errors.New("login-for-token response missing data")
	}
	return &authgrpc.LoginForTokenResponse{
		Result: &authgrpc.TokenResponse{
			AccessToken:  wrap.Data.AccessToken,
			RefreshToken: wrap.Data.RefreshToken,
			UserId:       wrap.Data.UserID,
			AccountId:    wrap.Data.AccountID,
		},
	}, nil
}

func (h *httpClient) AdminLoginForToken(ctx context.Context, in *authgrpc.AdminLoginForTokenRequest, opts ...grpc.CallOption) (*authgrpc.LoginForTokenResponse, error) {
	if in == nil || in.Input == nil {
		return nil, errors.New("AdminLoginForTokenRequest.Input is required")
	}
	// Same endpoint as LoginForToken
	body := loginForTokenBody{
		Username:  in.Input.Username,
		Password:  in.Input.Password,
		TOTPToken: in.Input.TotpToken,
	}
	res, err := h.do(ctx, http.MethodPost, "/api/login-for-token", body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("login-for-token returned %d: %s", res.StatusCode, string(bb))
	}
	var wrap apiResponse
	if err := json.NewDecoder(res.Body).Decode(&wrap); err != nil {
		return nil, err
	}
	if wrap.Data == nil {
		return nil, errors.New("login-for-token response missing data")
	}
	return &authgrpc.LoginForTokenResponse{
		Result: &authgrpc.TokenResponse{
			AccessToken:  wrap.Data.AccessToken,
			RefreshToken: wrap.Data.RefreshToken,
			UserId:       wrap.Data.UserID,
			AccountId:    wrap.Data.AccountID,
		},
	}, nil
}

// verifyTOTPBody is the JSON body for POST /api/verify-totp.
type verifyTOTPBody struct {
	UserID    string `json:"userID"`
	TOTPToken string `json:"totpToken"`
}

func (h *httpClient) VerifyTOTPSecret(ctx context.Context, in *authgrpc.VerifyTOTPSecretRequest, opts ...grpc.CallOption) (*authgrpc.VerifyTOTPSecretResponse, error) {
	if in == nil {
		return nil, errors.New("VerifyTOTPSecretRequest is required")
	}
	body := verifyTOTPBody{
		UserID:    in.UserId,
		TOTPToken: in.TotpToken,
	}
	res, err := h.do(ctx, http.MethodPost, "/api/verify-totp", body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		bb, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("verify-totp returned %d: %s", res.StatusCode, string(bb))
	}
	return &authgrpc.VerifyTOTPSecretResponse{}, nil
}
