package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/pflag"
)

var (
	checkOnlyFlag = pflag.Bool("check", false, "only check if files match")
	databaseFlag  = pflag.String("database", postgres, "what database to use")
)

const (
	postgres = "postgres"
)

func main() {
	pflag.Parse()

	runErrors := &multierror.Error{}
	databaseToUse := *databaseFlag

	queryOutput := map[string][]*Query{
		"internal/repositories/postgres/maintenance/sqlc_queries/maintenance.sql":                 buildMaintenanceQueries(databaseToUse),
		"internal/repositories/postgres/oauth/sqlc_queries/oauth2_client_tokens.sql":              buildOAuth2ClientTokensQueries(databaseToUse),
		"internal/repositories/postgres/oauth/sqlc_queries/oauth2_clients.sql":                    buildOAuth2ClientsQueries(databaseToUse),
		"internal/repositories/postgres/identity/sqlc_queries/account_invitations.sql":            buildAccountInvitationsQueries(databaseToUse),
		"internal/repositories/postgres/identity/sqlc_queries/account_user_memberships.sql":       buildAccountUserMembershipsQueries(databaseToUse),
		"internal/repositories/postgres/identity/sqlc_queries/accounts.sql":                       buildAccountsQueries(databaseToUse),
		"internal/repositories/postgres/auditlogentries/sqlc_queries/audit_logs.sql":              buildAuditLogEntryQueries(databaseToUse),
		"internal/repositories/postgres/identity/sqlc_queries/admin.sql":                          buildAdminQueries(databaseToUse),
		"internal/repositories/postgres/auth/sqlc_queries/password_reset_tokens.sql":              buildPasswordResetTokensQueries(databaseToUse),
		"internal/repositories/postgres/identity/sqlc_queries/users.sql":                          buildUsersQueries(databaseToUse),
		"internal/repositories/postgres/settings/sqlc_queries/service_settings.sql":               buildServiceSettingQueries(databaseToUse),
		"internal/repositories/postgres/settings/sqlc_queries/service_setting_configurations.sql": buildServiceSettingConfigurationQueries(databaseToUse),
		"internal/repositories/postgres/webhooks/sqlc_queries/webhooks.sql":                       buildWebhooksQueries(databaseToUse),
		"internal/repositories/postgres/webhooks/sqlc_queries/webhook_trigger_events.sql":         buildWebhookTriggerEventsQueries(databaseToUse),
		"internal/repositories/postgres/notifications/sqlc_queries/user_notifications.sql":        buildUserNotificationQueries(databaseToUse),
		"internal/repositories/postgres/waitlists/sqlc_queries/waitlists.sql":                     buildWaitlistsQueries(databaseToUse),
		"internal/repositories/postgres/waitlists/sqlc_queries/waitlist_signups.sql":              buildWaitlistSignupsQueries(databaseToUse),
		"internal/repositories/postgres/issuereports/sqlc_queries/issue_reports.sql":              buildIssueReportsQueries(databaseToUse),
		"internal/repositories/postgres/uploadedmedia/sqlc_queries/uploaded_media.sql":            buildUploadedMediaQueries(databaseToUse),
		"internal/repositories/postgres/dataprivacy/sqlc_queries/user_data_disclosures.sql":       buildUserDataDisclosuresQueries(databaseToUse),
		"internal/repositories/postgres/playscripts/sqlc_queries/play_scripts.sql":                buildPlayScriptsQueries(databaseToUse),
		"internal/repositories/postgres/playscripts/sqlc_queries/play_script_cast_members.sql":    buildPlayScriptCastMembersQueries(databaseToUse),
		"internal/repositories/postgres/playscripts/sqlc_queries/play_script_lines.sql":           buildPlayScriptLinesQueries(databaseToUse),
		"internal/repositories/postgres/playscripts/sqlc_queries/script_tags.sql":                 buildScriptTagsQueries(databaseToUse),
		"internal/repositories/postgres/playscripts/sqlc_queries/play_script_ratings.sql":         buildPlayScriptRatingsQueries(databaseToUse),
	}

	checkOnly := *checkOnlyFlag

	for filePath, queries := range queryOutput {
		existingFile, err := os.ReadFile(filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				if _, err = os.Create(filePath); err != nil {
					log.Fatal(fmt.Errorf("creating file: %w", err))
				}
			}
			if err != nil {
				log.Fatal(fmt.Errorf("opening existing file: %w", err))
			}
		}

		var fileContent string
		for i, query := range queries {
			if i != 0 {
				fileContent += "\n"
			}
			fileContent += query.Render()
		}

		fileOutput := ""
		for _, line := range strings.Split(strings.TrimSpace(fileContent), "\n") {
			fileOutput += strings.TrimSuffix(line, " ") + "\n"
		}

		if string(existingFile) != fileOutput && checkOnly {
			runErrors = multierror.Append(runErrors, fmt.Errorf("files don't match: %s", filePath))
		}

		if !checkOnly {
			if err = os.WriteFile(filePath, []byte(fileOutput), 0o600); err != nil {
				runErrors = multierror.Append(runErrors, err)
			}
		}
	}

	if runErrors.ErrorOrNil() != nil {
		log.Fatal(runErrors)
	}
}
