package app

import (
	"fmt"
	"strings"

	"github.com/ktsopanakis/redmine-tui/api"
)

// getFilteredIssues returns issues filtered by current filters
func (m *Model) getFilteredIssues() []api.Issue {
	// First apply multi-user and/or multi-project filters if set
	filteredBySelection := m.issues

	// Build user ID map if user filter is active
	var userIDMap map[string]bool
	if m.assigneeFilter != "" {
		userIDs := strings.Split(m.assigneeFilter, ",")
		userIDMap = make(map[string]bool)
		for _, id := range userIDs {
			userIDMap[id] = true
		}
	}

	// Build project ID map if project filter is active
	var projectIDMap map[string]bool
	if m.projectFilter != "" {
		projectIDs := strings.Split(m.projectFilter, ",")
		projectIDMap = make(map[string]bool)
		for _, id := range projectIDs {
			projectIDMap[id] = true
		}
	}

	// Apply filters with AND logic when both are set
	if userIDMap != nil || projectIDMap != nil {
		var filtered []api.Issue
		for _, issue := range m.issues {
			// Check user filter (OR across selected users)
			userMatch := userIDMap == nil // If no user filter, pass this check
			if !userMatch && issue.AssignedTo != nil {
				userMatch = userIDMap[fmt.Sprintf("%d", issue.AssignedTo.ID)]
			}

			// Check project filter (OR across selected projects)
			projectMatch := projectIDMap == nil // If no project filter, pass this check
			if !projectMatch {
				projectMatch = projectIDMap[fmt.Sprintf("%d", issue.Project.ID)]
			}

			// Include if both filters match (AND logic)
			if userMatch && projectMatch {
				filtered = append(filtered, issue)
			}
		}
		filteredBySelection = filtered
	}

	// Then apply text filter if present
	if m.filterText == "" {
		return filteredBySelection
	}

	filterLower := strings.ToLower(m.filterText)
	filtered := []api.Issue{}
	for _, issue := range filteredBySelection {
		// Search in ID, Subject, Status, Project, and Assignee
		if strings.Contains(strings.ToLower(fmt.Sprintf("%d", issue.ID)), filterLower) ||
			strings.Contains(strings.ToLower(issue.Subject), filterLower) ||
			strings.Contains(strings.ToLower(issue.Status.Name), filterLower) ||
			strings.Contains(strings.ToLower(issue.Project.Name), filterLower) ||
			(issue.AssignedTo != nil && strings.Contains(strings.ToLower(issue.AssignedTo.Name), filterLower)) {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}
