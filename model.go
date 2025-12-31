package main

import (
	"fmt"

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
	ready         bool
	width         int
	height        int
	leftPane      viewport.Model
	rightPane     viewport.Model
	activePane    int
	leftTitle     string
	rightTitle    string
	showHelp      bool
	client        *Client
	issues        []Issue
	selectedIndex int
	loading       bool
	err           error
}

func initialModel() model {
	client := NewClient(settings.Redmine.URL, settings.Redmine.APIKey)
	return model{
		leftTitle:     "Issues",
		rightTitle:    "Details",
		activePane:    0,
		client:        client,
		selectedIndex: 0,
		loading:       true,
	}
}

type issuesLoadedMsg struct {
	issues []Issue
	err    error
}

func fetchIssues(client *Client) tea.Cmd {
	return func() tea.Msg {
		resp, err := client.GetIssues(0, true, true, 100, 0)
		if err != nil {
			return issuesLoadedMsg{err: err}
		}
		return issuesLoadedMsg{issues: resp.Issues}
	}
}

func (m model) Init() tea.Cmd {
	return fetchIssues(m.client)
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
		}
		// Update panes with content if ready
		if m.ready {
			m.updatePaneContent()
		}
		return m, nil

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

		case "up", "k":
			if m.activePane == 0 && len(m.issues) > 0 {
				// Navigate issues list
				if m.selectedIndex > 0 {
					m.selectedIndex--
					m.updatePaneContent()
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
			if m.activePane == 0 && len(m.issues) > 0 {
				// Navigate issues list
				if m.selectedIndex < len(m.issues)-1 {
					m.selectedIndex++
					m.updatePaneContent()
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

		case "pgdown", "f":
			if m.activePane == 0 {
				m.leftPane, cmd = m.leftPane.Update(msg)
			} else {
				m.rightPane, cmd = m.rightPane.Update(msg)
			}
			cmds = append(cmds, cmd)
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
func (m *model) updatePaneContent() {
	if !m.ready {
		return
	}

	// Left pane: List of issues
	var leftContent string
	if m.loading {
		leftContent = "Loading issues..."
	} else if m.err != nil {
		leftContent = fmt.Sprintf("Error: %v", m.err)
	} else if len(m.issues) == 0 {
		leftContent = "No issues found."
	} else {
		for i, issue := range m.issues {
			prefix := "  "
			if i == m.selectedIndex {
				prefix = "> "
			}
			leftContent += fmt.Sprintf("%s#%d %s\n", prefix, issue.ID, issue.Subject)
		}
	}
	m.leftPane.SetContent(lipgloss.NewStyle().Width(m.leftPane.Width).Render(leftContent))

	// Right pane: Selected issue details
	var rightContent string
	if m.loading {
		rightContent = "Loading..."
	} else if m.err != nil {
		rightContent = fmt.Sprintf("Error: %v", m.err)
	} else if len(m.issues) == 0 {
		rightContent = "No issue selected."
	} else if m.selectedIndex >= 0 && m.selectedIndex < len(m.issues) {
		issue := m.issues[m.selectedIndex]
		rightContent = fmt.Sprintf("Issue #%d\n\n", issue.ID)
		rightContent += fmt.Sprintf("Subject: %s\n\n", issue.Subject)
		rightContent += fmt.Sprintf("Status: %s\n", issue.Status.Name)
		rightContent += fmt.Sprintf("Priority: %s\n", issue.Priority.Name)
		rightContent += fmt.Sprintf("Tracker: %s\n", issue.Tracker.Name)
		rightContent += fmt.Sprintf("Project: %s\n", issue.Project.Name)
		rightContent += fmt.Sprintf("Author: %s\n", issue.Author.Name)
		if issue.AssignedTo != nil {
			rightContent += fmt.Sprintf("Assigned To: %s\n", issue.AssignedTo.Name)
		}
		rightContent += fmt.Sprintf("Done: %d%%\n", issue.DoneRatio)
		if issue.StartDate != "" {
			rightContent += fmt.Sprintf("Start Date: %s\n", issue.StartDate)
		}
		if issue.DueDate != "" {
			rightContent += fmt.Sprintf("Due Date: %s\n", issue.DueDate)
		}
		rightContent += fmt.Sprintf("\nCreated: %s\n", issue.CreatedOn.Format("2006-01-02 15:04"))
		rightContent += fmt.Sprintf("Updated: %s\n\n", issue.UpdatedOn.Format("2006-01-02 15:04"))
		rightContent += "Description:\n"
		rightContent += fmt.Sprintf("%s\n", issue.Description)
	}
	m.rightPane.SetContent(lipgloss.NewStyle().Width(m.rightPane.Width).Render(rightContent))
}
