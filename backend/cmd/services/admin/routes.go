package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/v4/routing"

	ghttp "maragu.dev/gomponents/http"
)

const (
	assetsDir = "./cmd/services/admin/assets"
)

func (s *AdminFrontendServer) setupRoutes(router routing.Router) {
	r := router.WithMiddleware(s.authMiddleware)

	r.Get("/", ghttp.Adapt(s.HomePage))

	r.Get(fmt.Sprintf("/users/{%s}", userIDURLParamKey), ghttp.Adapt(s.UserPage))
	r.Get("/users", ghttp.Adapt(s.UsersList))
	r.Get("/api/users/search", ghttp.Adapt(s.UsersSearch))
	r.Get(fmt.Sprintf("/api/users/{%s}/accounts", userIDURLParamKey), ghttp.Adapt(s.UserAccountsList))

	r.Get(fmt.Sprintf("/accounts/{%s}", accountIDURLParamKey), ghttp.Adapt(s.AccountPage))
	r.Get("/accounts", ghttp.Adapt(s.AccountsList))
	r.Get("/api/accounts/search", ghttp.Adapt(s.AccountsSearch))
	r.Get(fmt.Sprintf("/api/accounts/{%s}/users", accountIDURLParamKey), ghttp.Adapt(s.AccountUsersList))

	r.Get(fmt.Sprintf("/oauth2_clients/{%s}", oauth2ClientIDURLParamKey), ghttp.Adapt(s.OAuth2ClientPage))
	r.Get("/oauth2_clients", ghttp.Adapt(s.OAuth2ClientsList))
	r.Get("/api/oauth2_clients/search", ghttp.Adapt(s.OAuth2ClientsSearch))

	r.Get("/settings/new", ghttp.Adapt(s.SettingNewPage))
	r.Post("/api/settings", ghttp.Adapt(s.SettingCreate))
	r.Get(fmt.Sprintf("/settings/{%s}", settingIDURLParamKey), ghttp.Adapt(s.SettingPage))
	r.Get("/settings", ghttp.Adapt(s.SettingsList))
	r.Get("/api/settings/search", ghttp.Adapt(s.SettingsSearch))

	// Waitlists
	r.Get(fmt.Sprintf("/waitlists/{%s}", waitlistIDURLParamKey), ghttp.Adapt(s.WaitlistPage))
	r.Get("/waitlists", ghttp.Adapt(s.WaitlistsList))
	r.Get("/api/waitlists/search", ghttp.Adapt(s.WaitlistsSearch))
	r.Get(fmt.Sprintf("/api/waitlists/{%s}/signups", waitlistIDURLParamKey), ghttp.Adapt(s.WaitlistSignupsForWaitlist))

	// Issue Reports
	r.Get(fmt.Sprintf("/issue_reports/{%s}", issueReportIDURLParamKey), ghttp.Adapt(s.IssueReportPage))
	r.Get("/issue_reports", ghttp.Adapt(s.IssueReportsList))
	r.Get("/api/issue_reports/search", ghttp.Adapt(s.IssueReportsSearch))

	// Script Tags
	r.Get("/script_tags", ghttp.Adapt(s.ScriptTagsList))
	r.Get("/script_tags/new", ghttp.Adapt(s.ScriptTagNewPage))
	r.Get("/api/script_tags/search", ghttp.Adapt(s.ScriptTagsSearch))
	r.Post("/api/script_tags", ghttp.Adapt(s.ScriptTagCreate))

	// Play Scripts
	r.Get("/play_scripts", ghttp.Adapt(s.PlayScriptsList))
	r.Get("/api/play_scripts/search", ghttp.Adapt(s.PlayScriptsSearch))
	r.Get(fmt.Sprintf("/play_scripts/{%s}", playScriptIDURLParamKey), ghttp.Adapt(s.PlayScriptPage))
	r.Get(fmt.Sprintf("/play_scripts/{%s}/lines/{%s}/edit", playScriptIDURLParamKey, lineIDURLParamKey), ghttp.Adapt(s.PlayScriptLineEditPage))
	r.Get(fmt.Sprintf("/api/play_scripts/{%s}/tags", playScriptIDURLParamKey), ghttp.Adapt(s.PlayScriptTagsSection))
	r.Post(fmt.Sprintf("/api/play_scripts/{%s}", playScriptIDURLParamKey), ghttp.Adapt(s.PlayScriptUpdate))
	r.Post(fmt.Sprintf("/api/play_scripts/{%s}/lines/{%s}", playScriptIDURLParamKey, lineIDURLParamKey), ghttp.Adapt(s.PlayScriptLineUpdate))
	r.Post(fmt.Sprintf("/api/play_scripts/{%s}/tags", playScriptIDURLParamKey), ghttp.Adapt(s.PlayScriptAddTag))
	r.Delete(fmt.Sprintf("/api/play_scripts/{%s}/tags/{%s}", playScriptIDURLParamKey, tagIDURLParamKey), ghttp.Adapt(s.PlayScriptRemoveTag))

	router.Get("/login", ghttp.Adapt(s.LoginPage))
	router.Post("/login/submit", ghttp.Adapt(s.LoginSubmission))

	// static files - NOTE: this must be registered last
	fileServer := http.FileServer(http.Dir(assetsDir))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Only serve root-level files (no subdirectories)
		if strings.Contains(r.URL.Path[1:], "/") {
			http.NotFound(w, r)
			return
		}

		// Check if the file exists in the assets directory
		filePath := filepath.Join(assetsDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			// File exists, serve it
			fileServer.ServeHTTP(w, r)
			return
		}

		// File doesn't exist
		http.NotFound(w, r)
	})
}
