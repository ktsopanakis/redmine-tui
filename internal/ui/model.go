package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ktsopanakis/redmine-tui/internal/redmine"
)

type viewMode int

const (
	viewSplitScreen viewMode = iota
	viewProjectList
	viewFilterPopup
	viewHelpPopup
	viewConfirmPopup
)

type editField int

const (
	fieldNone editField = iota
	fieldSubject
	fieldDescription
	fieldStatus
	fieldDoneRatio
	fieldNotes
)

type Model struct {
	client         *redmine.Client
	redmineURL     string
	currentView    viewMode
	previousView   viewMode
	issueList      list.Model
	projectList    list.Model
	issues         []redmine.Issue
	projects       []redmine.Project
	currentIssue   *redmine.Issue
	currentProject *redmine.Project
	viewport       viewport.Model
	statuses       []redmine.Status
	width          int
	height         int
	err            error
	loading        bool

	// Edit mode
	editMode       bool
	activeField    editField
	subjectInput   textinput.Model
	descInput      textinput.Model
	doneRatioInput textinput.Model
	notesInput     textinput.Model
	statusList     list.Model
	statusSelected int

	// Pending changes
	pendingChanges map[string]interface{}

	// Cache for issue details to prevent flickering
	issueCache map[int]*redmine.Issue

	// Filters
	filterAssignedToMe bool
	filterStatusOpen   bool

	// Quick action status IDs
	inProgressStatusID int
	closedStatusID     int
	rejectedStatusID   int

	// Confirmation dialog
	pendingQuickAction string
	pendingStatusID    int
	pendingNote        string
}

type issueItem struct {
	issue redmine.Issue
}

func (i issueItem) Title() string {
	return fmt.Sprintf("#%d - %s", i.issue.ID, i.issue.Subject)
}

func (i issueItem) Description() string {
	status := i.issue.Status.Name
	assignee := "Unassigned"
	if i.issue.AssignedTo != nil {
		assignee = i.issue.AssignedTo.Name
	}
	return fmt.Sprintf("%s | %s | %d%% done", status, assignee, i.issue.DoneRatio)
}

func (i issueItem) FilterValue() string {
	return i.issue.Subject
}

type projectItem struct {
	project redmine.Project
}

func (p projectItem) Title() string {
	return p.project.Name
}

func (p projectItem) Description() string {
	return fmt.Sprintf("ID: %d", p.project.ID)
}

func (p projectItem) FilterValue() string {
	return p.project.Name
}

type statusItem struct {
	status redmine.Status
}

func (s statusItem) Title() string {
	return s.status.Name
}

func (s statusItem) Description() string {
	return ""
}

func (s statusItem) FilterValue() string {
	return s.status.Name
}

func New(client *redmine.Client) Model {
	// Initialize issue list
	issueDelegate := list.NewDefaultDelegate()
	issueList := list.New([]list.Item{}, issueDelegate, 0, 0)
	issueList.Title = "" // No title, header shows Redmine URL
	issueList.SetShowStatusBar(false)
	issueList.SetFilteringEnabled(true)
	issueList.Styles.Title = titleStyle

	// Initialize project list
	projectDelegate := list.NewDefaultDelegate()
	projectList := list.New([]list.Item{}, projectDelegate, 0, 0)
	projectList.Title = "Projects"
	projectList.SetShowStatusBar(true)
	projectList.SetFilteringEnabled(true)
	projectList.Styles.Title = titleStyle

	// Initialize viewport for detail panel
	vp := viewport.New(0, 0)

	// Initialize status list
	statusDelegate := list.NewDefaultDelegate()
	statusList := list.New([]list.Item{}, statusDelegate, 0, 0)
	statusList.Title = "Status"
	statusList.SetShowStatusBar(false)
	statusList.SetFilteringEnabled(false)
	statusList.Styles.Title = lipgloss.NewStyle().Bold(true)

	// Initialize text inputs
	subjectInput := textinput.New()
	subjectInput.Placeholder = "Issue subject"
	subjectInput.CharLimit = 255

	descInput := textinput.New()
	descInput.Placeholder = "Description"
	descInput.CharLimit = 5000

	doneRatioInput := textinput.New()
	doneRatioInput.Placeholder = "0-100"
	doneRatioInput.CharLimit = 3

	notesInput := textinput.New()
	notesInput.Placeholder = "Add notes about your changes..."
	notesInput.CharLimit = 5000

	return Model{
		client:             client,
		redmineURL:         client.BaseURL,
		currentView:        viewSplitScreen,
		issueList:          issueList,
		projectList:        projectList,
		statusList:         statusList,
		viewport:           vp,
		subjectInput:       subjectInput,
		descInput:          descInput,
		doneRatioInput:     doneRatioInput,
		notesInput:         notesInput,
		pendingChanges:     make(map[string]interface{}),
		issueCache:         make(map[int]*redmine.Issue),
		filterAssignedToMe: true, // Default to showing only my tasks
	}
}

type issuesLoadedMsg struct {
	issues []redmine.Issue
}

type projectsLoadedMsg struct {
	projects []redmine.Project
}

type issueLoadedMsg struct {
	issue *redmine.Issue
}

type statusesLoadedMsg struct {
	statuses []redmine.Status
}

type errMsg struct {
	err error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadIssues(),
		m.loadStatuses(),
	)
}

func (m Model) loadIssues() tea.Cmd {
	return func() tea.Msg {
		projectID := 0
		if m.currentProject != nil {
			projectID = m.currentProject.ID
		}
		resp, err := m.client.GetIssues(projectID, m.filterAssignedToMe, m.filterStatusOpen, 100, 0)
		if err != nil {
			return errMsg{err}
		}
		return issuesLoadedMsg{issues: resp.Issues}
	}
}

func (m Model) loadProjects() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.GetProjects(100, 0)
		if err != nil {
			return errMsg{err}
		}
		return projectsLoadedMsg{projects: resp.Projects}
	}
}

func (m Model) loadIssueDetail(id int) tea.Cmd {
	return func() tea.Msg {
		issue, err := m.client.GetIssue(id)
		if err != nil {
			return errMsg{err}
		}
		return issueLoadedMsg{issue: issue}
	}
}

func (m Model) loadStatuses() tea.Cmd {
	return func() tea.Msg {
		statuses, err := m.client.GetStatuses()
		if err != nil {
			return errMsg{err}
		}
		return statusesLoadedMsg{statuses: statuses}
	}
}

func (m Model) quickUpdateStatus(issueID, statusID int, note string) tea.Cmd {
	return func() tea.Msg {
		updates := map[string]interface{}{
			"status_id": statusID,
			"notes":     note,
		}
		err := m.client.UpdateIssue(issueID, updates)
		if err != nil {
			return errMsg{err}
		}
		// Reload the issue to show updated status
		issue, err := m.client.GetIssue(issueID)
		if err != nil {
			return errMsg{err}
		}
		return issueLoadedMsg{issue: issue}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Split screen: 35% for list, 65% for detail - use full width
		listWidth := int(float64(msg.Width) * 0.35)
		detailWidth := msg.Width - listWidth

		// List: account for border (2), padding (4), header (1)
		m.issueList.SetSize(listWidth-6, msg.Height-8)
		m.projectList.SetSize(msg.Width, msg.Height-2)
		// Viewport: account for border (2), padding (4), header (1)
		m.viewport.Width = detailWidth - 6
		m.viewport.Height = msg.Height - 8
		m.statusList.SetSize(30, 10)

		m.subjectInput.Width = detailWidth - 20
		m.descInput.Width = detailWidth - 20
		m.notesInput.Width = detailWidth - 20

	case issuesLoadedMsg:
		m.issues = msg.issues
		items := make([]list.Item, len(msg.issues))

		// Cache all issues and create list items
		for i, issue := range msg.issues {
			items[i] = issueItem{issue: issue}
			// Cache the issue data
			issueCopy := issue
			m.issueCache[issue.ID] = &issueCopy
		}
		m.issueList.SetItems(items)

		// Auto-select first issue if none selected
		if m.currentIssue == nil && len(msg.issues) > 0 {
			m.currentIssue = &msg.issues[0]
			m.viewport.SetContent(m.renderIssueDetail())
			m.viewport.GotoTop()
		}
		m.loading = false

	case projectsLoadedMsg:
		m.projects = msg.projects
		// Add "All Projects" as first item
		items := make([]list.Item, len(msg.projects)+1)
		items[0] = projectItem{project: redmine.Project{ID: 0, Name: "All Projects"}}
		for i, project := range msg.projects {
			items[i+1] = projectItem{project: project}
		}
		m.projectList.SetItems(items)
		m.loading = false

	case issueLoadedMsg:
		m.currentIssue = msg.issue
		// Update cache with detailed issue data
		m.issueCache[msg.issue.ID] = msg.issue

		// Update the issue in the issues list as well
		for i, issue := range m.issues {
			if issue.ID == msg.issue.ID {
				m.issues[i] = *msg.issue
				break
			}
		}

		// Refresh the list items to show updated status
		items := make([]list.Item, len(m.issues))
		for i, issue := range m.issues {
			items[i] = issueItem{issue: issue}
		}
		currentIndex := m.issueList.Index()
		m.issueList.SetItems(items)
		m.issueList.Select(currentIndex) // Maintain selection

		m.resetEditFields()
		m.viewport.SetContent(m.renderIssueDetail())
		m.viewport.GotoTop()
		m.loading = false

	case statusesLoadedMsg:
		m.statuses = msg.statuses
		items := make([]list.Item, len(msg.statuses))
		for i, status := range msg.statuses {
			items[i] = statusItem{status: status}
			// Identify common status IDs by name (case-insensitive matching)
			nameLower := strings.ToLower(status.Name)
			if strings.Contains(nameLower, "in progress") || strings.Contains(nameLower, "in-progress") {
				m.inProgressStatusID = status.ID
			} else if strings.Contains(nameLower, "closed") || strings.Contains(nameLower, "close") {
				m.closedStatusID = status.ID
			} else if strings.Contains(nameLower, "reject") {
				m.rejectedStatusID = status.ID
			}
		}
		m.statusList.SetItems(items)

	case errMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}

		switch m.currentView {
		case viewSplitScreen:
			// Handle edit mode separately
			if m.editMode {
				return m.handleEditMode(msg)
			}

			// Check if list is filtering - if so, let it handle all keys
			if m.issueList.FilterState() == list.Filtering {
				m.issueList, cmd = m.issueList.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}

			// Normal navigation mode
			switch {
			case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down):
				m.issueList, cmd = m.issueList.Update(msg)
				cmds = append(cmds, cmd)

				// Update selected issue instantly from cache
				if item, ok := m.issueList.SelectedItem().(issueItem); ok {
					if m.currentIssue == nil || m.currentIssue.ID != item.issue.ID {
						// Use cached data for instant display
						if cachedIssue, exists := m.issueCache[item.issue.ID]; exists {
							m.currentIssue = cachedIssue
							m.viewport.SetContent(m.renderIssueDetail())
							m.viewport.GotoTop()
						} else {
							// Fallback: use the basic issue from list
							issueCopy := item.issue
							m.currentIssue = &issueCopy
							m.viewport.SetContent(m.renderIssueDetail())
							m.viewport.GotoTop()
						}
					}
				}

			case key.Matches(msg, keys.Projects):
				m.currentView = viewProjectList
				m.loading = true
				return m, m.loadProjects()

			case key.Matches(msg, keys.Edit):
				if m.currentIssue != nil {
					m.editMode = true
					m.activeField = fieldSubject
					m.resetEditFields()
					m.subjectInput.Focus()
					m.viewport.SetContent(m.renderIssueDetail())
				}

			case key.Matches(msg, keys.Refresh):
				m.loading = true
				m.issueCache = make(map[int]*redmine.Issue) // Clear cache on refresh
				return m, m.loadIssues()

			case key.Matches(msg, keys.Filter):
				m.previousView = m.currentView
				m.currentView = viewFilterPopup

			case key.Matches(msg, keys.Help):
				m.previousView = m.currentView
				m.currentView = viewHelpPopup

			case key.Matches(msg, keys.QuickClose):
				if m.currentIssue != nil && m.closedStatusID > 0 {
					m.previousView = m.currentView
					m.currentView = viewConfirmPopup
					m.pendingQuickAction = "Close"
					m.pendingStatusID = m.closedStatusID
					m.pendingNote = "Closed via quick action"
				}

			case key.Matches(msg, keys.QuickInProgress):
				if m.currentIssue != nil && m.inProgressStatusID > 0 {
					m.previousView = m.currentView
					m.currentView = viewConfirmPopup
					m.pendingQuickAction = "Set to In Progress"
					m.pendingStatusID = m.inProgressStatusID
					m.pendingNote = "Set to In Progress via quick action"
				}

			case key.Matches(msg, keys.QuickReject):
				if m.currentIssue != nil && m.rejectedStatusID > 0 {
					m.previousView = m.currentView
					m.currentView = viewConfirmPopup
					m.pendingQuickAction = "Reject"
					m.pendingStatusID = m.rejectedStatusID
					m.pendingNote = "Rejected via quick action"
				}
			default:
				// Let list handle other keys (including /)
				m.issueList, cmd = m.issueList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case viewProjectList:
			switch {
			case key.Matches(msg, keys.Back):
				m.currentView = viewSplitScreen
			case key.Matches(msg, keys.Enter):
				if item, ok := m.projectList.SelectedItem().(projectItem); ok {
					if item.project.ID == 0 {
						// "All Projects" selected - clear current project
						m.currentProject = nil
					} else {
						m.currentProject = &item.project
					}
					m.currentView = viewSplitScreen
					m.currentIssue = nil
					m.issueCache = make(map[int]*redmine.Issue) // Clear cache when switching projects
					m.loading = true
					return m, m.loadIssues()
				}
			default:
				m.projectList, cmd = m.projectList.Update(msg)
				cmds = append(cmds, cmd)
			}

		case viewFilterPopup:
			switch {
			case key.Matches(msg, keys.Back), key.Matches(msg, keys.Filter):
				m.currentView = m.previousView
			case msg.String() == "1":
				// Toggle "My Tasks" filter
				m.filterAssignedToMe = !m.filterAssignedToMe
				m.currentView = m.previousView
				m.loading = true
				m.issueCache = make(map[int]*redmine.Issue)
				return m, m.loadIssues()
			case msg.String() == "2":
				// Toggle "Open Only" filter
				m.filterStatusOpen = !m.filterStatusOpen
				m.currentView = m.previousView
				m.loading = true
				m.issueCache = make(map[int]*redmine.Issue)
				return m, m.loadIssues()
			}

		case viewHelpPopup:
			// Any key closes help popup
			m.currentView = m.previousView

		case viewConfirmPopup:
			switch {
			case key.Matches(msg, keys.Enter):
				// Confirm the quick action
				m.currentView = m.previousView
				m.loading = true
				return m, m.quickUpdateStatus(m.currentIssue.ID, m.pendingStatusID, m.pendingNote)
			case key.Matches(msg, keys.Back):
				// Cancel the action
				m.currentView = m.previousView
				m.pendingQuickAction = ""
			}
		}
	}

	// Update viewport when not in edit mode
	if m.currentView == viewSplitScreen && !m.editMode {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch {
	case key.Matches(msg, keys.Save):
		// Save all pending changes
		if len(m.pendingChanges) > 0 {
			if notes := m.notesInput.Value(); notes != "" {
				m.pendingChanges["notes"] = notes
			}

			err := m.client.UpdateIssue(m.currentIssue.ID, m.pendingChanges)
			if err != nil {
				m.err = err
			} else {
				m.editMode = false
				m.activeField = fieldNone
				m.pendingChanges = make(map[string]interface{})
				m.loading = true
				return m, m.loadIssueDetail(m.currentIssue.ID)
			}
		} else {
			m.editMode = false
			m.activeField = fieldNone
		}

	case key.Matches(msg, keys.Cancel):
		m.editMode = false
		m.activeField = fieldNone
		m.pendingChanges = make(map[string]interface{})
		m.resetEditFields()
		m.viewport.SetContent(m.renderIssueDetail())

	case key.Matches(msg, keys.Tab):
		// Move to next field
		m.activeField = m.nextField(m.activeField)
		m.focusActiveField()
		m.viewport.SetContent(m.renderIssueDetail())

	case key.Matches(msg, keys.ShiftTab):
		// Move to previous field
		m.activeField = m.prevField(m.activeField)
		m.focusActiveField()
		m.viewport.SetContent(m.renderIssueDetail())

	case key.Matches(msg, keys.Enter):
		// Handle enter for status list
		if m.activeField == fieldStatus {
			if item, ok := m.statusList.SelectedItem().(statusItem); ok {
				m.pendingChanges["status_id"] = item.status.ID
				m.statusSelected = item.status.ID
			}
		}

	default:
		// Update the active input field
		switch m.activeField {
		case fieldSubject:
			m.subjectInput, cmd = m.subjectInput.Update(msg)
			if m.subjectInput.Value() != m.currentIssue.Subject {
				m.pendingChanges["subject"] = m.subjectInput.Value()
			} else {
				delete(m.pendingChanges, "subject")
			}
			cmds = append(cmds, cmd)

		case fieldDescription:
			m.descInput, cmd = m.descInput.Update(msg)
			if m.descInput.Value() != m.currentIssue.Description {
				m.pendingChanges["description"] = m.descInput.Value()
			} else {
				delete(m.pendingChanges, "description")
			}
			cmds = append(cmds, cmd)

		case fieldStatus:
			m.statusList, cmd = m.statusList.Update(msg)
			cmds = append(cmds, cmd)

		case fieldDoneRatio:
			m.doneRatioInput, cmd = m.doneRatioInput.Update(msg)
			if val := m.doneRatioInput.Value(); val != "" {
				if ratio, err := strconv.Atoi(val); err == nil && ratio >= 0 && ratio <= 100 {
					if ratio != m.currentIssue.DoneRatio {
						m.pendingChanges["done_ratio"] = ratio
					} else {
						delete(m.pendingChanges, "done_ratio")
					}
				}
			}
			cmds = append(cmds, cmd)

		case fieldNotes:
			m.notesInput, cmd = m.notesInput.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update the detail view to show changes
	m.viewport.SetContent(m.renderIssueDetail())

	return m, tea.Batch(cmds...)
}

func (m *Model) resetEditFields() {
	if m.currentIssue != nil {
		m.subjectInput.SetValue(m.currentIssue.Subject)
		m.descInput.SetValue(m.currentIssue.Description)
		m.doneRatioInput.SetValue(fmt.Sprintf("%d", m.currentIssue.DoneRatio))
		m.notesInput.SetValue("")
		m.statusSelected = m.currentIssue.Status.ID

		// Set status list selection
		for i, item := range m.statusList.Items() {
			if s, ok := item.(statusItem); ok && s.status.ID == m.currentIssue.Status.ID {
				m.statusList.Select(i)
				break
			}
		}
	}
	m.pendingChanges = make(map[string]interface{})
}

func (m *Model) focusActiveField() {
	m.subjectInput.Blur()
	m.descInput.Blur()
	m.doneRatioInput.Blur()
	m.notesInput.Blur()

	switch m.activeField {
	case fieldSubject:
		m.subjectInput.Focus()
	case fieldDescription:
		m.descInput.Focus()
	case fieldDoneRatio:
		m.doneRatioInput.Focus()
	case fieldNotes:
		m.notesInput.Focus()
	}
}

func (m Model) nextField(current editField) editField {
	switch current {
	case fieldSubject:
		return fieldDescription
	case fieldDescription:
		return fieldStatus
	case fieldStatus:
		return fieldDoneRatio
	case fieldDoneRatio:
		return fieldNotes
	case fieldNotes:
		return fieldSubject
	default:
		return fieldSubject
	}
}

func (m Model) prevField(current editField) editField {
	switch current {
	case fieldSubject:
		return fieldNotes
	case fieldDescription:
		return fieldSubject
	case fieldStatus:
		return fieldDescription
	case fieldDoneRatio:
		return fieldStatus
	case fieldNotes:
		return fieldDoneRatio
	default:
		return fieldSubject
	}
}

func (m Model) View() string {
	if m.loading {
		return loadingStyle.Render("Loading...")
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress q to quit", m.err))
	}

	var content string
	var help string

	switch m.currentView {
	case viewSplitScreen:
		// Header with Redmine URL spanning full width at the top
		header := headerStyle.Width(m.width).Render(fmt.Sprintf("🔗 %s", m.redmineURL))

		// Calculate widths - use 35/65 split and full width
		listWidth := int(float64(m.width) * 0.35)
		detailWidth := m.width - listWidth

		// Left panel: issue list with border - same height as right
		leftPanelContent := m.issueList.View()
		leftPanel := listPaneStyle.Width(listWidth - 2).Height(m.height - 4).Render(leftPanelContent)

		// Right panel: detail with border - same height as left
		var rightPanelContent string
		if m.currentIssue != nil {
			rightPanelContent = m.viewport.View()
		} else {
			rightPanelContent = "Select an issue to view details"
		}
		rightPanel := detailStyle.Width(detailWidth - 2).Height(m.height - 4).Render(rightPanelContent)

		// Combine panels side by side
		panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

		// Stack header on top of panels
		content = lipgloss.JoinVertical(lipgloss.Left, header, panels)

		// Help text
		if m.editMode {
			changesInfo := ""
			if len(m.pendingChanges) > 0 {
				changesInfo = fmt.Sprintf(" [%d pending changes]", len(m.pendingChanges))
			}
			help = helpStyle.Render(fmt.Sprintf("EDIT MODE%s | tab/shift+tab: navigate • ctrl+s: save • esc: cancel • q: quit", changesInfo))
		} else {
			projectInfo := "All Projects"
			if m.currentProject != nil {
				projectInfo = m.currentProject.Name
			}
			filterInfo := ""
			if m.filterAssignedToMe {
				filterInfo = " [My Tasks]"
			}
			help = helpStyle.Render(fmt.Sprintf("Project: %s%s | f: filters • ?: help • c/i/x: quick actions • q: quit", projectInfo, filterInfo))
		}

	case viewProjectList:
		content = m.projectList.View()
		help = helpStyle.Render("enter: select • esc: back • q: quit")

	case viewFilterPopup:
		content = m.renderFilterPopup()
		help = helpStyle.Render("1-9: toggle filter • esc/f: close • q: quit")

	case viewHelpPopup:
		content = m.renderHelpPopup()
		help = helpStyle.Render("Any key to close")

	case viewConfirmPopup:
		content = m.renderConfirmPopup()
		help = helpStyle.Render("enter: confirm • esc: cancel")
	}

	return fmt.Sprintf("%s\n%s", content, help)
}

func (m Model) renderIssueDetail() string {
	if m.currentIssue == nil {
		return ""
	}

	var b strings.Builder

	// Title / Subject
	if m.editMode && m.activeField == fieldSubject {
		b.WriteString(editHeaderStyle.Render("Subject (editing):"))
		b.WriteString("\n")
		b.WriteString(m.subjectInput.View())
		b.WriteString("\n\n")
	} else {
		b.WriteString(titleStyle.Render(fmt.Sprintf("#%d - %s", m.currentIssue.ID, m.currentIssue.Subject)))
		b.WriteString("\n\n")
	}

	// Project (read-only)
	b.WriteString(labelStyle.Render("Project: "))
	b.WriteString(m.currentIssue.Project.Name)
	b.WriteString("\n")

	// Status
	if m.editMode && m.activeField == fieldStatus {
		b.WriteString(editHeaderStyle.Render("Status (use ↑/↓ to select, enter to confirm):"))
		b.WriteString("\n")
		b.WriteString(m.statusList.View())
		b.WriteString("\n")
	} else {
		b.WriteString(labelStyle.Render("Status: "))
		// Show pending change if exists
		if statusID, ok := m.pendingChanges["status_id"].(int); ok {
			for _, s := range m.statuses {
				if s.ID == statusID {
					b.WriteString(changedStyle.Render(s.Name + " *"))
					break
				}
			}
		} else {
			b.WriteString(m.currentIssue.Status.Name)
		}
		b.WriteString("\n")
	}

	// Priority (read-only)
	b.WriteString(labelStyle.Render("Priority: "))
	b.WriteString(m.currentIssue.Priority.Name)
	b.WriteString("\n")

	// Tracker (read-only)
	b.WriteString(labelStyle.Render("Tracker: "))
	b.WriteString(m.currentIssue.Tracker.Name)
	b.WriteString("\n")

	// Assigned to (read-only)
	if m.currentIssue.AssignedTo != nil {
		b.WriteString(labelStyle.Render("Assigned to: "))
		b.WriteString(m.currentIssue.AssignedTo.Name)
		b.WriteString("\n")
	}

	// Done ratio
	if m.editMode && m.activeField == fieldDoneRatio {
		b.WriteString(editHeaderStyle.Render("Done % (0-100):"))
		b.WriteString("\n")
		b.WriteString(m.doneRatioInput.View())
		b.WriteString("\n")
	} else {
		b.WriteString(labelStyle.Render("Done: "))
		if ratio, ok := m.pendingChanges["done_ratio"].(int); ok {
			b.WriteString(changedStyle.Render(fmt.Sprintf("%d%% *", ratio)))
		} else {
			b.WriteString(fmt.Sprintf("%d%%", m.currentIssue.DoneRatio))
		}
		b.WriteString("\n")
	}

	// Due date (read-only)
	if m.currentIssue.DueDate != "" {
		b.WriteString(labelStyle.Render("Due date: "))
		b.WriteString(m.currentIssue.DueDate)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Description
	if m.editMode && m.activeField == fieldDescription {
		b.WriteString(editHeaderStyle.Render("Description (editing):"))
		b.WriteString("\n")
		b.WriteString(m.descInput.View())
		b.WriteString("\n\n")
	} else {
		b.WriteString(labelStyle.Render("Description:"))
		b.WriteString("\n")
		if desc, ok := m.pendingChanges["description"].(string); ok {
			b.WriteString(changedStyle.Render(desc + " *"))
		} else if m.currentIssue.Description != "" {
			b.WriteString(m.currentIssue.Description)
		} else {
			b.WriteString(dimStyle.Render("No description"))
		}
		b.WriteString("\n\n")
	}

	// Notes field (only in edit mode)
	if m.editMode {
		if m.activeField == fieldNotes {
			b.WriteString(editHeaderStyle.Render("Notes (add comments about your changes):"))
		} else {
			b.WriteString(labelStyle.Render("Notes:"))
		}
		b.WriteString("\n")
		b.WriteString(m.notesInput.View())
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderFilterPopup() string {
	width := min(60, m.width-4)
	height := min(15, m.height-4)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Filters"))
	content.WriteString("\n\n")

	// My Tasks filter
	myTasksStatus := "[ ]"
	if m.filterAssignedToMe {
		myTasksStatus = "[✓]"
	}
	content.WriteString(fmt.Sprintf("1. %s My Tasks Only\n", myTasksStatus))
	content.WriteString("   Show only issues assigned to me\n\n")

	// Open Status filter
	openStatus := "[ ]"
	if m.filterStatusOpen {
		openStatus = "[✓]"
	}
	content.WriteString(fmt.Sprintf("2. %s Open Issues Only\n", openStatus))
	content.WriteString("   Hide closed/resolved issues\n\n")

	content.WriteString(dimStyle.Render("Press 1-2 to toggle filters\nPress ESC or F to close"))

	popup := popupStyle.Width(width).Height(height).Render(content.String())
	return placeInCenter(m.width, m.height, width+4, height+4, popup)
}

func (m Model) renderHelpPopup() string {
	width := min(70, m.width-4)
	height := min(25, m.height-4)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Keyboard Shortcuts"))
	content.WriteString("\n\n")

	content.WriteString(labelStyle.Render("Navigation"))
	content.WriteString("\n")
	content.WriteString("  ↑/↓, j/k    Navigate issues\n")
	content.WriteString("  p           Switch projects\n")
	content.WriteString("  r           Refresh issue list\n")
	content.WriteString("  /           Search/filter in lists\n\n")

	content.WriteString(labelStyle.Render("Editing"))
	content.WriteString("\n")
	content.WriteString("  e           Enter edit mode\n")
	content.WriteString("  Tab         Next field\n")
	content.WriteString("  Shift+Tab   Previous field\n")
	content.WriteString("  Ctrl+S      Save changes\n")
	content.WriteString("  Esc         Cancel/Go back\n\n")

	content.WriteString(labelStyle.Render("Quick Actions"))
	content.WriteString("\n")
	content.WriteString("  c           Close ticket\n")
	content.WriteString("  i           Set to In Progress\n")
	content.WriteString("  x           Reject ticket\n\n")

	content.WriteString(labelStyle.Render("Other"))
	content.WriteString("\n")
	content.WriteString("  f           Open filters\n")
	content.WriteString("  ?           Show this help\n")
	content.WriteString("  q           Quit\n\n")

	content.WriteString(dimStyle.Render("Press any key to close"))

	popup := popupStyle.Width(width).Height(height).Render(content.String())
	return placeInCenter(m.width, m.height, width+4, height+4, popup)
}

func (m Model) renderConfirmPopup() string {
	width := min(60, m.width-4)
	height := min(12, m.height-4)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Confirm Action"))
	content.WriteString("\n\n")

	if m.currentIssue != nil {
		content.WriteString(fmt.Sprintf("Issue: #%d - %s\n\n", m.currentIssue.ID, m.currentIssue.Subject))
	}

	content.WriteString(fmt.Sprintf("Action: %s\n\n", m.pendingQuickAction))

	// Find and show the target status name
	var statusName string
	for _, s := range m.statuses {
		if s.ID == m.pendingStatusID {
			statusName = s.Name
			break
		}
	}
	if statusName != "" {
		content.WriteString(fmt.Sprintf("New Status: %s\n\n", statusName))
	}

	content.WriteString(changedStyle.Render("Press ENTER to confirm\n"))
	content.WriteString(dimStyle.Render("Press ESC to cancel"))

	popup := popupStyle.Width(width).Height(height).Render(content.String())
	return placeInCenter(m.width, m.height, width+4, height+4, popup)
}

func placeInCenter(termWidth, termHeight, boxWidth, boxHeight int, content string) string {
	left := (termWidth - boxWidth) / 2
	top := (termHeight - boxHeight) / 2

	if left < 0 {
		left = 0
	}
	if top < 0 {
		top = 0
	}

	return lipgloss.Place(termWidth, termHeight, lipgloss.Center, lipgloss.Center, content)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575"))

	editHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFA500")).
			Underline(true)

	changedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Italic(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(1, 0, 0, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(1)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(1)

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)

	listPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#04B575")).
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FFA500")).
			Padding(2, 4).
			Background(lipgloss.Color("#1a1a1a"))
)

// Key bindings
type keyMap struct {
	Quit            key.Binding
	Enter           key.Binding
	Back            key.Binding
	Up              key.Binding
	Down            key.Binding
	Projects        key.Binding
	Edit            key.Binding
	Save            key.Binding
	Cancel          key.Binding
	Refresh         key.Binding
	Tab             key.Binding
	ShiftTab        key.Binding
	Filter          key.Binding
	Help            key.Binding
	QuickClose      key.Binding
	QuickInProgress key.Binding
	QuickReject     key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Projects: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "projects"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filters"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	QuickClose: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "close ticket"),
	),
	QuickInProgress: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "in progress"),
	),
	QuickReject: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "reject ticket"),
	),
}
