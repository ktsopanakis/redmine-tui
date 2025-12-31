package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ktsopanakis/redmine-tui/api"
)

const (
	headerHeight = 1
	footerHeight = 1
)

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
	client              *api.Client
	issues              []api.Issue
	selectedIndex       int
	selectedDisplayLine int // Line number where selected issue is displayed
	loading             bool
	err                 error
	currentUser         *api.User
	filterMode          bool
	filterInput         textinput.Model
	filterText          string
	viewMode            string // "my", "all", "user"
	assigneeFilter      string // username or "" for my/all modes
	projectFilter       string // project name or "" for all projects
	userInputMode       string // "", "user", "project" - which input is active

	// List selection state
	availableUsers       []api.User
	availableProjects    []api.Project
	selectedUsers        map[int]bool   // user ID -> selected
	selectedProjects     map[int]bool   // project ID -> selected
	selectedUserNames    map[int]string // user ID -> display name
	selectedProjectNames map[int]string // project ID -> display name
	listCursor           int            // cursor position in list
	listLoading          bool           // loading list data
	listFilterText       string         // filter text for list items
	filteredIndices      []int          // indices of filtered items in original list

	// Edit mode state
	editMode            bool              // whether edit mode is active
	editFieldIndex      int               // which field is currently selected for editing
	editInput           textinput.Model   // input for editing
	editingIssueID      int               // ID of the issue being edited
	availableStatuses   []api.Status      // available statuses for selection
	availablePriorities []api.Priority    // available priorities for selection
	hasUnsavedChanges   bool              // whether there are unsaved changes in edit mode
	editOriginalValue   string            // original value before editing
	pendingEdits        map[string]string // fieldName -> new value for all pending edits
	originalValues      map[string]string // fieldName -> original value for comparison
}

func initialModel() model {
	client := api.NewClient(settings.Redmine.URL, settings.Redmine.APIKey)
	filterInput := textinput.New()
	filterInput.Placeholder = "Type to filter issues..."
	filterInput.CharLimit = 100
	filterInput.Width = 50

	editInput := textinput.New()
	editInput.Placeholder = "Enter value..."
	editInput.CharLimit = 500
	editInput.Width = 50

	return model{
		leftTitle:        "Issues",
		rightTitle:       "Details",
		activePane:       0,
		client:           client,
		selectedIndex:    0,
		loading:          true,
		filterInput:      filterInput,
		editInput:        editInput,
		viewMode:         "my",
		selectedUsers:    make(map[int]bool),
		selectedProjects: make(map[int]bool),
		editMode:         false,
		editFieldIndex:   0,
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

	case usersLoadedMsg:
		m.listLoading = false
		if msg.err == nil {
			m.availableUsers = msg.users
			m.listCursor = 0
			// Build initial filtered list
			m.buildFilteredList()
		}
		return m, nil

	case projectsLoadedMsg:
		m.listLoading = false
		if msg.err == nil {
			m.availableProjects = msg.projects
			m.listCursor = 0
			// Build initial filtered list
			m.buildFilteredList()
		}
		return m, nil

	case statusesLoadedMsg:
		if msg.err == nil {
			m.availableStatuses = msg.statuses
		}
		return m, nil

	case prioritiesLoadedMsg:
		if msg.err == nil {
			m.availablePriorities = msg.priorities
		}
		return m, nil

	case issueUpdatedMsg:
		m.loading = false
		m.editMode = false
		m.editInput.Blur()
		if msg.err == nil {
			// Refresh the issue list and details
			return m, tea.Batch(
				fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues),
				fetchIssueDetail(m.client, msg.issueID),
			)
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
		// Check if we're in any input mode - if so, only handle esc, enter, and pass to input
		inInputMode := m.filterMode || m.userInputMode != "" || m.editMode

		switch msg.String() {
		case "ctrl+c", "q":
			if m.editMode && m.hasUnsavedChanges {
				// Warn about unsaved changes but allow quit
				m.editMode = false
				m.hasUnsavedChanges = false
				m.editInput.Blur()
				// Could add a confirmation dialog here, but for now just quit
				return m, tea.Quit
			} else if m.editMode {
				// Exit edit mode without saving
				m.editMode = false
				m.editInput.Blur()
				return m, nil
			}
			return m, tea.Quit

		case "esc":
			if m.editMode {
				// Exit edit mode without saving - clear all pending edits
				m.editMode = false
				m.hasUnsavedChanges = false
				m.pendingEdits = make(map[string]string)
				m.originalValues = make(map[string]string)
				m.editInput.Blur()
				m.updatePaneContent()
				return m, nil
			} else if m.filterMode {
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
				m.listFilterText = ""
				m.filterInput.Blur()
				m.filterInput.SetValue("")
				return m, nil
			}

		case "ctrl+s":
			if m.editMode && len(m.pendingEdits) > 0 {
				// Save current field to pending before submitting
				if m.editFieldIndex < len(editableFields) {
					field := editableFields[m.editFieldIndex]
					currentValue := m.editInput.Value()
					originalValue := m.originalValues[field.Name]
					if currentValue != originalValue {
						m.pendingEdits[field.Name] = currentValue
					} else {
						delete(m.pendingEdits, field.Name)
					}
				}

				// Save all pending changes at once
				if len(m.pendingEdits) > 0 {
					filteredIssues := m.getFilteredIssues()
					if m.selectedIndex >= 0 && m.selectedIndex < len(filteredIssues) {
						issueID := filteredIssues[m.selectedIndex].ID
						m.editingIssueID = issueID
						m.loading = true
						m.hasUnsavedChanges = false
						m.editMode = false
						m.editInput.Blur()
						return m, updateIssueMultiple(m.client, issueID, m.pendingEdits, m)
					}
				}
			}
			return m, nil

		case "enter":
			if m.editMode {
				// Save current field edit to pending edits before moving to next
				if m.editFieldIndex < len(editableFields) {
					field := editableFields[m.editFieldIndex]
					currentValue := m.editInput.Value()
					originalValue := m.originalValues[field.Name]
					if currentValue != originalValue {
						m.pendingEdits[field.Name] = currentValue
					} else {
						// Remove from pending if reverted to original
						delete(m.pendingEdits, field.Name)
					}
				}

				// Cycle to next field (like Tab)
				m.editFieldIndex = (m.editFieldIndex + 1) % len(editableFields)

				// Update input field with current value (from pending edits or original)
				filteredIssues := m.getFilteredIssues()
				if m.selectedIndex >= 0 && m.selectedIndex < len(filteredIssues) {
					field := editableFields[m.editFieldIndex]
					// Check if there's a pending edit for this field
					if pendingValue, exists := m.pendingEdits[field.Name]; exists {
						m.editInput.SetValue(pendingValue)
					} else {
						m.editInput.SetValue(m.originalValues[field.Name])
					}
					m.editOriginalValue = m.originalValues[field.Name]
				}
				m.hasUnsavedChanges = len(m.pendingEdits) > 0
				m.editInput.Focus()
				m.updatePaneContent()
				return m, nil
			} else if m.filterMode {
				// Apply filter and exit filter mode
				m.filterText = m.filterInput.Value()
				m.filterMode = false
				m.filterInput.Blur()
				m.selectedIndex = 0
				m.updatePaneContent()
				return m, nil
			} else if m.userInputMode == "user" {
				// Apply selected users - use client-side filtering for multiple users
				selectedUserIDs := []int{}
				for id, selected := range m.selectedUsers {
					if selected {
						selectedUserIDs = append(selectedUserIDs, id)
					}
				}

				if len(selectedUserIDs) > 0 {
					m.viewMode = "user-multi"
					// Store selected IDs as comma-separated string and store names
					if m.selectedUserNames == nil {
						m.selectedUserNames = make(map[int]string)
					}
					var idStrings []string
					for _, id := range selectedUserIDs {
						idStrings = append(idStrings, fmt.Sprintf("%d", id))
						// Store user name for display
						for _, user := range m.availableUsers {
							if user.ID == id {
								displayName := user.Name
								if displayName == "" {
									if user.Firstname != "" || user.Lastname != "" {
										displayName = strings.TrimSpace(user.Firstname + " " + user.Lastname)
									} else if user.Login != "" {
										displayName = user.Login
									}
								}
								m.selectedUserNames[id] = displayName
								break
							}
						}
					}
					m.assigneeFilter = strings.Join(idStrings, ",")
				} else {
					// No selection - clear filter but keep view mode if project filter is active
					m.assigneeFilter = ""
					m.selectedUserNames = make(map[int]string)
					if m.projectFilter == "" {
						m.viewMode = "all"
					}
				}

				m.userInputMode = ""
				m.listFilterText = ""
				m.filterInput.SetValue("")
				m.filterInput.Blur()
				m.loading = true
				m.selectedIndex = 0
				return m, fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues)
			} else if m.userInputMode == "project" {
				// Apply selected projects - use client-side filtering for multiple projects
				selectedProjectIDs := []int{}
				for id, selected := range m.selectedProjects {
					if selected {
						selectedProjectIDs = append(selectedProjectIDs, id)
					}
				}

				if len(selectedProjectIDs) > 0 {
					// Store selected IDs as comma-separated string and store names
					if m.selectedProjectNames == nil {
						m.selectedProjectNames = make(map[int]string)
					}
					var idStrings []string
					for _, id := range selectedProjectIDs {
						idStrings = append(idStrings, fmt.Sprintf("%d", id))
						// Store project name for display
						for _, project := range m.availableProjects {
							if project.ID == id {
								m.selectedProjectNames[id] = project.Name
								break
							}
						}
					}
					m.projectFilter = strings.Join(idStrings, ",")
					// Set view mode based on whether user filter is also active
					if m.assigneeFilter != "" {
						m.viewMode = "user-project-multi"
					} else {
						m.viewMode = "project-multi"
					}
				} else {
					// No selection - clear filter but keep view mode if user filter is active
					m.projectFilter = ""
					m.selectedProjectNames = make(map[int]string)
					if m.assigneeFilter == "" {
						m.viewMode = "all"
					} else {
						m.viewMode = "user-multi"
					}
				}

				m.userInputMode = ""
				m.listFilterText = ""
				m.filterInput.SetValue("")
				m.filterInput.Blur()
				m.loading = true
				m.selectedIndex = 0
				return m, fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues)
			} else if !inInputMode {
				// Open selected issue in browser or show details
				filteredIssues := m.getFilteredIssues()
				if len(filteredIssues) > 0 && m.selectedIndex < len(filteredIssues) {
					cmds = append(cmds, fetchIssueDetail(m.client, filteredIssues[m.selectedIndex].ID))
				}
			}

		case "up", "k":
			if m.editMode && m.editFieldIndex < len(editableFields) {
				field := editableFields[m.editFieldIndex]
				if field.Type == "select" {
					// Cycle through select options backwards
					options := field.GetOptions(&m)
					if len(options) > 0 {
						currentValue := m.editInput.Value()
						currentIndex := -1
						for i, opt := range options {
							if opt == currentValue {
								currentIndex = i
								break
							}
						}
						nextIndex := (currentIndex - 1 + len(options)) % len(options)
						m.editInput.SetValue(options[nextIndex])
						m.hasUnsavedChanges = true
						m.updatePaneContent()
					}
					return m, nil
				}
			} else if m.userInputMode == "user" && len(m.filteredIndices) > 0 {
				// Navigate user list
				if m.listCursor > 0 {
					m.listCursor--
				}
				return m, nil
			} else if m.userInputMode == "project" && len(m.filteredIndices) > 0 {
				// Navigate project list
				if m.listCursor > 0 {
					m.listCursor--
				}
				return m, nil
			} else if inInputMode {
				// Don't navigate in other input modes
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
			if m.editMode && m.editFieldIndex < len(editableFields) {
				field := editableFields[m.editFieldIndex]
				if field.Type == "select" {
					// Cycle through select options forwards
					options := field.GetOptions(&m)
					if len(options) > 0 {
						currentValue := m.editInput.Value()
						currentIndex := -1
						for i, opt := range options {
							if opt == currentValue {
								currentIndex = i
								break
							}
						}
						nextIndex := (currentIndex + 1) % len(options)
						m.editInput.SetValue(options[nextIndex])
						m.hasUnsavedChanges = true
						m.updatePaneContent()
					}
					return m, nil
				}
			} else if m.userInputMode == "user" && len(m.filteredIndices) > 0 {
				// Navigate user list
				if m.listCursor < len(m.filteredIndices)-1 {
					m.listCursor++
				}
				return m, nil
			} else if m.userInputMode == "project" && len(m.filteredIndices) > 0 {
				// Navigate project list
				if m.listCursor < len(m.filteredIndices)-1 {
					m.listCursor++
				}
				return m, nil
			} else if inInputMode {
				// Don't navigate in other input modes
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
			if inInputMode {
				// Don't page in input mode
				return m, nil
			}
			if m.activePane == 0 {
				m.leftPane, cmd = m.leftPane.Update(msg)
			} else {
				m.rightPane, cmd = m.rightPane.Update(msg)
			}
			cmds = append(cmds, cmd)

		case "pgdown":
			if inInputMode {
				// Don't page in input mode
				return m, nil
			}
			if m.activePane == 0 {
				m.leftPane, cmd = m.leftPane.Update(msg)
			} else {
				m.rightPane, cmd = m.rightPane.Update(msg)
			}
			cmds = append(cmds, cmd)

		case " ": // Space key
			if m.editMode {
				// Pass space to edit input
				m.editInput, cmd = m.editInput.Update(msg)
				cmds = append(cmds, cmd)
				m.hasUnsavedChanges = (m.editInput.Value() != m.editOriginalValue)
				m.updatePaneContent()
				return m, tea.Batch(cmds...)
			} else if m.userInputMode == "user" && len(m.filteredIndices) > 0 && m.listCursor < len(m.filteredIndices) {
				// Toggle selection of current user
				user := m.availableUsers[m.filteredIndices[m.listCursor]]
				m.selectedUsers[user.ID] = !m.selectedUsers[user.ID]
				return m, nil
			} else if m.userInputMode == "project" && len(m.filteredIndices) > 0 && m.listCursor < len(m.filteredIndices) {
				// Toggle selection of current project
				project := m.availableProjects[m.filteredIndices[m.listCursor]]
				m.selectedProjects[project.ID] = !m.selectedProjects[project.ID]
				return m, nil
			}

		default:
			// Handle text input in filter mode or user input mode or edit mode
			if m.editMode {
				// In edit mode, handle Tab separately
				if msg.String() == "tab" {
					// Save current field edit to pending edits before moving to next
					if m.editFieldIndex < len(editableFields) {
						field := editableFields[m.editFieldIndex]
						currentValue := m.editInput.Value()
						originalValue := m.originalValues[field.Name]
						if currentValue != originalValue {
							m.pendingEdits[field.Name] = currentValue
						} else {
							delete(m.pendingEdits, field.Name)
						}
					}

					// Cycle through editable fields
					m.editFieldIndex = (m.editFieldIndex + 1) % len(editableFields)

					// Update input field with current value (from pending edits or original)
					filteredIssues := m.getFilteredIssues()
					if m.selectedIndex >= 0 && m.selectedIndex < len(filteredIssues) {
						field := editableFields[m.editFieldIndex]
						// Check if there's a pending edit for this field
						if pendingValue, exists := m.pendingEdits[field.Name]; exists {
							m.editInput.SetValue(pendingValue)
						} else {
							m.editInput.SetValue(m.originalValues[field.Name])
						}
						m.editOriginalValue = m.originalValues[field.Name]
					}
					m.hasUnsavedChanges = len(m.pendingEdits) > 0
					m.editInput.Focus()
					m.updatePaneContent()
					return m, nil
				}
				// Pass other keys to edit input and update pane in real-time
				m.editInput, cmd = m.editInput.Update(msg)
				cmds = append(cmds, cmd)
				// Check if current field has changes
				currentFieldChanged := (m.editInput.Value() != m.editOriginalValue)
				// Update hasUnsavedChanges based on pending edits + current field
				m.hasUnsavedChanges = (len(m.pendingEdits) > 0 || currentFieldChanged)
				// Update pane immediately to show changes
				m.updatePaneContent()
			} else if inInputMode {
				m.filterInput, cmd = m.filterInput.Update(msg)
				cmds = append(cmds, cmd)
				// Update list filter text when in list mode
				if m.userInputMode == "user" || m.userInputMode == "project" {
					m.listFilterText = m.filterInput.Value()
					// Rebuild filtered indices immediately
					m.buildFilteredList()
				}
			} else {
				// Handle command keys when NOT in input mode
				switch msg.String() {
				case "e":
					// Enter edit mode
					filteredIssues := m.getFilteredIssues()
					if len(filteredIssues) > 0 && m.selectedIndex < len(filteredIssues) {
						m.editMode = true
						m.editFieldIndex = 0
						m.pendingEdits = make(map[string]string)
						m.originalValues = make(map[string]string)

						// Store all original values
						issue := filteredIssues[m.selectedIndex]
						for _, field := range editableFields {
							m.originalValues[field.Name] = field.GetValue(&issue)
						}

						// Load statuses and priorities if not already loaded
						var cmds []tea.Cmd
						if len(m.availableStatuses) == 0 {
							cmds = append(cmds, fetchStatuses(m.client))
						}
						if len(m.availablePriorities) == 0 {
							cmds = append(cmds, fetchPriorities(m.client))
						}
						if len(m.availableUsers) == 0 {
							cmds = append(cmds, fetchUsers(m.client))
						}

						// Set initial value in edit input
						field := editableFields[m.editFieldIndex]
						currentValue := field.GetValue(&issue)
						m.editInput.SetValue(currentValue)
						m.editOriginalValue = currentValue
						m.hasUnsavedChanges = false
						m.editInput.Focus()

						m.updatePaneContent()
						cmds = append(cmds, textinput.Blink)
						return m, tea.Batch(cmds...)
					}
					return m, nil
				case "f":
					// Enter filter mode
					m.filterMode = true
					m.filterInput.SetValue(m.filterText)
					m.filterInput.Placeholder = "Type to filter issues..."
					m.filterInput.Focus()
					return m, textinput.Blink
				case "m":
					// Toggle view mode: my -> all -> my
					if m.viewMode == "my" {
						m.viewMode = "all"
					} else {
						m.viewMode = "my"
					}
					m.loading = true
					m.selectedIndex = 0
					return m, fetchIssues(m.client, m.viewMode, m.assigneeFilter, m.projectFilter, m.issues)
				case "u":
					// Enter user selection mode
					m.userInputMode = "user"
					m.listCursor = 0
					m.listLoading = true
					m.listFilterText = ""
					m.filterInput.SetValue("")
					m.filterInput.Placeholder = "Type to filter users..."
					m.filterInput.Focus()
					return m, tea.Batch(fetchUsers(m.client), textinput.Blink)
				case "p":
					// Enter project selection mode
					m.userInputMode = "project"
					m.listCursor = 0
					m.listLoading = true
					m.listFilterText = ""
					m.filterInput.SetValue("")
					m.filterInput.Placeholder = "Type to filter projects..."
					m.filterInput.Focus()
					return m, tea.Batch(fetchProjects(m.client), textinput.Blink)
				case "?":
					m.showHelp = !m.showHelp
					return m, nil
				case "tab":
					// Switch between panes
					if m.activePane == 0 {
						m.activePane = 1
					} else {
						m.activePane = 0
					}
				}
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
