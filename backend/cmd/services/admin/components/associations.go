package components

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// AssociationItem represents a single association between two entities.
type AssociationItem struct {
	ID          string
	Name        string
	Description string
	Notes       string
}

// AssociationListProps defines the properties for displaying associations.
type AssociationListProps struct {
	Palette              *design.Palette
	Title                string
	EntityID             string
	AddSearchPlaceholder string
	AddSearchEndpoint    string
	CreateEndpoint       string
	DeleteEndpoint       string
	NoItemsMessage       string
	HTMXTarget           string
	Items                []*AssociationItem
}

// AssociationList renders a list of associations with add/remove functionality.
func AssociationList(props *AssociationListProps) g.Node {
	if props.Palette == nil {
		props.Palette = &design.StandardPalette
	}

	if props.HTMXTarget == "" {
		props.HTMXTarget = "#association-list-container"
	}

	if props.NoItemsMessage == "" {
		props.NoItemsMessage = "No associations found."
	}

	return ghtml.Div(
		ghtml.ID("association-list-container"),
		ghtml.Class("space-y-4"),

		// Title
		ghtml.H3(
			ghtml.Class("text-lg font-semibold text-gray-900"),
			g.Text(props.Title),
		),

		// Add new association section
		ghtml.Div(
			ghtml.Class("bg-gray-50 p-4 rounded-lg border border-gray-200"),
			ghtml.Div(
				ghtml.Class("flex gap-2"),

				// Search input with HTMX
				ghtml.Input(
					ghtml.Type("text"),
					ghtml.Name("q"),
					ghtml.Placeholder(props.AddSearchPlaceholder),
					ghtml.Class("flex-1 px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"),
					g.Attr("hx-get", props.AddSearchEndpoint),
					g.Attr("hx-trigger", "keyup changed delay:300ms"),
					g.Attr("hx-target", "#search-results"),
					g.Attr("hx-swap", "innerHTML"),
					g.Attr("autocomplete", "off"),
				),
			),

			// Search results container
			ghtml.Div(
				ghtml.ID("search-results"),
				ghtml.Class("mt-2"),
			),
		),

		// Current associations list
		ghtml.Div(
			ghtml.Class("mt-4"),
			g.If(
				len(props.Items) == 0,
				// Enhanced empty state
				ghtml.Div(
					ghtml.Class("text-center py-12 px-4 bg-gray-50 border-2 border-dashed border-gray-300 rounded-lg"),
					ghtml.Div(
						ghtml.Class("mx-auto w-16 h-16 mb-4 flex items-center justify-center rounded-full bg-gray-100"),
						// Icon for empty state
						g.Raw(`<svg class="w-8 h-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
						</svg>`),
					),
					ghtml.P(
						ghtml.Class("text-gray-600 font-medium mb-1"),
						g.Text(props.NoItemsMessage),
					),
					ghtml.P(
						ghtml.Class("text-sm text-gray-500"),
						g.Text("Use the search box above to add associations."),
					),
				),
			),
			g.If(
				len(props.Items) > 0,
				ghtml.Div(
					ghtml.Class("space-y-2"),
					g.Map(props.Items, func(item *AssociationItem) g.Node {
						return AssociationListItem(&AssociationListItemProps{
							Item:           item,
							DeleteEndpoint: props.DeleteEndpoint,
							HTMXTarget:     props.HTMXTarget,
							Palette:        props.Palette,
						})
					}),
				),
			),
		),
	)
}

// AssociationListItemProps defines the properties for a single association item.
type AssociationListItemProps struct {
	Palette        *design.Palette
	Item           *AssociationItem
	DeleteEndpoint string
	HTMXTarget     string
}

// AssociationListItem renders a single association item with delete button.
func AssociationListItem(props *AssociationListItemProps) g.Node {
	return ghtml.Div(
		ghtml.Class("flex items-start justify-between p-3 bg-white border border-gray-200 rounded-lg hover:border-gray-300 transition-colors"),

		// Item info
		ghtml.Div(
			ghtml.Class("flex-1"),
			ghtml.Div(
				ghtml.Class("font-medium text-gray-900"),
				g.Text(props.Item.Name),
			),
			g.If(
				props.Item.Description != "",
				ghtml.P(
					ghtml.Class("text-sm text-gray-600 mt-1"),
					g.Text(props.Item.Description),
				),
			),
			g.If(
				props.Item.Notes != "",
				ghtml.P(
					ghtml.Class("text-xs text-gray-500 mt-1 italic"),
					g.Text("Notes: "+props.Item.Notes),
				),
			),
		),

		// Delete button
		ghtml.Button(
			ghtml.Type("button"),
			ghtml.Class("ml-3 p-2 text-red-600 hover:bg-red-50 rounded-md transition-colors"),
			g.Attr("hx-delete", props.DeleteEndpoint+"/"+props.Item.ID),
			g.Attr("hx-target", props.HTMXTarget),
			g.Attr("hx-swap", "outerHTML"),
			g.Attr("hx-confirm", "Are you sure you want to remove this association?"),
			ghtml.Title("Remove association"),
			// SVG trash icon
			g.Raw(`<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
			</svg>`),
		),
	)
}

// SearchResultItem represents a search result for adding associations.
type SearchResultItem struct {
	ID          string
	Name        string
	Description string
	ExtraInfo   string // optional extra information to display
}

// AssociationSearchResultsProps defines the properties for search results.
type AssociationSearchResultsProps struct {
	CreateEndpoint string
	HTMXTarget     string
	EntityID       string
	NoResultsText  string
	Results        []*SearchResultItem
}

// AssociationSearchResults renders search results for adding associations.
func AssociationSearchResults(props *AssociationSearchResultsProps) g.Node {
	if props.NoResultsText == "" {
		props.NoResultsText = "No results found."
	}

	if len(props.Results) == 0 {
		return ghtml.Div(
			ghtml.Class("text-sm text-gray-500 py-2"),
			g.Text(props.NoResultsText),
		)
	}

	return ghtml.Div(
		ghtml.Class("max-h-60 overflow-y-auto space-y-1 border border-gray-200 rounded-md bg-white"),
		g.Map(props.Results, func(item *SearchResultItem) g.Node {
			return ghtml.Button(
				ghtml.Type("button"),
				ghtml.Class("w-full text-left px-3 py-2 hover:bg-gray-50 transition-colors border-b border-gray-100 last:border-b-0"),
				g.Attr("hx-post", props.CreateEndpoint),
				g.Attr("hx-target", props.HTMXTarget),
				g.Attr("hx-swap", "outerHTML"),
				g.Attr("hx-vals", `{"id": "`+item.ID+`"}`),
				ghtml.Div(
					ghtml.Class("font-medium text-gray-900"),
					g.Text(item.Name),
				),
				g.If(
					item.Description != "",
					ghtml.Div(
						ghtml.Class("text-sm text-gray-600 mt-0.5"),
						g.Text(item.Description),
					),
				),
			)
		}),
	)
}
