package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// ReadUserDataPermission is a service admin permission.
	ReadUserDataPermission Permission = "admin.read_user_data"
	// PublishArbitraryQueueMessagePermission is a service admin permission.
	PublishArbitraryQueueMessagePermission Permission = "queues.publish.message"
)

var (
	// AdminPermissions contains all admin-specific permissions.
	AdminPermissions = []gorbac.Permission{
		ReadUserDataPermission,
		PublishArbitraryQueueMessagePermission,
	}
)
