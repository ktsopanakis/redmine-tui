package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ModalConfig contains configuration for rendering a modal window
type ModalConfig struct {
	Title        string   // Title of the modal
	Content      []string // Content lines to display
	Width        int      // Container width for positioning
	Height       int      // Container height for positioning
	BorderColor  string   // Border color for the modal
	TitleColor   string   // Title color
	ScrollOffset int      // Current scroll position (line offset)
	MaxHeight    int      // Maximum height for modal content area (0 = auto)
}

// RenderModal renders a centered modal window with title and content
func RenderModal(cfg ModalConfig) string {
	// Calculate maximum content height if not specified
	maxContentHeight := cfg.MaxHeight
	if maxContentHeight == 0 {
		// Use 85% of container height, leaving room for padding and border
		maxContentHeight = int(float64(cfg.Height) * 0.85)
	}

	// Reserve space for title, separator, borders, and padding
	titleHeight := 0
	if cfg.Title != "" {
		titleHeight = 3 // title + separator + blank line
	}
	maxVisibleLines := maxContentHeight - titleHeight - 4 // 4 for borders and padding

	if maxVisibleLines < 5 {
		maxVisibleLines = 5
	}

	// Calculate visible content window
	totalLines := len(cfg.Content)
	startLine := cfg.ScrollOffset
	endLine := startLine + maxVisibleLines

	if startLine < 0 {
		startLine = 0
	}
	if endLine > totalLines {
		endLine = totalLines
	}
	if startLine > totalLines-maxVisibleLines && totalLines > maxVisibleLines {
		startLine = totalLines - maxVisibleLines
	}

	// Build content
	var content strings.Builder

	// Add title
	if cfg.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(cfg.TitleColor)).
			Bold(true)
		content.WriteString(titleStyle.Render(cfg.Title) + "\n")
		content.WriteString(strings.Repeat("─", 60) + "\n\n")
	}

	// Add visible content lines
	for i := startLine; i < endLine; i++ {
		content.WriteString(cfg.Content[i] + "\n")
	}

	// Add scroll indicator if content is scrollable
	if totalLines > maxVisibleLines {
		content.WriteString("\n")
		scrollInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render(strings.Repeat("─", 60))
		content.WriteString(scrollInfo + "\n")

		indicator := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Render("↑/↓ to scroll | ")

		position := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")).
			Render("Showing lines " + strings.Join([]string{
				strings.Repeat(" ", len("Showing lines ")),
			}, "") + " | Esc to close")

		// Simple position text
		posText := ""
		if startLine == 0 {
			posText = "Top"
		} else if endLine >= totalLines {
			posText = "Bottom"
		} else {
			posText = "Middle"
		}

		position = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")).
			Render(posText + " | Esc to close")

		content.WriteString(indicator + position)
	}

	// Create the modal box
	borderColor := cfg.BorderColor
	if borderColor == "" {
		borderColor = "#61AFEF"
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(1, 2).
		Width(66). // Fixed width for consistency
		Render(content.String())

	// Center the modal
	return centerModal(box, cfg.Width, cfg.Height)
}

// centerModal centers a modal box in the container with opaque background
func centerModal(box string, containerWidth, containerHeight int) string {
	boxHeight := lipgloss.Height(box)
	boxWidth := lipgloss.Width(box)

	// Calculate padding
	horizontalPadding := (containerWidth - boxWidth) / 2
	if horizontalPadding < 0 {
		horizontalPadding = 0
	}

	verticalPadding := (containerHeight - boxHeight) / 2
	if verticalPadding < 0 {
		verticalPadding = 0
	}

	// Create background overlay style
	overlayStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#000000")).
		Foreground(lipgloss.Color("#333333"))

	var result strings.Builder

	// Top padding with opaque background
	for i := 0; i < verticalPadding; i++ {
		result.WriteString(overlayStyle.Render(strings.Repeat(" ", containerWidth)) + "\n")
	}

	// Add the box with horizontal padding (opaque background on sides)
	boxLines := strings.Split(box, "\n")
	for _, line := range boxLines {
		leftPadding := overlayStyle.Render(strings.Repeat(" ", horizontalPadding))
		rightPadding := overlayStyle.Render(strings.Repeat(" ", containerWidth-horizontalPadding-lipgloss.Width(line)))
		result.WriteString(leftPadding + line + rightPadding + "\n")
	}

	// Bottom padding with opaque background
	bottomPadding := containerHeight - verticalPadding - boxHeight
	if bottomPadding < 0 {
		bottomPadding = 0
	}
	for i := 0; i < bottomPadding; i++ {
		result.WriteString(overlayStyle.Render(strings.Repeat(" ", containerWidth)) + "\n")
	}

	return result.String()
}

// OverlayInCorner overlays content in a specific corner of the base content
func OverlayInCorner(base, overlay string, width int, corner string) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Find the maximum visual width of overlay
	overlayWidth := 0
	for _, line := range overlayLines {
		lineWidth := lipgloss.Width(line)
		if lineWidth > overlayWidth {
			overlayWidth = lineWidth
		}
	}

	overlayHeight := len(overlayLines)

	// Calculate position based on corner
	var startRow, startCol int
	switch corner {
	case "top-right":
		startRow = 1 // Leave 1 line for header
		startCol = width - overlayWidth - 1
		if startCol < 0 {
			startCol = 0
		}
	case "top-left":
		startRow = 1
		startCol = 2
	case "bottom-right":
		startRow = len(baseLines) - overlayHeight - 2
		startCol = width - overlayWidth - 1
		if startCol < 0 {
			startCol = 0
		}
	case "bottom-left":
		startRow = len(baseLines) - overlayHeight - 2
		startCol = 2
	default:
		// Default to top-right
		startRow = 1
		startCol = width - overlayWidth - 1
		if startCol < 0 {
			startCol = 0
		}
	}

	if startRow < 0 {
		startRow = 0
	}

	// Overlay the content line by line
	result := make([]string, len(baseLines))
	copy(result, baseLines)

	for i, overlayLine := range overlayLines {
		row := startRow + i
		if row >= len(result) {
			break
		}

		baseLine := result[row]
		baseWidth := lipgloss.Width(baseLine)

		// Build the new line by padding and placing overlay at the right position
		var newLine string

		if startCol >= baseWidth {
			// Need to pad the base line to reach startCol
			padding := startCol - baseWidth
			newLine = baseLine + strings.Repeat(" ", padding) + overlayLine
		} else {
			// Truncate base line at startCol and add overlay
			truncated := lipgloss.NewStyle().MaxWidth(startCol).Render(baseLine)
			actualWidth := lipgloss.Width(truncated)
			if actualWidth < startCol {
				// Add padding if truncation resulted in shorter width
				newLine = truncated + strings.Repeat(" ", startCol-actualWidth) + overlayLine
			} else {
				newLine = truncated + overlayLine
			}
		}

		result[row] = newLine
	}

	return strings.Join(result, "\n")
}
