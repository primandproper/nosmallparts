package main

import (
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/filtering"
	identitysvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

const (
	accountIDURLParamKey = "accountID"
)

func (s *AdminFrontendServer) AccountPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Accounts", s.renderAccountsError("Error: No API client available")), nil
	}

	accountID := s.accountIDRouteParamFetcher(req)
	if accountID == "" {
		return page("Accounts", s.renderAccountsError("Error: No account ID provided")), nil
	}

	accountsRes, err := c.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: accountID})
	if err != nil {
		return page("Accounts", s.renderAccountsError(fmt.Sprintf("Error loading account: %v", err))), nil
	}

	if accountsRes == nil || accountsRes.Result == nil {
		return page("Accounts", s.renderAccountsError("Error: Account not found")), nil
	}

	account := accountsRes.Result

	// Build billing status options with IsDefault set based on account's current status
	currentBillingStatus := account.BillingStatus
	billingStatusOptions := []*components.SelectOption{
		{Value: identity.UnpaidAccountBillingStatus, Label: "Unpaid", IsDefault: currentBillingStatus == identity.UnpaidAccountBillingStatus || currentBillingStatus == ""},
		{Value: identity.PaidAccountBillingStatus, Label: "Paid", IsDefault: currentBillingStatus == identity.PaidAccountBillingStatus},
		{Value: identity.TrialAccountBillingStatus, Label: "Trial", IsDefault: currentBillingStatus == identity.TrialAccountBillingStatus},
		{Value: identity.SuspendedAccountBillingStatus, Label: "Suspended", IsDefault: currentBillingStatus == identity.SuspendedAccountBillingStatus},
	}

	// Use the new FormPage component for editing account data
	formPageResult, err := components.FormPage(&components.FormPageProps[*identitysvc.Account]{
		Title:        "Account Details",
		BaseSubtitle: "View and edit account information",
		Palette:      &design.StandardPalette,
		Data:         account,
		FormOptions: &components.FormOptions[*identitysvc.Account]{
			Palette: &design.StandardPalette,
			FormID:  "edit-account-form",
			Action:  fmt.Sprintf("/api/accounts/%s", account.Id),
			Method:  "PUT",
			// With the new auto-enable feature, we only need to explicitly list fields
			// that should be editable even when they already have a value.
			// Empty/zero-value fields are automatically editable.
			EnabledFields: []string{
				"Name",
				"BillingStatus",
			},
			FieldConfigs: map[string]*components.FieldConfig{
				"Name": {
					Validation: &components.FieldValidation{
						Required: true,
					},
				},
				"ContactPhone": {
					InputType:   "tel",
					Placeholder: "Enter contact phone number...",
				},
				"BillingStatus": {
					Options:     billingStatusOptions,
					Placeholder: "Select billing status...",
					Validation: &components.FieldValidation{
						Required:      true,
						CustomMessage: "Please select a billing status",
					},
				},
				"AddressLine1": {
					Placeholder: "Enter address...",
				},
				"AddressLine2": {
					Placeholder: "Enter address line 2...",
					Validation: &components.FieldValidation{
						CustomMessage: "(optional)",
					},
				},
				"City": {
					Placeholder: "Enter city...",
				},
				"State": {
					Placeholder: "Enter state/province...",
				},
				"ZipCode": {
					InputType:   "tel",
					Placeholder: "Enter zip/postal code...",
					Validation: &components.FieldValidation{
						Pattern:       "[0-9]{5}(-[0-9]{4})?",
						CustomMessage: "Enter a valid US zip code (e.g., 12345 or 12345-6789)",
					},
				},
				"Country": {
					Placeholder: "Enter country...",
				},
			},
			FormRows: []*components.FormRow{
				{
					Fields:  []string{"Name"},
					Columns: 1,
				},
				{
					Fields:  []string{"ContactPhone", "BillingStatus"},
					Columns: 2,
				},
				{
					Fields:  []string{"AddressLine1"},
					Columns: 1,
				},
				{
					Fields:  []string{"AddressLine2"},
					Columns: 1,
				},
				{
					Fields:  []string{"City", "State"},
					Columns: 2,
				},
				{
					Fields:  []string{"ZipCode", "Country"},
					Columns: 2,
				},
			},

			SubmitButtonText: "Save Changes",
			ShowCancelButton: true,
			CancelButtonText: "Back to Accounts",
			CancelURL:        "/accounts",

			HTMXTarget:  "body",
			HTMXSwap:    "innerHTML",
			HTMXPushURL: true,
		},

		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Accounts", URL: "/accounts"},
			{Text: account.Name, URL: ""},
		},

		// Dynamic subtitle showing account info
		SubtitleGenerator: func(u *identitysvc.Account) string {
			return fmt.Sprintf("Editing account: %s", u.Name)
		},
	})
	if err != nil {
		return page("Accounts", s.renderAccountsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	// Create users section that will be loaded via HTMX
	usersSection := components.CardWithHeader(
		"Account Users",
		&design.StandardPalette,
		nil, // No header actions for now
		g.El("div",
			g.Attr("id", "account-users-container"),
			g.Attr("hx-get", fmt.Sprintf("/api/accounts/%s/users", accountID)),
			g.Attr("hx-trigger", "load"),
			g.Attr("hx-swap", "innerHTML"),
			components.LoadingSpinner(&design.StandardPalette),
		),
	)

	// Combine form and users section
	return page("Accounts",
		ghtml.Div(
			ghtml.Class("space-y-6"),
			formPageResult.Node,
			usersSection,
		),
	), nil
}

func (s *AdminFrontendServer) AccountUsersList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	accountID := s.accountIDRouteParamFetcher(req)
	if accountID == "" {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text("Error: No account ID provided"),
			),
		), nil
	}

	// Get page parameter from query string
	var pageSize uint32 = 10 // Default page size

	usersRes, err := c.GetUsersForAccount(ctx, &identitysvc.GetUsersForAccountRequest{
		AccountId: accountID,
		Filter: &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		},
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading users: %v", err)),
			),
		), nil
	}

	// If no users, show empty state
	if len(usersRes.Results) == 0 {
		return g.El("div",
			g.Attr("class", "text-center py-8"),
			ghtml.P(
				ghtml.Class("text-sm text-gray-500"),
				g.Text("This account has no users."),
			),
		), nil
	}

	// Create compact table for users
	table, err := components.Table(usersRes.Results, &components.TableOptions[*identitysvc.User]{
		TableID: "account-users-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"ID",
			"Username",
			"CreatedAt",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"CreatedAt": renderTimestamp,
		},
		RowLinkGenerator: func(user *identitysvc.User) string {
			// Link to user details page
			return fmt.Sprintf("/users/%s", user.Id)
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
	if usersRes.Pagination != nil && usersRes.Pagination.Cursor != "" {
		paginationControls = append(paginationControls,
			ghtml.Div(
				ghtml.Class("flex justify-between items-center mt-4 pt-4 border-t border-gray-200"),
				ghtml.Div(
					ghtml.Class("text-sm text-gray-500"),
					g.Text(fmt.Sprintf("Showing %d users", len(usersRes.Results))),
				),
				ghtml.Button(
					ghtml.Class(fmt.Sprintf("px-4 py-2 text-sm font-medium rounded-md %s %s hover:%s",
						design.TextColor(design.Color{Value: "white"}),
						design.Background(design.StandardPalette.Primary),
						design.Background(design.Color{Value: design.StandardPalette.Primary.Value + "-700"}),
					)),
					g.Attr("hx-get", fmt.Sprintf("/api/accounts/%s/users?page=%s", accountID, usersRes.Pagination.Cursor)),
					g.Attr("hx-target", "#account-users-container"),
					g.Attr("hx-swap", "innerHTML"),
					g.Text("Load More"),
				),
			),
		)
	} else if len(usersRes.Results) > 0 {
		paginationControls = append(paginationControls,
			ghtml.Div(
				ghtml.Class("text-center mt-4 pt-4 border-t border-gray-200"),
				ghtml.P(
					ghtml.Class("text-sm text-gray-500"),
					g.Text(fmt.Sprintf("Showing all %d user(s)", len(usersRes.Results))),
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

func (s *AdminFrontendServer) AccountsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Accounts", s.renderAccountsError("Error: No API client available")), nil
	}

	// Extract QueryFilter and convert to gRPC filter
	queryFilter, grpcFilter := buildQueryFilterFromRequest(req)

	accountsRes, err := c.GetAccounts(ctx, &identitysvc.GetAccountsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return page("Accounts", s.renderAccountsError(fmt.Sprintf("Error loading accounts: %v", err))), nil
	}

	// Build pagination from response
	pagination := buildPaginationFromGRPCResponse(accountsRes.Pagination)
	// Use main page URL for pagination to enable deep linking
	// HTMX will handle partial updates via the target selector
	paginationURLGenerator := buildPaginationURLGeneratorForSearch(req, "/api/accounts/search", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/accounts", queryFilter)

	// Use the new integrated TablePage component
	tablePageResult, err := components.TablePage(&components.TablePageProps[*identitysvc.Account]{
		Title:             "Accounts",
		BaseSubtitle:      "Manage account accounts",
		Palette:           &design.StandardPalette,
		ShowSearch:        true,
		SearchPlaceholder: "Search accounts...",
		HTMXSearchTarget:  "/api/accounts/search",
		Data:              accountsRes.Results,
		Actions:           []g.Node{},
		TableOptions: &components.TableOptions[*identitysvc.Account]{
			TableID: "accounts-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"ID",
				"Name",
				"BillingStatus",
				"ContactPhone",
				"PaymentProcessorCustomerID",
				"SubscriptionPlanID",
				"CreatedAt",
				"LastUpdatedAt",
				"ArchivedAt",
			},
			FieldReplacements: map[string]string{
				"PaymentProcessorCustomerID": "Payment Processor ID",
				"SubscriptionPlanID":         "Subscription Plan",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"CreatedAt":     renderTimestamp,
				"LastUpdatedAt": renderTimestamp,
				"ArchivedAt":    renderTimestamp,
			},
			Pagination:             pagination,
			PaginationURLGenerator: paginationURLGenerator,
			DeepLinkURLGenerator:   deepLinkURLGenerator,
			PaginationHTMXTarget:   "#search-results",
		},
		RowLinkGenerator: func(data *identitysvc.Account) string {
			return fmt.Sprintf("/accounts/%s", data.Id)
		},
		EmptyStateTitle:       "No accounts found",
		EmptyStateDescription: "Get started by creating your first account account.",
		EmptyStateActions:     []g.Node{},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "Manage account accounts"
			}
			return fmt.Sprintf("Manage %d account accounts", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("Accounts", s.renderAccountsError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("Accounts", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) AccountsSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	accountsRes, err := c.GetAccounts(ctx, &identitysvc.GetAccountsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading accounts: %v", err)),
			),
		), nil
	}

	// Build pagination from response
	pagination := buildPaginationFromGRPCResponse(accountsRes.Pagination)
	// Use main page URL for pagination to enable deep linking
	// HTMX will handle partial updates via the target selector
	paginationURLGenerator := buildPaginationURLGeneratorForSearch(req, "/api/accounts/search", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/accounts", queryFilter)

	// Generate just the table (not the full page)
	if len(accountsRes.Results) == 0 {
		searchQuery := req.URL.Query().Get("search")
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No accounts found",
				fmt.Sprintf("No accounts match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{
					components.ActionButton("Add Account", "/accounts/new", &design.StandardPalette, true),
				},
			),
		), nil
	}

	table, err := components.Table(accountsRes.Results, &components.TableOptions[*identitysvc.Account]{
		TableID: "accounts-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"ID",
			"Name",
			"BillingStatus",
			"ContactPhone",
			"PaymentProcessorCustomerID",
			"SubscriptionPlanID",
			"CreatedAt",
			"LastUpdatedAt",
			"ArchivedAt",
		},
		FieldReplacements: map[string]string{
			"PaymentProcessorCustomerID": "Payment Processor ID",
			"SubscriptionPlanID":         "Subscription Plan",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"CreatedAt":     renderTimestamp,
			"LastUpdatedAt": renderTimestamp,
			"ArchivedAt":    renderTimestamp,
		},
		Pagination:             pagination,
		PaginationURLGenerator: paginationURLGenerator,
		DeepLinkURLGenerator:   deepLinkURLGenerator,
		PaginationHTMXTarget:   "#search-results",
		RowLinkGenerator: func(data *identitysvc.Account) string {
			return fmt.Sprintf("/accounts/%s", data.Id)
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

// renderAccountsError creates a consistent error display for the accounts page.
func (s *AdminFrontendServer) renderAccountsError(errorMsg string) g.Node {
	return components.ContentContainer(&components.ContentContainerProps{
		Title:    "Accounts",
		Subtitle: "Manage account accounts",
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
