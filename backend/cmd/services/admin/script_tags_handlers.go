package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

func (s *AdminFrontendServer) renderScriptTagsError(message string) g.Node {
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

func (s *AdminFrontendServer) ScriptTagsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Script Tags", s.renderScriptTagsError("Error: No API client available")), nil
	}

	tagsRes, err := c.ListScriptTags(ctx, &playscriptssvc.ListScriptTagsRequest{})
	if err != nil {
		return page("Script Tags", s.renderScriptTagsError(fmt.Sprintf("Error loading script tags: %v", err))), nil
	}

	tablePageResult, err := components.TablePage(&components.TablePageProps[*playscriptssvc.ScriptTag]{
		Title:        "Script Tags",
		BaseSubtitle: "Manage tags for categorizing play scripts",
		Palette:      &design.StandardPalette,
		ShowSearch:   true,
		SearchPlaceholder: "Search tags...",
		HTMXSearchTarget:  "/api/script_tags/search",
		Data:         tagsRes.Results,
		Actions: []g.Node{
			components.ActionButton("Create Tag", "/script_tags/new", &design.StandardPalette, true),
		},
		TableOptions: &components.TableOptions[*playscriptssvc.ScriptTag]{
			TableID: "script_tags-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"id",
				"name",
				"description",
				"created_at",
			},
			FieldReplacements: map[string]string{
				"created_at": "Created",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"created_at": renderTimestamp,
			},
		},
		RowLinkGenerator: nil, // No detail page; tags are view-only from list
		EmptyStateTitle:       "No script tags found",
		EmptyStateDescription: "Create tags to categorize play scripts (e.g., comedy, drama, one-act).",
		EmptyStateActions: []g.Node{
			components.ActionButton("Create Tag", "/script_tags/new", &design.StandardPalette, true),
		},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "Manage tags for categorizing play scripts"
			}
			return fmt.Sprintf("Manage %d script tags", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("Script Tags", s.renderScriptTagsError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("Script Tags", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) ScriptTagsSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	tagsRes, err := c.ListScriptTags(ctx, &playscriptssvc.ListScriptTagsRequest{})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading tags: %v", err)),
			),
		), nil
	}

	searchQuery := req.URL.Query().Get("search")
	var filtered []*playscriptssvc.ScriptTag
	if searchQuery == "" {
		filtered = tagsRes.Results
	} else {
		q := strings.ToLower(searchQuery)
		for _, t := range tagsRes.Results {
			if strings.Contains(strings.ToLower(t.Name), q) ||
				strings.Contains(strings.ToLower(t.Description), q) {
				filtered = append(filtered, t)
			}
		}
	}

	if len(filtered) == 0 {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No tags found",
				fmt.Sprintf("No tags match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{
					components.ActionButton("Create Tag", "/script_tags/new", &design.StandardPalette, true),
				},
			),
		), nil
	}

	table, err := components.Table(filtered, &components.TableOptions[*playscriptssvc.ScriptTag]{
		TableID: "script_tags-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"id",
			"name",
			"description",
			"created_at",
		},
		FieldReplacements: map[string]string{
			"created_at": "Created",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"created_at": renderTimestamp,
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

func (s *AdminFrontendServer) ScriptTagNewPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	if _, err := fetchClientFromContext(ctx); err != nil {
		return page("New Script Tag", s.renderScriptTagsError("Error: No API client available")), nil
	}

	emptyInput := &playscriptssvc.CreateScriptTagRequest{
		Name:        "",
		Description: "",
	}

	formPageResult, err := components.FormPage(&components.FormPageProps[*playscriptssvc.CreateScriptTagRequest]{
		Title:        "Create Script Tag",
		BaseSubtitle: "Add a new tag for categorizing play scripts",
		Palette:      &design.StandardPalette,
		Data:         emptyInput,
		FormOptions: &components.FormOptions[*playscriptssvc.CreateScriptTagRequest]{
			FormID:   "create-script-tag-form",
			Action:   "/api/script_tags",
			Method:   "POST",
			EnabledFields: []string{"name", "description"},
			FieldConfigs: map[string]*components.FieldConfig{
				"name": {
					DisplayName: "Name",
					Placeholder: "e.g., Comedy, Drama, One-Act",
					Validation: &components.FieldValidation{
						Required: true,
					},
				},
				"description": {
					DisplayName: "Description",
					Placeholder: "Optional description for this tag...",
				},
			},
			FormRows: []*components.FormRow{
				{Fields: []string{"name"}, Columns: 1},
				{Fields: []string{"description"}, Columns: 1},
			},
			SubmitButtonText: "Create Tag",
			ShowCancelButton: true,
			CancelButtonText: "Cancel",
			CancelURL:        "/script_tags",
			HTMXTarget:       "body",
			HTMXSwap:         "innerHTML",
			HTMXPushURL:      true,
		},
		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Script Tags", URL: "/script_tags"},
			{Text: "New Tag", URL: ""},
		},
	})
	if err != nil {
		return page("New Script Tag", s.renderScriptTagsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	return page("New Script Tag", formPageResult.Node), nil
}

func (s *AdminFrontendServer) ScriptTagCreate(res http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Script Tags", s.renderScriptTagsError("Error: No API client available")), nil
	}

	if err := req.ParseForm(); err != nil {
		return page("Script Tags", s.renderScriptTagsError(fmt.Sprintf("Error parsing form: %v", err))), nil
	}

	name := req.FormValue("name")
	description := req.FormValue("description")
	if name == "" {
		return page("New Script Tag", s.renderScriptTagsError("Name is required")), nil
	}

	_, err = c.CreateScriptTag(ctx, &playscriptssvc.CreateScriptTagRequest{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
	})
	if err != nil {
		return page("New Script Tag", s.renderScriptTagsError(fmt.Sprintf("Error creating tag: %v", err))), nil
	}

	http.Redirect(res, req, "/script_tags", http.StatusSeeOther)
	return g.El("div"), nil
}
