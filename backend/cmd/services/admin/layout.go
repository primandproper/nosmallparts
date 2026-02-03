package main

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	g "maragu.dev/gomponents"
	ghtml "maragu.dev/gomponents/html"
)

// LayoutConfig holds configuration for the admin layout.
type LayoutConfig struct {
	Palette     *design.Palette
	AppName     string
	MaxWidth    string // e.g., "7xl", "6xl", "full"
	Margin      string // e.g., "4", "6", "8"
	ShowSidebar bool
}

// DefaultLayoutConfig provides sensible defaults for the admin layout.
func DefaultLayoutConfig() *LayoutConfig {
	return &LayoutConfig{
		Palette:     &design.StandardPalette,
		AppName:     "Admin Dashboard",
		MaxWidth:    "7xl",
		Margin:      "4",
		ShowSidebar: true,
	}
}

func header(config *LayoutConfig) g.Node {
	if config == nil {
		config = DefaultLayoutConfig()
	}

	return ghtml.Header(
		ghtml.Class(fmt.Sprintf("sticky top-0 z-50 %s border-b %s shadow-sm",
			design.Background(config.Palette.Background),
			design.BorderColor(config.Palette.Text),
		)),
		ghtml.Div(
			ghtml.Class(fmt.Sprintf("max-w-%s mx-auto px-%s", config.MaxWidth, config.Margin)),
			ghtml.Div(
				ghtml.Class("flex items-center justify-between h-16"),
				// Left side - Logo and main nav
				ghtml.Div(
					ghtml.Class("flex items-center space-x-8"),
					ghtml.A(
						ghtml.Href("/"),
						ghtml.H1(
							ghtml.Class(fmt.Sprintf("text-xl font-bold %s hover:opacity-80 transition-opacity duration-200 cursor-pointer", design.TextColor(config.Palette.Primary))),
							g.Text(config.AppName),
						),
					),
					// Main navigation
					ghtml.Nav(
						ghtml.Class("hidden md:flex space-x-6"),
						navLink("Users", "/users", config.Palette),
						navLink("Accounts", "/accounts", config.Palette),
						navLink("OAuth2 Clients", "/oauth2_clients", config.Palette),
						navLink("Recipes", "/recipes", config.Palette),
						navLink("Waitlists", "/waitlists", config.Palette),
						navLink("Issue Reports", "/issue_reports", config.Palette),
						navLink("Settings", "/settings", config.Palette),
						navDropdown("Enumerations", config.Palette, []*dropdownItem{
							{Text: "Ingredients", Href: "/valid_ingredients"},
							{Text: "Instruments", Href: "/valid_instruments"},
							{Text: "Preparations", Href: "/valid_preparations"},
							{Text: "Measurement Units", Href: "/valid_measurement_units"},
							{Text: "Vessels", Href: "/valid_vessels"},
							{Text: "Ingredient States", Href: "/valid_ingredient_states"},
						}),
					),
				),
				// Right side - Mobile menu button
				ghtml.Div(
					ghtml.Class("flex items-center space-x-4"),
					// Mobile menu button
					ghtml.Button(
						ghtml.Class(fmt.Sprintf("md:hidden p-2 rounded-md %s hover:%s focus:outline-none focus:ring-2 focus:ring-%s",
							design.TextColor(config.Palette.Text),
							design.Background(config.Palette.Secondary),
							config.Palette.Primary.Value,
						)),
						ghtml.Type("button"),
						g.Attr("aria-label", "Open menu"),
						g.Attr("onclick", "toggleMobileMenu()"),
						// Hamburger icon
						ghtml.Div(
							ghtml.Class("w-6 h-6 flex flex-col justify-center items-center"),
							ghtml.Span(ghtml.Class(fmt.Sprintf("block w-5 h-0.5 %s mb-1", design.Background(config.Palette.Text)))),
							ghtml.Span(ghtml.Class(fmt.Sprintf("block w-5 h-0.5 %s mb-1", design.Background(config.Palette.Text)))),
							ghtml.Span(ghtml.Class(fmt.Sprintf("block w-5 h-0.5 %s", design.Background(config.Palette.Text)))),
						),
					),
				),
			),
			// Mobile navigation menu
			ghtml.Div(
				ghtml.ID("mobile-menu"),
				ghtml.Class("hidden md:hidden border-t border-gray-200 py-4"),
				ghtml.Nav(
					ghtml.Class("flex flex-col space-y-2"),
					mobileNavLink("Users", "/users", config.Palette),
					mobileNavLink("Accounts", "/accounts", config.Palette),
					mobileNavLink("OAuth2 Clients", "/oauth2_clients", config.Palette),
					mobileNavLink("Recipes", "/recipes", config.Palette),
					mobileNavLink("Waitlists", "/waitlists", config.Palette),
					mobileNavLink("Issue Reports", "/issue_reports", config.Palette),
					mobileNavLink("Settings", "/settings", config.Palette),
					mobileNavDropdown("Enumerations", config.Palette, []*dropdownItem{
						{Text: "Ingredients", Href: "/valid_ingredients"},
						{Text: "Instruments", Href: "/valid_instruments"},
						{Text: "Preparations", Href: "/valid_preparations"},
						{Text: "Measurement Units", Href: "/valid_measurement_units"},
						{Text: "Vessels", Href: "/valid_vessels"},
						{Text: "Ingredient States", Href: "/valid_ingredient_states"},
					}),
				),
			),
		),
	)
}

func navLink(text, href string, palette *design.Palette) g.Node {
	return ghtml.A(
		ghtml.Href(href),
		ghtml.Class(fmt.Sprintf("px-3 py-2 rounded-md text-sm font-medium %s hover:%s hover:%s transition-colors duration-200",
			design.TextColor(palette.Text),
			design.TextColor(palette.Primary),
			design.Background(palette.Secondary),
		)),
		g.Text(text),
	)
}

func mobileNavLink(text, href string, palette *design.Palette) g.Node {
	return ghtml.A(
		ghtml.Href(href),
		ghtml.Class(fmt.Sprintf("block px-3 py-2 rounded-md text-base font-medium %s hover:%s hover:%s transition-colors duration-200",
			design.TextColor(palette.Text),
			design.TextColor(palette.Primary),
			design.Background(palette.Secondary),
		)),
		g.Text(text),
	)
}

// dropdownItem represents a single item in a dropdown menu.
type dropdownItem struct {
	Text string
	Href string
}

// navDropdown creates a dropdown menu for desktop navigation.
func navDropdown(text string, palette *design.Palette, items []*dropdownItem) g.Node {
	dropdownItems := make([]g.Node, len(items))
	for i, item := range items {
		dropdownItems[i] = ghtml.A(
			ghtml.Href(item.Href),
			ghtml.Class(fmt.Sprintf("block px-4 py-2 text-sm %s hover:%s hover:%s transition-colors duration-200",
				design.TextColor(palette.Text),
				design.TextColor(palette.Primary),
				design.Background(palette.Secondary),
			)),
			g.Text(item.Text),
		)
	}

	return ghtml.Div(
		ghtml.Class("relative group"),
		ghtml.Button(
			ghtml.Class(fmt.Sprintf("px-3 py-2 rounded-md text-sm font-medium %s hover:%s hover:%s transition-colors duration-200 flex items-center",
				design.TextColor(palette.Text),
				design.TextColor(palette.Primary),
				design.Background(palette.Secondary),
			)),
			ghtml.Type("button"),
			g.Text(text),
			// Dropdown arrow
			g.El("svg",
				ghtml.Class("ml-1 h-4 w-4"),
				g.Attr("fill", "none"),
				g.Attr("viewBox", "0 0 24 24"),
				g.Attr("stroke", "currentColor"),
				g.El("path",
					g.Attr("stroke-linecap", "round"),
					g.Attr("stroke-linejoin", "round"),
					g.Attr("stroke-width", "2"),
					g.Attr("d", "M19 9l-7 7-7-7"),
				),
			),
		),
		// Dropdown menu
		ghtml.Div(
			ghtml.Class(fmt.Sprintf("absolute left-0 mt-2 w-56 rounded-md shadow-lg %s ring-1 ring-black ring-opacity-5 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200 z-50",
				design.Background(palette.Background),
			)),
			ghtml.Div(
				ghtml.Class("py-1"),
				g.Group(dropdownItems),
			),
		),
	)
}

// mobileNavDropdown creates an expanded list of menu items for mobile navigation (CSS-only, no JavaScript).
func mobileNavDropdown(text string, palette *design.Palette, items []*dropdownItem) g.Node {
	dropdownItems := make([]g.Node, len(items))
	for i, item := range items {
		dropdownItems[i] = ghtml.A(
			ghtml.Href(item.Href),
			ghtml.Class(fmt.Sprintf("block px-6 py-2 text-sm %s hover:%s hover:%s transition-colors duration-200",
				design.TextColor(palette.Text),
				design.TextColor(palette.Primary),
				design.Background(palette.Secondary),
			)),
			g.Text(item.Text),
		)
	}

	return ghtml.Div(
		ghtml.Class("space-y-1"),
		// Section header (not clickable, just a label)
		ghtml.Div(
			ghtml.Class(fmt.Sprintf("px-3 py-2 text-base font-medium %s opacity-75",
				design.TextColor(palette.Text),
			)),
			g.Text(text),
		),
		// Menu items (always visible on mobile)
		ghtml.Div(
			ghtml.Class("space-y-1"),
			g.Group(dropdownItems),
		),
	)
}

var (
	tailwindImport = ghtml.Script(ghtml.Src("https://cdn.tailwindcss.com?plugins=typography"))

	htmxImport = g.Group{
		ghtml.Script(
			ghtml.Src("https://cdn.jsdelivr.net/npm/htmx.org@2.0.7/dist/htmx.min.js"),
			ghtml.Integrity("sha384-ZBXiYtYQ6hJ2Y0ZNoYuI+Nq5MqWBr+chMrS/RkXpNzQCApHEhOt2aY8EJgqwHLkJ"),
			ghtml.CrossOrigin("anonymous"),
		),

		ghtml.Script(
			ghtml.Src("https://unpkg.com/htmx.org@2.0.7/dist/ext/json-enc.js"),
			ghtml.Integrity("sha384-j+tqxLrwDkbeOdjbpaVekgvQL/J7qm/yh/UqSEs6RjEtnBFHqlJViBWG/jBZ6I2p"),
			ghtml.CrossOrigin("anonymous"),
		),
	}

	// JavaScript for mobile menu toggle.
	mobileMenuScript = ghtml.Script(
		g.Raw(`function toggleMobileMenu() {
				const menu = document.getElementById('mobile-menu');
				menu.classList.toggle('hidden');
			}

			document.addEventListener('click', function(event) {
				const menu = document.getElementById('mobile-menu');
				const button = event.target.closest('button[aria-label="Open menu"]');
				
				if (!button && !menu.contains(event.target)) {
					menu.classList.add('hidden');
				}
			});`),
	)
)

func page(title string, children ...g.Node) g.Node {
	return pageWithConfig(title, nil, children...)
}

func pageWithConfig(title string, config *LayoutConfig, children ...g.Node) g.Node {
	if config == nil {
		config = DefaultLayoutConfig()
	}

	return ghtml.HTML(
		ghtml.Lang("en"),
		ghtml.Head(
			ghtml.Meta(ghtml.Charset("utf-8")),
			ghtml.Meta(ghtml.Name("viewport"), ghtml.Content("width=device-width, initial-scale=1")),
			ghtml.Title(fmt.Sprintf("%s - %s", title, config.AppName)),
			tailwindImport,
			htmxImport,
			mobileMenuScript,
		),
		ghtml.Body(
			ghtml.Class(fmt.Sprintf("min-h-screen flex flex-col %s %s",
				design.Background(config.Palette.Background),
				design.TextColor(config.Palette.Text),
			)),
			header(config),
			ghtml.Main(
				ghtml.Class("flex-1 overflow-hidden"),
				ghtml.Div(
					ghtml.Class(fmt.Sprintf("h-full max-w-%s mx-auto px-%s py-%s",
						config.MaxWidth, config.Margin, config.Margin)),
					ghtml.Div(
						ghtml.Class("h-full overflow-auto"),
						ghtml.Div(
							ghtml.ID("main-content"),
							ghtml.Class("min-h-full"),
							g.Group(children),
						),
					),
				),
			),
			footer(config),
		),
	)
}

func footer(config *LayoutConfig) g.Node {
	if config == nil {
		config = DefaultLayoutConfig()
	}

	return ghtml.Footer(
		ghtml.Class(fmt.Sprintf("border-t %s %s",
			design.BorderColor(config.Palette.Text),
			design.Background(config.Palette.Background),
		)),
		ghtml.Div(
			ghtml.Class(fmt.Sprintf("max-w-%s mx-auto px-%s py-4", config.MaxWidth, config.Margin)),
			ghtml.Div(
				ghtml.Class("flex flex-col sm:flex-row justify-between items-center space-y-2 sm:space-y-0"),
				ghtml.P(
					ghtml.Class(fmt.Sprintf("text-sm %s", design.TextColor(config.Palette.Text))),
					g.Textf("© 2025 %s. All rights reserved.", config.AppName),
				),
				ghtml.Div(
					ghtml.Class("flex space-x-4 text-sm"),
					ghtml.A(
						ghtml.Href("/privacy"),
						ghtml.Class(fmt.Sprintf("%s hover:%s transition-colors duration-200",
							design.TextColor(config.Palette.Text),
							design.TextColor(config.Palette.Primary),
						)),
						g.Text("Privacy Policy"),
					),
					ghtml.A(
						ghtml.Href("/terms"),
						ghtml.Class(fmt.Sprintf("%s hover:%s transition-colors duration-200",
							design.TextColor(config.Palette.Text),
							design.TextColor(config.Palette.Primary),
						)),
						g.Text("Terms of Service"),
					),
				),
			),
		),
	)
}
