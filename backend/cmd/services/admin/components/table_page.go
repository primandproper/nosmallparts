package components

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/cmd/services/admin/design"

	g "maragu.dev/gomponents"
)

// TablePageMetadata holds information about the processed table data.
type TablePageMetadata struct {
	Fields        []*fieldInfo
	TotalCount    int
	FilteredCount int
	EmptyState    bool
	HasData       bool
}

// TablePageProps holds configuration for a complete table-based page.
type TablePageProps[T any] struct {
	TableOptions          *TableOptions[T]
	SubtitleGenerator     func(metadata TablePageMetadata) string
	Palette               *design.Palette
	RowLinkGenerator      func(item T) string
	EmptyStateTitle       string
	Title                 string
	HTMXSearchTrigger     string
	SearchPlaceholder     string
	HTMXSearchTarget      string
	EmptyStateDescription string
	BaseSubtitle          string
	Actions               []g.Node
	Data                  []T
	EmptyStateActions     []g.Node
	SearchModifiers       []g.Node
	ShowSearch            bool
}

// TablePageResult contains both the rendered page and metadata about the table.
type TablePageResult struct {
	Node     g.Node
	Metadata TablePageMetadata
}

// TablePage creates a complete page with integrated table and context-aware layout.
func TablePage[T any](props *TablePageProps[T]) (*TablePageResult, error) {
	if props == nil {
		return nil, fmt.Errorf("props cannot be nil")
	}

	// Set defaults
	if props.Palette == nil {
		props.Palette = &design.StandardPalette
	}
	if props.TableOptions == nil {
		props.TableOptions = &TableOptions[T]{}
	}
	if props.TableOptions.Palette == nil {
		props.TableOptions.Palette = props.Palette
	}
	// If RowLinkGenerator is provided at the TablePageProps level, use it
	if props.RowLinkGenerator != nil && props.TableOptions.RowLinkGenerator == nil {
		props.TableOptions.RowLinkGenerator = props.RowLinkGenerator
	}

	// Generate metadata from the data
	metadata := generateMetadata(props.Data, props.TableOptions)

	// Handle empty state
	if metadata.EmptyState {
		return &TablePageResult{
			Node:     createEmptyTablePage(props),
			Metadata: metadata,
		}, nil
	}

	// Generate table
	table, err := Table(props.Data, props.TableOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate table: %w", err)
	}

	// Generate dynamic subtitle
	subtitle := generateSubtitle(props, metadata)

	// Generate page size selector if pagination is available and search is enabled
	var pageSizeSelector g.Node
	if props.ShowSearch && props.TableOptions != nil && props.TableOptions.Pagination != nil && props.HTMXSearchTarget != "" {
		pageSizeSelector = CreatePageSizeSelector(
			props.HTMXSearchTarget,
			props.TableOptions.Pagination.MaxResponseSize,
			props.TableOptions.TableID,
		)
	}

	// Create the complete page
	node := ContentContainer(&ContentContainerProps{
		Title:             props.Title,
		Subtitle:          subtitle,
		Palette:           props.Palette,
		ShowSearch:        props.ShowSearch,
		SearchPlaceholder: props.SearchPlaceholder,
		HTMXSearchTarget:  props.HTMXSearchTarget,
		HTMXSearchTrigger: props.HTMXSearchTrigger,
		Actions:           props.Actions,
		PageSizeSelector:  pageSizeSelector,
		SearchModifiers:   props.SearchModifiers,
	},
		// Wrap table in horizontally scrollable container for HTMX targeting
		g.El("div",
			g.Attr("id", "search-results"),
			g.Attr("class", "overflow-x-auto"),
			table,
		),
	)

	return &TablePageResult{
		Node:     node,
		Metadata: metadata,
	}, nil
}

// generateMetadata extracts metadata from the data and table options.
func generateMetadata[T any](data []T, options *TableOptions[T]) TablePageMetadata {
	metadata := TablePageMetadata{
		TotalCount:    len(data),
		FilteredCount: len(data), // Same as total for now, will change with filtering
		EmptyState:    len(data) == 0,
		HasData:       len(data) > 0,
	}

	// Extract field information if we have data
	if len(data) > 0 {
		if fields, err := extractFields(data[0]); err == nil {
			metadata.Fields = applyFieldOrdering(fields, options.Fields, options.FieldReplacements)
		}
	}

	return metadata
}

// generateSubtitle creates a dynamic subtitle based on the data.
func generateSubtitle[T any](props *TablePageProps[T], metadata TablePageMetadata) string {
	// Use custom generator if provided
	if props.SubtitleGenerator != nil {
		return props.SubtitleGenerator(metadata)
	}

	// Default subtitle generation
	if metadata.EmptyState {
		return props.BaseSubtitle
	}

	// Create a count-aware subtitle
	var countText string
	switch metadata.TotalCount {
	case 1:
		countText = "1 item"
	default:
		countText = fmt.Sprintf("%d items", metadata.TotalCount)
	}

	if props.BaseSubtitle != "" {
		return fmt.Sprintf("%s - %s", props.BaseSubtitle, countText)
	}

	return countText
}

// createEmptyTablePage creates the page for when there's no data.
func createEmptyTablePage[T any](props *TablePageProps[T]) g.Node {
	emptyActions := props.EmptyStateActions
	if len(emptyActions) == 0 && len(props.Actions) > 0 {
		// Use the first action as the empty state primary action
		emptyActions = props.Actions[:1]
	}

	return ContentContainer(&ContentContainerProps{
		Title:             props.Title,
		Subtitle:          props.BaseSubtitle,
		Palette:           props.Palette,
		ShowSearch:        false, // No search for empty state
		HTMXSearchTarget:  props.HTMXSearchTarget,
		HTMXSearchTrigger: props.HTMXSearchTrigger,
		Actions:           props.Actions,
	},
		// Wrap empty state in scrollable container with ID for HTMX targeting
		g.El("div",
			g.Attr("id", "search-results"),
			g.Attr("class", "overflow-x-auto"),
			EmptyState(
				getEmptyStateTitle(props),
				getEmptyStateDescription(props),
				props.Palette,
				emptyActions,
			),
		),
	)
}

// Helper functions for empty state.
func getEmptyStateTitle[T any](props *TablePageProps[T]) string {
	if props.EmptyStateTitle != "" {
		return props.EmptyStateTitle
	}
	return fmt.Sprintf("No %s found", props.Title)
}

func getEmptyStateDescription[T any](props *TablePageProps[T]) string {
	if props.EmptyStateDescription != "" {
		return props.EmptyStateDescription
	}
	return fmt.Sprintf("Get started by creating your first %s.", props.Title)
}

// TablePageWithCustomStats creates a table page with custom statistics.
func TablePageWithCustomStats[T any](props *TablePageProps[T], customStats []g.Node) (*TablePageResult, error) {
	result, err := TablePage(props)
	if err != nil {
		return nil, err
	}

	// If we have data, replace the default stats with custom ones
	if result.Metadata.HasData && len(customStats) > 0 {
		// Generate table
		var table g.Node
		table, err = Table(props.Data, props.TableOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to generate table: %w", err)
		}

		subtitle := generateSubtitle(props, result.Metadata)

		// Create page with custom stats
		node := ContentContainer(&ContentContainerProps{
			Title:             props.Title,
			Subtitle:          subtitle,
			Palette:           props.Palette,
			ShowSearch:        props.ShowSearch,
			SearchPlaceholder: props.SearchPlaceholder,
			HTMXSearchTarget:  props.HTMXSearchTarget,
			HTMXSearchTrigger: props.HTMXSearchTrigger,
			Actions:           props.Actions,
		},
			// Custom stats
			g.Group(customStats),

			// Wrap table in horizontally scrollable container for HTMX targeting
			g.El("div",
				g.Attr("id", "search-results"),
				g.Attr("class", "overflow-x-auto"),
				table,
			),
		)

		result.Node = node
	}

	return result, nil
}
