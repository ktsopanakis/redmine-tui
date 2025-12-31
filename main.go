package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

const (
	headerHeight = 1
	footerHeight = 1
)

type Settings struct {
	Colors struct {
		ActivePaneBorder   string `yaml:"active_pane_border"`
		InactivePaneBorder string `yaml:"inactive_pane_border"`
		HeaderBackground   string `yaml:"header_background"`
		HeaderText         string `yaml:"header_text"`
		FooterBackground   string `yaml:"footer_background"`
		FooterText         string `yaml:"footer_text"`
	} `yaml:"colors"`
}

var settings Settings

func loadSettings() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Printf("Warning: Could not load config.yaml, using defaults: %v\n", err)
		settings.Colors.ActivePaneBorder = "#FF00FF"
		settings.Colors.InactivePaneBorder = "#874BFD"
		settings.Colors.HeaderBackground = "#7D56F4"
		settings.Colors.HeaderText = "#FAFAFA"
		settings.Colors.FooterBackground = "#3C3C3C"
		settings.Colors.FooterText = "#FAFAFA"
		return
	}

	if err := yaml.Unmarshal(data, &settings); err != nil {
		fmt.Printf("Warning: Could not parse config.yaml: %v\n", err)
	}
}

func initStyles() {
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(settings.Colors.HeaderText)).
		Background(lipgloss.Color(settings.Colors.HeaderBackground)).
		PaddingLeft(1)

	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(settings.Colors.FooterText)).
		Background(lipgloss.Color(settings.Colors.FooterBackground)).
		PaddingLeft(1)

	paneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(settings.Colors.InactivePaneBorder)).
		Padding(0, 1)

	activePaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(settings.Colors.ActivePaneBorder)).
		Padding(0, 1)
}

var (
	headerStyle     lipgloss.Style
	footerStyle     lipgloss.Style
	paneStyle       lipgloss.Style
	activePaneStyle lipgloss.Style
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
	activePane int
	leftTitle  string
	rightTitle string
	showHelp   bool
}

func initialModel() model {
	return model{
		leftTitle:  "Le",
		rightTitle: "Right Pane long",
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
			paneWidth := (msg.Width / 3) - 4
			leftPaneTotal := paneWidth + 4
			rightPaneWidth := msg.Width - leftPaneTotal - 4
			paneHeight := msg.Height - headerHeight - footerHeight - 2

			// Initialize left pane
			m.leftPane = viewport.New(paneWidth, paneHeight)
			m.leftPane.SetContent(lipgloss.NewStyle().Width(paneWidth).Render(loremIpsum))

			// Initialize right pane
			m.rightPane = viewport.New(rightPaneWidth, paneHeight)
			m.rightPane.SetContent(lipgloss.NewStyle().Width(rightPaneWidth).Render(loremIpsum))

			m.ready = true
		} else {
			paneWidth := (msg.Width / 3) - 4
			leftPaneTotal := paneWidth + 4
			rightPaneWidth := msg.Width - leftPaneTotal - 4
			paneHeight := msg.Height - headerHeight - footerHeight - 2

			m.leftPane.Width = paneWidth
			m.leftPane.Height = paneHeight
			m.leftPane.SetContent(lipgloss.NewStyle().Width(paneWidth).Render(loremIpsum))
			m.rightPane.Width = rightPaneWidth
			m.rightPane.Height = paneHeight
			m.rightPane.SetContent(lipgloss.NewStyle().Width(rightPaneWidth).Render(loremIpsum))
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
		// Calculate the actual border width: viewport width + padding (2) + borders (2)
		borderWidth := m.leftPane.Width + 4
		// Add dot indicator if active, use border lines before title
		var titlePart string
		if m.activePane == 0 {
			// Active: border lines + dot + title
			titlePart = "─ ● " + m.leftTitle + " "
		} else {
			// Inactive: border lines + title
			titlePart = "─── " + m.leftTitle + " "
		}
		if borderWidth > len(titlePart)+2 {
			// Build new border line: corner + title + remaining border + corner
			remainingBorder := ""
			if m.activePane == 1 {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+4)
			} else {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+2)
			}
			newPlainLine := "╭" + titlePart + remainingBorder + "╮"
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
		// Calculate the actual border width: viewport width + padding (2) + borders (2)
		borderWidth := m.rightPane.Width + 2
		// Add dot indicator if active, use border lines before title
		var titlePart string
		if m.activePane == 1 {
			// Active: border lines + dot + title
			titlePart = "─ ● " + m.rightTitle + " "
		} else {
			// Inactive: border lines + title
			titlePart = "─── " + m.rightTitle + " "
		}
		if borderWidth > len(titlePart)+2 {
			// Build new border line: corner + title + remaining border + corner
			remainingBorder := ""
			if m.activePane == 1 {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+4)
			} else {
				remainingBorder = strings.Repeat("─", borderWidth-len(titlePart)+6)
			}
			newPlainLine := "╭" + titlePart + remainingBorder + "╮"
			// Apply the border color
			styledTopLine := lipgloss.NewStyle().Foreground(rightBorderColor).Render(newPlainLine)
			rightLines[0] = styledTopLine
			rightPane = strings.Join(rightLines, "\n")
		}
	}

	// Combine panes side by side
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Footer with adaptive options
	footer := footerStyle.Width(m.width).Render(m.getFooterText())

	// Combine all sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		panes,
		footer,
	)
}

func (m model) getFooterText() string {
	items := []string{
		"Tab: Switch",
		"↑↓/jk: Scroll",
		"PgUp/PgDn: Page",
		"?: Help",
		"q: Quit",
	}

	required := []string{"Tab: Switch", "q: Quit"}

	text := ""
	for _, item := range items {
		testText := text
		if testText != "" {
			testText += " | "
		}
		testText += item

		if lipgloss.Width(testText) > m.width-2 {
			isRequired := false
			for _, req := range required {
				if item == req {
					isRequired = true
					break
				}
			}
			if !isRequired {
				continue
			}
		}

		if text != "" {
			text += " | "
		}
		text += item
	}

	return text
}

func main() {
	loadSettings()
	initStyles()

	// Parse command-line flags
	altScreen := flag.Bool("alt-screen", false, "Use alternate screen buffer (clears on exit)")
	flag.Parse()

	// Build program options
	opts := []tea.ProgramOption{tea.WithMouseCellMotion()}
	if *altScreen {
		opts = append(opts, tea.WithAltScreen())
	}

	p := tea.NewProgram(
		initialModel(),
		opts...,
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
