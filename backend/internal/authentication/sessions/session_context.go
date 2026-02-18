package sessions

import (
	"context"
	"encoding/gob"
	"errors"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing"
)

func init() {
	gob.Register(&ContextData{})
}

var (
	// ErrNoSessionContextDataAvailable indicates no ContextData was attached to the request.
	ErrNoSessionContextDataAvailable = errors.New("no ContextData attached to session context data")
)

// FetchContextDataFromRequest fetches a ContextData from a request.
func FetchContextDataFromRequest(req *http.Request) (*ContextData, error) {
	if sessionCtxData, ok := req.Context().Value(SessionContextDataKey).(*ContextData); ok && sessionCtxData != nil {
		return sessionCtxData, nil
	}

	return nil, ErrNoSessionContextDataAvailable
}

func FetchContextDataFromContext(ctx context.Context) (*ContextData, error) {
	if sessionCtxData, ok := ctx.Value(SessionContextDataKey).(*ContextData); ok && sessionCtxData != nil {
		return sessionCtxData, nil
	}
	return nil, ErrAuthenticationNotFound
}

const SessionContextDataKey routing.ContextKey = "session_context_data"

// ContextData represents what we encode in our passwords cookies.
type ContextData struct {
	_ struct{} `json:"-"`

	AccountPermissions map[string]authorization.AccountRolePermissionsChecker `json:"-"`
	Requester          RequesterInfo                                          `json:"-"`
	ActiveAccountID    string                                                 `json:"-"`
}

// RequesterInfo contains data relevant to the user making a request.
type RequesterInfo struct {
	_ struct{} `json:"-"`

	ServicePermissions       authorization.ServiceRolePermissionChecker `json:"-"`
	AccountStatus            string                                     `json:"-"`
	AccountStatusExplanation string                                     `json:"-"`
	UserID                   string                                     `json:"-"`
	EmailAddress             string                                     `json:"-"`
	Username                 string                                     `json:"-"`
}

// GetUserID is a simple getter.
func (x *ContextData) GetUserID() string {
	return x.Requester.UserID
}

// GetServicePermissions is a simple getter.
func (x *ContextData) GetServicePermissions() authorization.ServiceRolePermissionChecker {
	return x.Requester.ServicePermissions
}

// GetActiveAccountID is a simple getter.
func (x *ContextData) GetActiveAccountID() string {
	return x.ActiveAccountID
}

// AccountRolePermissionsChecker returns the relevant AccountRolePermissionsChecker.
func (x *ContextData) AccountRolePermissionsChecker() authorization.AccountRolePermissionsChecker {
	return x.AccountPermissions[x.ActiveAccountID]
}

// ServiceRolePermissionChecker returns the relevant ServiceRolePermissionChecker.
func (x *ContextData) ServiceRolePermissionChecker() authorization.ServiceRolePermissionChecker {
	return x.Requester.ServicePermissions
}

// AttachToLogger provides a consistent way to attach a ContextData object to a logger.
func (x *ContextData) AttachToLogger(logger logging.Logger) logging.Logger {
	if x != nil {
		logger = logger.WithValue(keys.RequesterIDKey, x.GetUserID()).
			WithValue(keys.ActiveAccountIDKey, x.ActiveAccountID)
	}

	return logger
}
