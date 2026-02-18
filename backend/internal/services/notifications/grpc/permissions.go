package grpc

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authorization"
	notificationssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
)

// NotificationsMethodPermissions is a named type for Wire dependency injection.
type NotificationsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the notifications service's method permissions.
func ProvideMethodPermissions() NotificationsMethodPermissions {
	return NotificationsMethodPermissions{
		notificationssvc.UserNotificationsService_GetUserNotification_FullMethodName: {
			authorization.ReadUserNotificationsPermission,
		},
		notificationssvc.UserNotificationsService_GetUserNotifications_FullMethodName: {
			authorization.ReadUserNotificationsPermission,
		},
		notificationssvc.UserNotificationsService_UpdateUserNotification_FullMethodName: {
			authorization.UpdateUserNotificationsPermission,
		},
	}
}
