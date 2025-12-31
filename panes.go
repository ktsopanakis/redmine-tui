package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ktsopanakis/redmine-tui/config"
)

func (m *model) updatePaneContent() {
	if !m.ready {
		return
	}

	// Left pane: List of issues with smart roller-style navigation
	var leftContent string
	filteredIssues := m.getFilteredIssues()

	if m.loading {
		leftContent = "Loading issues..."
	} else if m.err != nil {
		leftContent = fmt.Sprintf("Error: %v", m.err)
	} else if len(filteredIssues) == 0 {
		if m.filterText != "" {
			leftContent = "No matching issues found."
		} else {
			leftContent = "No issues found."
		}
	} else {
		// Each issue takes 3 lines (ID+Title, Status+Project, Assignee) + 1 blank line = 4 total
		linesPerIssue := 4
		visibleLines := m.leftPane.Height
		// Calculate how many complete issues we can fit (excluding the trailing blank line for the last issue)
		visibleIssues := (visibleLines + 1) / linesPerIssue

		// Calculate start index based on position in list
		var startIdx int
		if m.selectedIndex < visibleIssues/2 {
			// Near the start - selection at top
			startIdx = 0
		} else if m.selectedIndex >= len(filteredIssues)-(visibleIssues/2) {
			// Near the end - selection at bottom
			startIdx = len(filteredIssues) - visibleIssues
			if startIdx < 0 {
				startIdx = 0
			}
		} else {
			// In the middle - keep selection centered
			startIdx = m.selectedIndex - (visibleIssues / 2)
		}

		endIdx := startIdx + visibleIssues
		if endIdx > len(filteredIssues) {
			endIdx = len(filteredIssues)
		}

		// Show active filters at the top if present
		filterStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#61AFEF")).
			Bold(true)
		filterValueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B"))

		filterLinesAdded := 0

		// Show selected users
		if m.assigneeFilter != "" {
			userNames := []string{}
			userIDs := strings.Split(m.assigneeFilter, ",")
			for _, idStr := range userIDs {
				var id int
				if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
					if name, ok := m.selectedUserNames[id]; ok {
						userNames = append(userNames, name)
					}
				}
			}
			if len(userNames) > 0 {
				leftContent += filterStyle.Render("Users: ") + filterValueStyle.Render(strings.Join(userNames, ", ")) + "\n"
				filterLinesAdded++
			}
		}

		// Show selected projects
		if m.projectFilter != "" {
			projectNames := []string{}
			projectIDs := strings.Split(m.projectFilter, ",")
			for _, idStr := range projectIDs {
				var id int
				if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil {
					if name, ok := m.selectedProjectNames[id]; ok {
						projectNames = append(projectNames, name)
					}
				}
			}
			if len(projectNames) > 0 {
				leftContent += filterStyle.Render("Projects: ") + filterValueStyle.Render(strings.Join(projectNames, ", ")) + "\n"
				filterLinesAdded++
			}
		}

		// Show text filter
		if m.filterText != "" {
			leftContent += filterStyle.Render("Filter: ") + filterValueStyle.Render(m.filterText) + "\n"
			filterLinesAdded++
		}

		// Add separator if any filters are active
		if filterLinesAdded > 0 {
			leftContent += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(strings.Repeat("─", m.leftPane.Width)) + "\n\n"
			filterLinesAdded += 2 // separator and blank line
			// Reduce visible lines to account for filter display
			visibleLines -= filterLinesAdded
			visibleIssues = (visibleLines + 1) / linesPerIssue

			// Recalculate indices with reduced space
			if m.selectedIndex < visibleIssues/2 {
				startIdx = 0
			} else if m.selectedIndex >= len(filteredIssues)-(visibleIssues/2) {
				startIdx = len(filteredIssues) - visibleIssues
				if startIdx < 0 {
					startIdx = 0
				}
			} else {
				startIdx = m.selectedIndex - (visibleIssues / 2)
			}
			endIdx = startIdx + visibleIssues
			if endIdx > len(filteredIssues) {
				endIdx = len(filteredIssues)
			}
		}

		// Build the content
		for i := startIdx; i < endIdx; i++ {
			issue := filteredIssues[i]
			isSelected := i == m.selectedIndex

			// Styles for different components with vibrant colors
			idStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF"))       // Cyan
			titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))    // White
			statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))   // Gold
			projectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))  // Green
			assigneeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD")) // Purple

			var linePrefix string
			spacerStyle := lipgloss.NewStyle() // For spaces/dots between elements
			if isSelected {
				// Subtle background tint + bold for selection
				subtleBg := lipgloss.Color("#2A2A3A") // Dark subtle background
				idStyle = idStyle.Background(subtleBg).Bold(true)
				titleStyle = titleStyle.Background(subtleBg).Bold(true)
				statusStyle = statusStyle.Background(subtleBg).Bold(true)
				projectStyle = projectStyle.Background(subtleBg).Bold(true)
				assigneeStyle = assigneeStyle.Background(subtleBg).Bold(true)
				spacerStyle = spacerStyle.Background(subtleBg) // Apply background to spacers too
				linePrefix = lipgloss.NewStyle().
					Foreground(lipgloss.Color(config.Current.Colors.ActivePaneBorder)).
					Background(subtleBg).
					Render("▌")
			} else {
				linePrefix = " "
			}

			// Line 1: ID and Subject
			line1 := linePrefix + idStyle.Render(fmt.Sprintf("#%d", issue.ID)) + spacerStyle.Render(" ") + titleStyle.Render(issue.Subject)
			if isSelected {
				// Pad to full width for complete background
				availableWidth := m.leftPane.Width
				currentLen := len(fmt.Sprintf("#%d %s", issue.ID, issue.Subject)) + 1
				if currentLen < availableWidth {
					line1 += spacerStyle.Render(strings.Repeat(" ", availableWidth-currentLen))
				}
			}
			leftContent += line1 + "\n"

			// Line 2: Status and Project
			line2 := linePrefix + statusStyle.Render(issue.Status.Name) + spacerStyle.Render(" • ") + projectStyle.Render(issue.Project.Name)
			if isSelected {
				availableWidth := m.leftPane.Width
				currentLen := len(issue.Status.Name) + 3 + len(issue.Project.Name) + 1
				if currentLen < availableWidth {
					line2 += spacerStyle.Render(strings.Repeat(" ", availableWidth-currentLen))
				}
			}
			leftContent += line2 + "\n"

			// Line 3: Assignee
			assignee := "Unassigned"
			if issue.AssignedTo != nil {
				assignee = issue.AssignedTo.Name
			}
			line3 := linePrefix + assigneeStyle.Render("→ "+assignee)
			if isSelected {
				availableWidth := m.leftPane.Width
				currentLen := len("→ "+assignee) - 1
				if currentLen < availableWidth {
					line3 += spacerStyle.Render(strings.Repeat(" ", availableWidth-currentLen))
				}
			}
			leftContent += line3 + "\n"

			// Blank line between issues
			leftContent += "\n"
		}
	}

	// Store which issue is at which line for border arrow placement
	if len(m.issues) > 0 {
		linesPerIssue := 4
		visibleLines := m.leftPane.Height
		visibleIssues := visibleLines / linesPerIssue

		var startIdx int
		if m.selectedIndex < visibleIssues/2 {
			startIdx = 0
		} else if m.selectedIndex >= len(m.issues)-(visibleIssues/2) {
			startIdx = len(m.issues) - visibleIssues
			if startIdx < 0 {
				startIdx = 0
			}
		} else {
			startIdx = m.selectedIndex - (visibleIssues / 2)
		}

		// Calculate the line number where selected issue appears (0-indexed)
		m.selectedDisplayLine = (m.selectedIndex - startIdx) * linesPerIssue
	}

	m.leftPane.SetContent(leftContent)

	// Update left title with filter info and view mode
	viewModeText := ""
	switch m.viewMode {
	case "my":
		viewModeText = "My Issues"
	case "all":
		viewModeText = "All Issues"
	case "user":
		if m.assigneeFilter != "" {
			viewModeText = m.assigneeFilter
		} else {
			viewModeText = "All Issues"
		}
	default:
		viewModeText = "Issues"
	}

	// Add project filter to title if set
	if m.projectFilter != "" {
		viewModeText = m.projectFilter + ": " + viewModeText
	}

	if m.filterText != "" {
		m.leftTitle = fmt.Sprintf("%s (%d/%d)", viewModeText, len(filteredIssues), len(m.issues))
	} else {
		m.leftTitle = viewModeText
	}

	// Right pane: Selected issue details
	var rightContent string
	if m.loading {
		rightContent = "Loading..."
		m.rightTitle = "Details"
	} else if m.err != nil {
		rightContent = fmt.Sprintf("Error: %v", m.err)
		m.rightTitle = "Details"
	} else if len(filteredIssues) == 0 {
		rightContent = "No issue selected."
		m.rightTitle = "Details"
	} else if m.selectedIndex >= 0 && m.selectedIndex < len(filteredIssues) {
		issue := filteredIssues[m.selectedIndex]

		// Update title to show issue ID (plain text, styling happens in border)
		m.rightTitle = fmt.Sprintf("#%d", issue.ID)

		// Color styles matching the left pane
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)   // Gold
		projectStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379")).Bold(true)  // Green
		assigneeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD")).Bold(true) // Purple
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF"))               // Blue
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)    // White, bold
		sectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")).Bold(true)  // Yellow

		// Highlight style for edit mode
		highlightStyle := getFieldHighlightStyle()
		currentField := ""
		editedValue := ""
		if m.editMode && m.editFieldIndex < len(editableFields) {
			currentField = editableFields[m.editFieldIndex].Name
			editedValue = m.editInput.Value()
		}

		// Helper function to get display value (edited or original)
		getDisplayValue := func(fieldName string, originalValue string) string {
			// Check pending edits first
			if pendingValue, exists := m.pendingEdits[fieldName]; exists {
				return pendingValue
			}
			// Then check if it's the currently edited field
			if currentField == fieldName && editedValue != "" {
				return editedValue
			}
			return originalValue
		}

		// Display pending edits summary at the top if any exist
		if m.editMode && len(m.pendingEdits) > 0 {
			pendingStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")).Bold(true) // Yellow
			oldValueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75"))           // Red
			newValueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))           // Green
			arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))              // Gray

			rightContent = pendingStyle.Render("PENDING CHANGES:") + "\n"
			for _, field := range editableFields {
				if newValue, exists := m.pendingEdits[field.Name]; exists {
					oldValue := m.originalValues[field.Name]
					rightContent += "  " + labelStyle.Render(field.DisplayName+":") + " " +
						oldValueStyle.Render(oldValue) + " " +
						arrowStyle.Render("→") + " " +
						newValueStyle.Render(newValue) + "\n"
				}
			}
			rightContent += "\n"
		}

		// Subject (title) - field 0
		subjectValue := getDisplayValue("subject", issue.Subject)
		if currentField == "subject" {
			rightContent = labelStyle.Render("Subject: ") + highlightStyle.Render(subjectValue) + "\n\n"
		} else {
			rightContent = labelStyle.Render("Subject: ") + titleStyle.Render(subjectValue) + "\n\n"
		}

		// Static information (color-coded)
		// Status - field 2
		statusValue := getDisplayValue("status_id", issue.Status.Name)
		if currentField == "status_id" {
			rightContent += labelStyle.Render("Status: ") + highlightStyle.Render(statusValue) + "  "
		} else {
			rightContent += labelStyle.Render("Status: ") + statusStyle.Render(statusValue) + "  "
		}

		// Priority - field 3
		priorityValue := getDisplayValue("priority_id", issue.Priority.Name)
		if currentField == "priority_id" {
			rightContent += labelStyle.Render("Priority: ") + highlightStyle.Render(priorityValue) + "\n"
		} else {
			rightContent += labelStyle.Render("Priority: ") + statusStyle.Render(priorityValue) + "\n"
		}

		rightContent += labelStyle.Render("Project: ") + projectStyle.Render(issue.Project.Name) + "  "
		rightContent += labelStyle.Render("Tracker: ") + projectStyle.Render(issue.Tracker.Name) + "\n"

		assignee := "Unassigned"
		if issue.AssignedTo != nil {
			assignee = issue.AssignedTo.Name
		}
		// Assigned - field 4
		assigneeValue := getDisplayValue("assigned_to_id", assignee)
		if currentField == "assigned_to_id" {
			rightContent += labelStyle.Render("Assigned: ") + highlightStyle.Render(assigneeValue) + "  "
		} else {
			rightContent += labelStyle.Render("Assigned: ") + assigneeStyle.Render(assigneeValue) + "  "
		}
		rightContent += labelStyle.Render("Author: ") + assigneeStyle.Render(issue.Author.Name) + "\n"

		// Progress - field 5
		progressText := fmt.Sprintf("%d%%", issue.DoneRatio)
		progressValue := getDisplayValue("done_ratio", progressText)
		if currentField == "done_ratio" {
			rightContent += labelStyle.Render("Progress: ") + highlightStyle.Render(progressValue)
		} else {
			rightContent += labelStyle.Render("Progress: ") + statusStyle.Render(progressValue)
		}
		if issue.StartDate != "" {
			rightContent += "  " + labelStyle.Render("Start: ") + issue.StartDate
		}
		// Due Date - field 6
		dueValue := getDisplayValue("due_date", issue.DueDate)
		if issue.DueDate != "" || currentField == "due_date" {
			if currentField == "due_date" {
				rightContent += "  " + labelStyle.Render("Due: ") + highlightStyle.Render(dueValue)
			} else {
				rightContent += "  " + labelStyle.Render("Due: ") + dueValue
			}
		}
		rightContent += "\n"

		rightContent += labelStyle.Render("Created: ") + issue.CreatedOn.Format("2006-01-02 15:04") + "  "
		rightContent += labelStyle.Render("Updated: ") + issue.UpdatedOn.Format("2006-01-02 15:04") + "\n\n"

		// Description section - field 1
		rightContent += sectionStyle.Render("━━━ DESCRIPTION ") + sectionStyle.Render(strings.Repeat("━", m.rightPane.Width-17)) + "\n\n"
		descValue := getDisplayValue("description", issue.Description)
		if descValue != "" {
			if currentField == "description" {
				rightContent += highlightStyle.Render(descValue) + "\n"
			} else {
				rightContent += descValue + "\n"
			}
		} else {
			if currentField == "description" {
				rightContent += highlightStyle.Render("No description provided.") + "\n"
			} else {
				rightContent += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("No description provided.") + "\n"
			}
		}

		// History and notes section
		rightContent += "\n" + sectionStyle.Render("━━━ HISTORY & NOTES ") + sectionStyle.Render(strings.Repeat("━", m.rightPane.Width-21)) + "\n\n"

		if len(issue.Journals) > 0 {
			for _, journal := range issue.Journals {
				userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#C678DD")).Bold(true)
				dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

				rightContent += userStyle.Render(journal.User.Name) + " " + dateStyle.Render(journal.CreatedOn.Format("2006-01-02 15:04")) + "\n"

				// Show property changes
				if len(journal.Details) > 0 {
					for _, detail := range journal.Details {
						fieldStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#61AFEF")).Bold(true) // Cyan for field name
						oldValueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B"))         // Yellow/orange for old value
						newValueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#98C379"))         // Green for new value
						arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))            // Gray arrow

						if detail.OldValue != "" && detail.NewValue != "" {
							rightContent += "  " + fieldStyle.Render(detail.Name+":") + " " +
								oldValueStyle.Render(detail.OldValue) + " " +
								arrowStyle.Render("→") + " " +
								newValueStyle.Render(detail.NewValue) + "\n"
						} else if detail.NewValue != "" {
							rightContent += "  " + fieldStyle.Render(detail.Name+":") + " " +
								newValueStyle.Render(detail.NewValue) + "\n"
						}
					}
				}

				// Show notes/comments
				if journal.Notes != "" {
					rightContent += "  " + journal.Notes + "\n"
				}

				rightContent += "\n"
			}
		} else {
			rightContent += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("No history available.") + "\n"
		}
	}
	m.rightPane.SetContent(lipgloss.NewStyle().Width(m.rightPane.Width).Render(rightContent))
}
