package app

import (
	"fmt"
	"strings"

	appui "github.com/ktsopanakis/redmine-tui/ui"
)

// buildUserListItems converts available users to ListItems
func (m *Model) buildUserListItems() []appui.ListItem {
	items := make([]appui.ListItem, len(m.availableUsers))
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

		// Add login if different from name
		text := displayName
		if user.Login != "" && displayName != user.Login {
			text += fmt.Sprintf(" (%s)", user.Login)
		}

		items[i] = appui.ListItem{
			ID:          user.ID,
			DisplayText: text,
			IsSelected:  m.selectedUsers[user.ID],
		}
	}
	return items
}

// buildProjectListItems converts available projects to ListItems
func (m *Model) buildProjectListItems() []appui.ListItem {
	items := make([]appui.ListItem, len(m.availableProjects))
	for i, project := range m.availableProjects {
		items[i] = appui.ListItem{
			ID:          project.ID,
			DisplayText: project.Name,
			IsSelected:  m.selectedProjects[project.ID],
		}
	}
	return items
}

// updateFilteredIndices updates the filteredIndices based on the filtered list result
func (m *Model) updateFilteredIndices(items []appui.ListItem) {
	m.filteredIndices = []int{}
	
	if m.userInputMode == "user" {
		// Map filtered items back to original indices
		for _, item := range items {
			for i, user := range m.availableUsers {
				if user.ID == item.ID {
					m.filteredIndices = append(m.filteredIndices, i)
					break
				}
			}
		}
	} else if m.userInputMode == "project" {
		// Map filtered items back to original indices
		for _, item := range items {
			for i, project := range m.availableProjects {
				if project.ID == item.ID {
					m.filteredIndices = append(m.filteredIndices, i)
					break
				}
			}
		}
	}
}

// customFilterFunc provides custom filtering logic that checks multiple fields
func customUserFilterFunc(item appui.ListItem, filter string) bool {
	filterLower := strings.ToLower(filter)
	textLower := strings.ToLower(item.DisplayText)
	return strings.Contains(textLower, filterLower)
}

// renderListOverlay renders the user/project selection list overlay
func (m Model) renderListOverlay() string {
	var title, borderColor, loadingMsg, emptyMsg string
	var items []appui.ListItem

	if m.userInputMode == "user" {
		title = "Select Users (↑/↓: Navigate, Space: Toggle, Enter: Apply, Esc: Cancel)"
		borderColor = "#61AFEF"
		loadingMsg = "Loading users..."
		emptyMsg = "No users found"
		
		// Create mutable copy to build list
		mutableModel := m
		items = mutableModel.buildUserListItems()
	} else if m.userInputMode == "project" {
		title = "Select Projects (↑/↓: Navigate, Space: Toggle, Enter: Apply, Esc: Cancel)"
		borderColor = "#98C379"
		loadingMsg = "Loading projects..."
		emptyMsg = "No projects found"
		
		// Create mutable copy to build list
		mutableModel := m
		items = mutableModel.buildProjectListItems()
	}

	cfg := appui.ListConfig{
		Title:           title,
		Items:           items,
		Cursor:          m.listCursor,
		FilterText:      m.listFilterText,
		BorderColor:     borderColor,
		Width:           m.width,
		Height:          m.height,
		MaxVisibleItems: 12,
		IsLoading:       m.listLoading,
		LoadingMessage:  loadingMsg,
		EmptyMessage:    emptyMsg,
		ShowScrollInfo:  true,
		FilterFunc:      customUserFilterFunc,
	}

	return appui.RenderListOverlay(cfg, headerHeight, footerHeight)
}

// buildFilteredList is kept for updating cursor and indices
func (m *Model) buildFilteredList() {
	var items []appui.ListItem
	
	if m.userInputMode == "user" {
		items = m.buildUserListItems()
	} else if m.userInputMode == "project" {
		items = m.buildProjectListItems()
	}

	cfg := appui.ListConfig{
		Items:      items,
		Cursor:     m.listCursor,
		FilterText: m.listFilterText,
		FilterFunc: customUserFilterFunc,
	}

	result := appui.BuildFilteredList(cfg)
	
	// Update cursor
	m.listCursor = result.UpdatedCursor
	
	// Update filtered indices
	m.updateFilteredIndices(result.Items)
}
