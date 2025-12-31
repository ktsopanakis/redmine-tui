package app

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ktsopanakis/redmine-tui/api"
)

// Message types for Bubble Tea update loop

type issuesLoadedMsg struct {
	issues []api.Issue
	err    error
}

type issueDetailMsg struct {
	issue *api.Issue
	err   error
}

type currentUserMsg struct {
	user *api.User
	err  error
}

type usersLoadedMsg struct {
	users []api.User
	err   error
}

type projectsLoadedMsg struct {
	projects []api.Project
	err      error
}

type tickMsg time.Time

// Fetch commands that return messages

func fetchIssues(client *api.Client, viewMode string, assigneeFilter string, projectFilter string, issues []api.Issue) tea.Cmd {
	return func() tea.Msg {
		var resp *api.IssuesResponse
		var err error

		// Determine project ID if projectFilter is set
		projectID := 0
		if projectFilter != "" && len(issues) > 0 {
			// Try to find project ID from existing issues
			for _, issue := range issues {
				if strings.EqualFold(issue.Project.Name, projectFilter) {
					projectID = issue.Project.ID
					break
				}
			}
		}

		// Determine user ID if assigneeFilter is set
		userID := 0
		if assigneeFilter != "" && viewMode == "user" {
			// Try to parse as number first (from list selection)
			var parseErr error
			userID, parseErr = strconv.Atoi(assigneeFilter)
			if parseErr != nil || userID == 0 {
				// Fall back to searching by name in existing issues
				for _, issue := range issues {
					if issue.AssignedTo != nil && strings.EqualFold(issue.AssignedTo.Name, assigneeFilter) {
						userID = issue.AssignedTo.ID
						break
					}
				}
			}
		}

		switch viewMode {
		case "my":
			// Fetch issues assigned to me
			resp, err = client.GetIssues(projectID, true, 0, true, 100, 0)
		case "all":
			// Fetch all open issues
			resp, err = client.GetIssues(projectID, false, 0, true, 100, 0)
		case "user":
			// Fetch issues for specific user
			resp, err = client.GetIssues(projectID, false, userID, true, 100, 0)
		case "user-multi":
			// Fetch all issues for client-side filtering by multiple users
			resp, err = client.GetIssues(projectID, false, 0, true, 100, 0)
		case "project-multi":
			// Fetch all issues for client-side filtering by multiple projects
			resp, err = client.GetIssues(0, false, 0, true, 100, 0)
		case "user-project-multi":
			// Fetch all issues for client-side filtering by both users and projects
			resp, err = client.GetIssues(0, false, 0, true, 100, 0)
		default:
			// Default to all issues
			resp, err = client.GetIssues(projectID, false, 0, true, 100, 0)
		}

		if err != nil {
			return issuesLoadedMsg{err: err}
		}
		return issuesLoadedMsg{issues: resp.Issues}
	}
}

func fetchIssueDetail(client *api.Client, issueID int) tea.Cmd {
	return func() tea.Msg {
		issue, err := client.GetIssue(issueID)
		if err != nil {
			return issueDetailMsg{err: err}
		}
		return issueDetailMsg{issue: issue}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchCurrentUser(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		user, err := client.GetCurrentUser()
		if err != nil {
			return currentUserMsg{err: err}
		}
		return currentUserMsg{user: user}
	}
}

func fetchUsers(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		users, err := client.GetUsers(100, 0)
		if err != nil {
			return usersLoadedMsg{err: err}
		}
		return usersLoadedMsg{users: users}
	}
}

func fetchProjects(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetProjects(100, 0)
		if err != nil {
			return projectsLoadedMsg{err: err}
		}
		return projectsLoadedMsg{projects: resp.Projects}
	}
}
