package components

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// SearchBoxProps defines the properties for a search box component.
type SearchBoxProps struct {
	Palette            *design.Palette
	ID                 string
	Placeholder        string
	SearchEndpoint     string
	OnSelectCallback   string
	HiddenFieldName    string
	InitialValue       string
	InitialDisplayText string
}

// SearchBox creates a search box with HTMX-powered autocomplete.
func SearchBox(props *SearchBoxProps) g.Node {
	if props.Palette == nil {
		props.Palette = &design.StandardPalette
	}

	searchInputID := props.ID + "-input"
	selectedDisplayID := props.ID + "-selected"
	hiddenFieldID := props.ID + "-hidden"

	return ghtml.Div(
		ghtml.Class("relative"),

		// Hidden field that stores the selected ID
		ghtml.Input(
			ghtml.Type("hidden"),
			ghtml.ID(hiddenFieldID),
			ghtml.Name(props.HiddenFieldName),
			g.If(props.InitialValue != "", ghtml.Value(props.InitialValue)),
		),

		// Display area for selected item
		g.If(props.InitialValue != "",
			ghtml.Div(
				ghtml.ID(selectedDisplayID),
				ghtml.Class("mb-2 px-3 py-2 bg-blue-50 border border-blue-200 rounded-md flex items-center justify-between"),
				ghtml.Span(
					ghtml.Class("text-sm text-gray-700"),
					g.Text(props.InitialDisplayText),
				),
				ghtml.Button(
					ghtml.Type("button"),
					ghtml.Class("text-sm text-blue-600 hover:text-blue-800 font-medium"),
					g.Text("Clear"),
					g.Attr("onclick", fmt.Sprintf(`
						document.getElementById('%s').value = '';
						document.getElementById('%s').style.display = 'none';
						document.getElementById('%s').style.display = 'block';
						document.getElementById('%s').value = '';
						document.getElementById('%s').focus();
					`, hiddenFieldID, selectedDisplayID, searchInputID, searchInputID, searchInputID)),
				),
			),
		),

		// Search input container
		ghtml.Div(
			ghtml.ID(searchInputID),
			g.If(props.InitialValue != "", g.Attr("style", "display: none;")),
			ghtml.Input(
				ghtml.Type("text"),
				ghtml.ID(props.ID+"-query"),
				ghtml.Name("q"),
				ghtml.Placeholder(props.Placeholder),
				ghtml.Class(fmt.Sprintf("w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-%s focus:border-%s",
					props.Palette.Primary.Value,
					props.Palette.Primary.Value,
				)),
				g.Attr("hx-get", props.SearchEndpoint),
				g.Attr("hx-trigger", "keyup changed delay:300ms"),
				g.Attr("hx-target", "#"+searchInputID),
				g.Attr("hx-swap", "outerHTML"),
				g.Attr("hx-include", "this"),
				g.Attr("hx-vals", fmt.Sprintf(`{"search_box_id": %q}`, props.ID)),
				g.Attr("autocomplete", "off"),
			),
		),
	)
}

// SearchInputWithResults renders the search input (with preserved query) and results dropdown.
func SearchInputWithResults(searchBoxID, searchEndpoint, query string, results []*SearchResultItem, onSelectJS string, palette *design.Palette) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	searchInputID := searchBoxID + "-input"
	resultsID := searchBoxID + "-results"

	var resultNodes []g.Node
	if len(results) > 0 {
		for _, result := range results {
			resultNodes = append(resultNodes, SearchResultItem_(result, onSelectJS, palette))
		}
	}

	return ghtml.Div(
		ghtml.ID(searchInputID),
		// Search input with preserved query
		ghtml.Input(
			ghtml.Type("text"),
			ghtml.ID(searchBoxID+"-query"),
			ghtml.Name("q"),
			ghtml.Value(query),
			ghtml.Placeholder("Search..."),
			ghtml.Class(fmt.Sprintf("w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-%s focus:border-%s",
				palette.Primary.Value,
				palette.Primary.Value,
			)),
			g.Attr("hx-get", searchEndpoint),
			g.Attr("hx-trigger", "keyup changed delay:300ms"),
			g.Attr("hx-target", "#"+searchInputID),
			g.Attr("hx-swap", "outerHTML"),
			g.Attr("hx-include", "this"),
			g.Attr("hx-vals", fmt.Sprintf(`{"search_box_id": %q}`, searchBoxID)),
			g.Attr("autocomplete", "off"),
		),
		// Results dropdown
		g.If(len(results) > 0,
			ghtml.Div(
				ghtml.ID(resultsID),
				ghtml.Class("absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-60 overflow-auto"),
				g.Group(resultNodes),
			),
		),
	)
}

// SearchResults creates a list of search results and manages visibility.
func SearchResults(results []*SearchResultItem, onSelectJS string, palette *design.Palette) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	// If no results, hide the container
	if len(results) == 0 {
		return ghtml.Div(
			ghtml.Style("display: none;"),
		)
	}

	var items []g.Node
	for _, result := range results {
		items = append(items, SearchResultItem_(result, onSelectJS, palette))
	}

	// Return results with inline script to show the container
	return ghtml.Div(
		g.Raw(`<script>document.currentScript.parentElement.style.display = 'block';</script>`),
		g.Group(items),
	)
}

// SearchResultItem_ creates a single search result item.
func SearchResultItem_(item *SearchResultItem, onSelectJS string, palette *design.Palette) g.Node {
	if palette == nil {
		palette = &design.StandardPalette
	}

	// Build the onclick handler - needs ID, Name, Name (for hidden value, display text, and innerHTML)
	// Also hide the results dropdown after selection
	onclick := fmt.Sprintf(onSelectJS, item.ID, item.Name, item.Name)

	return ghtml.Button(
		ghtml.Type("button"),
		ghtml.Class(fmt.Sprintf("w-full text-left px-4 py-3 hover:bg-%s hover:bg-opacity-10 border-b border-gray-100 last:border-b-0 focus:outline-none focus:bg-%s focus:bg-opacity-10",
			palette.Primary.Value,
			palette.Primary.Value,
		)),
		g.Attr("onclick", onclick),
		ghtml.Div(
			ghtml.Class("flex flex-col"),
			ghtml.Div(
				ghtml.Class("font-medium text-gray-900"),
				g.Text(item.Name),
			),
			g.If(item.Description != "",
				ghtml.Div(
					ghtml.Class("text-sm text-gray-600 mt-1"),
					g.Text(item.Description),
				),
			),
			g.If(item.ExtraInfo != "",
				ghtml.Div(
					ghtml.Class("text-xs text-gray-500 mt-1"),
					g.Text(item.ExtraInfo),
				),
			),
		),
	)
}
