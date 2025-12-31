package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ktsopanakis/redmine-tui/config"
	appui "github.com/ktsopanakis/redmine-tui/ui"
)

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Build header sections
	leftSections := []appui.HeaderSection{
		{Text: "◆", Color: "#FFFFFF", Bold: true},
		{Text: config.Current.Redmine.URL, Color: "#FFD700", Bold: true},
	}

	dayOfWeek, dateTime := appui.FormatDateTime()
	rightSections := []appui.HeaderSection{}

	if m.currentUser != nil {
		rightSections = append(rightSections,
			appui.HeaderSection{Text: m.currentUser.Name, Color: "#C678DD", Bold: false},
			appui.HeaderSection{Text: "|", Color: "#666666", Bold: false},
		)
	}
	rightSections = append(rightSections,
		appui.HeaderSection{Text: dayOfWeek, Color: "#61AFEF", Bold: false},
		appui.HeaderSection{Text: dateTime, Color: "#98C379", Bold: false},
	)

	header := appui.RenderHeader(leftSections, rightSections, m.width)

	// Left pane with title embedded in border
	leftBorderColor := lipgloss.Color(config.Current.Colors.InactivePaneBorder)
	if m.activePane == 0 {
		leftBorderColor = lipgloss.Color(config.Current.Colors.ActivePaneBorder)
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
	rightBorderColor := lipgloss.Color(config.Current.Colors.InactivePaneBorder)
	if m.editMode && m.hasUnsavedChanges {
		// Red border when editing with unsaved changes
		rightBorderColor = lipgloss.Color("#E06C75") // Red
	} else if m.activePane == 1 {
		rightBorderColor = lipgloss.Color(config.Current.Colors.ActivePaneBorder)
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
			// Active: border lines + dot + title (white)
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
			// Build border with styled title (white for ID)
			whiteTitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
			borderOnlyStyle := lipgloss.NewStyle().Foreground(rightBorderColor)

			// Build the top line with proper coloring
			newTopLine := borderOnlyStyle.Render("╭")
			if m.activePane == 1 {
				newTopLine += borderOnlyStyle.Render("─ ● ") + whiteTitleStyle.Render(m.rightTitle) + borderOnlyStyle.Render(" "+remainingBorder+"╮")
			} else {
				newTopLine += borderOnlyStyle.Render("─── ") + whiteTitleStyle.Render(m.rightTitle) + borderOnlyStyle.Render(" "+remainingBorder+"╮")
			}
			rightLines[0] = newTopLine
			rightPane = strings.Join(rightLines, "\n")
		}
	}

	// Combine panes side by side
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// If in list selection mode, overlay the list on top
	if m.userInputMode == "user" || m.userInputMode == "project" {
		listOverlay := m.renderListOverlay()
		// Overlay on top of the panes
		panes = m.overlayListOnPanes(panes, listOverlay)
	}

	// Footer with adaptive options
	var footer string
	if m.filterMode {
		footer = appui.RenderPromptFooter("Filter: ", m.filterInput.View(), m.width, "#61AFEF")
	} else if m.editMode {
		footer = appui.RenderFooter(m.renderEditFooter(), m.width)
	} else if m.userInputMode == "user" {
		footer = appui.RenderPromptFooter("Filter Users: ", m.filterInput.View(), m.width, "#61AFEF")
	} else if m.userInputMode == "project" {
		footer = appui.RenderPromptFooter("Filter Projects: ", m.filterInput.View(), m.width, "#98C379")
	} else {
		menuText := appui.BuildAdaptiveMenu(m.getFooterItems(), m.width-2, " | ")
		footer = appui.RenderFooter(menuText, m.width)
	}

	// Combine all sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		panes,
		footer,
	)
}

// getFooterItems returns footer menu items with required status
func (m Model) getFooterItems() []appui.FooterItem {
	return []appui.FooterItem{
		{Text: "↑↓/jk: Nav", Required: false},
		{Text: "Tab: Switch", Required: false},
		{Text: "f: Filter", Required: true},
		{Text: "m: My/All", Required: true},
		{Text: "e: Edit", Required: true},
		{Text: "q: Quit", Required: true},
	}
}

// overlayListOnPanes overlays the list selection on top of the panes
func (m Model) overlayListOnPanes(panes, listOverlay string) string {
	panesLines := strings.Split(panes, "\n")
	overlayLines := strings.Split(listOverlay, "\n")

	// Calculate where to start overlaying (from bottom)
	startLine := len(panesLines) - len(overlayLines)
	if startLine < 0 {
		startLine = 0
	}

	// Keep all background lines, only overlay where the list actually is
	result := make([]string, len(panesLines))
	copy(result, panesLines) // Keep all original lines

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

// renderListOverlay renders the user/project selection list overlay
