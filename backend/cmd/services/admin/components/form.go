package components

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	"google.golang.org/protobuf/types/known/timestamppb"
	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// SelectOption represents a single option in a select dropdown.
type SelectOption struct {
	Value     string
	Label     string
	Disabled  bool
	IsDefault bool
}

// FieldConfig holds configuration for a single form field.
type FieldConfig struct {
	Validation     *FieldValidation
	CustomRenderer func(fieldName string, value any, config *FieldConfig, palette *design.Palette) g.Node
	Name           string
	DisplayName    string
	InputType      string
	Placeholder    string
	Options        []*SelectOption
	Enabled        bool
}

// FieldValidation holds HTMX-compatible validation rules.
type FieldValidation struct {
	Min           *float64
	Max           *float64
	Pattern       string
	HTMXValidate  string
	CustomMessage string
	MinLength     int
	MaxLength     int
	Required      bool
}

// FormRow represents a row of fields in the form.
type FormRow struct {
	Fields      []string
	ColumnSpans []int
	Columns     int
}

// FormOptions holds configuration options for the form.
type FormOptions[T any] struct {
	FieldConfigs                map[string]*FieldConfig
	Palette                     *design.Palette
	CancelButtonText            string
	FormID                      string
	CSSClasses                  string
	CancelURL                   string
	Action                      string
	HTMXSwap                    string
	HTMXTarget                  string
	Method                      string
	SubmitButtonText            string
	FormRows                    []*FormRow
	AdditionalButtons           []g.Node
	EnabledFields               []string
	ShowCancelButton            bool
	HTMXPushURL                 bool
	DisableAutoEnableZeroValues bool // If true, disables automatic enabling of zero-value fields
}

// Form creates a generic HTML form from a struct with HTMX support.
func Form[T any](data T, options *FormOptions[T]) (g.Node, error) {
	if options == nil {
		options = &FormOptions[T]{}
	}

	// Get palette (use standard if not provided)
	palette := options.Palette
	if palette == nil {
		palette = &design.StandardPalette
	}

	// Extract field information from the data
	fields, values, err := extractFieldsAndValues(data)
	if err != nil {
		return nil, fmt.Errorf("failed to extract fields: %w", err)
	}

	// Apply field configurations
	fields = applyFieldConfigs(fields, options)

	// Build form attributes
	formAttrs := buildFormAttributes(options)

	// Generate form rows
	formRows := generateFormRows(fields, values, options, palette)

	// Generate buttons section
	buttonsSection := generateFormButtons(options, palette)

	return ghtml.Form(
		append(formAttrs,
			ghtml.Class(buildFormClasses(options.CSSClasses)),
			ghtml.Div(
				ghtml.Class("space-y-6"),
				g.Group(formRows),
			),
			buttonsSection,
		)...,
	), nil
}

// extractFieldsAndValues uses reflection to get field information and values from a struct.
func extractFieldsAndValues(item any) ([]*fieldInfo, map[string]any, error) {
	var fields []*fieldInfo
	values := make(map[string]any)

	v := reflect.ValueOf(item)
	if !v.IsValid() {
		return nil, nil, fmt.Errorf("invalid value")
	}

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, nil, fmt.Errorf("nil pointer")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("expected struct or pointer to struct, got %s", v.Kind())
	}

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Extract JSON tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			// Skip fields without JSON tags or with "-" tag
			continue
		}

		// Parse JSON tag (handle "fieldName,omitempty" format)
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" || jsonName == "-" {
			continue
		}

		// Create a display name
		displayName := camelCaseToTitleCase(field.Name)

		fields = append(fields, &fieldInfo{
			Name:        jsonName,   // Use JSON tag name
			GoFieldName: field.Name, // Keep Go struct field name
			DisplayName: displayName,
			Type:        field.Type,
		})

		// Extract value
		fieldValue := v.Field(i)
		if fieldValue.CanInterface() {
			values[jsonName] = fieldValue.Interface() // Use JSON tag name as key
		}
	}

	return fields, values, nil
}

// applyFieldConfigs applies field configurations and enabled fields list.
func applyFieldConfigs[T any](fields []*fieldInfo, options *FormOptions[T]) []*fieldInfo {
	// Create a set of enabled fields for quick lookup
	enabledSet := make(map[string]bool)
	for _, name := range options.EnabledFields {
		enabledSet[name] = true
	}

	// Apply custom display names from configs
	for i := range fields {
		if config, exists := options.FieldConfigs[fields[i].Name]; exists {
			if config.DisplayName != "" {
				fields[i].DisplayName = config.DisplayName
			}
		}
	}

	return fields
}

// buildFormAttributes constructs the form HTML attributes.
func buildFormAttributes[T any](options *FormOptions[T]) []g.Node {
	attrs := []g.Node{}

	if options.FormID != "" {
		attrs = append(attrs, ghtml.ID(options.FormID))
	}

	if options.Action != "" {
		attrs = append(attrs, ghtml.Action(options.Action))
	}

	method := options.Method
	if method == "" {
		method = "POST"
	}
	attrs = append(attrs, ghtml.Method(method))

	// HTMX attributes
	if options.Action != "" {
		attrs = append(attrs, g.Attr("hx-post", options.Action))
	}

	if options.HTMXTarget != "" {
		attrs = append(attrs, g.Attr("hx-target", options.HTMXTarget))
	}

	swap := options.HTMXSwap
	if swap == "" {
		swap = "innerHTML"
	}
	attrs = append(attrs, g.Attr("hx-swap", swap))

	if options.HTMXPushURL {
		attrs = append(attrs, g.Attr("hx-push-url", "true"))
	}

	return attrs
}

// generateFormRows generates the form field rows based on layout configuration.
func generateFormRows[T any](fields []*fieldInfo, values map[string]any, options *FormOptions[T], palette *design.Palette) []g.Node {
	if palette == nil {
		p := design.StandardPalette
		palette = &p
	}

	var rows []g.Node

	// If no custom layout specified, create one field per row
	if len(options.FormRows) == 0 {
		for _, field := range fields {
			config := getFieldConfig(field.Name, values[field.Name], options)
			fieldNode := generateFormField(field, values[field.Name], config, palette)
			rows = append(rows, ghtml.Div(
				ghtml.Class("grid grid-cols-1 gap-4"),
				fieldNode,
			))
		}
		return rows
	}

	// Create a map for quick field lookup
	fieldMap := make(map[string]*fieldInfo)
	for _, field := range fields {
		fieldMap[field.Name] = field
	}

	// Generate rows based on custom layout
	for _, formRow := range options.FormRows {
		var rowFields []g.Node
		columns := formRow.Columns
		if columns == 0 {
			columns = len(formRow.Fields)
		}

		// If ColumnSpans are specified, use a 12-column grid with col-span classes
		if len(formRow.ColumnSpans) > 0 {
			for i, fieldName := range formRow.Fields {
				if field, exists := fieldMap[fieldName]; !exists {
					continue
				} else {
					config := getFieldConfig(fieldName, values[fieldName], options)
					fieldNode := generateFormField(field, values[fieldName], config, palette)

					// Get the column span for this field (default to auto if not enough spans specified)
					span := 12 / len(formRow.Fields) // equal distribution as fallback
					if i < len(formRow.ColumnSpans) {
						span = formRow.ColumnSpans[i]
					}

					// Wrap field in a div with col-span class
					colSpanClass := fmt.Sprintf("col-span-12 md:col-span-%d", span)
					wrappedField := ghtml.Div(
						ghtml.Class(colSpanClass),
						fieldNode,
					)
					rowFields = append(rowFields, wrappedField)
				}
			}

			if len(rowFields) > 0 {
				rows = append(rows, ghtml.Div(
					ghtml.Class("grid grid-cols-12 gap-4"),
					g.Group(rowFields),
				))
			}
		} else {
			// Use default grid layout for equal-width columns
			for _, fieldName := range formRow.Fields {
				if field, exists := fieldMap[fieldName]; exists {
					config := getFieldConfig(fieldName, values[fieldName], options)
					fieldNode := generateFormField(field, values[fieldName], config, palette)
					rowFields = append(rowFields, fieldNode)
				}
			}

			if len(rowFields) > 0 {
				gridClass := fmt.Sprintf("grid grid-cols-1 md:grid-cols-%d gap-4", columns)
				rows = append(rows, ghtml.Div(
					ghtml.Class(gridClass),
					g.Group(rowFields),
				))
			}
		}
	}

	return rows
}

// isZeroValue checks if a value is a zero value (nil, empty string, zero number, zero timestamp, etc.)
func isZeroValue(value any) bool {
	if value == nil {
		return true
	}

	// Handle protobuf Timestamp specially
	if ts, ok := value.(*timestamppb.Timestamp); ok {
		return ts == nil || !ts.IsValid() || ts.AsTime().IsZero()
	}

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Struct:
		// For time.Time
		if t, ok := value.(time.Time); ok {
			return t.IsZero()
		}
		// For other structs, check if all fields are zero (expensive, so we'll just return false)
		return false
	default:
		return false
	}
}

// getFieldConfig retrieves the field configuration or returns defaults.
// It automatically enables fields with zero values unless explicitly disabled via DisableAutoEnableZeroValues.
func getFieldConfig[T any](fieldName string, fieldValue any, options *FormOptions[T]) *FieldConfig {
	// Check if field is explicitly in enabled list
	explicitlyEnabled := false
	for _, name := range options.EnabledFields {
		if name == fieldName {
			explicitlyEnabled = true
			break
		}
	}

	// Get or create config
	var config *FieldConfig
	if options.FieldConfigs != nil {
		if c, exists := options.FieldConfigs[fieldName]; exists {
			config = c
		}
	}

	// Apply enabled status:
	// - Always enabled if explicitly in EnabledFields list
	// - OR enabled if the value is zero/empty AND auto-enable is not disabled
	//   (making empty fields editable by default unless DisableAutoEnableZeroValues is set)
	if explicitlyEnabled {
		config.Enabled = true
	} else if !options.DisableAutoEnableZeroValues && isZeroValue(fieldValue) {
		config.Enabled = true
	}

	config.Name = fieldName

	return config
}

// generateFormField generates a single form field.
func generateFormField(field *fieldInfo, value any, config *FieldConfig, palette *design.Palette) g.Node {
	// Ensure palette is not nil
	if palette == nil {
		p := design.StandardPalette
		palette = &p
	}

	// Use custom renderer if provided
	if config.CustomRenderer != nil {
		return config.CustomRenderer(field.Name, value, config, palette)
	}

	displayName := field.DisplayName
	if config.DisplayName != "" {
		displayName = config.DisplayName
	}

	// Build label content
	labelContent := []g.Node{
		ghtml.For(field.Name),
		ghtml.Class(fmt.Sprintf("block text-sm font-medium %s mb-2", design.TextColor(palette.Text))),
		g.Text(displayName),
	}

	if config.Validation != nil && config.Validation.Required {
		labelContent = append(labelContent, ghtml.Span(
			ghtml.Class("text-red-500 ml-1"),
			g.Text("*"),
		))
	}

	// Build field content
	var inputElement g.Node

	// Render select element if options are provided
	if len(config.Options) > 0 {
		inputElement = buildSelectElement(field.Name, value, config, palette)
	} else {
		inputElement = buildInputElement(field.Name, field.Type, value, config, palette)
	}

	fieldContent := []g.Node{
		ghtml.Class("flex flex-col"),
		ghtml.Label(labelContent...),
		inputElement,
	}

	if config.Validation != nil && config.Validation.CustomMessage != "" {
		fieldContent = append(fieldContent, ghtml.P(
			ghtml.Class("mt-1 text-sm text-gray-500"),
			g.Text(config.Validation.CustomMessage),
		))
	}

	return ghtml.Div(fieldContent...)
}

// buildInputElement creates an HTML input element.
func buildInputElement(fieldName string, fieldType reflect.Type, value any, config *FieldConfig, palette *design.Palette) g.Node {
	inputType := config.InputType
	if inputType == "" {
		if fieldType != nil {
			inputType = inferInputType(fieldType)
		} else {
			inputType = "text"
		}
	}

	// Build input attributes
	inputAttrs := []g.Node{
		ghtml.Type(inputType),
		ghtml.ID(fieldName),
		ghtml.Name(fieldName),
		ghtml.Class(buildInputClasses(palette)),
	}

	// Handle checkbox specially
	if inputType == "checkbox" {
		// For checkboxes, set value to "true" and add checked attribute if the boolean is true
		inputAttrs = append(inputAttrs, ghtml.Value("true"))
		if value != nil {
			if boolVal, ok := value.(bool); ok && boolVal {
				inputAttrs = append(inputAttrs, ghtml.Checked())
			}
		}
	} else if value != nil {
		inputAttrs = append(inputAttrs, ghtml.Value(formatValue(value)))
	}

	// Add disabled attribute if not enabled
	if !config.Enabled {
		inputAttrs = append(inputAttrs, ghtml.Disabled())
	}

	// Add placeholder
	if config.Placeholder != "" {
		inputAttrs = append(inputAttrs, ghtml.Placeholder(config.Placeholder))
	}

	// Add validation attributes
	if config.Validation != nil {
		inputAttrs = append(inputAttrs, buildValidationAttributes(config.Validation)...)
	}

	return ghtml.Input(inputAttrs...)
}

// buildSelectElement creates an HTML select element with options.
func buildSelectElement(fieldName string, value any, config *FieldConfig, palette *design.Palette) g.Node {
	// Build select attributes
	selectAttrs := []g.Node{
		ghtml.ID(fieldName),
		ghtml.Name(fieldName),
		ghtml.Class(buildSelectClasses(palette)),
	}

	// Add disabled attribute if not enabled
	if !config.Enabled {
		selectAttrs = append(selectAttrs, ghtml.Disabled())
	}

	// Add validation attributes
	if config.Validation != nil && config.Validation.Required {
		selectAttrs = append(selectAttrs, ghtml.Required())
	}

	// Format the current value for comparison
	currentValue := formatValue(value)

	// Build options
	var options []g.Node

	// Add placeholder option if provided
	if config.Placeholder != "" {
		options = append(options, ghtml.Option(
			ghtml.Value(""),
			ghtml.Disabled(),
			g.If(currentValue == "", ghtml.Selected()),
			g.Text(config.Placeholder),
		))
	}

	// Add configured options
	for _, opt := range config.Options {
		optionAttrs := []g.Node{
			ghtml.Value(opt.Value),
			g.Text(opt.Label),
		}

		// Check if this option should be selected
		if opt.IsDefault && currentValue == "" {
			optionAttrs = append(optionAttrs, ghtml.Selected())
		} else if currentValue == opt.Value {
			optionAttrs = append(optionAttrs, ghtml.Selected())
		}

		if opt.Disabled {
			optionAttrs = append(optionAttrs, ghtml.Disabled())
		}

		options = append(options, ghtml.Option(optionAttrs...))
	}

	return ghtml.Select(append(selectAttrs, g.Group(options))...)
}

// buildSelectClasses constructs the CSS classes for select fields.
func buildSelectClasses(palette *design.Palette) string {
	if palette == nil {
		p := design.StandardPalette
		palette = &p
	}
	return fmt.Sprintf("block w-full px-3 py-2 bg-white border %s rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-%s focus:border-%s sm:text-sm disabled:bg-gray-100 disabled:cursor-not-allowed",
		design.BorderColor(palette.Background),
		palette.Primary.Value,
		palette.Primary.Value,
	)
}

// inferInputType infers the HTML input type from the Go type.
func inferInputType(t reflect.Type) string {
	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "number"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "checkbox"
	default:
		// Check if it's a time.Time or timestamppb.Timestamp
		if t.String() == "time.Time" || t.String() == "timestamppb.Timestamp" {
			return "datetime-local"
		}
		return "text"
	}
}

// formatValue formats a value for display in an input field.
func formatValue(value any) string {
	if value == nil {
		return ""
	}

	// Handle protobuf Timestamp pointers specially
	if ts, ok := value.(*timestamppb.Timestamp); ok {
		if ts == nil || !ts.IsValid() {
			return ""
		}
		return ts.AsTime().Format("2006-01-02T15:04")
	}

	switch v := value.(type) {
	case time.Time:
		if v.IsZero() {
			return ""
		}
		// Format for datetime-local input
		return v.Format("2006-01-02T15:04")
	case *time.Time:
		if v == nil || v.IsZero() {
			return ""
		}
		return v.Format("2006-01-02T15:04")
	case string:
		return v
	case *string:
		if v == nil {
			return ""
		}
		return *v
	case *float32:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", *v)
	case *float64:
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", *v)
	case bool:
		if v {
			return "checked"
		}
		return ""
	case fmt.Stringer:
		return v.String()
	default:
		// Handle slices and arrays - don't try to display them in form fields
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
			return ""
		}
		return fmt.Sprintf("%v", v)
	}
}

// buildValidationAttributes constructs HTML5 and HTMX validation attributes.
func buildValidationAttributes(validation *FieldValidation) []g.Node {
	var attrs []g.Node

	if validation.Required {
		attrs = append(attrs, ghtml.Required())
	}

	if validation.MinLength > 0 {
		attrs = append(attrs, ghtml.MinLength(fmt.Sprintf("%d", validation.MinLength)))
	}

	if validation.MaxLength > 0 {
		attrs = append(attrs, ghtml.MaxLength(fmt.Sprintf("%d", validation.MaxLength)))
	}

	if validation.Pattern != "" {
		attrs = append(attrs, ghtml.Pattern(validation.Pattern))
	}

	if validation.Min != nil {
		attrs = append(attrs, ghtml.Min(fmt.Sprintf("%f", *validation.Min)))
	}

	if validation.Max != nil {
		attrs = append(attrs, ghtml.Max(fmt.Sprintf("%f", *validation.Max)))
	}

	if validation.HTMXValidate != "" {
		attrs = append(attrs,
			g.Attr("hx-post", validation.HTMXValidate),
			g.Attr("hx-trigger", "blur"),
			g.Attr("hx-target", "next .validation-message"),
		)
	}

	return attrs
}

// buildInputClasses constructs the CSS classes for input fields.
func buildInputClasses(palette *design.Palette) string {
	if palette == nil {
		p := design.StandardPalette
		palette = &p
	}
	return fmt.Sprintf("block w-full px-3 py-2 bg-white border %s rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-%s focus:border-%s sm:text-sm disabled:bg-gray-100 disabled:cursor-not-allowed",
		design.BorderColor(palette.Background),
		palette.Primary.Value,
		palette.Primary.Value,
	)
}

// generateFormButtons generates the form buttons section.
func generateFormButtons[T any](options *FormOptions[T], palette *design.Palette) g.Node {
	if palette == nil {
		p := design.StandardPalette
		palette = &p
	}

	var buttons []g.Node

	// Submit button
	submitText := options.SubmitButtonText
	if submitText == "" {
		submitText = "Submit"
	}

	buttons = append(buttons, ghtml.Button(
		ghtml.Type("submit"),
		ghtml.Class(fmt.Sprintf("inline-flex justify-center px-4 py-2 text-sm font-medium %s %s border border-transparent rounded-md shadow-sm hover:%s focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-%s",
			design.TextColor(design.Color{Value: "white"}),
			design.Background(palette.Primary),
			design.Background(design.Color{Value: palette.Primary.Value + "-700"}),
			palette.Primary.Value,
		)),
		g.Text(submitText),
	))

	// Cancel button
	if options.ShowCancelButton {
		cancelText := options.CancelButtonText
		if cancelText == "" {
			cancelText = "Cancel"
		}

		cancelAttrs := []g.Node{
			ghtml.Type("button"),
			ghtml.Class(fmt.Sprintf("inline-flex justify-center px-4 py-2 text-sm font-medium %s %s border %s rounded-md shadow-sm hover:%s focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-%s",
				design.TextColor(palette.Text),
				design.Background(design.Color{Value: "white"}),
				design.BorderColor(palette.Text),
				design.Background(palette.Background),
				palette.Primary.Value,
			)),
			g.Text(cancelText),
		}

		if options.CancelURL != "" {
			cancelAttrs = append(cancelAttrs,
				g.Attr("hx-get", options.CancelURL),
				g.Attr("hx-target", "body"),
				g.Attr("hx-swap", "innerHTML"),
				g.Attr("hx-push-url", "true"),
			)
		}

		buttons = append(buttons, ghtml.Button(cancelAttrs...))
	}

	// Additional buttons
	buttons = append(buttons, options.AdditionalButtons...)

	return ghtml.Div(
		ghtml.Class("flex justify-end space-x-3 mt-6 pt-6 border-t border-gray-200"),
		g.Group(buttons),
	)
}

// buildFormClasses constructs the CSS classes for the form.
func buildFormClasses(customClasses string) string {
	baseClasses := "bg-white shadow rounded-lg p-6"
	if customClasses != "" {
		return fmt.Sprintf("%s %s", baseClasses, customClasses)
	}
	return baseClasses
}
