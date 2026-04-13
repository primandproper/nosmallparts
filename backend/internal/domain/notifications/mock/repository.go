package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"

	"github.com/stretchr/testify/mock"
)

var _ notifications.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// UserNotificationExists is a mock function.
func (m *Repository) UserNotificationExists(ctx context.Context, userID, userNotificationID string) (bool, error) {
	args := m.Called(ctx, userID, userNotificationID)
	return args.Bool(0), args.Error(1)
}

// GetUserNotification is a mock function.
func (m *Repository) GetUserNotification(ctx context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
	args := m.Called(ctx, userID, userNotificationID)
	return args.Get(0).(*notifications.UserNotification), args.Error(1)
}

// GetUserNotifications is a mock function.
func (m *Repository) GetUserNotifications(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[notifications.UserNotification], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[notifications.UserNotification]), args.Error(1)
}

// CreateUserNotification is a mock function.
func (m *Repository) CreateUserNotification(ctx context.Context, input *notifications.UserNotificationDatabaseCreationInput) (*notifications.UserNotification, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*notifications.UserNotification), args.Error(1)
}

// UpdateUserNotification is a mock function.
func (m *Repository) UpdateUserNotification(ctx context.Context, updated *notifications.UserNotification) error {
	args := m.Called(ctx, updated)
	return args.Error(0)
}
