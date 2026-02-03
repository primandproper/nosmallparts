package main

import (
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

const (
	userIDURLParamKey = "userID"
)

func (s *AdminFrontendServer) UserPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Users", s.renderUsersError("Error: No API client available")), nil
	}

	userID := s.userIDRouteParamFetcher(req)
	if userID == "" {
		return page("Users", s.renderUsersError("Error: No user ID provided")), nil
	}

	usersRes, err := c.GetUser(ctx, &identitysvc.GetUserRequest{UserId: userID})
	if err != nil {
		return page("Users", s.renderUsersError(fmt.Sprintf("Error loading users: %v", err))), nil
	}
	user := usersRes.Result

	// Use the new FormPage component for editing user data
	formPageResult, err := components.FormPage(&components.FormPageProps[*identitysvc.User]{
		Title:        "User Details",
		BaseSubtitle: "View and edit user information",
		Palette:      &design.StandardPalette,
		Data:         user,
		FormOptions: &components.FormOptions[*identitysvc.User]{
			FormID: "edit-user-form",
			Action: fmt.Sprintf("/api/users/%s", user.Id),
			Method: "PUT",

			// Enable editable fields
			EnabledFields: []string{
				"AccountStatus",
				"AccountStatusExplanation",
			},

			// Configure field validation
			FieldConfigs: map[string]*components.FieldConfig{
				"AccountStatus": {
					Options: []*components.SelectOption{
						{Value: "good", Label: "Good Standing"},
						{Value: "unverified", Label: "Unverified"},
						{Value: "banned", Label: "Banned"},
						{Value: "terminated", Label: "Terminated"},
					},
					Validation: &components.FieldValidation{
						Required:      true,
						CustomMessage: "Please select an account status",
					},
				},
				"FirstName": {
					Validation: &components.FieldValidation{
						Required:      true,
						MinLength:     2,
						MaxLength:     50,
						CustomMessage: "First name must be between 2 and 50 characters",
					},
				},
				"LastName": {
					Validation: &components.FieldValidation{
						Required:      true,
						MinLength:     2,
						MaxLength:     50,
						CustomMessage: "Last name must be between 2 and 50 characters",
					},
				},
				"EmailAddress": {
					InputType: "email",
					Validation: &components.FieldValidation{
						Required:      true,
						Pattern:       `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
						CustomMessage: "Please enter a valid email address",
					},
				},
				"AccountStatusExplanation": {
					Placeholder: "Enter explanation for account status...",
					Validation: &components.FieldValidation{
						MaxLength:     500,
						CustomMessage: "Maximum 500 characters",
					},
				},
			},

			// Layout configuration: FirstName and LastName together, others separate
			FormRows: []*components.FormRow{
				{
					Fields:  []string{"FirstName", "LastName"},
					Columns: 2,
				},
				{
					Fields:  []string{"EmailAddress"},
					Columns: 1,
				},
				{
					Fields:  []string{"AccountStatus"},
					Columns: 1,
				},
				{
					Fields:  []string{"AccountStatusExplanation"},
					Columns: 1,
				},
			},

			SubmitButtonText: "Save Changes",
			ShowCancelButton: true,
			CancelButtonText: "Back to Users",
			CancelURL:        "/users",

			// HTMX configuration
			HTMXTarget:  "body",
			HTMXSwap:    "innerHTML",
			HTMXPushURL: true,
		},

		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Users", URL: "/users"},
			{Text: user.Username, URL: ""},
		},

		// Dynamic subtitle showing user info
		SubtitleGenerator: func(u *identitysvc.User) string {
			return fmt.Sprintf("Editing user: %s", u.Username)
		},
	})
	if err != nil {
		return page("Users", s.renderUsersError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	// Create accounts section that will be loaded via HTMX
	accountsSection := components.CardWithHeader(
		"User Accounts",
		&design.StandardPalette,
		nil,
		g.El("div",
			g.Attr("id", "user-accounts-container"),
			g.Attr("hx-get", fmt.Sprintf("/api/users/%s/accounts", userID)),
			g.Attr("hx-trigger", "load"),
			g.Attr("hx-swap", "innerHTML"),
			components.LoadingSpinner(&design.StandardPalette),
		),
	)

	// Combine form and accounts section
	return page("Users",
		ghtml.Div(
			ghtml.Class("space-y-6"),
			formPageResult.Node,
			accountsSection,
		),
	), nil
}

func (s *AdminFrontendServer) UserAccountsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text("Error: No API client available"),
			),
		), nil
	}

	userID := s.userIDRouteParamFetcher(req)
	if userID == "" {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text("Error: No user ID provided"),
			),
		), nil
	}

	// Get page parameter from query string
	pageParam := req.URL.Query().Get("page")
	var pageSize uint32 = 10 // Default page size
	var nextCursor *string

	if pageParam != "" {
		nextCursor = &pageParam
	}

	accountsRes, err := c.GetAccountsForUser(ctx, &identitysvc.GetAccountsForUserRequest{
		UserId: userID,
		Filter: &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
			Cursor:          nextCursor,
		},
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading accounts: %v", err)),
			),
		), nil
	}

	// If no accounts, show empty state
	if len(accountsRes.Results) == 0 {
		return g.El("div",
			g.Attr("class", "text-center py-8"),
			ghtml.P(
				ghtml.Class("text-sm text-gray-500"),
				g.Text("This user is not associated with any accounts."),
			),
		), nil
	}

	// Create compact table for accounts
	table, err := components.Table(accountsRes.Results, &components.TableOptions[*identitysvc.Account]{
		TableID: "user-accounts-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"ID",
			"Name",
			"BillingStatus",
			"ContactPhone",
			"City",
			"State",
			"Country",
			"CreatedAt",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"CreatedAt": renderTimestamp,
		},
		RowLinkGenerator: func(account *identitysvc.Account) string {
			return fmt.Sprintf("/accounts/%s", account.Id)
		},
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error creating table: %v", err)),
			),
		), nil
	}

	var paginationControls []g.Node

	// Add pagination controls if there's a next page
	if accountsRes.Pagination != nil && accountsRes.Pagination.Cursor != "" {
		paginationControls = append(paginationControls,
			ghtml.Div(
				ghtml.Class("flex justify-between items-center mt-4 pt-4 border-t border-gray-200"),
				ghtml.Div(
					ghtml.Class("text-sm text-gray-500"),
					g.Text(fmt.Sprintf("Showing %d accounts", len(accountsRes.Results))),
				),
				ghtml.Button(
					ghtml.Class(fmt.Sprintf("px-4 py-2 text-sm font-medium rounded-md %s %s hover:%s",
						design.TextColor(design.Color{Value: "white"}),
						design.Background(design.StandardPalette.Primary),
						design.Background(design.Color{Value: design.StandardPalette.Primary.Value + "-700"}),
					)),
					g.Attr("hx-get", fmt.Sprintf("/api/users/%s/accounts?page=%s", userID, accountsRes.Pagination.Cursor)),
					g.Attr("hx-target", "#user-accounts-container"),
					g.Attr("hx-swap", "innerHTML"),
					g.Text("Load More"),
				),
			),
		)
	} else if len(accountsRes.Results) > 0 {
		paginationControls = append(paginationControls,
			ghtml.Div(
				ghtml.Class("text-center mt-4 pt-4 border-t border-gray-200"),
				ghtml.P(
					ghtml.Class("text-sm text-gray-500"),
					g.Text(fmt.Sprintf("Showing all %d account(s)", len(accountsRes.Results))),
				),
			),
		)
	}

	return g.El("div",
		g.Attr("class", "space-y-4"),
		table,
		g.Group(paginationControls),
	), nil
}

func (s *AdminFrontendServer) UsersList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Users", s.renderUsersError("Error: No API client available")), nil
	}

	// Extract QueryFilter and convert to gRPC filter
	queryFilter, grpcFilter := buildQueryFilterFromRequest(req)

	usersRes, err := c.GetUsers(ctx, &identitysvc.GetUsersRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return page("Users", s.renderUsersError(fmt.Sprintf("Error loading users: %v", err))), nil
	}

	// Build pagination from response
	pagination := buildPaginationFromGRPCResponse(usersRes.Pagination)
	// Use search endpoint for pagination buttons to return just the table content
	// The main page URL is still used for deep linking via hx-push-url
	paginationURLGenerator := buildPaginationURLGeneratorForSearch(req, "/api/users/search", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/users", queryFilter)

	// Use the new integrated TablePage component
	tablePageResult, err := components.TablePage(&components.TablePageProps[*identitysvc.User]{
		Title:             "Users",
		BaseSubtitle:      "Manage user accounts",
		Palette:           &design.StandardPalette,
		ShowSearch:        true,
		SearchPlaceholder: "Search users...",
		HTMXSearchTarget:  "/api/users/search",
		Data:              usersRes.Results,
		Actions:           []g.Node{},
		TableOptions: &components.TableOptions[*identitysvc.User]{
			TableID: "users-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"ID",
				"Username",
				"FirstName",
				"LastName",
				"EmailAddress",
				"ServiceRole",
				"AccountStatus",
				"AccountStatusExplanation",
				"Birthday",
				"PasswordLastChangedAt",
				"LastAcceptedTermsOfService",
				"LastAcceptedPrivacyPolicy",
				"TwoFactorSecretVerifiedAt",
				"EmailAddressVerifiedAt",
				"CreatedAt",
				"LastUpdatedAt",
				"ArchivedAt",
			},
			FieldReplacements: map[string]string{
				"EmailAddressVerifiedAt": "Email Verified At",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"Birthday":                  renderTimestamp,
				"CreatedAt":                 renderTimestamp,
				"TwoFactorSecretVerifiedAt": renderTimestamp,
				"LastUpdatedAt":             renderTimestamp,
				"ArchivedAt":                renderTimestamp,
			},
			Pagination:             pagination,
			PaginationURLGenerator: paginationURLGenerator,
			DeepLinkURLGenerator:   deepLinkURLGenerator,
			PaginationHTMXTarget:   "#search-results",
		},
		RowLinkGenerator: func(data *identitysvc.User) string {
			return fmt.Sprintf("/users/%s", data.Id)
		},
		EmptyStateTitle:       "No users found",
		EmptyStateDescription: "Get started by creating your first user account.",
		EmptyStateActions:     []g.Node{},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "Manage user accounts"
			}
			return fmt.Sprintf("Manage %d user accounts", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("Users", s.renderUsersError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("Users", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) UsersSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text("Error: No API client available"),
			),
		), nil
	}

	// Extract QueryFilter and convert to gRPC filter
	queryFilter, grpcFilter := buildQueryFilterFromRequest(req)

	usersRes, err := c.GetUsers(ctx, &identitysvc.GetUsersRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading users: %v", err)),
			),
		), nil
	}

	// Build pagination from response
	pagination := buildPaginationFromGRPCResponse(usersRes.Pagination)
	// Use search endpoint for pagination buttons to return just the table content
	// The main page URL is still used for deep linking via hx-push-url
	paginationURLGenerator := buildPaginationURLGeneratorForSearch(req, "/api/users/search", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/users", queryFilter)

	// Generate just the table (not the full page)
	if len(usersRes.Results) == 0 {
		searchQuery := req.URL.Query().Get("search")
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No users found",
				fmt.Sprintf("No users match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{
					components.ActionButton("Add User", "/users/new", &design.StandardPalette, true),
				},
			),
		), nil
	}

	table, err := components.Table(usersRes.Results, &components.TableOptions[*identitysvc.User]{
		TableID: "users-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"ID",
			"Username",
			"FirstName",
			"LastName",
			"EmailAddress",
			"ServiceRole",
			"AccountStatus",
			"AccountStatusExplanation",
			"Birthday",
			"PasswordLastChangedAt",
			"LastAcceptedTermsOfService",
			"LastAcceptedPrivacyPolicy",
			"TwoFactorSecretVerifiedAt",
			"EmailAddressVerifiedAt",
			"CreatedAt",
			"LastUpdatedAt",
			"ArchivedAt",
		},
		FieldReplacements: map[string]string{
			"EmailAddressVerifiedAt": "Email Verified At",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"Birthday":                  renderTimestamp,
			"CreatedAt":                 renderTimestamp,
			"TwoFactorSecretVerifiedAt": renderTimestamp,
			"LastUpdatedAt":             renderTimestamp,
			"ArchivedAt":                renderTimestamp,
		},
		Pagination:             pagination,
		PaginationURLGenerator: paginationURLGenerator,
		DeepLinkURLGenerator:   deepLinkURLGenerator,
		PaginationHTMXTarget:   "#search-results",
		RowLinkGenerator: func(data *identitysvc.User) string {
			return fmt.Sprintf("/users/%s", data.Id)
		},
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error creating table: %v", err)),
			),
		), nil
	}

	// Wrap table in the same scrollable container structure for consistency
	return g.El("div",
		g.Attr("class", "overflow-x-auto"),
		table,
	), nil
}

// renderUsersError creates a consistent error display for the users page.
func (s *AdminFrontendServer) renderUsersError(errorMsg string) g.Node {
	return components.ContentContainer(&components.ContentContainerProps{
		Title:    "Users",
		Subtitle: "Manage user accounts",
		Palette:  &design.StandardPalette,
	},
		components.Card(&design.StandardPalette,
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(errorMsg),
			),
		),
	)
}
