package app

import (
	"fmt"

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

	// Build left pane title with position indicator
	var leftTitle string
	if len(m.issues) > 0 {
		leftTitle = fmt.Sprintf("%s (%d/%d)", m.leftTitle, m.selectedIndex+1, len(m.issues))
	} else {
		leftTitle = m.leftTitle
	}

	// Render left pane
	leftArrowLine := 0
	if len(m.issues) > 0 && m.activePane == 0 {
		leftArrowLine = m.selectedDisplayLine + 1 // +1 for top border
	}

	leftPane := appui.RenderPane(appui.PaneConfig{
		Content:   m.leftPane.View(),
		Title:     leftTitle,
		Width:     m.leftPane.Width,
		IsActive:  m.activePane == 0,
		ShowDot:   m.activePane == 0,
		ArrowLine: leftArrowLine,
		ShowArrow: m.activePane == 0 && len(m.issues) > 0,
	})

	// Render right pane with custom color if editing with unsaved changes
	rightCustomColor := ""
	if m.editMode && m.hasUnsavedChanges {
		rightCustomColor = "#E06C75" // Red border for unsaved changes
	}

	rightPane := appui.RenderPaneWithColoredTitle(appui.PaneConfig{
		Content:     m.rightPane.View(),
		Title:       m.rightTitle,
		Width:       m.rightPane.Width,
		IsActive:    m.activePane == 1,
		ShowDot:     m.activePane == 1,
		CustomColor: rightCustomColor,
	}, "#FFFFFF") // White title color

	// Combine panes side by side
	panes := appui.CombinePanes(leftPane, rightPane)

	// If in list selection mode, overlay the list on top
	if m.userInputMode == "user" || m.userInputMode == "project" {
		listOverlay := m.renderListOverlay()
		panes = appui.OverlayOnContent(panes, listOverlay)
	}

	// If modal is active, overlay the modal on top
	if m.showModal {
		var modal string
		switch m.modalType {
		case "help":
			modal = m.renderHelpModal()
		}
		if modal != "" {
			panes = appui.OverlayOnContent(panes, modal)
		}
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
		{Text: "?: Help", Required: false},
		{Text: "q: Quit", Required: true},
	}
}

// renderListOverlay renders the user/project selection list overlay
