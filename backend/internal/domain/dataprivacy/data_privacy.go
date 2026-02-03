package dataprivacy

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/webhooks"
)

type (
	// DataDeletionResponse is returned when a user requests their data be deleted.
	DataDeletionResponse struct {
		Successful bool `json:"Successful"`
	}

	// UserDataAggregationRequest represents a message queue event meant to aggregate data for a user.
	UserDataAggregationRequest struct {
		_ struct{} `json:"-"`

		RequestID string `json:"id"`
		ReportID  string `json:"reportID"`
		UserID    string `json:"userID"`
	}

	// UserDataCollectionResponse represents the response to a UserDataAggregationRequest.
	UserDataCollectionResponse struct {
		_ struct{} `json:"-"`

		ReportID string `json:"reportID"`
	}

	UserDataCollection struct {
		Webhooks      webhooks.UserDataCollection      `json:"webhooks,omitempty"`
		Identity      identity.UserDataCollection      `json:"identity"`
		Settings      settings.UserDataCollection      `json:"settings,omitempty"`
		Notifications notifications.UserDataCollection `json:"notifications,omitempty"`
	}

	// DataPrivacyDataManager contains data privacy management functions.
	DataPrivacyDataManager interface {
		FetchUserDataCollection(ctx context.Context, userID string) (*UserDataCollectionResponse, error)
		DeleteUser(ctx context.Context, userID string) error
	}
)
