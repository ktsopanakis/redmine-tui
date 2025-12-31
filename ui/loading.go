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
	Message string
}

type LoadingModel struct {
	spinner  spinner.Model
	messages []string
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
		messages: []string{},
		maxLines: 6,
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
		timestamp := time.Now().Format("15:04:05")
		m.messages = append(m.messages, fmt.Sprintf("[%s] %s", timestamp, msg.Message))
		if len(m.messages) > m.maxLines {
			m.messages = m.messages[1:]
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
		MaxWidth(60).
		Background(lipgloss.Color("235"))

	b.WriteString(fmt.Sprintf("%s Loading API calls...\n", m.spinner.View()))
	for _, msg := range m.messages {
		// Truncate long messages
		if len(msg) > 55 {
			msg = msg[:52] + "..."
		}
		b.WriteString(fmt.Sprintf("• %s\n", msg))
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
		return LoadingMsg{Message: message}
	}
}
