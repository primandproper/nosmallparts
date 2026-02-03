package oauth

import (
	"context"
	"time"
)

type (
	// OAuth2ClientToken represents a user-authorized OAuth2 client's token.
	OAuth2ClientToken struct {
		_ struct{} `json:"-"`

		RefreshCreatedAt    time.Time     `json:"refreshCreatedAt"`
		AccessCreatedAt     time.Time     `json:"accessCreatedAt"`
		CodeCreatedAt       time.Time     `json:"codeCreatedAt"`
		RedirectURI         string        `json:"redirectURI"`
		Code                string        `json:"code"`
		CodeChallenge       string        `json:"codeChallenge"`
		CodeChallengeMethod string        `json:"codeChallengeMethod"`
		BelongsToUser       string        `json:"belongsToUser"`
		Access              string        `json:"access"`
		ClientID            string        `json:"clientID"`
		Refresh             string        `json:"refresh"`
		ID                  string        `json:"id"`
		CodeExpiresAt       time.Duration `json:"codeExpiresIn"`
		AccessExpiresAt     time.Duration `json:"accessExpiresIn"`
		RefreshExpiresAt    time.Duration `json:"refreshExpiresIn"`
	}

	// OAuth2ClientTokenDatabaseCreationInput represents a user-authorized OAuth2 client's token's database creation input.
	OAuth2ClientTokenDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		RefreshCreatedAt    time.Time     `json:"-"`
		AccessCreatedAt     time.Time     `json:"-"`
		CodeCreatedAt       time.Time     `json:"-"`
		RedirectURI         string        `json:"-"`
		Code                string        `json:"-"`
		CodeChallenge       string        `json:"-"`
		CodeChallengeMethod string        `json:"-"`
		BelongsToUser       string        `json:"-"`
		Access              string        `json:"-"`
		ClientID            string        `json:"-"`
		Refresh             string        `json:"-"`
		ID                  string        `json:"-"`
		CodeExpiresIn       time.Duration `json:"-"`
		AccessExpiresIn     time.Duration `json:"-"`
		RefreshExpiresIn    time.Duration `json:"-"`
	}

	OAuth2ClientTokenDataManager interface {
		GetOAuth2ClientTokenByCode(ctx context.Context, code string) (*OAuth2ClientToken, error)
		GetOAuth2ClientTokenByAccess(ctx context.Context, access string) (*OAuth2ClientToken, error)
		GetOAuth2ClientTokenByRefresh(ctx context.Context, refresh string) (*OAuth2ClientToken, error)
		CreateOAuth2ClientToken(ctx context.Context, input *OAuth2ClientTokenDatabaseCreationInput) (*OAuth2ClientToken, error)
		DeleteOAuth2ClientTokenByAccess(ctx context.Context, access string) error
		DeleteOAuth2ClientTokenByCode(ctx context.Context, code string) error
		DeleteOAuth2ClientTokenByRefresh(ctx context.Context, refresh string) error
	}
)
