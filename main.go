package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	headerHeight = 1
	footerHeight = 1
)

var (
	// Styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(0, 1)

	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF00FF")).
			Padding(0, 1)
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
	ready      bool
	width      int
	height     int
	leftPane   viewport.Model
	rightPane  viewport.Model
	activePane int // 0 for left, 1 for right
	leftTitle  string
	rightTitle string
}

func initialModel() model {
	return model{
		leftTitle:  "Left Pane",
		rightTitle: "Right Pane",
		activePane: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			// Calculate pane dimensions
			// Total width per pane = width/2
			// Content width = total - border(2) - padding(2) = total - 4
			paneWidth := (msg.Width / 2) - 4
			paneHeight := msg.Height - headerHeight - footerHeight - 4

			// Initialize left pane
			m.leftPane = viewport.New(paneWidth, paneHeight)
			m.leftPane.SetContent(lipgloss.NewStyle().Width(paneWidth).Render(loremIpsum))

			// Initialize right pane
			m.rightPane = viewport.New(paneWidth, paneHeight)
			m.rightPane.SetContent(lipgloss.NewStyle().Width(paneWidth).Render(loremIpsum))

			m.ready = true
		} else {
			// Update pane dimensions on resize
			paneWidth := (msg.Width / 2) - 4
			paneHeight := msg.Height - headerHeight - footerHeight - 4

			m.leftPane.Width = paneWidth
			m.leftPane.Height = paneHeight
			m.leftPane.SetContent(lipgloss.NewStyle().Width(paneWidth).Render(loremIpsum))
			m.rightPane.Width = paneWidth
			m.rightPane.Height = paneHeight
			m.rightPane.SetContent(lipgloss.NewStyle().Width(paneWidth).Render(loremIpsum))
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Switch between panes
			if m.activePane == 0 {
				m.activePane = 1
			} else {
				m.activePane = 0
			}

		case "up", "k":
			if m.activePane == 0 {
				m.leftPane, cmd = m.leftPane.Update(msg)
			} else {
				m.rightPane, cmd = m.rightPane.Update(msg)
			}
			cmds = append(cmds, cmd)

		case "down", "j":
			if m.activePane == 0 {
				m.leftPane, cmd = m.leftPane.Update(msg)
			} else {
				m.rightPane, cmd = m.rightPane.Update(msg)
			}
			cmds = append(cmds, cmd)

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

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	// Header
	header := headerStyle.Width(m.width).Render("Redmine TUI")

	// Left pane with title embedded in border
	leftBorderColor := lipgloss.Color("#874BFD")
	if m.activePane == 0 {
		leftBorderColor = lipgloss.Color("#FF00FF")
	}

	// Render pane with border
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor).
		Padding(0, 1).
		Render(m.leftPane.View())

	// Embed title in the top border line
	leftLines := strings.Split(leftPane, "\n")
	if len(leftLines) > 0 {
		// Calculate the actual border width: viewport width + padding (2)
		borderWidth := m.leftPane.Width + 2
		titlePart := " " + m.leftTitle + " "
		if borderWidth > len(titlePart)+2 {
			// Build new border line: corner + title + remaining border + corner
			remainingBorder := strings.Repeat("─", borderWidth-len(titlePart))
			newPlainLine := "╭" + titlePart + remainingBorder + "─"
			// Apply the border color
			styledTopLine := lipgloss.NewStyle().Foreground(leftBorderColor).Render(newPlainLine)
			leftLines[0] = styledTopLine
			leftPane = strings.Join(leftLines, "\n")
		}
	}

	// Right pane with title embedded in border
	rightBorderColor := lipgloss.Color("#874BFD")
	if m.activePane == 1 {
		rightBorderColor = lipgloss.Color("#FF00FF")
	}

	// Render pane with border
	rightPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor).
		Padding(0, 1).
		Render(m.rightPane.View())

	// Embed title in the top border line
	rightLines := strings.Split(rightPane, "\n")
	if len(rightLines) > 0 {
		// Calculate the actual border width: viewport width + padding (2)
		borderWidth := m.rightPane.Width + 2
		titlePart := " " + m.rightTitle + " "
		if borderWidth > len(titlePart)+2 {
			// Build new border line: corner + title + remaining border + corner
			remainingBorder := strings.Repeat("─", borderWidth-len(titlePart))
			newPlainLine := "╭" + titlePart + remainingBorder + "─"
			// Apply the border color
			styledTopLine := lipgloss.NewStyle().Foreground(rightBorderColor).Render(newPlainLine)
			rightLines[0] = styledTopLine
			rightPane = strings.Join(rightLines, "\n")
		}
	}

	// Combine panes side by side
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Footer with options
	footerText := "Tab: Switch Panes | ↑↓/jk: Scroll | PgUp/PgDn: Page | q: Quit"
	footer := footerStyle.Width(m.width).Render(footerText)

	// Combine all sections
	ui := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		panes,
		footer,
	)
	return strings.TrimRight(ui, "\n")
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
