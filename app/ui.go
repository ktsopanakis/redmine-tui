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

	// If note mode is active, overlay the note input on top
	if m.noteMode {
		panes = appui.OverlayOnContent(panes, m.renderNoteOverlay())
	}

	// If the description editor is open, overlay it on top
	if m.descEditMode {
		panes = appui.OverlayOnContent(panes, m.renderDescEditor())
	}

	// If the quick status picker is open, overlay it on top
	if m.statusPickMode {
		panes = appui.OverlayOnContent(panes, m.renderStatusPicker())
	}

	// If the quick-actions popup is open, overlay it on top
	if m.quickMode {
		panes = appui.OverlayOnContent(panes, m.renderQuickActions())
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

	// If loading indicator is visible, overlay it in top-right corner
	if m.loadingIndicator.Visible {
		loadingView := m.loadingIndicator.View()
		if loadingView != "" {
			panes = appui.OverlayInCorner(panes, loadingView, m.width, "top-right")
		}
	}

	// Footer with adaptive options
	var footer string
	if m.filterMode {
		footer = appui.RenderPromptFooter("Filter: ", m.filterInput.View(), m.width, "#61AFEF")
	} else if m.noteMode {
		footer = appui.RenderFooter("Ctrl+S: Post note  |  Esc: Cancel", m.width)
	} else if m.descEditMode {
		footer = appui.RenderFooter("Ctrl+S: Save description  |  Esc: Cancel", m.width)
	} else if m.statusPickMode {
		footer = appui.RenderFooter("↑↓/1-9: Select  |  Enter: Apply  |  Esc: Cancel", m.width)
	} else if m.quickMode {
		footer = appui.RenderFooter("Tab: Next field  |  Ctrl+S: Apply all  |  Esc: Cancel", m.width)
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

// renderNoteOverlay renders the add-note input as a centered modal
func (m Model) renderNoteOverlay() string {
	return appui.RenderInputModal(appui.InputModalConfig{
		Title:       fmt.Sprintf("Add note to #%d", m.noteIssueID),
		Body:        m.noteInput.View(),
		Hint:        "Ctrl+S: Post   Esc: Cancel",
		Width:       m.width,
		Height:      m.height,
		BorderColor: "#98C379",
		TitleColor:  "#FFFFFF",
		BoxWidth:    66,
	})
}

// renderQuickActions renders the combined status/assignee/note popup
func (m Model) renderQuickActions() string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")).Bold(true)
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Status row
	statusName := "(none)"
	if len(m.availableStatuses) > 0 && m.quickStatusIdx < len(m.availableStatuses) {
		statusName = m.availableStatuses[m.quickStatusIdx].Name
	}
	statusVal := "‹ " + statusName + " ›"
	if m.quickField == 0 {
		statusVal = activeStyle.Render(statusVal)
	} else {
		statusVal = valueStyle.Render(statusVal)
	}
	statusLine := labelStyle.Render("Status:   ") + statusVal

	// Assignee row (type-to-filter)
	opts := m.quickFilteredAssignees()
	selName := "(no match)"
	if len(opts) > 0 && m.quickAssigneeSel < len(opts) {
		selName = opts[m.quickAssigneeSel].Name
	}
	filterPart := m.quickAssigneeFilter
	if m.quickField == 1 {
		filterPart += "▌"
	}
	var assigneeVal string
	if filterPart != "" {
		assigneeVal = filterPart + "  → " + selName
	} else {
		assigneeVal = selName
	}
	if m.quickField == 1 {
		assigneeVal = activeStyle.Render(assigneeVal)
	} else {
		assigneeVal = valueStyle.Render(assigneeVal)
	}
	assigneeLine := labelStyle.Render("Assignee: ") + assigneeVal

	// Note row
	noteLabel := "Note:"
	if m.quickField == 2 {
		noteLabel = activeStyle.Render(noteLabel)
	} else {
		noteLabel = labelStyle.Render(noteLabel)
	}

	body := statusLine + "\n" + assigneeLine + "\n\n" + noteLabel + "\n" + m.quickNote.View()

	return appui.RenderInputModal(appui.InputModalConfig{
		Title:       fmt.Sprintf("Quick actions · #%d", m.quickIssueID),
		Body:        body,
		Hint:        "Tab: next field   ←/→: change   type: filter assignee   Ctrl+S: apply   Esc: cancel",
		Width:       m.width,
		Height:      m.height,
		BorderColor: "#C678DD",
		TitleColor:  "#FFFFFF",
		BoxWidth:    66,
	})
}

// renderStatusPicker renders the quick status picker as a centered modal
func (m Model) renderStatusPicker() string {
	var lines []string
	if len(m.availableStatuses) == 0 {
		lines = append(lines, "Loading statuses...")
	}
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
	currentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))
	for i, st := range m.availableStatuses {
		prefix := "  "
		if i == m.statusPickCursor {
			prefix = "→ "
		}
		num := "  "
		if i < 9 {
			num = fmt.Sprintf("%d ", i+1)
		}
		line := prefix + num + st.Name
		if st.ID == m.statusPickCurrentID {
			line += currentStyle.Render("  (current)")
		}
		if i == m.statusPickCursor {
			line = cursorStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return appui.RenderModal(appui.ModalConfig{
		Title:       fmt.Sprintf("Status for #%d", m.statusPickIssueID),
		Content:     lines,
		Width:       m.width,
		Height:      m.height,
		BorderColor: "#FFD700",
		TitleColor:  "#FFFFFF",
	})
}

// renderDescEditor renders the multi-line description editor as a centered modal
func (m Model) renderDescEditor() string {
	title := "Edit Description"
	if m.editFieldIndex < len(editableFields) {
		title = "Edit " + editableFields[m.editFieldIndex].DisplayName
	}
	return appui.RenderInputModal(appui.InputModalConfig{
		Title:       title,
		Body:        m.descInput.View(),
		Hint:        "Ctrl+S: Save   Esc: Cancel   (Enter inserts a new line)",
		Width:       m.width,
		Height:      m.height,
		BorderColor: "#61AFEF",
		TitleColor:  "#FFFFFF",
		BoxWidth:    66,
	})
}

// getFooterItems returns footer menu items with required status
func (m Model) getFooterItems() []appui.FooterItem {
	return []appui.FooterItem{
		{Text: "↑↓/jk: Nav", Required: false},
		{Text: "Tab: Switch", Required: false},
		{Text: "f: Filter", Required: true},
		{Text: "m: My/All", Required: true},
		{Text: "r: Reload", Required: true},
		{Text: "a: Actions", Required: true},
		{Text: "e: Edit", Required: true},
		{Text: "s: Status", Required: true},
		{Text: "c: Note", Required: true},
		{Text: "?: Help", Required: false},
		{Text: "q: Quit", Required: true},
	}
}

// renderListOverlay renders the user/project selection list overlay
