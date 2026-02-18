package converters

import (
	"log"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications"
	grpcconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/converters"
	notificationssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"
)

func ConvertStringToUserNotificationStatus(s string) notificationssvc.UserNotificationStatus {
	switch s {
	case notifications.UserNotificationStatusTypeRead:
		return notificationssvc.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ
	case notifications.UserNotificationStatusTypeUnread:
		return notificationssvc.UserNotificationStatus_USER_NOTIFICATION_STATUS_UNREAD
	case notifications.UserNotificationStatusTypeDismissed:
		return notificationssvc.UserNotificationStatus_USER_NOTIFICATION_STATUS_DISMISSED
	default:
		log.Printf("UNKNOWN USER NOTIFICATION STATUS: %q", s)
		return notificationssvc.UserNotificationStatus_USER_NOTIFICATION_STATUS_UNREAD
	}
}

func ConvertUserNotificationStatusToString(s notificationssvc.UserNotificationStatus) string {
	switch s {
	case notificationssvc.UserNotificationStatus_USER_NOTIFICATION_STATUS_READ:
		return notifications.UserNotificationStatusTypeRead
	case notificationssvc.UserNotificationStatus_USER_NOTIFICATION_STATUS_UNREAD:
		return notifications.UserNotificationStatusTypeUnread
	case notificationssvc.UserNotificationStatus_USER_NOTIFICATION_STATUS_DISMISSED:
		return notifications.UserNotificationStatusTypeDismissed
	default:
		log.Printf("UNKNOWN USER NOTIFICATION STATUS: %q", s)
		return notifications.UserNotificationStatusTypeUnread
	}
}

func ConvertUserNotificationToGRPCUserNotification(notification *notifications.UserNotification) *notificationssvc.UserNotification {
	return &notificationssvc.UserNotification{
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(notification.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(notification.LastUpdatedAt),
		Id:            notification.ID,
		Content:       notification.Content,
		Status:        ConvertStringToUserNotificationStatus(notification.Status),
		BelongsToUser: notification.BelongsToUser,
	}
}

func ConvertGRPCUserNotificationToUserNotification(notification *notificationssvc.UserNotification) *notifications.UserNotification {
	return &notifications.UserNotification{
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(notification.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(notification.LastUpdatedAt),
		ID:            notification.Id,
		Content:       notification.Content,
		Status:        ConvertUserNotificationStatusToString(notification.Status),
		BelongsToUser: notification.BelongsToUser,
	}
}

func ConvertUserNotificationUpdateRequestInputToGRPCUserNotificationUpdateRequestInput(input *notifications.UserNotificationUpdateRequestInput) *notificationssvc.UserNotificationUpdateRequestInput {
	var newStatus *notificationssvc.UserNotificationStatus
	if input.Status != nil {
		newStatus = pointer.To(ConvertStringToUserNotificationStatus(*input.Status))
	}

	return &notificationssvc.UserNotificationUpdateRequestInput{
		Status: newStatus,
	}
}

// ConvertUserDataCollectionToGRPCDataCollection converts a domain notifications UserDataCollection to a proto DataCollection.
func ConvertUserDataCollectionToGRPCDataCollection(input *notifications.UserDataCollection) *notificationssvc.DataCollection {
	result := &notificationssvc.DataCollection{}

	for i := range input.Data {
		result.Notifications = append(result.Notifications, ConvertUserNotificationToGRPCUserNotification(&input.Data[i]))
	}

	return result
}
