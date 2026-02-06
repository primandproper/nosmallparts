-- name: DeleteExpiredOAuth2ClientTokens :execrows
DELETE FROM oauth2_client_tokens WHERE code_expires_at < (NOW() - interval '1 day') AND access_expires_at < (NOW() - interval '1 day') AND refresh_expires_at < (NOW() - interval '1 day');

-- name: DestroyAllData :exec
TRUNCATE account_invitations, account_user_memberships, accounts, audit_log_entries, issue_reports, oauth2_client_tokens, oauth2_clients, password_reset_tokens, play_script_cast_members, play_script_lines, play_scripts, service_setting_configurations, service_settings, uploaded_media, user_data_disclosures, user_notifications, users, waitlist_signups, waitlists, webhook_trigger_events, webhooks CASCADE;
