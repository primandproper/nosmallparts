package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	oauthsvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/oauth"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

const (
	oauth2ClientIDURLParamKey = "oauth2ClientID"
)

func (s *AdminFrontendServer) OAuth2ClientPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError("Error: No API client available")), nil
	}

	oauth2ClientID := s.oauth2ClientIDRouteParamFetcher(req)
	if oauth2ClientID == "" {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError("Error: No OAuth2 client ID provided")), nil
	}

	oauth2ClientRes, err := c.GetOAuth2Client(ctx, &oauthsvc.GetOAuth2ClientRequest{Oauth2ClientId: oauth2ClientID})
	if err != nil {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError(fmt.Sprintf("Error loading OAuth2 client: %v", err))), nil
	}

	if oauth2ClientRes == nil || oauth2ClientRes.Result == nil {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError("Error: OAuth2 client not found")), nil
	}

	oauth2Client := oauth2ClientRes.Result

	// Use the new FormPage component for viewing OAuth2 client data
	formPageResult, err := components.FormPage(&components.FormPageProps[*oauthsvc.OAuth2Client]{
		Title:        "OAuth2 Client Details",
		BaseSubtitle: "View OAuth2 client information",
		Palette:      &design.StandardPalette,
		Data:         oauth2Client,
		FormOptions: &components.FormOptions[*oauthsvc.OAuth2Client]{
			Palette: &design.StandardPalette,
			FormID:  "view-oauth2-client-form",
			Action:  fmt.Sprintf("/api/oauth2_clients/%s", oauth2Client.Id),
			Method:  "PUT",

			// All fields are read-only for OAuth2 clients
			EnabledFields: []string{},

			FieldConfigs: map[string]*components.FieldConfig{
				"Name": {
					Validation: &components.FieldValidation{
						Required: true,
					},
				},
				"Description": {
					Placeholder: "Description of the OAuth2 client...",
				},
				"ClientID": {
					Placeholder: "OAuth2 Client ID",
				},
				"ClientSecret": {
					InputType:   "password",
					Placeholder: "OAuth2 Client Secret (hidden)",
				},
			},

			FormRows: []*components.FormRow{
				{
					Fields:  []string{"Name"},
					Columns: 1,
				},
				{
					Fields:  []string{"Description"},
					Columns: 1,
				},
				{
					Fields:  []string{"ClientID"},
					Columns: 1,
				},
				{
					Fields:  []string{"ClientSecret"},
					Columns: 1,
				},
			},

			SubmitButtonText: "Save Changes",
			ShowCancelButton: true,
			CancelButtonText: "Back to OAuth2 Clients",
			CancelURL:        "/oauth2_clients",

			HTMXTarget:  "body",
			HTMXSwap:    "innerHTML",
			HTMXPushURL: true,
		},

		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "OAuth2 Clients", URL: "/oauth2_clients"},
			{Text: oauth2Client.Name, URL: ""},
		},

		// Dynamic subtitle showing OAuth2 client info
		SubtitleGenerator: func(c *oauthsvc.OAuth2Client) string {
			return fmt.Sprintf("Viewing OAuth2 client: %s", c.Name)
		},
	})
	if err != nil {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	// Return just the form page
	return page("OAuth2 Clients", formPageResult.Node), nil
}

func (s *AdminFrontendServer) OAuth2ClientsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError("Error: No API client available")), nil
	}

	oauth2ClientsRes, err := c.GetOAuth2Clients(ctx, &oauthsvc.GetOAuth2ClientsRequest{})
	if err != nil {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError(fmt.Sprintf("Error loading OAuth2 clients: %v", err))), nil
	}

	// Use the new integrated TablePage component
	tablePageResult, err := components.TablePage(&components.TablePageProps[*oauthsvc.OAuth2Client]{
		Title:             "OAuth2 Clients",
		BaseSubtitle:      "Manage OAuth2 client applications",
		Palette:           &design.StandardPalette,
		ShowSearch:        true,
		SearchPlaceholder: "Search OAuth2 clients...",
		HTMXSearchTarget:  "/api/oauth2_clients/search",
		Data:              oauth2ClientsRes.Results,
		Actions:           []g.Node{},
		TableOptions: &components.TableOptions[*oauthsvc.OAuth2Client]{
			TableID: "oauth2-clients-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"ID",
				"Name",
				"Description",
				"ClientID",
				"CreatedAt",
				"ArchivedAt",
			},
			FieldReplacements: map[string]string{
				"ClientID": "Client ID",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"CreatedAt":  renderTimestamp,
				"ArchivedAt": renderTimestamp,
			},
		},
		RowLinkGenerator: func(data *oauthsvc.OAuth2Client) string {
			return fmt.Sprintf("/oauth2_clients/%s", data.Id)
		},
		EmptyStateTitle:       "No OAuth2 clients found",
		EmptyStateDescription: "No OAuth2 clients have been created yet.",
		EmptyStateActions:     []g.Node{},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "Manage OAuth2 client applications"
			}
			return fmt.Sprintf("Manage %d OAuth2 client applications", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("OAuth2 Clients", s.renderOAuth2ClientsError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("OAuth2 Clients", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) OAuth2ClientsSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	// Get search query from request
	searchQuery := req.URL.Query().Get("search")

	oauth2ClientsRes, err := c.GetOAuth2Clients(ctx, &oauthsvc.GetOAuth2ClientsRequest{})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading OAuth2 clients: %v", err)),
			),
		), nil
	}

	// Filter OAuth2 clients based on search query
	var filteredOAuth2Clients []*oauthsvc.OAuth2Client
	if searchQuery == "" {
		// No search query, return all OAuth2 clients
		filteredOAuth2Clients = oauth2ClientsRes.Results
	} else {
		// Filter OAuth2 clients by search query (case insensitive)
		searchQueryLower := strings.ToLower(searchQuery)
		for _, oauth2Client := range oauth2ClientsRes.Results {
			if strings.Contains(strings.ToLower(oauth2Client.Name), searchQueryLower) ||
				strings.Contains(strings.ToLower(oauth2Client.Description), searchQueryLower) ||
				strings.Contains(strings.ToLower(oauth2Client.ClientId), searchQueryLower) {
				filteredOAuth2Clients = append(filteredOAuth2Clients, oauth2Client)
			}
		}
	}

	// Generate just the table (not the full page)
	if len(filteredOAuth2Clients) == 0 {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No OAuth2 clients found",
				fmt.Sprintf("No OAuth2 clients match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{},
			),
		), nil
	}

	table, err := components.Table(filteredOAuth2Clients, &components.TableOptions[*oauthsvc.OAuth2Client]{
		TableID: "oauth2-clients-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"ID",
			"Name",
			"Description",
			"ClientID",
			"CreatedAt",
			"ArchivedAt",
		},
		FieldReplacements: map[string]string{
			"ClientID": "Client ID",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"CreatedAt":  renderTimestamp,
			"ArchivedAt": renderTimestamp,
		},
		RowLinkGenerator: func(data *oauthsvc.OAuth2Client) string {
			return fmt.Sprintf("/oauth2_clients/%s", data.Id)
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

// renderOAuth2ClientsError creates a consistent error display for the OAuth2 clients page.
func (s *AdminFrontendServer) renderOAuth2ClientsError(errorMsg string) g.Node {
	return components.ContentContainer(&components.ContentContainerProps{
		Title:    "OAuth2 Clients",
		Subtitle: "Manage OAuth2 client applications",
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
