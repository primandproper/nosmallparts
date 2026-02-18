package components

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	g "maragu.dev/gomponents"
)

// FormPageProps holds the configuration for a complete form page.
type FormPageProps[T any] struct {
	Data              T
	SubtitleGenerator func(data T) string
	Palette           *design.Palette
	FormOptions       *FormOptions[T]
	Title             string
	BaseSubtitle      string
	Actions           []g.Node
	Breadcrumbs       []*Breadcrumb
	AdditionalContent []g.Node
	ShowBreadcrumbs   bool
}

// Breadcrumb represents a breadcrumb navigation item.
type Breadcrumb struct {
	Text string
	URL  string
}

// FormPageResult contains the rendered form page and metadata.
type FormPageResult struct {
	Node g.Node
	// Add any metadata as needed
}

// FormPage creates a complete page with form, header, and optional breadcrumbs.
func FormPage[T any](props *FormPageProps[T]) (*FormPageResult, error) {
	if props.Palette == nil {
		props.Palette = &design.StandardPalette
	}

	if props.FormOptions == nil {
		props.FormOptions = &FormOptions[T]{}
	}

	// Ensure form options use the same palette
	if props.FormOptions.Palette == nil {
		props.FormOptions.Palette = props.Palette
	}

	// Generate subtitle
	subtitle := props.BaseSubtitle
	if props.SubtitleGenerator != nil {
		subtitle = props.SubtitleGenerator(props.Data)
	}

	// Create the form
	formNode, err := Form(props.Data, props.FormOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create form: %w", err)
	}

	// Build page content
	var pageContent []g.Node

	// Breadcrumbs (if enabled)
	if props.ShowBreadcrumbs && len(props.Breadcrumbs) > 0 {
		pageContent = append(pageContent, renderBreadcrumbs(props.Breadcrumbs, props.Palette))
	}

	// Content container with header
	contentContainer := ContentContainer(
		&ContentContainerProps{
			Title:    props.Title,
			Subtitle: subtitle,
			Palette:  props.Palette,
			Actions:  props.Actions,
		},
		// Form wrapped in a card
		Card(props.Palette, formNode),
	)

	pageContent = append(pageContent, contentContainer)

	// Additional content
	if len(props.AdditionalContent) > 0 {
		pageContent = append(pageContent, g.Group(props.AdditionalContent))
	}

	return &FormPageResult{
		Node: g.Group(pageContent),
	}, nil
}

// renderBreadcrumbs creates a breadcrumb navigation component.
func renderBreadcrumbs(breadcrumbs []*Breadcrumb, palette *design.Palette) g.Node {
	if len(breadcrumbs) == 0 {
		return g.Group(nil)
	}

	var items []g.Node

	for i, crumb := range breadcrumbs {
		isLast := i == len(breadcrumbs)-1

		// Breadcrumb item
		var itemContent g.Node
		if crumb.URL != "" && !isLast {
			itemContent = g.El("a",
				g.Attr("href", crumb.URL),
				g.Attr("class", fmt.Sprintf("%s hover:%s transition-colors",
					design.TextColor(palette.Primary),
					design.TextColor(design.Color{Value: palette.Primary.Value + "-700"}),
				)),
				g.Attr("hx-get", crumb.URL),
				g.Attr("hx-target", "body"),
				g.Attr("hx-swap", "innerHTML"),
				g.Attr("hx-push-url", "true"),
				g.Text(crumb.Text),
			)
		} else {
			textColor := design.TextColor(palette.Primary)
			if isLast {
				textColor = design.TextColor(palette.Text)
			}
			itemContent = g.El("span",
				g.Attr("class", textColor),
				g.Text(crumb.Text),
			)
		}

		items = append(items, g.El("li",
			g.Attr("class", "flex items-center"),
			itemContent,
			g.If(!isLast, g.El("svg",
				g.Attr("class", fmt.Sprintf("flex-shrink-0 h-5 w-5 %s mx-2", design.TextColor(palette.Text))),
				g.Attr("xmlns", "http://www.w3.org/2000/svg"),
				g.Attr("fill", "currentColor"),
				g.Attr("viewBox", "0 0 20 20"),
				g.Attr("aria-hidden", "true"),
				g.Raw(`<path d="M5.555 17.776l8-16 .894.448-8 16-.894-.448z"/>`),
			)),
		))
	}

	return g.El("nav",
		g.Attr("class", "flex mb-4"),
		g.Attr("aria-label", "Breadcrumb"),
		g.El("ol",
			g.Attr("class", "flex items-center space-x-2"),
			g.Group(items),
		),
	)
}
