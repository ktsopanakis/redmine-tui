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

	// Focus management
	focusedPane paneType // "list" or "detail"
}

type paneType int

const (
	focusList paneType = iota
	focusDetail
)

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
	// Return all searchable fields for comprehensive filtering
	assignee := "unassigned"
	if i.issue.AssignedTo != nil {
		assignee = i.issue.AssignedTo.Name
	}
	author := i.issue.Author.Name
	desc := i.issue.Description
	if len(desc) > 200 {
		desc = desc[:200]
	}
	// Combine all searchable fields
	return fmt.Sprintf("#%d %s %s %s %s %s %s %s",
		i.issue.ID,
		i.issue.Subject,
		i.issue.Status.Name,
		i.issue.Project.Name,
		i.issue.Tracker.Name,
		assignee,
		author,
		desc,
	)
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
	issueList.Title = ""
	issueList.SetShowStatusBar(false)
	issueList.SetFilteringEnabled(true)
	issueList.Styles.Title = titleStyle
	issueList.Styles.TitleBar = lipgloss.NewStyle()
	issueList.SetShowTitle(false)

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
		focusedPane:        focusList,
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

		// Split screen: 40% for list, 60% for detail
		listWidth := int(float64(msg.Width) * 0.40)
		detailWidth := msg.Width - listWidth

		// List: account for border and padding
		m.issueList.SetSize(listWidth-3, msg.Height-5)
		m.projectList.SetSize(msg.Width, msg.Height-2)
		// Viewport: account for padding
		m.viewport.Width = detailWidth - 3
		m.viewport.Height = msg.Height - 5
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
			case key.Matches(msg, keys.Tab):
				// Switch focus between panes
				if m.focusedPane == focusList {
					m.focusedPane = focusDetail
				} else {
					m.focusedPane = focusList
				}

			case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down):
				if m.focusedPane == focusList {
					// Navigate issue list
					m.issueList, cmd = m.issueList.Update(msg)
					cmds = append(cmds, cmd)

					// Load full issue details including journals
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
							// Always load full details with journals in the background
							cmds = append(cmds, m.loadIssueDetail(item.issue.ID))
						}
					}
				} else {
					// Scroll detail viewport
					if key.Matches(msg, keys.Up) {
						m.viewport.LineUp(1)
					} else {
						m.viewport.LineDown(1)
					}
				}

			case msg.String() == "pgup":
				if m.focusedPane == focusDetail {
					m.viewport.HalfViewUp()
				}

			case msg.String() == "pgdown":
				if m.focusedPane == focusDetail {
					m.viewport.HalfViewDown()
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
				if item.status.ID != m.currentIssue.Status.ID {
					m.pendingChanges["status_id"] = item.status.ID
					m.statusSelected = item.status.ID
				} else {
					delete(m.pendingChanges, "status_id")
				}
			}
		}

	case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down):
		// Handle up/down - either for status list or viewport scrolling
		if m.activeField == fieldStatus {
			m.statusList, cmd = m.statusList.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}

	default:
		// Update only the active input field
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

	switch m.currentView {
	case viewSplitScreen:
		// Header bar
		headerLeft := fmt.Sprintf("🔗 %s", m.redmineURL)
		headerRight := ""
		if m.currentProject != nil {
			headerRight = fmt.Sprintf("Project: %s", m.currentProject.Name)
		} else {
			headerRight = "All Projects"
		}
		if m.filterAssignedToMe {
			headerRight += " • My Tasks"
		}
		if m.filterStatusOpen {
			headerRight += " • Open Only"
		}

		headerPadding := m.width - lipgloss.Width(headerLeft) - lipgloss.Width(headerRight) - 4
		if headerPadding < 0 {
			headerPadding = 0
		}
		header := headerStyle.Width(m.width).Render(headerLeft + strings.Repeat(" ", headerPadding) + headerRight)

		// Calculate widths - adjust for better spacing
		listWidth := int(float64(m.width) * 0.40)
		detailWidth := m.width - listWidth

		// Left panel title - match exact panel width
		leftTitleText := "ISSUES"
		if m.focusedPane == focusList {
			leftTitleText = "● " + leftTitleText // Add indicator when focused
		}
		leftTitle := panelTitleStyle.Width(listWidth).Render(leftTitleText)

		// Left panel content - highlight border when focused (more padding on right)
		leftContent := m.issueList.View()
		leftPanelStyle := panelStyle
		if m.focusedPane == focusList {
			leftPanelStyle = leftPanelStyle.BorderForeground(lipgloss.Color(colorPrimary))
		}
		leftPanel := leftPanelStyle.Width(listWidth - 3).Height(m.height - 4).Render(leftContent)

		// Right panel title - match exact panel width
		rightTitleText := ""
		if m.currentIssue != nil {
			rightTitleText = fmt.Sprintf("ISSUE #%d", m.currentIssue.ID)
		} else {
			rightTitleText = "ISSUE DETAILS"
		}
		if m.focusedPane == focusDetail {
			rightTitleText = "● " + rightTitleText // Add indicator when focused
		}
		rightTitle := panelTitleStyle.Width(detailWidth).Render(rightTitleText)

		// Right panel content - highlight border when focused (less padding on right)
		var rightContent string
		if m.currentIssue != nil {
			rightContent = m.viewport.View()
		} else {
			rightContent = dimStyle.Render("← Select an issue to view details")
		}
		rightPanelStyleCurrent := rightPanelStyle
		if m.focusedPane == focusDetail {
			rightPanelStyleCurrent = rightPanelStyleCurrent.BorderForeground(lipgloss.Color(colorPrimary))
		}
		rightPanel := rightPanelStyleCurrent.Width(detailWidth - 2).Height(m.height - 4).Render(rightContent)

		// Combine panels with titles
		leftSection := lipgloss.JoinVertical(lipgloss.Left, leftTitle, leftPanel)
		rightSection := lipgloss.JoinVertical(lipgloss.Left, rightTitle, rightPanel)
		panels := lipgloss.JoinHorizontal(lipgloss.Top, leftSection, rightSection)

		// Footer with shortcuts
		footer := m.renderFooter()

		// Combine all
		content = lipgloss.JoinVertical(lipgloss.Left, header, panels, footer)

	case viewProjectList:
		header := headerStyle.Width(m.width).Render("SELECT PROJECT")
		body := m.projectList.View()
		footer := footerStyle.Width(m.width).Render(" enter: select │ esc: back │ q: quit")
		content = lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	case viewFilterPopup:
		header := headerStyle.Width(m.width).Render("FILTERS")
		body := m.renderFilterPopup()
		footer := footerStyle.Width(m.width).Render(" 1-9: toggle │ esc/f: close │ q: quit")
		content = lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	case viewHelpPopup:
		header := headerStyle.Width(m.width).Render("HELP")
		body := m.renderHelpPopup()
		footer := footerStyle.Width(m.width).Render(" Any key to close")
		content = lipgloss.JoinVertical(lipgloss.Left, header, body, footer)

	case viewConfirmPopup:
		header := headerStyle.Width(m.width).Render("CONFIRM ACTION")
		body := m.renderConfirmPopup()
		footer := footerStyle.Width(m.width).Render(" enter: confirm │ esc: cancel")
		content = lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	}

	return content
}

func (m Model) renderFooter() string {
	if m.editMode {
		shortcuts := []string{}
		if len(m.pendingChanges) > 0 {
			shortcuts = append(shortcuts, footerKeyStyle.Render("ctrl+s")+" save")
		}
		shortcuts = append(shortcuts,
			footerKeyStyle.Render("tab")+" next field",
			footerKeyStyle.Render("↑↓")+" scroll/select",
			footerKeyStyle.Render("esc")+" cancel",
			footerKeyStyle.Render("q")+" quit",
		)
		return footerStyle.Width(m.width).Render(" " + strings.Join(shortcuts, " │ "))
	}

	shortcuts := []string{
		footerKeyStyle.Render("↑↓/jk") + " navigate",
		footerKeyStyle.Render("tab") + " switch pane",
		footerKeyStyle.Render("enter") + " view",
		footerKeyStyle.Render("e") + " edit",
		footerKeyStyle.Render("p") + " projects",
		footerKeyStyle.Render("f") + " filters",
		footerKeyStyle.Render("/") + " search",
		footerKeyStyle.Render("c/i/x") + " quick actions",
		footerKeyStyle.Render("r") + " refresh",
		footerKeyStyle.Render("?") + " help",
		footerKeyStyle.Render("q") + " quit",
	}
	return footerStyle.Width(m.width).Render(" " + strings.Join(shortcuts, " │ "))
}

// wrapText wraps text to fit within the given width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var result strings.Builder
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	lineLen := 0
	for i, word := range words {
		wordLen := len(word)
		if lineLen == 0 {
			// First word on line
			result.WriteString(word)
			lineLen = wordLen
		} else if lineLen+1+wordLen <= width {
			// Word fits on current line
			result.WriteString(" " + word)
			lineLen += 1 + wordLen
		} else {
			// Start new line
			result.WriteString("\n" + word)
			lineLen = wordLen
		}
		if i < len(words)-1 && lineLen >= width {
			result.WriteString("\n")
			lineLen = 0
		}
	}
	return result.String()
}

func (m Model) renderIssueDetail() string {
	if m.currentIssue == nil {
		return ""
	}

	// Calculate available width for wrapping
	detailWidth := int(float64(m.width)*0.60) - 6
	if detailWidth < 40 {
		detailWidth = 40
	}

	var b strings.Builder

	// EDIT MODE - Show as editable form
	if m.editMode {
		b.WriteString(titleStyle.Render("EDIT ISSUE #" + fmt.Sprintf("%d", m.currentIssue.ID)))
		b.WriteString("\n\n")

		// Subject
		b.WriteString(labelStyle.Render("Subject:"))
		b.WriteString("\n")
		b.WriteString(m.subjectInput.View())
		b.WriteString("\n\n")

		// Status
		b.WriteString(labelStyle.Render("Status: (↑↓ to select)"))
		b.WriteString("\n")
		b.WriteString(m.statusList.View())
		b.WriteString("\n\n")

		// Done %
		b.WriteString(labelStyle.Render("Done % (0-100):"))
		b.WriteString("\n")
		b.WriteString(m.doneRatioInput.View())
		b.WriteString("\n\n")

		// Description
		b.WriteString(labelStyle.Render("Description:"))
		b.WriteString("\n")
		b.WriteString(m.descInput.View())
		b.WriteString("\n\n")

		// Notes
		b.WriteString(labelStyle.Render("Update Notes:"))
		b.WriteString("\n")
		b.WriteString(m.notesInput.View())
		b.WriteString("\n\n")

		if len(m.pendingChanges) > 0 {
			b.WriteString(changedStyle.Render(fmt.Sprintf("✓ %d field(s) modified - Press Ctrl+S to save", len(m.pendingChanges))))
		} else {
			b.WriteString(dimStyle.Render("No changes yet"))
		}

		return b.String()
	}

	// VIEW MODE - Compact information display
	b.WriteString(titleStyle.Render(m.currentIssue.Subject))
	b.WriteString("\n\n")

	// Metadata grid - Row 1
	metaItems := []string{
		labelStyle.Render("Project:") + " " + valueStyle.Render(m.currentIssue.Project.Name),
		labelStyle.Render("Tracker:") + " " + valueStyle.Render(m.currentIssue.Tracker.Name),
		labelStyle.Render("Priority:") + " " + valueStyle.Render(m.currentIssue.Priority.Name),
	}
	b.WriteString(wrapText(strings.Join(metaItems, "  │  "), detailWidth))
	b.WriteString("\n")

	// Metadata grid - Row 2
	metaItems = []string{
		labelStyle.Render("Status:") + " " + statusStyle.Render(m.currentIssue.Status.Name),
		labelStyle.Render("Done:") + " " + valueStyle.Render(fmt.Sprintf("%d%%", m.currentIssue.DoneRatio)),
	}
	if m.currentIssue.AssignedTo != nil {
		metaItems = append(metaItems, labelStyle.Render("Assigned:")+" "+valueStyle.Render(m.currentIssue.AssignedTo.Name))
	} else {
		metaItems = append(metaItems, labelStyle.Render("Assigned:")+" "+dimStyle.Render("unassigned"))
	}
	b.WriteString(wrapText(strings.Join(metaItems, "  │  "), detailWidth))
	b.WriteString("\n")

	// Metadata grid - Row 3 (Author and dates)
	metaItems = []string{
		labelStyle.Render("Author:") + " " + valueStyle.Render(m.currentIssue.Author.Name),
	}
	if m.currentIssue.StartDate != "" {
		metaItems = append(metaItems, labelStyle.Render("Start:")+" "+valueStyle.Render(m.currentIssue.StartDate))
	}
	if m.currentIssue.DueDate != "" {
		metaItems = append(metaItems, labelStyle.Render("Due:")+" "+valueStyle.Render(m.currentIssue.DueDate))
	}
	b.WriteString(wrapText(strings.Join(metaItems, "  │  "), detailWidth))
	b.WriteString("\n")

	// Timestamps
	timeItems := []string{
		labelStyle.Render("Created:") + " " + dimStyle.Render(m.currentIssue.CreatedOn.Format("2006-01-02 15:04")),
		labelStyle.Render("Updated:") + " " + dimStyle.Render(m.currentIssue.UpdatedOn.Format("2006-01-02 15:04")),
	}
	b.WriteString(wrapText(strings.Join(timeItems, "  │  "), detailWidth))
	b.WriteString("\n\n")

	// Description section with wrapping
	b.WriteString(labelStyle.Render("━━ DESCRIPTION "))
	b.WriteString(dimStyle.Render(strings.Repeat("━", max(0, detailWidth-15))))
	b.WriteString("\n")
	if m.currentIssue.Description != "" {
		wrappedDesc := wrapText(m.currentIssue.Description, detailWidth)
		b.WriteString(valueStyle.Render(wrappedDesc))
	} else {
		b.WriteString(dimStyle.Render("(no description)"))
	}
	b.WriteString("\n\n")

	// Notes/History section
	if len(m.currentIssue.Journals) > 0 {
		b.WriteString(labelStyle.Render("━━ NOTES & HISTORY "))
		b.WriteString(dimStyle.Render(strings.Repeat("━", max(0, detailWidth-19))))
		b.WriteString("\n\n")

		for _, journal := range m.currentIssue.Journals {
			hasContent := journal.Notes != "" || len(journal.Details) > 0
			if hasContent {
				// Note header with user and date
				noteHeader := labelStyle.Render(journal.User.Name) + " " +
					dimStyle.Render(journal.CreatedOn.Format("2006-01-02 15:04"))
				b.WriteString(noteHeader)
				b.WriteString("\n")

				// Show field changes
				if len(journal.Details) > 0 {
					for _, detail := range journal.Details {
						changeText := fmt.Sprintf("  • %s: ", detail.Name)
						if detail.OldValue != "" {
							changeText += fmt.Sprintf("%s → %s", detail.OldValue, detail.NewValue)
						} else {
							changeText += fmt.Sprintf("set to %s", detail.NewValue)
						}
						b.WriteString(dimStyle.Render(changeText))
						b.WriteString("\n")
					}
				}

				// Show note content if present
				if journal.Notes != "" {
					if len(journal.Details) > 0 {
						b.WriteString("\n")
					}
					wrappedNote := wrapText(journal.Notes, detailWidth)
					b.WriteString(valueStyle.Render(wrappedNote))
				}
				b.WriteString("\n\n")
			}
		}
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Styles - btop inspired
var (
	// Colors
	colorPrimary   = lipgloss.Color("#50fa7b")
	colorSecondary = lipgloss.Color("#8be9fd")
	colorAccent    = lipgloss.Color("#ffb86c")
	colorDanger    = lipgloss.Color("#ff5555")
	colorMuted     = lipgloss.Color("#6272a4")
	colorBg        = lipgloss.Color("#282a36")
	colorBgLight   = lipgloss.Color("#44475a")

	// Header bar
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282a36")).
			Background(colorPrimary).
			Padding(0, 2)

	// Footer bar
	footerStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Background(colorBgLight).
			Padding(0, 1)

	footerKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	// Panel styles
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary).
			Background(colorBgLight).
			Padding(0, 1)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(colorMuted).
			PaddingLeft(1)

	rightPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, true, false, false).
			BorderForeground(colorMuted).
			PaddingLeft(2)

	// Content styles
	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent)

	statusStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	changedStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Italic(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	editHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Underline(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true).
			Padding(1)

	loadingStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true).
			Padding(1)

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(2, 4).
			Background(colorBg)
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
