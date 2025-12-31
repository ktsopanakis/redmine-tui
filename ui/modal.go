package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ModalConfig contains configuration for rendering a modal window
type ModalConfig struct {
	Title       string   // Title of the modal
	Content     []string // Content lines to display
	Width       int      // Container width for positioning
	Height      int      // Container height for positioning
	BorderColor string   // Border color for the modal
	TitleColor  string   // Title color
}

// RenderModal renders a centered modal window with title and content
func RenderModal(cfg ModalConfig) string {
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
	
	// Add content lines
	for _, line := range cfg.Content {
		content.WriteString(line + "\n")
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
		Render(content.String())

	// Center the modal
	return centerModal(box, cfg.Width, cfg.Height)
}

// centerModal centers a modal box in the container
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

	var result strings.Builder
	
	// Top padding
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
