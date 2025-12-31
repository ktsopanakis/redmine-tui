package main

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/ktsopanakis/redmine-tui/config"
)

var (
	headerStyle     lipgloss.Style
	footerStyle     lipgloss.Style
	paneStyle       lipgloss.Style
	activePaneStyle lipgloss.Style
)

func initStyles() {
	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(config.Current.Colors.HeaderText)).
		Background(lipgloss.Color(config.Current.Colors.HeaderBackground)).
		PaddingLeft(1)

	footerStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(config.Current.Colors.FooterText)).
		Background(lipgloss.Color(config.Current.Colors.FooterBackground)).
		PaddingLeft(1)

	paneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.Current.Colors.InactivePaneBorder)).
		Padding(0, 1)

	activePaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(config.Current.Colors.ActivePaneBorder)).
		Padding(0, 1)
}
