package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	headerHeight = 1
	footerHeight = 1
)

// Lorem ipsum content for testing scrolling
var loremIpsum = `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.

Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.

Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.

Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.

Ut enim ad minima veniam, quis nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut aliquid ex ea commodi consequatur?

Quis autem vel eum iure reprehenderit qui in ea voluptate velit esse quam nihil molestiae consequatur, vel illum qui dolorem eum fugiat quo voluptas nulla pariatur?

At vero eos et accusamus et iusto odio dignissimos ducimus qui blanditiis praesentium voluptatum deleniti atque corrupti quos dolores et quas molestias excepturi sint occaecati cupiditate non provident.

Similique sunt in culpa qui officia deserunt mollitia animi, id est laborum et dolorum fuga. Et harum quidem rerum facilis est et expedita distinctio.

Nam libero tempore, cum soluta nobis est eligendi optio cumque nihil impedit quo minus id quod maxime placeat facere possimus, omnis voluptas assumenda est, omnis dolor repellendus.

Temporibus autem quibusdam et aut officiis debitis aut rerum necessitatibus saepe eveniet ut et voluptates repudiandae sint et molestiae non recusandae.

Itaque earum rerum hic tenetur a sapiente delectus, ut aut reiciendis voluptatibus maiores alias consequatur aut perferendis doloribus asperiores repellat.`

type model struct {
	ready               bool
	width               int
	height              int
	leftPane            viewport.Model
	rightPane           viewport.Model
	activePane          int
	leftTitle           string
	rightTitle          string
	showHelp            bool
	client              *Client
	issues              []Issue
	selectedIndex       int
	selectedDisplayLine int // Line number where selected issue is displayed
	loading             bool
	err                 error
	currentUser         *User
	filterMode          bool
	filterInput         textinput.Model
	filterText          string
	viewMode            string // "my", "all", "user"
	assigneeFilter      string // username or "" for my/all modes
	projectFilter       string // project name or "" for all projects
	userInputMode       string // "", "user", "project" - which input is active
}

func initialModel() model {
	client := NewClient(settings.Redmine.URL, settings.Redmine.APIKey)
	filterInput := textinput.New()
	filterInput.Placeholder = "Type to filter issues..."
	filterInput.CharLimit = 100
	filterInput.Width = 50
	return model{
		leftTitle:     "Issues",
		rightTitle:    "Details",
		activePane:    0,
		client:        client,
		selectedIndex: 0,
		loading:       true,
		filterInput:   filterInput,
		viewMode:      "my",
	}
}

type issuesLoadedMsg struct {
	issues []Issue
	err    error
}

type issueDetailMsg struct {
	issue *Issue
	err   error
}
type currentUserMsg struct {
	user *User
	err  error
}
type tickMsg time.Time

func fetchIssues(client *Client, viewMode string, assigneeFilter string, projectFilter string, issues []Issue) tea.Cmd {
	return func() tea.Msg {
		var resp *IssuesResponse
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
		if assigneeFilter != "" && viewMode == "user" && len(issues) > 0 {
			// Try to find user ID from existing issues
			for _, issue := range issues {
				if issue.AssignedTo != nil && strings.EqualFold(issue.AssignedTo.Name, assigneeFilter) {
					userID = issue.AssignedTo.ID
					break
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
		}

		if err != nil {
			return issuesLoadedMsg{err: err}
		}
		return issuesLoadedMsg{issues: resp.Issues}
	}
}

func fetchIssueDetail(client *Client, issueID int) tea.Cmd {
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

func fetchCurrentUser(client *Client) tea.Cmd {
	return func() tea.Msg {
		user, err := client.GetCurrentUser()
		if err != nil {
			return currentUserMsg{err: err}
		}
		return currentUserMsg{user: user}
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues), fetchCurrentUser(m.client), tickCmd())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case issuesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.issues = msg.issues
		if len(m.issues) > 0 {
			m.selectedIndex = 0
			// Fetch details for first issue
			cmds = append(cmds, fetchIssueDetail(m.client, m.issues[0].ID))
		}
		// Update panes with content if ready
		if m.ready {
			m.updatePaneContent()
		}
		return m, tea.Batch(cmds...)

	case issueDetailMsg:
		if msg.err == nil && msg.issue != nil {
			// Update the issue in the list with full details including journals
			for i, issue := range m.issues {
				if issue.ID == msg.issue.ID {
					m.issues[i] = *msg.issue
					break
				}
			}
			if m.ready {
				m.updatePaneContent()
			}
		}
		return m, nil

	case currentUserMsg:
		if msg.err == nil && msg.user != nil {
			m.currentUser = msg.user
		}
		return m, nil

	case tickMsg:
		// Time update - schedule next tick
		return m, tickCmd()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			paneWidth := (msg.Width / 3) - 4
			leftPaneTotal := paneWidth + 4
			rightPaneWidth := msg.Width - leftPaneTotal - 4
			paneHeight := msg.Height - headerHeight - footerHeight - 2

			// Initialize left pane
			m.leftPane = viewport.New(paneWidth, paneHeight)

			// Initialize right pane
			m.rightPane = viewport.New(rightPaneWidth, paneHeight)

			m.ready = true
			m.updatePaneContent()
		} else {
			paneWidth := (msg.Width / 3) - 4
			leftPaneTotal := paneWidth + 4
			rightPaneWidth := msg.Width - leftPaneTotal - 4
			paneHeight := msg.Height - headerHeight - footerHeight - 2

			m.leftPane.Width = paneWidth
			m.leftPane.Height = paneHeight
			m.rightPane.Width = rightPaneWidth
			m.rightPane.Height = paneHeight
			m.updatePaneContent()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "f":
			if !m.filterMode {
				// Enter filter mode
				m.filterMode = true
				// Keep current filter value in input for editing
				m.filterInput.SetValue(m.filterText)
				m.filterInput.Focus()
				return m, textinput.Blink
			}

		case "esc":
			if m.filterMode {
				// Exit filter mode
				m.filterMode = false
				m.filterInput.Blur()
				// Clear filter
				m.filterText = ""
				m.filterInput.SetValue("")
				m.selectedIndex = 0
				m.updatePaneContent()
				return m, nil
			} else if m.userInputMode != "" {
				// Exit user/project input mode without applying
				m.userInputMode = ""
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				return m, nil
			}

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "m":
			if !m.filterMode && m.userInputMode == "" {
				// Toggle view mode: my -> all -> my
				if m.viewMode == "my" {
					m.viewMode = "all"
				} else {
					m.viewMode = "my"
				}
				m.loading = true
				m.selectedIndex = 0
				return m, fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues)
			}

		case "u":
			if !m.filterMode && m.userInputMode == "" {
				// Enter user selection mode
				m.userInputMode = "user"
				m.filterInput.Placeholder = "Enter username (or leave empty for all)..."
				m.filterInput.SetValue(m.assigneeFilter)
				m.filterInput.Focus()
				return m, textinput.Blink
			}

		case "p":
			if !m.filterMode && m.userInputMode == "" {
				// Enter project selection mode
				m.userInputMode = "project"
				m.filterInput.Placeholder = "Enter project name (or leave empty for all)..."
				m.filterInput.SetValue(m.projectFilter)
				m.filterInput.Focus()
				return m, textinput.Blink
			}

		case "tab":
			if !m.filterMode {
				// Switch between panes
				if m.activePane == 0 {
					m.activePane = 1
				} else {
					m.activePane = 0
				}
			}

		case "enter":
			if m.filterMode {
				// Apply filter and exit filter mode
				m.filterText = m.filterInput.Value()
				m.filterMode = false
				m.filterInput.Blur()
				m.selectedIndex = 0
				m.updatePaneContent()
				return m, nil
			} else if m.userInputMode == "user" {
				// Apply user filter
				m.assigneeFilter = m.filterInput.Value()
				if m.assigneeFilter != "" {
					m.viewMode = "user"
				} else {
					m.viewMode = "all"
				}
				m.userInputMode = ""
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.loading = true
				m.selectedIndex = 0
				return m, fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues)
			} else if m.userInputMode == "project" {
				// Apply project filter
				m.projectFilter = m.filterInput.Value()
				m.userInputMode = ""
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				m.loading = true
				m.selectedIndex = 0
				return m, fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues)
			}

		case "up", "k":
			if m.filterMode {
				// Don't navigate in filter mode
				return m, nil
			}
			if m.activePane == 0 {
				filteredIssues := m.getFilteredIssues()
				if len(filteredIssues) > 0 {
					// Navigate issues list
					if m.selectedIndex > 0 {
						m.selectedIndex--
						m.updatePaneContent()
						// Fetch details for selected issue
						cmds = append(cmds, fetchIssueDetail(m.client, filteredIssues[m.selectedIndex].ID))
					}
				}
			} else {
				if m.activePane == 0 {
					m.leftPane, cmd = m.leftPane.Update(msg)
				} else {
					m.rightPane, cmd = m.rightPane.Update(msg)
				}
				cmds = append(cmds, cmd)
			}

		case "down", "j":
			if m.filterMode {
				// Don't navigate in filter mode
				return m, nil
			}
			if m.activePane == 0 {
				filteredIssues := m.getFilteredIssues()
				if len(filteredIssues) > 0 {
					// Navigate issues list
					if m.selectedIndex < len(filteredIssues)-1 {
						m.selectedIndex++
						m.updatePaneContent()
						// Fetch details for selected issue
						cmds = append(cmds, fetchIssueDetail(m.client, filteredIssues[m.selectedIndex].ID))
					}
				}
			} else {
				if m.activePane == 0 {
					m.leftPane, cmd = m.leftPane.Update(msg)
				} else {
					m.rightPane, cmd = m.rightPane.Update(msg)
				}
				cmds = append(cmds, cmd)
			}

		case "pgup", "b":
			if m.activePane == 0 {
				m.leftPane, cmd = m.leftPane.Update(msg)
			} else {
				m.rightPane, cmd = m.rightPane.Update(msg)
			}
			cmds = append(cmds, cmd)

		case "pgdown":
			if m.activePane == 0 {
				m.leftPane, cmd = m.leftPane.Update(msg)
			} else {
				m.rightPane, cmd = m.rightPane.Update(msg)
			}
			cmds = append(cmds, cmd)

		default:
			// Handle text input in filter mode or user input mode
			if m.filterMode || m.userInputMode != "" {
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case tea.MouseMsg:
		// Handle click to switch panes
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			// Determine which pane was clicked based on X position (1/3 split)
			if msg.X < m.width/3 {
				m.activePane = 0
			} else {
				m.activePane = 1
			}
		}

		// Forward mouse events to viewports for scrolling
		if m.activePane == 0 {
			m.leftPane, cmd = m.leftPane.Update(msg)
		} else {
			m.rightPane, cmd = m.rightPane.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// updatePaneContent updates the viewport content based on current state
func (m *model) getFilteredIssues() []Issue {
	if m.filterText == "" {
		return m.issues
	}

	filterLower := strings.ToLower(m.filterText)
	filtered := []Issue{}
	for _, issue := range m.issues {
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

		// Show active filter at the top if present
		if m.filterText != "" {
			filterStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#61AFEF")).
				Bold(true)
			filterValueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E5C07B"))
			leftContent += filterStyle.Render("Filter: ") + filterValueStyle.Render(m.filterText) + "\n"
			leftContent += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(strings.Repeat("─", m.leftPane.Width)) + "\n\n"
			// Reduce visible lines by 3 to account for filter display
			visibleLines -= 3
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
					Foreground(lipgloss.Color(settings.Colors.ActivePaneBorder)).
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

		// Subject (title)
		rightContent = labelStyle.Render("Subject: ") + titleStyle.Render(issue.Subject) + "\n\n"

		// Static information (color-coded)
		rightContent += labelStyle.Render("Status: ") + statusStyle.Render(issue.Status.Name) + "  "
		rightContent += labelStyle.Render("Priority: ") + statusStyle.Render(issue.Priority.Name) + "\n"

		rightContent += labelStyle.Render("Project: ") + projectStyle.Render(issue.Project.Name) + "  "
		rightContent += labelStyle.Render("Tracker: ") + projectStyle.Render(issue.Tracker.Name) + "\n"

		assignee := "Unassigned"
		if issue.AssignedTo != nil {
			assignee = issue.AssignedTo.Name
		}
		rightContent += labelStyle.Render("Assigned: ") + assigneeStyle.Render(assignee) + "  "
		rightContent += labelStyle.Render("Author: ") + assigneeStyle.Render(issue.Author.Name) + "\n"

		rightContent += labelStyle.Render("Progress: ") + statusStyle.Render(fmt.Sprintf("%d%%", issue.DoneRatio))
		if issue.StartDate != "" {
			rightContent += "  " + labelStyle.Render("Start: ") + issue.StartDate
		}
		if issue.DueDate != "" {
			rightContent += "  " + labelStyle.Render("Due: ") + issue.DueDate
		}
		rightContent += "\n"

		rightContent += labelStyle.Render("Created: ") + issue.CreatedOn.Format("2006-01-02 15:04") + "  "
		rightContent += labelStyle.Render("Updated: ") + issue.UpdatedOn.Format("2006-01-02 15:04") + "\n\n"

		// Description section
		rightContent += sectionStyle.Render("━━━ DESCRIPTION ") + sectionStyle.Render(strings.Repeat("━", m.rightPane.Width-17)) + "\n\n"
		if issue.Description != "" {
			rightContent += issue.Description + "\n"
		} else {
			rightContent += lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("No description provided.") + "\n"
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
