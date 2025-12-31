package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Header
	header := headerStyle.Width(m.width).Render("Redmine TUI")

	// Left pane with title embedded in border
	leftBorderColor := lipgloss.Color(settings.Colors.InactivePaneBorder)
	if m.activePane == 0 {
		leftBorderColor = lipgloss.Color(settings.Colors.ActivePaneBorder)
	}

	// Render pane with border
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(0, 1).
		Render(m.leftPane.View())

	// Embed title in the top border line
	leftLines := strings.Split(leftPane, "\n")
	if len(leftLines) > 0 {
		// Calculate the actual border width: viewport width + padding (2) + borders (2)
		borderWidth := m.leftPane.Width + 4
		// Add dot indicator if active, use border lines before title
		var titlePart string
		if m.activePane == 0 {
			// Active: border lines + dot + title
			titlePart = "─ ● " + m.leftTitle + " "
		} else {
			// Inactive: border lines + title
			titlePart = "─── " + m.leftTitle + " "
		}
		if borderWidth > len(titlePart)+2 {
			// Build new border line: corner + title + remaining border + corner
			remainingBorder := ""
			if m.activePane == 1 {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+4)
			} else {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+2)
			}
			newPlainLine := "╭" + titlePart + remainingBorder + "╮"
			// Apply the border color
			styledTopLine := lipgloss.NewStyle().Foreground(leftBorderColor).Render(newPlainLine)
			leftLines[0] = styledTopLine
			leftPane = strings.Join(leftLines, "\n")
		}
	}

	// Right pane with title embedded in border
	rightBorderColor := lipgloss.Color(settings.Colors.InactivePaneBorder)
	if m.activePane == 1 {
		rightBorderColor = lipgloss.Color(settings.Colors.ActivePaneBorder)
	}

	// Render pane with border
	rightPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(0, 1).
		Render(m.rightPane.View())

	// Embed title in the top border line
	rightLines := strings.Split(rightPane, "\n")
	if len(rightLines) > 0 {
		// Calculate the actual border width: viewport width + padding (2) + borders (2)
		borderWidth := m.rightPane.Width + 2
		// Add dot indicator if active, use border lines before title
		var titlePart string
		if m.activePane == 1 {
			// Active: border lines + dot + title
			titlePart = "─ ● " + m.rightTitle + " "
		} else {
			// Inactive: border lines + title
			titlePart = "─── " + m.rightTitle + " "
		}
		if borderWidth > len(titlePart)+2 {
			// Build new border line: corner + title + remaining border + corner
			remainingBorder := ""
			if m.activePane == 1 {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+4)
			} else {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+6)
			}
			newPlainLine := "╭" + titlePart + remainingBorder + "╮"
			// Apply the border color
			styledTopLine := lipgloss.NewStyle().Foreground(rightBorderColor).Render(newPlainLine)
			rightLines[0] = styledTopLine
			rightPane = strings.Join(rightLines, "\n")
		}
	}

	// Combine panes side by side
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Footer with adaptive options
	footer := footerStyle.Width(m.width).Render(m.getFooterText())

	// Combine all sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		panes,
		footer,
	)
}

// getFooterText returns footer text that adapts to available width
func (m model) getFooterText() string {
	items := []string{
		"Tab: Switch",
		"↑↓/jk: Scroll",
		"PgUp/PgDn: Page",
		"?: Help",
		"q: Quit",
	}

	required := []string{"Tab: Switch", "q: Quit"}

	text := ""
	for _, item := range items {
		testText := text
		if testText != "" {
			testText += " | "
		}
		testText += item

		// Check if adding this item would exceed width
		// Measure plain text width (not styled)
		if len(testText) > m.width-2 {
			isRequired := false
			for _, req := range required {
				if item == req {
					isRequired = true
					break
				}
			}
			if !isRequired {
				continue
			}
		}

		if text != "" {
			text += " | "
		}
		text += item
	}

	return text
}
