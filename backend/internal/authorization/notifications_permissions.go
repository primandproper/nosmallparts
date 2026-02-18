package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// CreateUserNotificationsPermission is an admin user permission.
	CreateUserNotificationsPermission Permission = "create.user_notifications"
	// ReadUserNotificationsPermission is a permission.
	ReadUserNotificationsPermission Permission = "read.user_notifications"
	// UpdateUserNotificationsPermission is a permission.
	UpdateUserNotificationsPermission Permission = "update.user_notifications"
)

var (
	// NotificationsPermissions contains all notification-related permissions.
	NotificationsPermissions = []gorbac.Permission{
		CreateUserNotificationsPermission,
		ReadUserNotificationsPermission,
		UpdateUserNotificationsPermission,
	}
)
