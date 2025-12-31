package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ktsopanakis/redmine-tui/config"
)

// PaneConfig contains configuration for rendering a pane
type PaneConfig struct {
	Content     string // The content to display in the pane
	Title       string // The title to embed in the border
	Width       int    // Width of the viewport (content area)
	IsActive    bool   // Whether this pane is active
	ShowDot     bool   // Whether to show the dot indicator (usually same as IsActive)
	CustomColor string // Custom border color (empty string uses default active/inactive colors)
	ArrowLine   int    // Line number to show arrow indicator (0 = no arrow, -1 = disabled)
	ShowArrow   bool   // Whether to show arrow at all
}

// RenderPane renders a pane with border, title, and optional indicators
func RenderPane(cfg PaneConfig) string {
	// Determine border color
	var borderColor lipgloss.Color
	if cfg.CustomColor != "" {
		borderColor = lipgloss.Color(cfg.CustomColor)
	} else if cfg.IsActive {
		borderColor = lipgloss.Color(config.Current.Colors.ActivePaneBorder)
	} else {
		borderColor = lipgloss.Color(config.Current.Colors.InactivePaneBorder)
	}

	// Render pane with border
	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(cfg.Content)

	// Split into lines for title and arrow embedding
	lines := strings.Split(pane, "\n")
	if len(lines) == 0 {
		return pane
	}

	// Embed title in the top border line
	borderWidth := cfg.Width + 4 // viewport width + padding (2) + borders (2)

	// Build title part with optional dot indicator
	var titlePart string
	if cfg.ShowDot {
		titlePart = "─ ● " + cfg.Title + " "
	} else {
		titlePart = "─── " + cfg.Title + " "
	}

	if borderWidth > len(titlePart)+2 {
		// Calculate remaining border length
		remainingLen := borderWidth - len(titlePart)
		if cfg.ShowDot {
			remainingLen += 2
		} else {
			remainingLen += 4
		}
		remainingBorder := strings.Repeat("─", remainingLen)

		// Build new top border line
		newPlainLine := "╭" + titlePart + remainingBorder + "╮"
		styledTopLine := lipgloss.NewStyle().Foreground(borderColor).Render(newPlainLine)
		lines[0] = styledTopLine
	}

	// Add arrow indicator if requested
	if cfg.ShowArrow && cfg.ArrowLine > 0 && cfg.ArrowLine < len(lines)-1 {
		line := lines[cfg.ArrowLine]
		if len(line) > 0 {
			runes := []rune(line)
			if len(runes) > 0 && runes[0] == '│' {
				lines[cfg.ArrowLine] = lipgloss.NewStyle().
					Foreground(borderColor).
					Render("→") + string(runes[1:])
			}
		}
	}

	return strings.Join(lines, "\n")
}

// RenderPaneWithColoredTitle renders a pane with a title that has custom color
func RenderPaneWithColoredTitle(cfg PaneConfig, titleColor string) string {
	// Determine border color
	var borderColor lipgloss.Color
	if cfg.CustomColor != "" {
		borderColor = lipgloss.Color(cfg.CustomColor)
	} else if cfg.IsActive {
		borderColor = lipgloss.Color(config.Current.Colors.ActivePaneBorder)
	} else {
		borderColor = lipgloss.Color(config.Current.Colors.InactivePaneBorder)
	}

	// Render pane with border
	pane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(cfg.Content)

	// Split into lines for title embedding
	lines := strings.Split(pane, "\n")
	if len(lines) == 0 {
		return pane
	}

	// Embed title in the top border line
	borderWidth := cfg.Width + 3 // viewport width + padding (2) + borders (2)

	// Build title part with optional dot indicator
	var titlePrefix string
	if cfg.ShowDot {
		titlePrefix = "─ ● "
	} else {
		titlePrefix = "─── "
	}

	if borderWidth > len(titlePrefix)+len(cfg.Title)+2 {
		// Calculate remaining border length
		remainingLen := borderWidth - len(titlePrefix) - len(cfg.Title) - 2
		if cfg.ShowDot {
			remainingLen += 4
		} else {
			remainingLen += 6
		}
		remainingBorder := strings.Repeat("─", remainingLen)

		// Build the top line with proper coloring
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(titleColor))
		borderStyle := lipgloss.NewStyle().Foreground(borderColor)

		newTopLine := borderStyle.Render("╭"+titlePrefix) +
			titleStyle.Render(cfg.Title) +
			borderStyle.Render(" "+remainingBorder+"╮")

		lines[0] = newTopLine
	}

	return strings.Join(lines, "\n")
}

// CombinePanes joins two panes horizontally
func CombinePanes(leftPane, rightPane string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}

// OverlayOnContent overlays content on top of background content
// This is useful for showing modal-like lists or dialogs on top of panes
func OverlayOnContent(background, overlay string) string {
	backgroundLines := strings.Split(background, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Calculate where to start overlaying (from bottom)
	startLine := len(backgroundLines) - len(overlayLines)
	if startLine < 0 {
		startLine = 0
	}

	// Keep all background lines, only overlay where content actually exists
	result := make([]string, len(backgroundLines))
	copy(result, backgroundLines) // Keep all original lines

	// Only replace the lines where the overlay actually appears
	for i, overlayLine := range overlayLines {
		lineIdx := startLine + i
		if lineIdx >= 0 && lineIdx < len(result) {
			// Only replace non-empty overlay lines to preserve background
			if strings.TrimSpace(overlayLine) != "" {
				result[lineIdx] = overlayLine
			}
		}
	}

	return strings.Join(result, "\n")
}
