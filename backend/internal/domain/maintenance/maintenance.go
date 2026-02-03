package maintenance

import "context"

type (
	MaintenanceDataManager interface {
		DeleteExpiredOAuth2ClientTokens(context.Context) (int64, error)
	}
)
