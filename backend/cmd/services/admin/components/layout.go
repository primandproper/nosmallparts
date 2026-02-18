package components

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// ContentContainerProps holds configuration for content containers.
type ContentContainerProps struct {
	PageSizeSelector  g.Node
	SearchModifiers   []g.Node
	Palette           *design.Palette
	Title             string
	Subtitle          string
	SearchPlaceholder string
	HTMXSearchTarget  string
	HTMXSearchTrigger string
	Actions           []g.Node
	ShowSearch        bool
}

// ContentContainer creates a responsive content container with optional search.
func ContentContainer(props *ContentContainerProps, children ...g.Node) g.Node {
	if props.Palette == nil {
		props.Palette = &design.StandardPalette
	}

	var headerContent []g.Node

	// Title section - takes only needed space
	titleSection := ghtml.Div(
		ghtml.Class("flex-shrink-0"),
		ghtml.H1(
			ghtml.Class(fmt.Sprintf("text-2xl font-bold %s", design.TextColor(props.Palette.Primary))),
			g.Text(props.Title),
		),
		g.If(props.Subtitle != "", ghtml.P(
			ghtml.Class(fmt.Sprintf("mt-1 text-sm %s", design.TextColor(props.Palette.Text))),
			g.Text(props.Subtitle),
		)),
	)

	headerContent = append(headerContent, titleSection)

	// Search section (if enabled) - takes remaining space
	if props.ShowSearch {
		placeholder := props.SearchPlaceholder
		if placeholder == "" {
			placeholder = "Search..."
		}

		searchSection := ghtml.Div(
			ghtml.Class("flex-1 mx-4"),
			SearchInput(&SearchInputProps{
				Placeholder: placeholder,
				Palette:     props.Palette,
				HTMXTarget:  props.HTMXSearchTarget,
				HTMXTrigger: props.HTMXSearchTrigger,
			}),
		)
		headerContent = append(headerContent, searchSection)
	}

	// Page size selector (if provided) - between search and actions
	if props.PageSizeSelector != nil {
		headerContent = append(headerContent, props.PageSizeSelector)
	}

	// Search modifiers (if provided) - between page size selector and actions
	if len(props.SearchModifiers) > 0 {
		headerContent = append(headerContent, ghtml.Div(
			ghtml.Class("flex items-center gap-2"),
			g.Group(props.SearchModifiers),
		))
	}

	// Actions section - takes only needed space
	if len(props.Actions) > 0 {
		headerContent = append(headerContent, ghtml.Div(
			ghtml.Class("flex items-center space-x-3 flex-shrink-0"),
			g.Group(props.Actions),
		))
	}

	return ghtml.Div(
		ghtml.Class("space-y-6"),
		// Header with title, search, and actions in one row
		ghtml.Div(
			ghtml.Class("flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-3 lg:space-y-0"),
			g.Group(headerContent),
		),
		// Content
		ghtml.Div(
			ghtml.Class("space-y-4"),
			g.Group(children),
		),
	)
}

// SearchInputProps holds configuration for search inputs.
type SearchInputProps struct {
	Placeholder string
	Palette     *design.Palette
	ID          string
	Name        string
	HTMXTarget  string // HTMX target for live search
	HTMXTrigger string // HTMX trigger event
}

// SearchInput creates a styled search input with optional HTMX integration.
func SearchInput(props *SearchInputProps) g.Node {
	if props.Palette == nil {
		props.Palette = &design.StandardPalette
	}

	id := props.ID
	if id == "" {
		id = "search"
	}

	name := props.Name
	if name == "" {
		name = "search"
	}

	placeholder := props.Placeholder
	if placeholder == "" {
		placeholder = "Search..."
	}

	inputAttrs := []g.Node{
		ghtml.Type("text"),
		ghtml.ID(id),
		ghtml.Name(name),
		ghtml.Placeholder(placeholder),
		ghtml.Class(fmt.Sprintf("block w-full pl-10 pr-3 py-2 border %s rounded-md leading-5 %s placeholder-%s focus:outline-none focus:placeholder-%s focus:ring-1 focus:ring-%s focus:border-%s sm:text-sm",
			design.BorderColor(props.Palette.Background),
			design.Background(design.Color{Value: "white"}),
			props.Palette.Text.Value,
			props.Palette.Text.Value,
			props.Palette.Primary.Value,
			props.Palette.Primary.Value,
		)),
	}

	// Add HTMX attributes if specified
	if props.HTMXTarget != "" {
		inputAttrs = append(inputAttrs,
			g.Attr("hx-get", props.HTMXTarget),
			g.Attr("hx-target", "#search-results"),
			g.Attr("hx-swap", "innerHTML"),
		)

		trigger := props.HTMXTrigger
		if trigger == "" {
			trigger = "keyup changed delay:300ms"
		}
		inputAttrs = append(inputAttrs, g.Attr("hx-trigger", trigger))
	}

	return ghtml.Div(
		ghtml.Class("relative w-full"),
		ghtml.Div(
			ghtml.Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
			// Search icon (using a simple SVG)
			ghtml.Div(
				ghtml.Class(fmt.Sprintf("h-5 w-5 %s", design.TextColor(props.Palette.Text))),
				g.Raw(`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>`),
			),
		),
		ghtml.Input(inputAttrs...),
	)
}

// ActionButton creates a styled action button.
func ActionButton(text, href string, palette *design.Palette, isPrimary bool) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	var classes string
	if isPrimary {
		classes = fmt.Sprintf("inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm %s %s hover:%s focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-%s",
			design.TextColor(design.Color{Value: "white"}),
			design.Background(palette.Primary),
			design.Background(design.Color{Value: palette.Primary.Value + "-700"}), // Darker shade for hover
			palette.Primary.Value,
		)
	} else {
		classes = fmt.Sprintf("inline-flex items-center px-4 py-2 border %s text-sm font-medium rounded-md %s %s hover:%s focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-%s",
			design.BorderColor(palette.Text),
			design.TextColor(palette.Text),
			design.Background(design.Color{Value: "white"}),
			design.Background(palette.Secondary),
			palette.Primary.Value,
		)
	}

	return ghtml.A(
		ghtml.Href(href),
		ghtml.Class(classes),
		g.Text(text),
	)
}

// Card creates a styled card container.
func Card(palette *design.Palette, children ...g.Node) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	return ghtml.Div(
		ghtml.Class(fmt.Sprintf("%s shadow rounded-lg border %s",
			design.Background(design.Color{Value: "white"}),
			design.BorderColor(palette.Background),
		)),
		ghtml.Div(
			ghtml.Class("p-6"),
			g.Group(children),
		),
	)
}

// CardWithHeader creates a card with a distinct header section.
func CardWithHeader(title string, palette *design.Palette, headerActions []g.Node, children ...g.Node) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	var headerContent []g.Node

	// Title
	headerContent = append(headerContent, ghtml.H3(
		ghtml.Class(fmt.Sprintf("text-lg font-medium %s", design.TextColor(palette.Primary))),
		g.Text(title),
	))

	// Actions (if any)
	if len(headerActions) > 0 {
		headerContent = append(headerContent, ghtml.Div(
			ghtml.Class("flex items-center space-x-2"),
			g.Group(headerActions),
		))
	}

	return ghtml.Div(
		ghtml.Class(fmt.Sprintf("%s shadow rounded-lg border %s overflow-hidden",
			design.Background(design.Color{Value: "white"}),
			design.BorderColor(palette.Background),
		)),
		// Header
		ghtml.Div(
			ghtml.Class(fmt.Sprintf("px-6 py-4 border-b %s %s",
				design.BorderColor(palette.Background),
				design.Background(palette.Background),
			)),
			ghtml.Div(
				ghtml.Class("flex items-center justify-between"),
				g.Group(headerContent),
			),
		),
		// Content
		ghtml.Div(
			ghtml.Class("p-6"),
			g.Group(children),
		),
	)
}

// EmptyState creates an empty state component.
func EmptyState(title, description string, palette *design.Palette, actions []g.Node) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	return ghtml.Div(
		ghtml.Class("text-center py-12"),
		ghtml.Div(
			ghtml.Class("mx-auto h-12 w-12 text-gray-400"),
			// Empty state icon
			g.Raw(`<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
				<path vector-effect="non-scaling-stroke" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
			</svg>`),
		),
		ghtml.H3(
			ghtml.Class(fmt.Sprintf("mt-2 text-sm font-medium %s", design.TextColor(palette.Text))),
			g.Text(title),
		),
		ghtml.P(
			ghtml.Class(fmt.Sprintf("mt-1 text-sm %s", design.TextColor(palette.Text))),
			g.Text(description),
		),
		g.If(len(actions) > 0, ghtml.Div(
			ghtml.Class("mt-6 flex justify-center space-x-3"),
			g.Group(actions),
		)),
	)
}

// LoadingSpinner creates a loading spinner component.
func LoadingSpinner(palette *design.Palette) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	return ghtml.Div(
		ghtml.Class("flex justify-center items-center py-8"),
		ghtml.Div(
			ghtml.Class(fmt.Sprintf("animate-spin rounded-full h-8 w-8 border-b-2 %s", design.BorderColor(palette.Primary))),
		),
	)
}
