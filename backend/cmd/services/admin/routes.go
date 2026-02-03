package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing"

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
