package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/components"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"
	settingssvc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/settings"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

const (
	settingIDURLParamKey = "settingID"
)

func (s *AdminFrontendServer) SettingCreate(res http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("New Setting", s.renderSettingsError("Error: No API client available")), nil
	}

	// Decode JSON request body
	var input *settingssvc.ServiceSettingCreationRequestInput
	if err = s.encoder.DecodeRequest(ctx, req, &input); err != nil {
		return page("New Setting", s.renderSettingsError(fmt.Sprintf("Error decoding request: %v", err))), nil
	}

	// Call gRPC service to create the setting
	createRes, err := c.CreateServiceSetting(ctx, &settingssvc.CreateServiceSettingRequest{
		Input: input,
	})
	if err != nil {
		return page("New Setting", s.renderSettingsError(fmt.Sprintf("Error creating setting: %v", err))), nil
	}

	if createRes == nil || createRes.Created == nil {
		return page("New Setting", s.renderSettingsError("Error: No setting returned from server")), nil
	}

	// Redirect to the newly created setting's page
	settingID := createRes.Created.Id
	http.Redirect(res, req, fmt.Sprintf("/settings/%s", settingID), http.StatusSeeOther)

	return g.El("div"), nil
}

func (s *AdminFrontendServer) SettingNewPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	if _, err := fetchClientFromContext(ctx); err != nil {
		return page("New Setting", s.renderSettingsError("Error: No API client available")), nil
	}

	// Create an empty ServiceSettingCreationRequestInput for the form
	emptyInput := &settingssvc.ServiceSettingCreationRequestInput{}

	// Use the new FormPage component for creating a new setting
	formPageResult, err := components.FormPage(&components.FormPageProps[*settingssvc.ServiceSettingCreationRequestInput]{
		Title:        "Create New Service Setting",
		BaseSubtitle: "Add a new service setting",
		Palette:      &design.StandardPalette,
		Data:         emptyInput,
		FormOptions: &components.FormOptions[*settingssvc.ServiceSettingCreationRequestInput]{
			FormID: "create-setting-form",
			Action: "/api/settings",
			Method: "POST",

			// All fields should be enabled for creation
			EnabledFields: []string{
				"Name",
				"Type",
				"Description",
				"DefaultValue",
				"Enumeration",
				"AdminsOnly",
			},

			// Configure field validation
			FieldConfigs: map[string]*components.FieldConfig{
				"Name": {
					Placeholder: "Enter setting name...",
					Validation: &components.FieldValidation{
						Required: true,
					},
				},
				"Type": {
					Options: []*components.SelectOption{
						{Value: "string", Label: "String", IsDefault: true},
						{Value: "number", Label: "Number"},
						{Value: "boolean", Label: "Boolean"},
						{Value: "enum", Label: "Enumeration"},
					},
					Placeholder: "Select setting type...",
					Validation: &components.FieldValidation{
						Required:      true,
						CustomMessage: "Please select a setting type",
					},
				},
				"Description": {
					Placeholder: "Enter description...",
					Validation: &components.FieldValidation{
						Required: true,
					},
				},
				"DefaultValue": {
					Placeholder: "Enter default value (optional)...",
				},
				"Enumeration": {
					Placeholder: "Enter comma-separated enumeration values (e.g., option1,option2,option3)...",
					Validation: &components.FieldValidation{
						CustomMessage: "Required when Type is 'enum'. Enter comma-separated values.",
					},
				},
				"AdminsOnly": {
					Placeholder: "Check if this setting is admin-only",
				},
			},

			// Layout configuration
			FormRows: []*components.FormRow{
				{
					Fields:      []string{"Name", "AdminsOnly"},
					Columns:     2,
					ColumnSpans: []int{9, 3},
				},
				{
					Fields:  []string{"Type"},
					Columns: 1,
				},
				{
					Fields:  []string{"Description"},
					Columns: 1,
				},
				{
					Fields:  []string{"DefaultValue"},
					Columns: 1,
				},
				{
					Fields:  []string{"Enumeration"},
					Columns: 1,
				},
			},

			SubmitButtonText: "Create Setting",
			ShowCancelButton: true,
			CancelButtonText: "Cancel",
			CancelURL:        "/settings",

			// HTMX configuration
			HTMXTarget:  "body",
			HTMXSwap:    "innerHTML",
			HTMXPushURL: true,
		},

		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Settings", URL: "/settings"},
			{Text: "New Setting", URL: ""},
		},
	})
	if err != nil {
		return page("New Setting", s.renderSettingsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	return page("New Setting", formPageResult.Node), nil
}

func (s *AdminFrontendServer) SettingPage(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Settings", s.renderSettingsError("Error: No API client available")), nil
	}

	settingID := s.settingIDRouteParamFetcher(req)
	if settingID == "" {
		return page("Settings", s.renderSettingsError("Error: No setting ID provided")), nil
	}

	settingsRes, err := c.GetServiceSetting(ctx, &settingssvc.GetServiceSettingRequest{ServiceSettingId: settingID})
	if err != nil {
		return page("Settings", s.renderSettingsError(fmt.Sprintf("Error loading setting: %v", err))), nil
	}
	setting := settingsRes.Result

	// Use the new FormPage component for editing setting data
	formPageResult, err := components.FormPage(&components.FormPageProps[*settingssvc.ServiceSetting]{
		Title:        "Service Setting Details",
		BaseSubtitle: "View and edit service setting information",
		Palette:      &design.StandardPalette,
		Data:         setting,
		FormOptions: &components.FormOptions[*settingssvc.ServiceSetting]{
			FormID: "edit-setting-form",
			Action: fmt.Sprintf("/api/settings/%s", setting.Id),
			Method: "PUT",

			// Enable editable fields
			EnabledFields: []string{
				"Name",
				"Type",
				"Description",
				"DefaultValue",
				"Enumeration",
				"AdminsOnly",
			},

			// Configure field validation
			FieldConfigs: map[string]*components.FieldConfig{
				"Name": {
					Validation: &components.FieldValidation{
						Required:      true,
						MinLength:     2,
						MaxLength:     100,
						CustomMessage: "Setting name must be between 2 and 100 characters",
					},
				},
				"Type": {
					Options: []*components.SelectOption{
						{Value: "string", Label: "String"},
						{Value: "number", Label: "Number"},
						{Value: "boolean", Label: "Boolean"},
						{Value: "enum", Label: "Enumeration"},
					},
					Validation: &components.FieldValidation{
						Required:      true,
						CustomMessage: "Please select a setting type",
					},
				},
				"Description": {
					Placeholder: "Enter description...",
					Validation: &components.FieldValidation{
						Required:      true,
						MinLength:     10,
						MaxLength:     500,
						CustomMessage: "Description must be between 10 and 500 characters",
					},
				},
				"DefaultValue": {
					Placeholder: "Enter default value...",
					Validation: &components.FieldValidation{
						MaxLength:     200,
						CustomMessage: "Maximum 200 characters",
					},
				},
				"Enumeration": {
					Placeholder: "Comma-separated enumeration values (e.g., option1,option2,option3)...",
					Validation: &components.FieldValidation{
						CustomMessage: "Required when Type is 'enum'. Enter comma-separated values.",
					},
				},
			},

			// Layout configuration
			FormRows: []*components.FormRow{
				{
					Fields:      []string{"Name", "AdminsOnly"},
					Columns:     2,
					ColumnSpans: []int{9, 3}, // Name gets 9/12 (75%), AdminsOnly gets 3/12 (25%)
				},
				{
					Fields:  []string{"Type"},
					Columns: 1,
				},
				{
					Fields:  []string{"Description"},
					Columns: 1,
				},
				{
					Fields:  []string{"DefaultValue"},
					Columns: 1,
				},
				{
					Fields:  []string{"Enumeration"},
					Columns: 1,
				},
			},

			SubmitButtonText: "Save Changes",
			ShowCancelButton: true,
			CancelButtonText: "Back to Settings",
			CancelURL:        "/settings",

			// HTMX configuration
			HTMXTarget:  "body",
			HTMXSwap:    "innerHTML",
			HTMXPushURL: true,
		},

		ShowBreadcrumbs: true,
		Breadcrumbs: []*components.Breadcrumb{
			{Text: "Dashboard", URL: "/"},
			{Text: "Settings", URL: "/settings"},
			{Text: setting.Name, URL: ""},
		},

		// Dynamic subtitle showing setting info
		SubtitleGenerator: func(s *settingssvc.ServiceSetting) string {
			return fmt.Sprintf("Editing setting: %s", s.Name)
		},
	})
	if err != nil {
		return page("Settings", s.renderSettingsError(fmt.Sprintf("Error creating form: %v", err))), nil
	}

	return page("Settings", formPageResult.Node), nil
}

func (s *AdminFrontendServer) SettingsList(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	c, err := fetchClientFromContext(ctx)
	if err != nil {
		return page("Settings", s.renderSettingsError("Error: No API client available")), nil
	}

	settingsRes, err := c.GetServiceSettings(ctx, &settingssvc.GetServiceSettingsRequest{})
	if err != nil {
		return page("Settings", s.renderSettingsError(fmt.Sprintf("Error loading settings: %v", err))), nil
	}

	// Use the new integrated TablePage component
	tablePageResult, err := components.TablePage(&components.TablePageProps[*settingssvc.ServiceSetting]{
		Title:             "Service Settings",
		BaseSubtitle:      "Manage service settings",
		Palette:           &design.StandardPalette,
		ShowSearch:        true,
		SearchPlaceholder: "Search settings...",
		HTMXSearchTarget:  "/api/settings/search",
		Data:              settingsRes.Results,
		Actions: []g.Node{
			components.ActionButton("Create New Setting", "/settings/new", &design.StandardPalette, true),
		},
		TableOptions: &components.TableOptions[*settingssvc.ServiceSetting]{
			TableID: "settings-table",
			Palette: &design.StandardPalette,
			Fields: []string{
				"ID",
				"Name",
				"Type",
				"Description",
				"DefaultValue",
				"Enumeration",
				"AdminsOnly",
				"CreatedAt",
				"LastUpdatedAt",
				"ArchivedAt",
			},
			FieldReplacements: map[string]string{
				"AdminsOnly": "Admins Only",
			},
			FieldRenderers: map[string]components.FieldRenderer{
				"CreatedAt":     renderTimestamp,
				"LastUpdatedAt": renderTimestamp,
				"ArchivedAt":    renderTimestamp,
				"AdminsOnly": func(value any) g.Node {
					if value == nil {
						return g.Text("-")
					}
					if boolVal, ok := value.(bool); ok {
						if boolVal {
							return g.Text("Yes")
						}
						return g.Text("No")
					}
					return g.Text("-")
				},
				"Enumeration": func(value any) g.Node {
					if value == nil {
						return g.Text("-")
					}
					if enumSlice, ok := value.([]string); ok {
						if len(enumSlice) == 0 {
							return g.Text("-")
						}
						return g.Text(strings.Join(enumSlice, ", "))
					}
					return g.Text("-")
				},
			},
		},
		RowLinkGenerator: func(data *settingssvc.ServiceSetting) string {
			return fmt.Sprintf("/settings/%s", data.Id)
		},
		EmptyStateTitle:       "No settings found",
		EmptyStateDescription: "Get started by creating your first service setting.",
		EmptyStateActions: []g.Node{
			components.ActionButton("Create New Setting", "/settings/new", &design.StandardPalette, true),
		},
		SubtitleGenerator: func(metadata components.TablePageMetadata) string {
			if metadata.EmptyState {
				return "Manage service settings"
			}
			return fmt.Sprintf("Manage %d service settings", metadata.TotalCount)
		},
	})
	if err != nil {
		return page("Settings", s.renderSettingsError(fmt.Sprintf("Error creating table: %v", err))), nil
	}

	return page("Settings", tablePageResult.Node), nil
}

func (s *AdminFrontendServer) SettingsSearch(_ http.ResponseWriter, req *http.Request) (g.Node, error) {
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

	settingsRes, err := c.GetServiceSettings(ctx, &settingssvc.GetServiceSettingsRequest{})
	if err != nil {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			ghtml.P(
				ghtml.Class(fmt.Sprintf("text-center py-8 %s", design.TextColor(design.StandardPalette.Warning))),
				g.Text(fmt.Sprintf("Error loading settings: %v", err)),
			),
		), nil
	}

	// Filter settings based on search query
	var filteredSettings []*settingssvc.ServiceSetting
	if searchQuery == "" {
		// No search query, return all settings
		filteredSettings = settingsRes.Results
	} else {
		// Filter settings by search query (case insensitive)
		searchQueryLower := strings.ToLower(searchQuery)
		for _, setting := range settingsRes.Results {
			if strings.Contains(strings.ToLower(setting.Name), searchQueryLower) ||
				strings.Contains(strings.ToLower(setting.Description), searchQueryLower) ||
				strings.Contains(strings.ToLower(setting.Type), searchQueryLower) {
				filteredSettings = append(filteredSettings, setting)
			}
		}
	}

	// Generate just the table (not the full page)
	if len(filteredSettings) == 0 {
		return g.El("div",
			g.Attr("class", "overflow-x-auto"),
			components.EmptyState(
				"No settings found",
				fmt.Sprintf("No settings match the search term '%s'.", searchQuery),
				&design.StandardPalette,
				[]g.Node{
					components.ActionButton("Create New Setting", "/settings/new", &design.StandardPalette, true),
				},
			),
		), nil
	}

	table, err := components.Table(filteredSettings, &components.TableOptions[*settingssvc.ServiceSetting]{
		TableID: "settings-table",
		Palette: &design.StandardPalette,
		Fields: []string{
			"ID",
			"Name",
			"Type",
			"Description",
			"DefaultValue",
			"Enumeration",
			"AdminsOnly",
			"CreatedAt",
			"LastUpdatedAt",
			"ArchivedAt",
		},
		FieldReplacements: map[string]string{
			"AdminsOnly": "Admins Only",
		},
		FieldRenderers: map[string]components.FieldRenderer{
			"CreatedAt":     renderTimestamp,
			"LastUpdatedAt": renderTimestamp,
			"ArchivedAt":    renderTimestamp,
			"AdminsOnly": func(value any) g.Node {
				if value == nil {
					return g.Text("-")
				}
				if boolVal, ok := value.(bool); ok {
					if boolVal {
						return g.Text("Yes")
					}
					return g.Text("No")
				}
				return g.Text("-")
			},
			"Enumeration": func(value any) g.Node {
				if value == nil {
					return g.Text("-")
				}
				if enumSlice, ok := value.([]string); ok {
					if len(enumSlice) == 0 {
						return g.Text("-")
					}
					return g.Text(strings.Join(enumSlice, ", "))
				}
				return g.Text("-")
			},
		},
		RowLinkGenerator: func(data *settingssvc.ServiceSetting) string {
			return fmt.Sprintf("/settings/%s", data.Id)
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

// renderSettingsError creates a consistent error display for the settings page.
func (s *AdminFrontendServer) renderSettingsError(errorMsg string) g.Node {
	return components.ContentContainer(&components.ContentContainerProps{
		Title:    "Settings",
		Subtitle: "Manage service settings",
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
