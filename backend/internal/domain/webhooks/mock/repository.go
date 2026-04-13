package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"

	"github.com/stretchr/testify/mock"
)

var _ webhooks.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// WebhookExists is a mock function.
func (m *Repository) WebhookExists(ctx context.Context, webhookID, accountID string) (bool, error) {
	args := m.Called(ctx, webhookID, accountID)
	return args.Bool(0), args.Error(1)
}

// GetWebhook is a mock function.
func (m *Repository) GetWebhook(ctx context.Context, webhookID, accountID string) (*webhooks.Webhook, error) {
	args := m.Called(ctx, webhookID, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.Webhook), args.Error(1)
}

// GetWebhooks is a mock function.
func (m *Repository) GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[webhooks.Webhook]), args.Error(1)
}

// GetWebhooksForAccountAndEvent is a mock function.
func (m *Repository) GetWebhooksForAccountAndEvent(ctx context.Context, accountID, eventType string) ([]*webhooks.Webhook, error) {
	args := m.Called(ctx, accountID, eventType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*webhooks.Webhook), args.Error(1)
}

// CreateWebhook is a mock function.
func (m *Repository) CreateWebhook(ctx context.Context, input *webhooks.WebhookDatabaseCreationInput) (*webhooks.Webhook, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.Webhook), args.Error(1)
}

// ArchiveWebhook is a mock function.
func (m *Repository) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	args := m.Called(ctx, webhookID, accountID)
	return args.Error(0)
}

// AddWebhookTriggerEvent is a mock function.
func (m *Repository) AddWebhookTriggerEvent(ctx context.Context, accountID string, input *webhooks.WebhookTriggerEventDatabaseCreationInput) (*webhooks.WebhookTriggerEvent, error) {
	args := m.Called(ctx, accountID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*webhooks.WebhookTriggerEvent), args.Error(1)
}

// ArchiveWebhookTriggerEvent is a mock function.
func (m *Repository) ArchiveWebhookTriggerEvent(ctx context.Context, webhookID, webhookTriggerEventID string) error {
	args := m.Called(ctx, webhookID, webhookTriggerEventID)
	return args.Error(0)
}
