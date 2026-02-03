package identity

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
		User                   User                    `json:"user"`
		Accounts               []Account               `json:"accounts"`
		AccountInvitations     []AccountInvitation     `json:"account_invitations"`
		AccountUserMemberships []AccountUserMembership `json:"account_user_memberships"`
	}
)
