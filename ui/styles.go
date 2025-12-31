package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/ktsopanakis/redmine-tui/config"
)

var (
	HeaderStyle     lipgloss.Style
	FooterStyle     lipgloss.Style
	PaneStyle       lipgloss.Style
	ActivePaneStyle lipgloss.Style
)

func InitStyles() {
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(config.Current.Colors.HeaderText)).
		Background(lipgloss.Color(config.Current.Colors.HeaderBackground)).
		PaddingLeft(1)

	FooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Current.Colors.FooterText)).
		Background(lipgloss.Color(config.Current.Colors.FooterBackground)).
		PaddingLeft(1)

	PaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.Current.Colors.InactivePaneBorder)).
		Padding(0, 1)

	ActivePaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.Current.Colors.ActivePaneBorder)).
		Padding(0, 1)
}
