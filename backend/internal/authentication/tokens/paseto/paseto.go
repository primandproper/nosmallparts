package paseto

import (
	"context"
	"fmt"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/identifiers"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/o1egl/paseto/v2"
)

const (
	issuer = "no-small-parts"
)

type (
	signer struct {
		tracer     tracing.Tracer
		logger     logging.Logger
		audience   string
		signingKey []byte
	}
)

func NewPASETOSigner(logger logging.Logger, tracerProvider tracing.TracerProvider, audience string, signingKey []byte) (tokens.Issuer, error) {
	s := &signer{
		audience:   audience,
		signingKey: signingKey,
		logger:     logging.EnsureLogger(logger),
		tracer:     tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("paseto_signer")),
	}

	return s, nil
}

type tokenPayload struct {
	Expiration time.Time
	IssuedAt   time.Time
	NotBefore  time.Time
	Audience   string
	Issuer     string
	JTI        string
	Subject    string
}

// IssueToken issues a new PASETO token.
func (s *signer) IssueToken(ctx context.Context, user tokens.User, expiry time.Duration) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	if expiry <= 0 {
		expiry = time.Minute * 10
	}

	t := tokenPayload{
		Audience:   s.audience,
		Issuer:     issuer,
		JTI:        identifiers.New(),
		Subject:    user.GetID(),
		IssuedAt:   time.Now().UTC(),
		Expiration: time.Now().Add(expiry).UTC(),
		NotBefore:  time.Now().Add(-1 * time.Minute).UTC(),
	}

	tokenString, err := paseto.NewV2().Encrypt(s.signingKey, t, "footer")
	if err != nil {
		return "", fmt.Errorf("signing token with key length %d: %w", len(s.signingKey), err)
	}

	return tokenString, nil
}

// ParseUserIDFromToken parses a AccessToken and returns the associated user ID.
func (s *signer) ParseUserIDFromToken(ctx context.Context, providedToken string) (string, error) {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	var (
		parsedToken tokenPayload
		footer      string
	)
	if err := paseto.NewV2().Decrypt(providedToken, s.signingKey, &parsedToken, &footer); err != nil {
		s.logger.Error("parsing JWT", err)
		return "", err
	}

	return parsedToken.Subject, nil
}
