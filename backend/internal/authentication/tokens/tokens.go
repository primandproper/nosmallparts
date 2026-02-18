package tokens

import (
	"context"
	"time"
)

type (
	User interface {
		GetID() string
		GetEmail() string
		GetUsername() string
		GetFirstName() string
		GetLastName() string
		FullName() string
	}

	Issuer interface {
		IssueToken(ctx context.Context, user User, expiry time.Duration) (string, error)
		ParseUserIDFromToken(ctx context.Context, token string) (string, error)
	}
)
