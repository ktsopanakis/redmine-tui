package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type filteredListItem struct {
	originalIndex int
	displayText   string
	isSelected    bool
}

// buildFilteredList creates a filtered list with proper indexing
func (m *model) buildFilteredList() []filteredListItem {
	var items []filteredListItem
	m.filteredIndices = []int{} // Reset filtered indices

	if m.userInputMode == "user" {
		for i, user := range m.availableUsers {
			// Build display name
			displayName := user.Name
			if displayName == "" {
				if user.Firstname != "" || user.Lastname != "" {
					displayName = strings.TrimSpace(user.Firstname + " " + user.Lastname)
				} else if user.Login != "" {
					displayName = user.Login
				} else {
					displayName = fmt.Sprintf("User %d", user.ID)
				}
			}

			// Apply filter - but always show selected items
			isSelected := m.selectedUsers[user.ID]
			if m.listFilterText != "" && !isSelected {
				filterLower := strings.ToLower(m.listFilterText)
				nameLower := strings.ToLower(displayName)
				loginLower := strings.ToLower(user.Login)
				if !strings.Contains(nameLower, filterLower) && !strings.Contains(loginLower, filterLower) {
					continue
				}
			}

			text := displayName
			if user.Login != "" && displayName != user.Login {
				text += fmt.Sprintf(" (%s)", user.Login)
			}

			items = append(items, filteredListItem{
				originalIndex: i,
				displayText:   text,
				isSelected:    isSelected,
			})
			m.filteredIndices = append(m.filteredIndices, i)
		}
	} else if m.userInputMode == "project" {
		for i, project := range m.availableProjects {
			// Apply filter - but always show selected items
			isSelected := m.selectedProjects[project.ID]
			if m.listFilterText != "" && !isSelected {
				filterLower := strings.ToLower(m.listFilterText)
				nameLower := strings.ToLower(project.Name)
				if !strings.Contains(nameLower, filterLower) {
					continue
				}
			}

			items = append(items, filteredListItem{
				originalIndex: i,
				displayText:   project.Name,
				isSelected:    isSelected,
			})
			m.filteredIndices = append(m.filteredIndices, i)
		}
	}

	// Ensure cursor is within bounds
	if m.listCursor >= len(items) {
		m.listCursor = len(items) - 1
	}
	if m.listCursor < 0 && len(items) > 0 {
		m.listCursor = 0
	}

	return items
}

// renderListOverlay renders the user/project selection list overlay
func (m model) renderListOverlay() string {
	var title string
	var items []filteredListItem

	// Create a mutable copy for building the filtered list
	mutableModel := m
	items = mutableModel.buildFilteredList()

	if m.userInputMode == "user" {
		title = "Select Users (↑/↓: Navigate, Space: Toggle, Enter: Apply, Esc: Cancel)"
		if m.listLoading {
			loadingBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#61AFEF")).
				Padding(1, 2).
				Width(m.width - 6).
				Render(title + "\n\nLoading users...")
			return m.positionAtBottom(loadingBox)
		}

		if len(m.availableUsers) == 0 {
			emptyBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#61AFEF")).
				Padding(1, 2).
				Width(m.width - 6).
				Render(title + "\n\nNo users found")
			return m.positionAtBottom(emptyBox)
		}
	} else if m.userInputMode == "project" {
		title = "Select Projects (↑/↓: Navigate, Space: Toggle, Enter: Apply, Esc: Cancel)"
		if m.listLoading {
			loadingBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#98C379")).
				Padding(1, 2).
				Width(m.width - 6).
				Render(title + "\n\nLoading projects...")
			return m.positionAtBottom(loadingBox)
		}

		if len(m.availableProjects) == 0 {
			emptyBox := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#98C379")).
				Padding(1, 2).
				Width(m.width - 6).
				Render(title + "\n\nNo projects found")
			return m.positionAtBottom(emptyBox)
		}
	}

	// Build list content
	var content strings.Builder
	content.WriteString(title + "\n")
	content.WriteString(strings.Repeat("─", m.width-8) + "\n\n")

	// Calculate visible items
	maxVisibleItems := 12
	if maxVisibleItems > len(items) {
		maxVisibleItems = len(items)
	}

	// Show items with scrolling if needed
	startIdx := 0
	endIdx := len(items)

	if len(items) > maxVisibleItems {
		// Calculate visible window centered on cursor
		startIdx = m.listCursor - maxVisibleItems/2
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx = startIdx + maxVisibleItems
		if endIdx > len(items) {
			endIdx = len(items)
			startIdx = endIdx - maxVisibleItems
			if startIdx < 0 {
				startIdx = 0
			}
		}
	}

	for i := startIdx; i < endIdx; i++ {
		checkbox := "[ ]"
		if items[i].isSelected {
			checkbox = "[✓]"
		}
		cursor := "  "
		if i == m.listCursor {
			cursor = "→ "
		}
		content.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, items[i].displayText))
	}

	// Show position indicator if scrolling
	if len(items) > maxVisibleItems {
		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Showing %d-%d of %d", startIdx+1, endIdx, len(items)))
	} else if len(items) == 0 {
		content.WriteString("No matching items")
	}

	borderColor := lipgloss.Color("#61AFEF")
	if m.userInputMode == "project" {
		borderColor = lipgloss.Color("#98C379")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(m.width - 6).
		Render(content.String())

	return m.positionAtBottom(box)
}

// positionAtBottom positions a box at the bottom of the screen
func (m model) positionAtBottom(box string) string {
	boxHeight := lipgloss.Height(box)
	boxWidth := lipgloss.Width(box)
	horizontalPadding := (m.width - boxWidth) / 2

	if horizontalPadding < 0 {
		horizontalPadding = 0
	}

	// Create a box that fills the pane space with the list at bottom
	emptySpace := m.height - headerHeight - footerHeight - 2
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
