package main

import (
	"fmt"
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

	// Embed title in the top border line with position indicator
	leftLines := strings.Split(leftPane, "\n")
	if len(leftLines) > 0 {
		// Calculate the actual border width: viewport width + padding (2) + borders (2)
		borderWidth := m.leftPane.Width + 4

		// Build title with position indicator
		var titleText string
		if len(m.issues) > 0 {
			titleText = fmt.Sprintf("%s (%d/%d)", m.leftTitle, m.selectedIndex+1, len(m.issues))
		} else {
			titleText = m.leftTitle
		}

		// Add dot indicator if active, use border lines before title
		var titlePart string
		if m.activePane == 0 {
			// Active: border lines + dot + title
			titlePart = "─ ● " + titleText + " "
		} else {
			// Inactive: border lines + title
			titlePart = "─── " + titleText + " "
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
		}

		// Add arrow indicator on the left border for selected issue
		if len(m.issues) > 0 && m.activePane == 0 {
			// Calculate which border line corresponds to the selected issue
			// +1 for top border
			arrowLine := m.selectedDisplayLine + 1
			if arrowLine > 0 && arrowLine < len(leftLines)-1 {
				// Replace the │ character with →
				line := leftLines[arrowLine]
				if len(line) > 0 {
					// The border character is at the start
					runes := []rune(line)
					if len(runes) > 0 && runes[0] == '│' {
						leftLines[arrowLine] = lipgloss.NewStyle().
							Foreground(leftBorderColor).
							Render("→") + string(runes[1:])
					}
				}
			}
		}

		leftPane = strings.Join(leftLines, "\n")
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
