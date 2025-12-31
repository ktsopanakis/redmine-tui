package main

import (
	"encoding/base64"
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
	selectionMode bool
	selectionLine int // Line number in the active viewport
	selectionCol  int // Column position (character offset)
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

		case "v":
			// Toggle selection mode
			m.selectionMode = !m.selectionMode
			if m.selectionMode {
				// Reset selection to start of visible content
				m.selectionLine = 0
				m.selectionCol = 0
			}
			return m, nil

		case "y":
			// Copy selected text to clipboard
			if m.selectionMode {
				text := m.getSelectedText()
				if text != "" {
					copyToClipboard(text)
				}
				m.selectionMode = false
			}
			return m, nil

		case "tab":
			// Switch between panes
			if !m.selectionMode {
				if m.activePane == 0 {
					m.activePane = 1
				} else {
					m.activePane = 0
				}
			}

		case "up", "k":
			if m.selectionMode {
				// Move selection up
				if m.selectionLine > 0 {
					m.selectionLine--
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
			if m.selectionMode {
				// Move selection down
				m.selectionLine++
			} else {
				if m.activePane == 0 {
					m.leftPane, cmd = m.leftPane.Update(msg)
				} else {
					m.rightPane, cmd = m.rightPane.Update(msg)
				}
				cmds = append(cmds, cmd)
			}

		case "left", "h":
			if m.selectionMode {
				// Move selection left
				if m.selectionCol > 0 {
					m.selectionCol--
				}
			}

		case "right", "l":
			if m.selectionMode {
				// Move selection right
				m.selectionCol++
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

// copyToClipboard copies text to the system clipboard using OSC 52 escape sequence
func copyToClipboard(text string) {
	// OSC 52 is a terminal escape sequence for copying to clipboard
	// Format: \033]52;c;<base64-encoded-text>\007
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	osc52 := fmt.Sprintf("\033]52;c;%s\007", encoded)
	fmt.Print(osc52)
	os.Stdout.Sync()
}

// getSelectedText extracts the text at the current selection position
func (m model) getSelectedText() string {
	var content string
	if m.activePane == 0 {
		content = m.leftPane.View()
	} else {
		content = m.rightPane.View()
	}

	lines := strings.Split(content, "\n")
	if m.selectionLine < 0 || m.selectionLine >= len(lines) {
		return ""
	}

	// Return the entire current line
	return lines[m.selectionLine]
}
