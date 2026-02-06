package converters

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/dataprivacy"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	auditconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/audit/grpc/converters"
	identityconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/grpc/converters"
	issuereportsconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/issuereports/grpc/converters"
	notificationsconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/notifications/grpc/converters"
	settingsconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/settings/grpc/converters"
	uploadedmediaconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/uploadedmedia/grpc/converters"
	waitlistsconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/waitlists/grpc/converters"
	webhooksconverters "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/webhooks/grpc/converters"
)

// ConvertUserDataCollectionToGRPCUserDataCollection converts a domain UserDataCollection to a proto UserDataCollection.
func ConvertUserDataCollectionToGRPCUserDataCollection(collection *dataprivacy.UserDataCollection, reportID string) *dataprivacysvc.UserDataCollection {
	result := &dataprivacysvc.UserDataCollection{
		ReportId: reportID,
	}

	// Convert identity data
	result.IdentityDataCollection = identityconverters.ConvertUserDataCollectionToGRPCDataCollection(&collection.Identity)

	// Convert settings data
	result.SettingsDataCollection = settingsconverters.ConvertUserDataCollectionToGRPCDataCollection(&collection.Settings)

	// Convert webhooks data
	result.WebhooksDataCollection = webhooksconverters.ConvertUserDataCollectionToGRPCDataCollection(&collection.Webhooks)

	// Convert notifications data
	result.NotificationsDataCollection = notificationsconverters.ConvertUserDataCollectionToGRPCDataCollection(&collection.Notifications)

	// Convert audit log entries
	for i := range collection.AuditLogEntries {
		result.AuditLogEntries = append(result.AuditLogEntries, auditconverters.ConvertAuditLogEntryToGRPCAuditLogEntry(&collection.AuditLogEntries[i]))
	}

	// Convert issue reports
	for i := range collection.IssueReports {
		result.IssueReports = append(result.IssueReports, issuereportsconverters.ConvertIssueReportToGRPCIssueReport(&collection.IssueReports[i]))
	}

	// Convert uploaded media
	for i := range collection.UploadedMedia {
		result.UploadedMedia = append(result.UploadedMedia, uploadedmediaconverters.ConvertUploadedMediaToGRPCUploadedMedia(&collection.UploadedMedia[i]))
	}

	// Convert waitlist signups
	for i := range collection.WaitlistSignups {
		result.WaitlistSignups = append(result.WaitlistSignups, waitlistsconverters.ConvertWaitlistSignupToGRPCWaitlistSignup(&collection.WaitlistSignups[i]))
	}

	return result
}
