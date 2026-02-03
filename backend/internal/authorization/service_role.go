package authorization

import (
	"encoding/gob"

	"github.com/mikespook/gorbac/v2"
)

func init() {
	gob.Register(serviceRoleCollection{})
}

const (
	serviceAdminRoleName     = "service_admin"
	serviceDataAdminRoleName = "service_data_admin"
	serviceUserRoleName      = "service_user"

	invalidServiceRoleWarning = "INVALID_SERVICE_ROLE"

	// invalidServiceRole is a service role to apply for non-admin users to have one.
	invalidServiceRole ServiceRole = iota
	// ServiceUserRole is a service role to apply for non-admin users to have one.
	ServiceUserRole ServiceRole = iota
	// ServiceAdminRole is a role that allows a user to do basically anything.
	ServiceAdminRole ServiceRole = iota
)

var (
	serviceUser      = gorbac.NewStdRole(serviceUserRoleName)
	serviceAdmin     = gorbac.NewStdRole(serviceAdminRoleName)
	serviceDataAdmin = gorbac.NewStdRole(serviceDataAdminRoleName)
)

type (
	// ServiceRole describes a role a user has for the Service context.
	ServiceRole role

	// ServiceRolePermissionChecker checks permissions for one or more service Roles.
	ServiceRolePermissionChecker interface {
		HasPermission(Permission) bool

		AsAccountRolePermissionChecker() AccountRolePermissionsChecker
		IsServiceAdmin() bool
		CanUpdateUserAccountStatuses() bool
		CanImpersonateUsers() bool
	}

	serviceRoleCollection struct {
		Roles []string
	}
)

func (r ServiceRole) String() string {
	switch r {
	case invalidServiceRole:
		return invalidServiceRoleWarning
	case ServiceUserRole:
		return serviceUserRoleName
	case ServiceAdminRole:
		return serviceAdminRoleName
	default:
		return ""
	}
}

// NewServiceRolePermissionChecker returns a new checker for a set of Roles.
func NewServiceRolePermissionChecker(roles ...string) ServiceRolePermissionChecker {
	return &serviceRoleCollection{
		Roles: roles,
	}
}

func (r serviceRoleCollection) AsAccountRolePermissionChecker() AccountRolePermissionsChecker {
	return NewAccountRolePermissionChecker(r.Roles...)
}

// HasPermission returns whether a user can do something or not.
func (r serviceRoleCollection) HasPermission(p Permission) bool {
	return hasPermission(p, r.Roles...)
}

// IsServiceAdmin returns if a role is an admin.
func (r serviceRoleCollection) IsServiceAdmin() bool {
	for _, x := range r.Roles {
		if x == ServiceAdminRole.String() {
			return true
		}
	}

	return false
}

// CanUpdateUserAccountStatuses returns whether a user can update user account statuses.
func (r serviceRoleCollection) CanUpdateUserAccountStatuses() bool {
	return hasPermission(UpdateUserStatusPermission, r.Roles...)
}

// CanImpersonateUsers returns whether a user can impersonate others.
func (r serviceRoleCollection) CanImpersonateUsers() bool {
	return hasPermission(ImpersonateUserPermission, r.Roles...)
}
