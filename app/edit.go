package app

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ktsopanakis/redmine-tui/api"
)

// EditableField represents a field that can be edited
type EditableField struct {
	Name        string
	DisplayName string
	Type        string // "text", "number", "select", "date", "multiline"
	GetValue    func(*api.Issue) string
	GetOptions  func(*Model) []string // for select fields
}

// Define editable fields
var editableFields = []EditableField{
	{
		Name:        "subject",
		DisplayName: "Subject",
		Type:        "text",
		GetValue:    func(i *api.Issue) string { return i.Subject },
	},
	{
		Name:        "description",
		DisplayName: "Description",
		Type:        "multiline",
		GetValue:    func(i *api.Issue) string { return i.Description },
	},
	{
		Name:        "status_id",
		DisplayName: "Status",
		Type:        "select",
		GetValue:    func(i *api.Issue) string { return i.Status.Name },
		GetOptions: func(m *Model) []string {
			options := []string{}
			for _, s := range m.availableStatuses {
				options = append(options, s.Name)
			}
			return options
		},
	},
	{
		Name:        "priority_id",
		DisplayName: "Priority",
		Type:        "select",
		GetValue:    func(i *api.Issue) string { return i.Priority.Name },
		GetOptions: func(m *Model) []string {
			options := []string{}
			for _, p := range m.availablePriorities {
				options = append(options, p.Name)
			}
			return options
		},
	},
	{
		Name:        "assigned_to_id",
		DisplayName: "Assigned To",
		Type:        "select",
		GetValue: func(i *api.Issue) string {
			if i.AssignedTo != nil {
				return i.AssignedTo.Name
			}
			return "Unassigned"
		},
		GetOptions: func(m *Model) []string {
			options := []string{"Unassigned"}
			for _, u := range m.availableUsers {
				displayName := u.Name
				if displayName == "" {
					if u.Firstname != "" || u.Lastname != "" {
						displayName = strings.TrimSpace(u.Firstname + " " + u.Lastname)
					} else if u.Login != "" {
						displayName = u.Login
					}
				}
				options = append(options, displayName)
			}
			return options
		},
	},
	{
		Name:        "done_ratio",
		DisplayName: "Progress",
		Type:        "number",
		GetValue:    func(i *api.Issue) string { return fmt.Sprintf("%d", i.DoneRatio) },
	},
	{
		Name:        "due_date",
		DisplayName: "Due Date",
		Type:        "date",
		GetValue:    func(i *api.Issue) string { return i.DueDate },
	},
}

// Message types for edit operations
type statusesLoadedMsg struct {
	statuses []api.Status
	err      error
}

type prioritiesLoadedMsg struct {
	priorities []api.Priority
	err        error
}

type issueUpdatedMsg struct {
	issueID int
	err     error
}

// Commands for edit operations
func fetchStatuses(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		statuses, err := client.GetStatuses()
		return statusesLoadedMsg{statuses: statuses, err: err}
	}
}

func fetchPriorities(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		priorities, err := client.GetPriorities()
		return prioritiesLoadedMsg{priorities: priorities, err: err}
	}
}

func updateIssue(client *api.Client, issueID int, field EditableField, value string, m *Model) tea.Cmd {
	return func() tea.Msg {
		updates := make(map[string]interface{})

		switch field.Name {
		case "subject":
			updates["subject"] = value
		case "description":
			updates["description"] = value
		case "status_id":
			// Find status ID by name
			for _, s := range m.availableStatuses {
				if s.Name == value {
					updates["status_id"] = s.ID
					break
				}
			}
		case "priority_id":
			// Find priority ID by name
			for _, p := range m.availablePriorities {
				if p.Name == value {
					updates["priority_id"] = p.ID
					break
				}
			}
		case "assigned_to_id":
			if value == "Unassigned" {
				updates["assigned_to_id"] = nil
			} else {
				// Find user ID by name
				for _, u := range m.availableUsers {
					displayName := u.Name
					if displayName == "" {
						if u.Firstname != "" || u.Lastname != "" {
							displayName = strings.TrimSpace(u.Firstname + " " + u.Lastname)
						} else if u.Login != "" {
							displayName = u.Login
						}
					}
					if displayName == value {
						updates["assigned_to_id"] = u.ID
						break
					}
				}
			}
		case "done_ratio":
			ratio, err := strconv.Atoi(value)
			if err == nil && ratio >= 0 && ratio <= 100 {
				updates["done_ratio"] = ratio
			}
		case "due_date":
			if value != "" {
				updates["due_date"] = value
			} else {
				updates["due_date"] = nil
			}
		}

		err := client.UpdateIssue(issueID, updates)
		return issueUpdatedMsg{issueID: issueID, err: err}
	}
}

// updateIssueMultiple sends all pending edits to the API in one request
func updateIssueMultiple(client *api.Client, issueID int, pendingEdits map[string]string, m Model) tea.Cmd {
	return func() tea.Msg {
		updates := make(map[string]interface{})

		for fieldName, value := range pendingEdits {
			switch fieldName {
			case "subject":
				if value != "" {
					updates["subject"] = value
				}
			case "description":
				updates["description"] = value
			case "status_id":
				// Find status ID by name
				for _, s := range m.availableStatuses {
					if s.Name == value {
						updates["status_id"] = s.ID
						break
					}
				}
			case "priority_id":
				// Find priority ID by name
				for _, p := range m.availablePriorities {
					if p.Name == value {
						updates["priority_id"] = p.ID
						break
					}
				}
			case "assigned_to_id":
				if value == "Unassigned" {
					updates["assigned_to_id"] = nil
				} else {
					// Find user ID by name
					for _, u := range m.availableUsers {
						displayName := u.Name
						if displayName == "" {
							if u.Firstname != "" || u.Lastname != "" {
								displayName = strings.TrimSpace(u.Firstname + " " + u.Lastname)
							} else if u.Login != "" {
								displayName = u.Login
							}
						}
						if displayName == value {
							updates["assigned_to_id"] = u.ID
							break
						}
					}
				}
			case "done_ratio":
				ratio, err := strconv.Atoi(value)
				if err == nil && ratio >= 0 && ratio <= 100 {
					updates["done_ratio"] = ratio
				}
			case "due_date":
				if value != "" {
					updates["due_date"] = value
				} else {
					updates["due_date"] = nil
				}
			}
		}

		err := client.UpdateIssue(issueID, updates)
		return issueUpdatedMsg{issueID: issueID, err: err}
	}
}

// renderEditFooter renders the footer when in edit mode
func (m Model) renderEditFooter() string {
	if !m.editMode || len(editableFields) == 0 {
		return ""
	}

	field := editableFields[m.editFieldIndex]
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#61AFEF")).
		Bold(true)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#98C379"))

	var footer string

	// Show navigation hint with Ctrl+S to save
	footer += style.Render("EDIT MODE") + " "
	unsavedIndicator := ""
	if m.hasUnsavedChanges {
		unsavedIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")).Render(" [UNSAVED] ")
	}
	footer += unsavedIndicator
	footer += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("Tab/Enter: Next | ↑↓: Select | Ctrl+S: Save | Esc: Cancel") + "\n"

	// Show current field being edited
	footer += promptStyle.Render(fmt.Sprintf("Editing %s: ", field.DisplayName))

	// Show input or selection options
	if field.Type == "select" {
		options := field.GetOptions(&m)
		footer += lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B")).Render(m.editInput.Value()) + " "
		footer += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(
			fmt.Sprintf("[%d options]", len(options)),
		)
	} else {
		footer += m.editInput.View()
	}

	return footer
}

// getFieldHighlightStyle returns style for highlighting the selected field
func getFieldHighlightStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#61AFEF")).
		Bold(true)
}
