package main

import (
	"fmt"
	"net/http"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

const (
	issueReportIDURLParamKey = "issueReportID"
)

func (s *AdminFrontendServer) renderIssueReportsError(message string) g.Node {
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

func (s *AdminFrontendServer) IssueReportsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Issue Reports", s.renderIssueReportsError("Error: No API client available")), nil
	}

	queryFilter, grpcFilter := buildQueryFilterFromRequest(req)

	issueReportsRes, err := c.GetIssueReports(ctx, &issuereportssvc.GetIssueReportsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return page("Issue Reports", s.renderIssueReportsError(fmt.Sprintf("Error loading issue reports: %v", err))), nil
	}

	pagination := buildPaginationFromGRPCResponse(issueReportsRes.Pagination)
	paginationURLGenerator := buildPaginationURLGenerator(req, "/issue_reports", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/issue_reports", queryFilter)

	tablePageResult, err := components.TablePage(&components.TablePageProps[*issuereportssvc.IssueReport]{
		Title:             "Issue Reports",
		BaseSubtitle:      "View and manage user-submitted issue reports",
		Palette:           &design.StandardPalette,
		ShowSearch:        true,
		SearchPlaceholder: "Search issue reports...",
		HTMXSearchTarget:  "/api/issue_reports/search",
		Data:              issueReportsRes.Results,
		Actions:           []g.Node{},
		TableOptions: &components.TableOptions[*issuereportssvc.IssueReport]{
			TableID: "issue-reports-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"ID",
				"IssueType",
				"Details",
				"RelevantTable",
				"RelevantRecordId",
				"CreatedByUser",
				"BelongsToAccount",
				"CreatedAt",
				"LastUpdatedAt",
				"ArchivedAt",
			},
			FieldReplacements: map[string]string{
				"IssueType":        "Issue Type",
				"RelevantTable":    "Table",
				"RelevantRecordId": "Record ID",
				"CreatedByUser":    "Reported By",
				"BelongsToAccount": "Account",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"Details": func(data any) g.Node {
					details, ok := data.(string)
					if !ok {
						return g.Text("Invalid data")
					}
					// Truncate long details text
					if len(details) > 100 {
						details = details[:97] + "..."
					}
					return g.Text(details)
				},
				"CreatedAt":     renderTimestamp,
				"LastUpdatedAt": renderTimestamp,
				"ArchivedAt":    renderTimestamp,
			},
			Pagination:             pagination,
			PaginationURLGenerator: paginationURLGenerator,
			DeepLinkURLGenerator:   deepLinkURLGenerator,
			PaginationHTMXTarget:   "#search-results",
		},
		RowLinkGenerator: func(data *issuereportssvc.IssueReport) string {
			return fmt.Sprintf("/issue_reports/%s", data.Id)
		},
		EmptyStateTitle:       "No issue reports found",
		EmptyStateDescription: "No users have submitted issue reports yet.",
		EmptyStateActions:     []g.Node{},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "View and manage user-submitted issue reports"
			}
			return fmt.Sprintf("Viewing %d issue reports", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("Issue Reports", s.renderIssueReportsError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("Issue Reports", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) IssueReportsSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	issueReportsRes, err := c.GetIssueReports(ctx, &issuereportssvc.GetIssueReportsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading issue reports: %v", err)),
			),
		), nil
	}

	if len(issueReportsRes.Results) == 0 {
		searchQuery := req.URL.Query().Get("search")
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No issue reports found",
				fmt.Sprintf("No issue reports match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{},
			),
		), nil
	}

	pagination := buildPaginationFromGRPCResponse(issueReportsRes.Pagination)
	paginationURLGenerator := buildPaginationURLGeneratorForSearch(req, "/api/issue_reports/search", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/issue_reports", queryFilter)

	table, err := components.Table(issueReportsRes.Results, &components.TableOptions[*issuereportssvc.IssueReport]{
		TableID: "issue-reports-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"ID",
			"IssueType",
			"Details",
			"RelevantTable",
			"RelevantRecordId",
			"CreatedByUser",
			"BelongsToAccount",
			"CreatedAt",
			"LastUpdatedAt",
			"ArchivedAt",
		},
		FieldReplacements: map[string]string{
			"IssueType":        "Issue Type",
			"RelevantTable":    "Table",
			"RelevantRecordId": "Record ID",
			"CreatedByUser":    "Reported By",
			"BelongsToAccount": "Account",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"Details": func(data any) g.Node {
				details, ok := data.(string)
				if !ok {
					return g.Text("Invalid data")
				}
				// Truncate long details text
				if len(details) > 100 {
					details = details[:97] + "..."
				}
				return g.Text(details)
			},
			"CreatedAt":     renderTimestamp,
			"LastUpdatedAt": renderTimestamp,
			"ArchivedAt":    renderTimestamp,
		},
		Pagination:             pagination,
		PaginationURLGenerator: paginationURLGenerator,
		DeepLinkURLGenerator:   deepLinkURLGenerator,
		PaginationHTMXTarget:   "#search-results",
		RowLinkGenerator: func(data *issuereportssvc.IssueReport) string {
			return fmt.Sprintf("/issue_reports/%s", data.Id)
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

func (s *AdminFrontendServer) IssueReportPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Issue Report", s.renderIssueReportsError("Error: No API client available")), nil
	}

	issueReportID := s.issueReportIDRouteParamFetcher(req)
	if issueReportID == "" {
		return page("Issue Report", s.renderIssueReportsError("Error: No issue report ID provided")), nil
	}

	issueReportRes, err := c.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: issueReportID})
	if err != nil {
		return page("Issue Report", s.renderIssueReportsError(fmt.Sprintf("Error loading issue report: %v", err))), nil
	}

	if issueReportRes == nil || issueReportRes.Result == nil {
		return page("Issue Report", s.renderIssueReportsError("Error: Issue report not found")), nil
	}

	issueReport := issueReportRes.Result

	// Use FormPage for viewing issue report details (read-only)
	formPageResult, err := components.FormPage(&components.FormPageProps[*issuereportssvc.IssueReport]{
		Title:        "Issue Report Details",
		BaseSubtitle: "View issue report information",
		Palette:      &design.StandardPalette,
		Data:         issueReport,
		FormOptions: &components.FormOptions[*issuereportssvc.IssueReport]{
			FormID: "view-issue-report-form",

			// All fields are read-only (no EnabledFields)
			EnabledFields: []string{},

			FieldConfigs: map[string]*components.FieldConfig{
				"Details": {
					InputType: "textarea",
				},
			},

			FormRows: []*components.FormRow{
				{
					Fields:  []string{"IssueType"},
					Columns: 1,
				},
				{
					Fields:  []string{"Details"},
					Columns: 1,
				},
				{
					Fields:  []string{"RelevantTable", "RelevantRecordId"},
					Columns: 2,
				},
				{
					Fields:  []string{"CreatedByUser", "BelongsToAccount"},
					Columns: 2,
				},
			},

			SubmitButtonText: "", // No submit button for read-only
			ShowCancelButton: true,
			CancelButtonText: "Back to Issue Reports",
			CancelURL:        "/issue_reports",
		},

		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Issue Reports", URL: "/issue_reports"},
			{Text: issueReport.Id, URL: ""},
		},

		SubtitleGenerator: func(ir *issuereportssvc.IssueReport) string {
			return fmt.Sprintf("Issue Report: %s", ir.IssueType)
		},
	})
	if err != nil {
		return page("Issue Report", s.renderIssueReportsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	return page("Issue Report", formPageResult.Node), nil
}
