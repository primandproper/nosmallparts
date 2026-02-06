package main

import (
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/filtering"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

const (
	waitlistIDURLParamKey = "waitlistID"
)

func (s *AdminFrontendServer) renderWaitlistsError(message string) g.Node {
	return ghtml.Div(
		ghtml.Class("max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8"),
		ghtml.Div(
			ghtml.Class(fmt.Sprintf("rounded-md p-4 %s", design.Background(design.StandardPalette.Warning))),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-sm font-medium %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(message),
			),
		),
	)
}

func (s *AdminFrontendServer) WaitlistsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Waitlists", s.renderWaitlistsError("Error: No API client available")), nil
	}

	queryFilter, grpcFilter := buildQueryFilterFromRequest(req)

	waitlistsRes, err := c.GetWaitlists(ctx, &waitlistssvc.GetWaitlistsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return page("Waitlists", s.renderWaitlistsError(fmt.Sprintf("Error loading waitlists: %v", err))), nil
	}

	pagination := buildPaginationFromGRPCResponse(waitlistsRes.Pagination)
	paginationURLGenerator := buildPaginationURLGenerator(req, "/waitlists", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/waitlists", queryFilter)

	tablePageResult, err := components.TablePage(&components.TablePageProps[*waitlistssvc.Waitlist]{
		Title:             "Waitlists",
		BaseSubtitle:      "Manage waitlists",
		Palette:           &design.StandardPalette,
		ShowSearch:        true,
		SearchPlaceholder: "Search waitlists...",
		HTMXSearchTarget:  "/api/waitlists/search",
		Data:              waitlistsRes.Results,
		Actions:           []g.Node{},
		TableOptions: &components.TableOptions[*waitlistssvc.Waitlist]{
			TableID: "waitlists-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"MealPlanTaskID",
				"Name",
				"Description",
				"ValidUntil",
				"CreatedAt",
				"LastUpdatedAt",
				"ArchivedAt",
			},
			FieldReplacements: map[string]string{
				"ValidUntil": "Valid Until",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"ValidUntil":    renderTimestamp,
				"CreatedAt":     renderTimestamp,
				"LastUpdatedAt": renderTimestamp,
				"ArchivedAt":    renderTimestamp,
			},
			Pagination:             pagination,
			PaginationURLGenerator: paginationURLGenerator,
			DeepLinkURLGenerator:   deepLinkURLGenerator,
			PaginationHTMXTarget:   "#search-results",
		},
		RowLinkGenerator: func(data *waitlistssvc.Waitlist) string {
			return fmt.Sprintf("/waitlists/%s", data.Id)
		},
		EmptyStateTitle:       "No waitlists found",
		EmptyStateDescription: "Get started by creating your first waitlist.",
		EmptyStateActions:     []g.Node{},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "Manage waitlists"
			}
			return fmt.Sprintf("Manage %d waitlists", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("Waitlists", s.renderWaitlistsError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("Waitlists", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) WaitlistsSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	queryFilter, grpcFilter := buildQueryFilterFromRequest(req)

	waitlistsRes, err := c.GetWaitlists(ctx, &waitlistssvc.GetWaitlistsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading waitlists: %v", err)),
			),
		), nil
	}

	if len(waitlistsRes.Results) == 0 {
		searchQuery := req.URL.Query().Get("search")
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No waitlists found",
				fmt.Sprintf("No waitlists match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{},
			),
		), nil
	}

	pagination := buildPaginationFromGRPCResponse(waitlistsRes.Pagination)
	paginationURLGenerator := buildPaginationURLGeneratorForSearch(req, "/api/waitlists/search", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/waitlists", queryFilter)

	table, err := components.Table(waitlistsRes.Results, &components.TableOptions[*waitlistssvc.Waitlist]{
		TableID: "waitlists-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"MealPlanTaskID",
			"Name",
			"Description",
			"ValidUntil",
			"CreatedAt",
			"LastUpdatedAt",
			"ArchivedAt",
		},
		FieldReplacements: map[string]string{
			"ValidUntil": "Valid Until",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"ValidUntil":    renderTimestamp,
			"CreatedAt":     renderTimestamp,
			"LastUpdatedAt": renderTimestamp,
			"ArchivedAt":    renderTimestamp,
		},
		Pagination:             pagination,
		PaginationURLGenerator: paginationURLGenerator,
		DeepLinkURLGenerator:   deepLinkURLGenerator,
		PaginationHTMXTarget:   "#search-results",
		RowLinkGenerator: func(data *waitlistssvc.Waitlist) string {
			return fmt.Sprintf("/waitlists/%s", data.Id)
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

	return g.El("div",
		g.Attr("class", "overflow-x-auto"),
		table,
	), nil
}

func (s *AdminFrontendServer) WaitlistPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Waitlists", s.renderWaitlistsError("Error: No API client available")), nil
	}

	waitlistID := s.waitlistIDRouteParamFetcher(req)
	if waitlistID == "" {
		return page("Waitlists", s.renderWaitlistsError("Error: No waitlist MealPlanTaskID provided")), nil
	}

	waitlistRes, err := c.GetWaitlist(ctx, &waitlistssvc.GetWaitlistRequest{WaitlistId: waitlistID})
	if err != nil {
		return page("Waitlists", s.renderWaitlistsError(fmt.Sprintf("Error loading waitlist: %v", err))), nil
	}

	if waitlistRes == nil || waitlistRes.Result == nil {
		return page("Waitlists", s.renderWaitlistsError("Error: Waitlist not found")), nil
	}

	waitlist := waitlistRes.Result

	// Use FormPage for viewing/editing waitlist details
	formPageResult, err := components.FormPage(&components.FormPageProps[*waitlistssvc.Waitlist]{
		Title:        "Waitlist Details",
		BaseSubtitle: "View and edit waitlist information",
		Palette:      &design.StandardPalette,
		Data:         waitlist,
		FormOptions: &components.FormOptions[*waitlistssvc.Waitlist]{
			FormID: "edit-waitlist-form",
			Action: fmt.Sprintf("/api/waitlists/%s", waitlist.Id),
			Method: "PUT",

			// Enable editable fields
			EnabledFields: []string{
				"Name",
				"Description",
				"ValidUntil",
			},

			FieldConfigs: map[string]*components.FieldConfig{
				"Name": {
					Placeholder: "Enter waitlist name...",
					Validation: &components.FieldValidation{
						Required:      true,
						MinLength:     1,
						MaxLength:     255,
						CustomMessage: "Name must be between 1 and 255 characters",
					},
				},
				"Description": {
					Placeholder: "Enter waitlist description...",
					Validation: &components.FieldValidation{
						MaxLength:     1000,
						CustomMessage: "Maximum 1000 characters",
					},
				},
				"ValidUntil": {
					InputType: "datetime-local",
					Validation: &components.FieldValidation{
						Required:      true,
						CustomMessage: "Please select a valid until date",
					},
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
					Fields:  []string{"ValidUntil"},
					Columns: 1,
				},
			},

			SubmitButtonText: "Save Changes",
			ShowCancelButton: true,
			CancelButtonText: "Back to Waitlists",
			CancelURL:        "/waitlists",

			HTMXTarget:  "body",
			HTMXSwap:    "innerHTML",
			HTMXPushURL: true,
		},

		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Waitlists", URL: "/waitlists"},
			{Text: waitlist.Name, URL: ""},
		},

		SubtitleGenerator: func(w *waitlistssvc.Waitlist) string {
			return fmt.Sprintf("Editing waitlist: %s", w.Name)
		},
	})
	if err != nil {
		return page("Waitlists", s.renderWaitlistsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	// Create signups section that will be loaded via HTMX
	signupsSection := components.CardWithHeader(
		"Waitlist Signups",
		&design.StandardPalette,
		nil,
		g.El("div",
			g.Attr("id", "waitlist-signups-container"),
			g.Attr("hx-get", fmt.Sprintf("/api/waitlists/%s/signups", waitlistID)),
			g.Attr("hx-trigger", "load"),
			g.Attr("hx-swap", "innerHTML"),
			components.LoadingSpinner(&design.StandardPalette),
		),
	)

	// Combine form and signups section
	return page("Waitlists",
		ghtml.Div(
			ghtml.Class("space-y-6"),
			formPageResult.Node,
			signupsSection,
		),
	), nil
}

func (s *AdminFrontendServer) WaitlistSignupsForWaitlist(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	waitlistID := s.waitlistIDRouteParamFetcher(req)
	if waitlistID == "" {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text("Error: No waitlist MealPlanTaskID provided"),
			),
		), nil
	}

	// Get page parameter from query string
	var pageSize uint32 = 10 // Default page size
	var nextCursor *string

	pageParam := req.URL.Query().Get("cursor")
	if pageParam != "" {
		nextCursor = &pageParam
	}

	signupsRes, err := c.GetWaitlistSignupsForWaitlist(ctx, &waitlistssvc.GetWaitlistSignupsForWaitlistRequest{
		WaitlistId: waitlistID,
		Filter: &grpcfiltering.QueryFilter{
			MaxResponseSize: &pageSize,
			Cursor:          nextCursor,
		},
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "text-center py-4"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("%s text-sm", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading signups: %v", err)),
			),
		), nil
	}

	// If no signups, show empty state
	if len(signupsRes.Results) == 0 {
		return g.El("div",
			g.Attr("class", "text-center py-8"),
			ghtml.P(
				ghtml.Class("text-sm text-gray-500"),
				g.Text("This waitlist has no signups."),
			),
		), nil
	}

	// Create compact table for signups
	table, err := components.Table(signupsRes.Results, &components.TableOptions[*waitlistssvc.WaitlistSignup]{
		TableID: "waitlist-signups-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"MealPlanTaskID",
			"Notes",
			"BelongsToUser",
			"BelongsToAccount",
			"CreatedAt",
			"LastUpdatedAt",
			"ArchivedAt",
		},
		FieldReplacements: map[string]string{
			"BelongsToUser":    "User MealPlanTaskID",
			"BelongsToAccount": "Account MealPlanTaskID",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"CreatedAt":     renderTimestamp,
			"LastUpdatedAt": renderTimestamp,
			"ArchivedAt":    renderTimestamp,
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
	if signupsRes.Pagination != nil && signupsRes.Pagination.Cursor != "" {
		paginationControls = append(paginationControls,
			ghtml.Div(
				ghtml.Class("flex justify-between items-center mt-4 pt-4 border-t border-gray-200"),
				ghtml.Div(
					ghtml.Class("text-sm text-gray-500"),
					g.Text(fmt.Sprintf("Showing %d signups", len(signupsRes.Results))),
				),
				ghtml.Button(
					ghtml.Class(fmt.Sprintf("px-4 py-2 text-sm font-medium rounded-md %s %s hover:%s",
						design.TextColor(design.Color{Value: "white"}),
						design.Background(design.StandardPalette.Primary),
						design.Background(design.Color{Value: design.StandardPalette.Primary.Value + "-700"}),
					)),
					g.Attr("hx-get", fmt.Sprintf("/api/waitlists/%s/signups?cursor=%s", waitlistID, signupsRes.Pagination.Cursor)),
					g.Attr("hx-target", "#waitlist-signups-container"),
					g.Attr("hx-swap", "innerHTML"),
					g.Text("Load More"),
				),
			),
		)
	} else if len(signupsRes.Results) > 0 {
		// Show count if no more pages
		paginationControls = append(paginationControls,
			ghtml.Div(
				ghtml.Class("flex justify-center items-center mt-4 pt-4 border-t border-gray-200"),
				ghtml.Div(
					ghtml.Class("text-sm text-gray-500"),
					g.Text(fmt.Sprintf("Showing all %d signups", len(signupsRes.Results))),
				),
			),
		)
	}

	return g.El("div",
		g.Attr("class", "overflow-x-auto"),
		table,
		ghtml.Div(paginationControls...),
	), nil
}
