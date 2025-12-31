package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ListItem represents an item in a selectable list
type ListItem struct {
	ID          int    // Unique identifier
	DisplayText string // Text to display
	IsSelected  bool   // Whether the item is currently selected
}

// ListConfig contains configuration for rendering a selection list
type ListConfig struct {
	Title            string            // Title/instructions for the list
	Items            []ListItem        // All available items
	Cursor           int               // Current cursor position
	FilterText       string            // Current filter text
	BorderColor      string            // Border color for the list box
	Width            int               // Total width of the container
	Height           int               // Total height of the container (for positioning)
	MaxVisibleItems  int               // Maximum items to show before scrolling
	IsLoading        bool              // Whether data is loading
	LoadingMessage   string            // Message to show when loading
	EmptyMessage     string            // Message to show when no items
	ShowScrollInfo   bool              // Whether to show scroll position info
	FilterFunc       func(item ListItem, filter string) bool // Custom filter function
}

// FilteredListResult contains the filtered list and updated indices
type FilteredListResult struct {
	Items           []ListItem
	UpdatedCursor   int
}

// BuildFilteredList filters and sorts items based on the filter text
// Selected items are shown first, then unselected items, both sorted alphabetically
func BuildFilteredList(cfg ListConfig) FilteredListResult {
	var selectedItems []ListItem
	var unselectedItems []ListItem

	// Default filter function if none provided
	filterFunc := cfg.FilterFunc
	if filterFunc == nil {
		filterFunc = func(item ListItem, filter string) bool {
			filterLower := strings.ToLower(filter)
			textLower := strings.ToLower(item.DisplayText)
			return strings.Contains(textLower, filterLower)
		}
	}

	// Filter and separate items
	for _, item := range cfg.Items {
		// Always show selected items, apply filter to unselected
		if cfg.FilterText != "" && !item.IsSelected {
			if !filterFunc(item, cfg.FilterText) {
				continue
			}
		}

		if item.IsSelected {
			selectedItems = append(selectedItems, item)
		} else {
			unselectedItems = append(unselectedItems, item)
		}
	}

	// Sort both groups alphabetically
	sort.Slice(selectedItems, func(i, j int) bool {
		return selectedItems[i].DisplayText < selectedItems[j].DisplayText
	})
	sort.Slice(unselectedItems, func(i, j int) bool {
		return unselectedItems[i].DisplayText < unselectedItems[j].DisplayText
	})

	// Combine: selected on top, then unselected
	items := append(selectedItems, unselectedItems...)

	// Adjust cursor to be within bounds
	cursor := cfg.Cursor
	if cursor >= len(items) {
		cursor = len(items) - 1
	}
	if cursor < 0 && len(items) > 0 {
		cursor = 0
	}

	return FilteredListResult{
		Items:         items,
		UpdatedCursor: cursor,
	}
}

// RenderListOverlay renders a selection list overlay (modal-like)
func RenderListOverlay(cfg ListConfig, headerHeight, footerHeight int) string {
	// Handle loading state
	if cfg.IsLoading {
		loadingBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(cfg.BorderColor)).
			Padding(1, 2).
			Width(cfg.Width - 6).
			Render(cfg.Title + "\n\n" + cfg.LoadingMessage)
		return positionAtBottom(loadingBox, cfg.Width, cfg.Height, headerHeight, footerHeight)
	}

	// Handle empty state
	if len(cfg.Items) == 0 {
		emptyBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(cfg.BorderColor)).
			Padding(1, 2).
			Width(cfg.Width - 6).
			Render(cfg.Title + "\n\n" + cfg.EmptyMessage)
		return positionAtBottom(emptyBox, cfg.Width, cfg.Height, headerHeight, footerHeight)
	}

	// Build filtered list
	filtered := BuildFilteredList(cfg)
	items := filtered.Items

	// Build list content
	var content strings.Builder
	content.WriteString(cfg.Title + "\n")
	content.WriteString(strings.Repeat("─", cfg.Width-8) + "\n\n")

	// Calculate visible items window
	maxVisible := cfg.MaxVisibleItems
	if maxVisible > len(items) {
		maxVisible = len(items)
	}

	startIdx := 0
	endIdx := len(items)

	if len(items) > maxVisible {
		// Calculate visible window centered on cursor
		startIdx = filtered.UpdatedCursor - maxVisible/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxVisible
		if endIdx > len(items) {
			endIdx = len(items)
			startIdx = endIdx - maxVisible
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	// Render visible items
	for i := startIdx; i < endIdx; i++ {
		checkbox := "[ ]"
		if items[i].IsSelected {
			checkbox = "[✓]"
		}
		cursor := "  "
		if i == filtered.UpdatedCursor {
			cursor = "→ "
		}
		content.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, items[i].DisplayText))
	}

	// Show scroll position indicator if needed
	if cfg.ShowScrollInfo && len(items) > maxVisible {
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(items)))
	} else if len(items) == 0 {
		content.WriteString("No matching items")
	}

	// Render the box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(cfg.BorderColor)).
		Padding(1, 2).
		Width(cfg.Width - 6).
		Render(content.String())

	return positionAtBottom(box, cfg.Width, cfg.Height, headerHeight, footerHeight)
}

// positionAtBottom positions content at the bottom with centered horizontal padding
func positionAtBottom(box string, containerWidth, containerHeight, headerHeight, footerHeight int) string {
	boxHeight := lipgloss.Height(box)
	boxWidth := lipgloss.Width(box)
	horizontalPadding := (containerWidth - boxWidth) / 2

	if horizontalPadding < 0 {
		horizontalPadding = 0
	}

	// Calculate available space
	emptySpace := containerHeight - headerHeight - footerHeight - 2

	var result strings.Builder

	// Top padding (push to bottom)
	verticalPadding := emptySpace - boxHeight - 1
	if verticalPadding < 0 {
		verticalPadding = 0
	}
	for i := 0; i < verticalPadding; i++ {
		result.WriteString("\n")
	}

	// Add the box with horizontal padding
	boxLines := strings.Split(box, "\n")
	for _, line := range boxLines {
		result.WriteString(strings.Repeat(" ", horizontalPadding) + line + "\n")
	}

	return result.String()
}
