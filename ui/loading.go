package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LoadingMsg struct {
	Message   string
	Completed bool // true if marking as completed
}

type loadingMessage struct {
	text        string
	timestamp   time.Time
	completed   bool
	completedAt time.Time
}

type removeCompletedMsg struct {
	timestamp time.Time
}

type LoadingModel struct {
	spinner  spinner.Model
	messages []loadingMessage
	maxLines int
	Visible  bool
	width    int
	height   int
}

func NewLoadingModel() LoadingModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return LoadingModel{
		spinner:  s,
		messages: []loadingMessage{},
		maxLines: 8,
		Visible:  true,
		width:    0,
		height:   0,
	}
}

func (m LoadingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m LoadingModel) Update(msg tea.Msg) (LoadingModel, tea.Cmd) {
	switch msg := msg.(type) {
	case LoadingMsg:
		now := time.Now()
		if msg.Completed {
			// Mark the last in-progress message as completed
			for i := len(m.messages) - 1; i >= 0; i-- {
				if !m.messages[i].completed {
					m.messages[i].completed = true
					m.messages[i].completedAt = now
					// Schedule removal after 5 seconds
					return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
						return removeCompletedMsg{timestamp: m.messages[i].completedAt}
					})
				}
			}
		} else {
			// Add new in-progress message
			m.messages = append(m.messages, loadingMessage{
				text:      msg.Message,
				timestamp: now,
				completed: false,
			})
			if len(m.messages) > m.maxLines*2 { // Keep more to handle cleanup
				m.messages = m.messages[1:]
			}
		}
		return m, nil
	case removeCompletedMsg:
		// Remove messages completed at the specified time
		filtered := []loadingMessage{}
		for _, lm := range m.messages {
			if !lm.completed || !lm.completedAt.Equal(msg.timestamp) {
				filtered = append(filtered, lm)
			}
		}
		m.messages = filtered

		// Auto-hide if no messages remain
		if len(m.messages) == 0 {
			m.Visible = false
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m LoadingModel) View() string {
	if !m.Visible || len(m.messages) == 0 {
		return ""
	}

	var b strings.Builder

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		Width(58). // Fixed width for consistent positioning
		Background(lipgloss.Color("235"))

	// Count in-progress messages
	inProgress := 0
	for _, msg := range m.messages {
		if !msg.completed {
			inProgress++
		}
	}

	if inProgress > 0 {
		b.WriteString(fmt.Sprintf("%s Loading API calls...\n", m.spinner.View()))
	} else {
		greenCheck := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
		b.WriteString(fmt.Sprintf("%s All operations complete\n", greenCheck))
	}

	// Show only the last maxLines messages
	displayStart := 0
	if len(m.messages) > m.maxLines {
		displayStart = len(m.messages) - m.maxLines
	}

	for i := displayStart; i < len(m.messages); i++ {
		msg := m.messages[i]
		timestamp := msg.timestamp.Format("15:04:05")
		text := msg.text

		// Truncate long messages to fit width
		maxTextLen := 36
		if len(text) > maxTextLen {
			text = text[:maxTextLen-3] + "..."
		}

		if msg.completed {
			// Green checkmark for completed
			checkmark := lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓")
			b.WriteString(fmt.Sprintf("%s [%s] %s\n", checkmark, timestamp, text))
		} else {
			// Yellow dot for in-progress
			dot := lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("●")
			b.WriteString(fmt.Sprintf("%s [%s] %s\n", dot, timestamp, text))
		}
	}

	return style.Render(b.String())
}

func (m *LoadingModel) Hide() {
	m.Visible = false
}

func (m *LoadingModel) Show() {
	m.Visible = true
}

func (m *LoadingModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func SendLoadingMsg(message string) tea.Cmd {
	return func() tea.Msg {
		return LoadingMsg{Message: message, Completed: false}
	}
}

func SendLoadingCompleteMsg() tea.Cmd {
	return func() tea.Msg {
		return LoadingMsg{Completed: true}
	}
}
