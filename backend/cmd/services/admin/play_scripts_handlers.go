package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	playscriptssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/play_scripts"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

const (
	playScriptIDURLParamKey = "playScriptID"
	lineIDURLParamKey       = "lineID"
	tagIDURLParamKey        = "tagID"
)

func (s *AdminFrontendServer) renderPlayScriptsError(message string) g.Node {
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

func (s *AdminFrontendServer) PlayScriptsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Play Scripts", s.renderPlayScriptsError("Error: No API client available")), nil
	}

	queryFilter, grpcFilter := buildQueryFilterFromRequest(req)

	playsRes, err := c.GetPlayScripts(ctx, &playscriptssvc.GetPlayScriptsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return page("Play Scripts", s.renderPlayScriptsError(fmt.Sprintf("Error loading play scripts: %v", err))), nil
	}

	pagination := buildPaginationFromGRPCResponse(playsRes.Pagination)
	paginationURLGenerator := buildPaginationURLGenerator(req, "/play_scripts", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/play_scripts", queryFilter)

	tablePageResult, err := components.TablePage(&components.TablePageProps[*playscriptssvc.PlayScript]{
		Title:             "Play Scripts",
		BaseSubtitle:      "Manage play scripts",
		Palette:           &design.StandardPalette,
		ShowSearch:        true,
		SearchPlaceholder: "Search play scripts...",
		HTMXSearchTarget:  "/api/play_scripts/search",
		Data:              playsRes.Results,
		Actions:           []g.Node{},
		TableOptions: &components.TableOptions[*playscriptssvc.PlayScript]{
			TableID: "play_scripts-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"id",
				"title",
				"author",
				"is_public",
				"created_at",
			},
			FieldReplacements: map[string]string{
				"id":         "ID",
				"is_public":  "Public",
				"created_at": "Created",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"created_at": renderTimestamp,
				"is_public": func(value any) g.Node {
					if b, ok := value.(bool); ok && b {
						return g.Text("Yes")
					}
					return g.Text("No")
				},
			},
			Pagination:             pagination,
			PaginationURLGenerator: paginationURLGenerator,
			DeepLinkURLGenerator:   deepLinkURLGenerator,
			PaginationHTMXTarget:   "#search-results",
		},
		RowLinkGenerator: func(data *playscriptssvc.PlayScript) string {
			return fmt.Sprintf("/play_scripts/%s", data.Id)
		},
		EmptyStateTitle:       "No play scripts found",
		EmptyStateDescription: "Play scripts will appear here when created.",
		EmptyStateActions:     []g.Node{},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "Manage play scripts"
			}
			return fmt.Sprintf("Manage %d play scripts", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("Play Scripts", s.renderPlayScriptsError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("Play Scripts", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) PlayScriptsSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	playsRes, err := c.GetPlayScripts(ctx, &playscriptssvc.GetPlayScriptsRequest{
		Filter: grpcFilter,
	})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading play scripts: %v", err)),
			),
		), nil
	}

	if len(playsRes.Results) == 0 {
		searchQuery := req.URL.Query().Get("search")
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No play scripts found",
				fmt.Sprintf("No play scripts match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{},
			),
		), nil
	}

	pagination := buildPaginationFromGRPCResponse(playsRes.Pagination)
	paginationURLGenerator := buildPaginationURLGeneratorForSearch(req, "/api/play_scripts/search", queryFilter)
	deepLinkURLGenerator := buildPaginationURLGenerator(req, "/play_scripts", queryFilter)

	table, err := components.Table(playsRes.Results, &components.TableOptions[*playscriptssvc.PlayScript]{
		TableID: "play_scripts-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"id",
			"title",
			"author",
			"is_public",
			"created_at",
		},
		FieldReplacements: map[string]string{
			"id":         "ID",
			"is_public":  "Public",
			"created_at": "Created",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"created_at": renderTimestamp,
			"is_public": func(value any) g.Node {
				if b, ok := value.(bool); ok && b {
					return g.Text("Yes")
				}
				return g.Text("No")
			},
		},
		Pagination:             pagination,
		PaginationURLGenerator: paginationURLGenerator,
		DeepLinkURLGenerator:   deepLinkURLGenerator,
		PaginationHTMXTarget:   "#search-results",
		RowLinkGenerator: func(data *playscriptssvc.PlayScript) string {
			return fmt.Sprintf("/play_scripts/%s", data.Id)
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

func (s *AdminFrontendServer) PlayScriptPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Play Script", s.renderPlayScriptsError("Error: No API client available")), nil
	}

	playScriptID := s.playScriptIDRouteParamFetcher(req)
	if playScriptID == "" {
		return page("Play Script", s.renderPlayScriptsError("Error: No play script ID provided")), nil
	}

	playRes, err := c.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{
		PlayScriptId:         playScriptID,
		IncludeLinkedScripts: false,
	})
	if err != nil {
		return page("Play Script", s.renderPlayScriptsError(fmt.Sprintf("Error loading play script: %v", err))), nil
	}

	if playRes == nil || playRes.Result == nil {
		return page("Play Script", s.renderPlayScriptsError("Error: Play script not found")), nil
	}

	play := playRes.Result

	// Play metadata form
	formPageResult, err := components.FormPage(&components.FormPageProps[*playscriptssvc.PlayScript]{
		Title:        "Play Script Details",
		BaseSubtitle: "View and edit play script information",
		Palette:      &design.StandardPalette,
		Data:         play,
		FormOptions: &components.FormOptions[*playscriptssvc.PlayScript]{
			FormID: "edit-play-script-form",
			Action: fmt.Sprintf("/api/play_scripts/%s", play.Id),
			Method: "POST",
			EnabledFields: []string{
				"title",
				"description",
				"author",
				"is_public",
				"preceded_by_id",
				"precedes_id",
			},
			FieldConfigs: map[string]*components.FieldConfig{
				"title": {
					Placeholder: "Enter title...",
					Validation: &components.FieldValidation{
						Required: true,
					},
				},
				"description": {
					Placeholder: "Enter description...",
				},
				"author": {
					Placeholder: "Enter author...",
				},
				"is_public": {
					InputType: "checkbox",
				},
				"preceded_by_id": {
					Placeholder: "Preceded by (play script ID)...",
				},
				"precedes_id": {
					Placeholder: "Precedes (play script ID)...",
				},
			},
			FormRows: []*components.FormRow{
				{Fields: []string{"title"}, Columns: 1},
				{Fields: []string{"description"}, Columns: 1},
				{Fields: []string{"author", "is_public"}, Columns: 2},
				{Fields: []string{"preceded_by_id", "precedes_id"}, Columns: 2},
			},
			SubmitButtonText: "Save Changes",
			ShowCancelButton: true,
			CancelButtonText: "Back to Play Scripts",
			CancelURL:        "/play_scripts",
			HTMXTarget:       "body",
			HTMXSwap:         "innerHTML",
			HTMXPushURL:      true,
		},
		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Play Scripts", URL: "/play_scripts"},
			{Text: play.Title, URL: ""},
		},
		SubtitleGenerator: func(p *playscriptssvc.PlayScript) string {
			return fmt.Sprintf("Editing: %s", p.Title)
		},
	})
	if err != nil {
		return page("Play Script", s.renderPlayScriptsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	// Lines section
	linesSection := s.renderPlayScriptLinesSection(play)
	// Tags section - loaded via HTMX
	tagsSection := components.CardWithHeader(
		"Tags",
		&design.StandardPalette,
		nil,
		g.El("div",
			g.Attr("id", "play-script-tags-container"),
			g.Attr("hx-get", fmt.Sprintf("/api/play_scripts/%s/tags", playScriptID)),
			g.Attr("hx-trigger", "load"),
			g.Attr("hx-swap", "innerHTML"),
			components.LoadingSpinner(&design.StandardPalette),
		),
	)

	return page("Play Script",
		ghtml.Div(
			ghtml.Class("space-y-6"),
			formPageResult.Node,
			linesSection,
			tagsSection,
		),
	), nil
}

func (s *AdminFrontendServer) renderPlayScriptLinesSection(play *playscriptssvc.PlayScript) g.Node {
	castMemberNames := make(map[string]string)
	for _, cm := range play.CastMembers {
		castMemberNames[cm.Id] = cm.Name
	}

	var rows []g.Node
	for _, line := range play.Lines {
		castName := castMemberNames[line.CastMemberId]
		if castName == "" {
			castName = line.CastMemberId
		}
		rows = append(rows,
			ghtml.Tr(
				ghtml.Class("border-b border-gray-200 hover:bg-gray-50"),
				ghtml.Td(ghtml.Class("px-4 py-2 text-sm"), g.Text(castName)),
				ghtml.Td(ghtml.Class("px-4 py-2 text-sm"), g.Text(line.LineText)),
				ghtml.Td(ghtml.Class("px-4 py-2 text-sm"), g.Text(fmt.Sprintf("%d", line.LineIndex))),
				ghtml.Td(ghtml.Class("px-4 py-2"),
					ghtml.A(
						ghtml.Href(fmt.Sprintf("/play_scripts/%s/lines/%s/edit", play.Id, line.Id)),
						ghtml.Class(fmt.Sprintf("text-sm %s hover:%s", design.TextColor(design.StandardPalette.Primary), design.TextColor(design.Color{Value: design.StandardPalette.Primary.Value + "-700"}))),
						g.Text("Edit"),
					),
				),
			),
		)
	}

	tableHeader := ghtml.THead(
		ghtml.Tr(
			ghtml.Th(ghtml.Class("px-4 py-2 text-left text-sm font-medium text-gray-700"), g.Text("Character")),
			ghtml.Th(ghtml.Class("px-4 py-2 text-left text-sm font-medium text-gray-700"), g.Text("Line")),
			ghtml.Th(ghtml.Class("px-4 py-2 text-left text-sm font-medium text-gray-700"), g.Text("Index")),
			ghtml.Th(ghtml.Class("px-4 py-2 text-left text-sm font-medium text-gray-700"), g.Text("Actions")),
		),
	)

	tableBody := ghtml.TBody(g.Group(rows))
	if len(rows) == 0 {
		tableBody = ghtml.TBody(
			ghtml.Tr(
				ghtml.Td(
					ghtml.ColSpan("4"),
					ghtml.Class("px-4 py-8 text-center text-sm text-gray-500"),
					g.Text("No lines in this play script."),
				),
			),
		)
	}

	return components.CardWithHeader(
		"Script Lines",
		&design.StandardPalette,
		nil,
		ghtml.Div(
			ghtml.Class("overflow-x-auto"),
			ghtml.Table(
				ghtml.Class("min-w-full divide-y divide-gray-200"),
				tableHeader,
				tableBody,
			),
		),
	)
}

func (s *AdminFrontendServer) PlayScriptTagsSection(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: No API client available"))), nil
	}

	playScriptID := s.playScriptIDRouteParamFetcher(req)
	if playScriptID == "" {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: No play script ID"))), nil
	}

	tagsRes, err := c.ListScriptTags(ctx, &playscriptssvc.ListScriptTagsRequest{})
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text(fmt.Sprintf("Error loading tags: %v", err)))), nil
	}

	playRes, err := c.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{PlayScriptId: playScriptID, IncludeLinkedScripts: false})
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text(fmt.Sprintf("Error loading play: %v", err)))), nil
	}

	assignedTagIDs := make(map[string]bool)
	for _, t := range playRes.Result.Tags {
		assignedTagIDs[t.Id] = true
	}

	var assignedTags []g.Node
	for _, t := range playRes.Result.Tags {
		tag := t
		assignedTags = append(assignedTags,
			ghtml.Span(
				ghtml.Class(fmt.Sprintf("inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium mr-2 mb-2 %s %s",
					design.Background(design.StandardPalette.Primary), design.TextColor(design.Color{Value: "white"}))),
				g.Text(tag.Name),
				ghtml.Button(
					ghtml.Type("button"),
					ghtml.Class("ml-1 hover:opacity-80 cursor-pointer bg-transparent border-none text-inherit text-lg leading-none"),
					g.Attr("hx-delete", fmt.Sprintf("/api/play_scripts/%s/tags/%s", playScriptID, tag.Id)),
					g.Attr("hx-target", "#play-script-tags-container"),
					g.Attr("hx-swap", "innerHTML"),
					g.Attr("hx-confirm", "Remove this tag?"),
					g.Text("×"),
				),
			),
		)
	}

	var selectOptions []g.Node
	selectOptions = append(selectOptions, ghtml.Option(g.Attr("value", ""), g.Text("-- Add tag --")))
	for _, t := range tagsRes.Results {
		if !assignedTagIDs[t.Id] {
			selectOptions = append(selectOptions,
				ghtml.Option(g.Attr("value", t.Id), g.Text(t.Name)),
			)
		}
	}

	addTagForm := ghtml.Form(
		g.Attr("hx-post", fmt.Sprintf("/api/play_scripts/%s/tags", playScriptID)),
		g.Attr("hx-target", "#play-script-tags-container"),
		g.Attr("hx-swap", "innerHTML"),
		ghtml.Class("flex gap-2 mt-2"),
		ghtml.Select(
			ghtml.Class("rounded-md border-gray-300 text-sm"),
			g.Attr("name", "script_tag_id"),
			g.Attr("required"),
			g.Group(selectOptions),
		),
		ghtml.Button(
			ghtml.Class(fmt.Sprintf("px-3 py-1 text-sm rounded-md %s %s", design.Background(design.StandardPalette.Primary), design.TextColor(design.Color{Value: "white"}))),
			ghtml.Type("submit"),
			g.Text("Add Tag"),
		),
	)

	return ghtml.Div(
		ghtml.Class("space-y-2"),
		ghtml.Div(g.Group(assignedTags)),
		addTagForm,
	), nil
}

func (s *AdminFrontendServer) PlayScriptUpdate(res http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Play Scripts", s.renderPlayScriptsError("Error: No API client available")), nil
	}

	playScriptID := s.playScriptIDRouteParamFetcher(req)
	if playScriptID == "" {
		return page("Play Scripts", s.renderPlayScriptsError("Error: No play script ID provided")), nil
	}

	if err := req.ParseForm(); err != nil {
		return page("Play Scripts", s.renderPlayScriptsError(fmt.Sprintf("Error parsing form: %v", err))), nil
	}

	title := req.FormValue("title")
	description := req.FormValue("description")
	author := req.FormValue("author")
	isPublic := req.FormValue("is_public") == "true" || req.FormValue("is_public") == "1"
	precededByID := req.FormValue("preceded_by_id")
	precedesID := req.FormValue("precedes_id")

	input := &playscriptssvc.PlayScriptUpdateRequestInput{
		Title:       &title,
		Description: &description,
		Author:      &author,
		IsPublic:    &isPublic,
	}
	if precededByID != "" {
		input.PrecededById = &precededByID
	}
	if precedesID != "" {
		input.PrecedesId = &precedesID
	}

	_, err = c.UpdatePlayScript(ctx, &playscriptssvc.UpdatePlayScriptRequest{
		PlayScriptId: playScriptID,
		Input:        input,
	})
	if err != nil {
		return page("Play Scripts", s.renderPlayScriptsError(fmt.Sprintf("Error updating play script: %v", err))), nil
	}

	http.Redirect(res, req, fmt.Sprintf("/play_scripts/%s", playScriptID), http.StatusSeeOther)
	return g.El("div"), nil
}

func (s *AdminFrontendServer) PlayScriptLineEditPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Edit Line", s.renderPlayScriptsError("Error: No API client available")), nil
	}

	playScriptID := s.playScriptIDRouteParamFetcher(req)
	lineID := s.lineIDRouteParamFetcher(req)
	if playScriptID == "" || lineID == "" {
		return page("Edit Line", s.renderPlayScriptsError("Error: Missing play script or line ID")), nil
	}

	playRes, err := c.GetPlayScript(ctx, &playscriptssvc.GetPlayScriptRequest{
		PlayScriptId:         playScriptID,
		IncludeLinkedScripts: false,
	})
	if err != nil {
		return page("Edit Line", s.renderPlayScriptsError(fmt.Sprintf("Error loading play script: %v", err))), nil
	}

	var targetLine *playscriptssvc.PlayScriptLine
	for _, l := range playRes.Result.Lines {
		if l.Id == lineID {
			targetLine = l
			break
		}
	}
	if targetLine == nil {
		return page("Edit Line", s.renderPlayScriptsError("Line not found")), nil
	}

	castMemberNames := make(map[string]string)
	for _, cm := range playRes.Result.CastMembers {
		castMemberNames[cm.Id] = cm.Name
	}

	var castOptions []g.Node
	for _, cm := range playRes.Result.CastMembers {
		castOptions = append(castOptions,
			ghtml.Option(
				g.If(cm.Id == targetLine.CastMemberId, ghtml.Selected()),
				g.Attr("value", cm.Id),
				g.Text(cm.Name),
			),
		)
	}

	scrollModeOptions := []struct {
		Value string
		Label string
	}{
		{"auto", "Auto"},
		{"manual", "Manual"},
	}

	var scrollOptions []g.Node
	for _, opt := range scrollModeOptions {
		selected := (opt.Value == "auto" && targetLine.ScrollMode == playscriptssvc.ScrollMode_SCROLL_MODE_AUTO) ||
			(opt.Value == "manual" && targetLine.ScrollMode == playscriptssvc.ScrollMode_SCROLL_MODE_MANUAL)
		scrollOptions = append(scrollOptions,
			ghtml.Option(
				g.If(selected, ghtml.Selected()),
				g.Attr("value", opt.Value),
				g.Text(opt.Label),
			),
		)
	}

	form := ghtml.Form(
		g.Attr("action", fmt.Sprintf("/api/play_scripts/%s/lines/%s", playScriptID, lineID)),
		g.Attr("method", "POST"),
		g.Attr("class", "space-y-4"),
		ghtml.Div(
			ghtml.Class("flex flex-col"),
			ghtml.Label(ghtml.Class("block text-sm font-medium mb-2"), g.Text("Character")),
			ghtml.Select(
				ghtml.Class("rounded-md border-gray-300"),
				g.Attr("name", "cast_member_id"),
				g.Attr("required"),
				g.Group(castOptions),
			),
		),
		ghtml.Div(
			ghtml.Class("flex flex-col"),
			ghtml.Label(ghtml.Class("block text-sm font-medium mb-2"), g.Text("Line Text")),
			ghtml.Textarea(
				ghtml.Class("rounded-md border-gray-300"),
				g.Attr("name", "line_text"),
				g.Attr("rows", "3"),
				g.Attr("required"),
				g.Text(targetLine.LineText),
			),
		),
		ghtml.Div(
			ghtml.Class("flex flex-col"),
			ghtml.Label(ghtml.Class("block text-sm font-medium mb-2"), g.Text("Line Index")),
			ghtml.Input(
				ghtml.Type("number"),
				ghtml.Class("rounded-md border-gray-300"),
				g.Attr("name", "line_index"),
				g.Attr("value", fmt.Sprintf("%d", targetLine.LineIndex)),
			),
		),
		ghtml.Div(
			ghtml.Class("flex flex-col"),
			ghtml.Label(ghtml.Class("block text-sm font-medium mb-2"), g.Text("Direction")),
			ghtml.Input(
				ghtml.Type("text"),
				ghtml.Class("rounded-md border-gray-300"),
				g.Attr("name", "direction"),
				g.Attr("value", targetLine.Direction),
			),
		),
		ghtml.Div(
			ghtml.Class("flex flex-col"),
			ghtml.Label(ghtml.Class("block text-sm font-medium mb-2"), g.Text("Scroll Mode")),
			ghtml.Select(
				ghtml.Class("rounded-md border-gray-300"),
				g.Attr("name", "scroll_mode"),
				g.Group(scrollOptions),
			),
		),
		ghtml.Div(
			ghtml.Class("flex flex-col"),
			ghtml.Label(ghtml.Class("block text-sm font-medium mb-2"), g.Text("Auto Scroll WPM")),
			ghtml.Input(
				ghtml.Type("number"),
				ghtml.Class("rounded-md border-gray-300"),
				g.Attr("name", "auto_scroll_wpm"),
				g.Attr("value", fmt.Sprintf("%d", targetLine.AutoScrollWpm)),
			),
		),
		ghtml.Div(
			ghtml.Class("flex gap-2 mt-4"),
			ghtml.Button(
				ghtml.Class(fmt.Sprintf("px-4 py-2 rounded-md %s %s", design.Background(design.StandardPalette.Primary), design.TextColor(design.Color{Value: "white"}))),
				ghtml.Type("submit"),
				g.Text("Save"),
			),
			ghtml.A(
				ghtml.Href(fmt.Sprintf("/play_scripts/%s", playScriptID)),
				ghtml.Class("px-4 py-2 rounded-md border border-gray-300"),
				g.Text("Cancel"),
			),
		),
	)

	return page("Edit Line",
		components.ContentContainer(&components.ContentContainerProps{
			Title:    "Edit Line",
			Subtitle: fmt.Sprintf("Editing line %d in %s", targetLine.LineIndex, playRes.Result.Title),
			Palette:  &design.StandardPalette,
		},
			components.Card(&design.StandardPalette, form),
		),
	), nil
}

func (s *AdminFrontendServer) PlayScriptLineUpdate(res http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: No API client available"))), nil
	}

	playScriptID := s.playScriptIDRouteParamFetcher(req)
	lineID := s.lineIDRouteParamFetcher(req)
	if playScriptID == "" || lineID == "" {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: Missing play script or line ID"))), nil
	}

	if err := req.ParseForm(); err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text(fmt.Sprintf("Error: %v", err)))), nil
	}

	lineText := req.FormValue("line_text")
	castMemberID := req.FormValue("cast_member_id")
	lineIndex, _ := strconv.ParseInt(req.FormValue("line_index"), 10, 32)
	direction := req.FormValue("direction")
	scrollModeStr := req.FormValue("scroll_mode")
	autoScrollWpm, _ := strconv.ParseInt(req.FormValue("auto_scroll_wpm"), 10, 32)

	scrollMode := playscriptssvc.ScrollMode_SCROLL_MODE_AUTO
	if scrollModeStr == "manual" || scrollModeStr == "SCROLL_MODE_MANUAL" {
		scrollMode = playscriptssvc.ScrollMode_SCROLL_MODE_MANUAL
	}

	line := &playscriptssvc.PlayScriptLine{
		Id:            lineID,
		PlayScriptId:  playScriptID,
		CastMemberId:  castMemberID,
		LineText:      lineText,
		LineIndex:     int32(lineIndex),
		Direction:     direction,
		ScrollMode:    scrollMode,
		AutoScrollWpm: int32(autoScrollWpm),
	}

	_, err = c.UpdatePlayScriptLine(ctx, &playscriptssvc.UpdatePlayScriptLineRequest{Line: line})
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text(fmt.Sprintf("Error: %v", err)))), nil
	}

	http.Redirect(res, req, fmt.Sprintf("/play_scripts/%s", playScriptID), http.StatusSeeOther)
	return g.El("div"), nil
}

func (s *AdminFrontendServer) PlayScriptAddTag(res http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: No API client available"))), nil
	}

	playScriptID := s.playScriptIDRouteParamFetcher(req)
	if playScriptID == "" {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: No play script ID"))), nil
	}

	if err := req.ParseForm(); err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error parsing form"))), nil
	}

	scriptTagID := req.FormValue("script_tag_id")
	if scriptTagID == "" {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: Tag ID required"))), nil
	}

	_, err = c.AddTagToPlayScript(ctx, &playscriptssvc.AddTagToPlayScriptRequest{
		PlayScriptId: playScriptID,
		ScriptTagId:  scriptTagID,
	})
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text(fmt.Sprintf("Error: %v", err)))), nil
	}

	return s.PlayScriptTagsSection(res, req)
}

func (s *AdminFrontendServer) PlayScriptRemoveTag(res http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: No API client available"))), nil
	}

	playScriptID := s.playScriptIDRouteParamFetcher(req)
	tagID := s.tagIDRouteParamFetcher(req)
	if playScriptID == "" || tagID == "" {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text("Error: Missing IDs"))), nil
	}

	_, err = c.RemoveTagFromPlayScript(ctx, &playscriptssvc.RemoveTagFromPlayScriptRequest{
		PlayScriptId: playScriptID,
		ScriptTagId:  tagID,
	})
	if err != nil {
		return g.El("div", ghtml.P(ghtml.Class("text-sm text-red-600"), g.Text(fmt.Sprintf("Error: %v", err)))), nil
	}

	return s.PlayScriptTagsSection(res, req)
}
